package vless

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/djalben/xplr-core/backend/shop"
)

// ══════════════════════════════════════════════════════════════
// VlessProvider — ProductProvider for VLESS+Reality VPN keys.
// Communicates with a 3X-UI panel (MHSanaei) via its REST API.
// ══════════════════════════════════════════════════════════════

// Config holds all settings loaded from environment variables.
type Config struct {
	PanelURL   string // e.g. "https://109.120.157.144:2053"
	Username   string // 3X-UI admin username
	Password   string // 3X-UI admin password
	InboundID  int    // ID of the VLESS+Reality inbound
	ServerIP   string // public IP of the VPN server
	ServerPort string // listening port (typically "443")
	SNI        string // TLS SNI / dest domain (e.g. "google.com")
	PublicKey  string // Reality public key (x25519)
	ShortID    string // Reality short ID
	Flow       string // typically "xtls-rprx-vision"
}

// VlessProvider implements shop.ProductProvider.
type VlessProvider struct {
	cfg    Config
	client *http.Client
	mu     sync.Mutex
	cookie string // session cookie from 3X-UI login
}

// NewVlessProvider creates a provider from environment variables.
// Returns nil if XPANEL_URL is not configured.
func NewVlessProvider() *VlessProvider {
	panelURL := os.Getenv("XPANEL_URL")
	if panelURL == "" {
		log.Println("[VLESS] ⚠️ XPANEL_URL not set — VlessProvider disabled")
		return nil
	}

	cfg := Config{
		PanelURL:   strings.TrimRight(panelURL, "/"),
		Username:   os.Getenv("XPANEL_USERNAME"),
		Password:   os.Getenv("XPANEL_PASSWORD"),
		InboundID:  1, // default; override via XPANEL_INBOUND_ID
		ServerIP:   getEnvOr("XPANEL_SERVER_IP", "109.120.157.144"),
		ServerPort: getEnvOr("XPANEL_SERVER_PORT", "443"),
		SNI:        getEnvOr("XPANEL_SNI", "google.com"),
		PublicKey:  os.Getenv("XPANEL_REALITY_PUBLIC_KEY"),
		ShortID:    os.Getenv("XPANEL_REALITY_SHORT_ID"),
		Flow:       getEnvOr("XPANEL_FLOW", "xtls-rprx-vision"),
	}

	if envID := os.Getenv("XPANEL_INBOUND_ID"); envID != "" {
		fmt.Sscanf(envID, "%d", &cfg.InboundID)
	}

	// 3X-UI uses self-signed certs — skip TLS verification for panel API
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
		Jar:       jar,
	}

	p := &VlessProvider{cfg: cfg, client: client}

	// Authenticate on startup
	if err := p.login(); err != nil {
		log.Printf("[VLESS] ❌ Initial login failed: %v (will retry on demand)", err)
	} else {
		log.Printf("[VLESS] ✅ Authenticated to 3X-UI panel at %s", cfg.PanelURL)
	}

	return p
}

// ── ProductProvider interface ──

func (v *VlessProvider) Name() string { return "vless" }

func (v *VlessProvider) GetCatalog() ([]shop.CatalogProduct, error) {
	// VPN keys are generated on-demand, so we return predefined plans
	plans := []shop.CatalogProduct{
		{
			ExternalID:  "vless-stockholm-7d",
			Name:        "Безопасный доступ — 7 дней",
			Description: "VLESS+Reality VPN ключ (Швеция). Лимит 15 ГБ, 7 дней.",
			Category:    "vpn",
			Country:     "Швеция",
			CountryCode: "SE",
			CostPrice:   decimal.NewFromFloat(0.88),
			Currency:    "EUR",
			InStock:     true,
			Meta:        map[string]any{"duration_days": 7, "server": "Stockholm", "retail_price": 5.00},
		},
		{
			ExternalID:  "vless-stockholm-30d",
			Name:        "Безопасный доступ — 30 дней",
			Description: "VLESS+Reality VPN ключ (Швеция). Лимит 60 ГБ, 30 дней.",
			Category:    "vpn",
			Country:     "Швеция",
			CountryCode: "SE",
			CostPrice:   decimal.NewFromFloat(5.30),
			Currency:    "EUR",
			InStock:     true,
			Meta:        map[string]any{"duration_days": 30, "server": "Stockholm", "retail_price": 10.00},
		},
		{
			ExternalID:  "vless-stockholm-180d",
			Name:        "Безопасный доступ — 180 дней",
			Description: "VLESS+Reality VPN ключ (Швеция). Лимит 300 ГБ, 180 дней.",
			Category:    "vpn",
			Country:     "Швеция",
			CountryCode: "SE",
			CostPrice:   decimal.NewFromFloat(26.50),
			Currency:    "EUR",
			InStock:     true,
			Meta:        map[string]any{"duration_days": 180, "server": "Stockholm", "retail_price": 35.00},
		},
		{
			ExternalID:  "vless-stockholm-365d",
			Name:        "Безопасный доступ — 365 дней",
			Description: "VLESS+Reality VPN ключ (Швеция). Лимит 600 ГБ, 365 дней.",
			Category:    "vpn",
			Country:     "Швеция",
			CountryCode: "SE",
			CostPrice:   decimal.NewFromFloat(48.00),
			Currency:    "EUR",
			InStock:     true,
			Meta:        map[string]any{"duration_days": 365, "server": "Stockholm", "retail_price": 55.00},
		},
	}
	return plans, nil
}

func (v *VlessProvider) CreateOrder(externalProductID string) (*shop.OrderResult, error) {
	// Determine duration from product ID
	durationDays := 30
	if strings.Contains(externalProductID, "7d") {
		durationDays = 7
	} else if strings.Contains(externalProductID, "180d") {
		durationDays = 180
	} else if strings.Contains(externalProductID, "365d") {
		durationDays = 365
	}

	// Traffic quota per plan (bytes): protect against abuse
	trafficQuotas := map[int]int64{
		7:   15 * 1024 * 1024 * 1024,  // 15 GB
		30:  60 * 1024 * 1024 * 1024,  // 60 GB
		180: 300 * 1024 * 1024 * 1024, // 300 GB
		365: 600 * 1024 * 1024 * 1024, // 600 GB
	}
	totalBytes := trafficQuotas[durationDays]
	if totalBytes == 0 {
		totalBytes = 60 * 1024 * 1024 * 1024 // fallback 60 GB
	}

	// Generate unique client UUID and email tag
	clientUUID := uuid.New().String()
	clientEmail := fmt.Sprintf("xplr-%s", clientUUID[:8])
	expiryMs := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour).UnixMilli()

	// Add client to 3X-UI inbound (limitIp=1, traffic quota enforced)
	if err := v.addClient(clientUUID, clientEmail, expiryMs, totalBytes); err != nil {
		return nil, fmt.Errorf("failed to create VPN key: %w", err)
	}

	// Build the vless:// connection link
	connLink := v.buildVlessLink(clientUUID, clientEmail)

	log.Printf("[VLESS] ✅ Created key: email=%s uuid=%s...%s expires=%dd limitIp=1 quota=%dGB",
		clientEmail, clientUUID[:8], clientUUID[len(clientUUID)-4:], durationDays, totalBytes/(1024*1024*1024))

	return &shop.OrderResult{
		ProviderRef:   clientEmail,
		ActivationKey: connLink,
		QRData:        connLink, // QR scanners can use the vless:// link directly
		Status:        "completed",
	}, nil
}

func (v *VlessProvider) CheckStatus(providerRef string) (*shop.OrderStatus, error) {
	stats, err := v.getClientTraffic(providerRef)
	if err != nil {
		return &shop.OrderStatus{
			ProviderRef:  providerRef,
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, nil
	}

	status := "completed"
	if !stats.Enable {
		status = "expired"
	}

	return &shop.OrderStatus{
		ProviderRef: providerRef,
		Status:      status,
	}, nil
}

func (v *VlessProvider) GetBalance() (*shop.BalanceInfo, error) {
	// Self-hosted — no external balance
	return nil, nil
}

// ── 3X-UI API Methods ──

// login authenticates with the 3X-UI panel and stores the session cookie.
func (v *VlessProvider) login() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	payload := fmt.Sprintf(`{"username":"%s","password":"%s"}`, v.cfg.Username, v.cfg.Password)
	resp, err := v.client.Post(
		v.cfg.PanelURL+"/login",
		"application/json",
		strings.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("login parse error: %w (body: %s)", err, string(body))
	}
	if !result.Success {
		return fmt.Errorf("login rejected: %s", result.Msg)
	}

	// Session cookie is stored in the cookie jar automatically
	log.Printf("[VLESS] 🔑 Logged into 3X-UI panel")
	return nil
}

// addClient adds a new VLESS client to the configured inbound.
func (v *VlessProvider) addClient(clientUUID, email string, expiryMs int64, totalBytes int64) error {
	// Build the client settings JSON
	// limitIp=1 → strict 1 IP per key policy (revenue protection)
	// total    → traffic quota in bytes (anti-abuse)
	clientSettings := []map[string]any{
		{
			"id":         clientUUID,
			"flow":       v.cfg.Flow,
			"email":      email,
			"limitIp":    1,
			"total":      totalBytes,
			"expiryTime": expiryMs,
			"enable":     true,
			"tgId":       "",
			"subId":      "",
		},
	}

	settingsJSON, _ := json.Marshal(clientSettings)

	// 3X-UI expects form data
	formData := url.Values{}
	formData.Set("id", fmt.Sprintf("%d", v.cfg.InboundID))
	formData.Set("settings", fmt.Sprintf(`{"clients":%s}`, string(settingsJSON)))

	// Try the request; re-login if session expired
	resp, err := v.doAPIRequest("POST", "/panel/api/inbounds/addClient", formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("addClient parse error: %w (body: %s)", err, string(body))
	}
	if !result.Success {
		return fmt.Errorf("addClient failed: %s", result.Msg)
	}

	return nil
}

// clientTrafficStats holds traffic info returned by the panel.
type clientTrafficStats struct {
	Email  string `json:"email"`
	Enable bool   `json:"enable"`
	Up     int64  `json:"up"`
	Down   int64  `json:"down"`
}

// getClientTraffic queries traffic stats for a client by email tag.
func (v *VlessProvider) getClientTraffic(email string) (*clientTrafficStats, error) {
	resp, err := v.doAPIRequest("GET", "/panel/api/inbounds/getClientTraffics/"+email, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool                `json:"success"`
		Msg     string              `json:"msg"`
		Obj     *clientTrafficStats `json:"obj"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("getClientTraffic parse error: %w", err)
	}
	if !result.Success || result.Obj == nil {
		return nil, fmt.Errorf("client %q not found: %s", email, result.Msg)
	}

	return result.Obj, nil
}

// doAPIRequest makes an authenticated request to the 3X-UI API.
// Automatically re-authenticates on 401.
func (v *VlessProvider) doAPIRequest(method, path string, formData url.Values) (*http.Response, error) {
	makeReq := func() (*http.Response, error) {
		var body io.Reader
		contentType := "application/json"
		if formData != nil {
			body = strings.NewReader(formData.Encode())
			contentType = "application/x-www-form-urlencoded"
		}

		req, err := http.NewRequest(method, v.cfg.PanelURL+path, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		return v.client.Do(req)
	}

	resp, err := makeReq()
	if err != nil {
		return nil, err
	}

	// If unauthorized, re-login and retry once
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		if err := v.login(); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		return makeReq()
	}

	return resp, nil
}

// buildVlessLink constructs the vless:// URI for client apps.
func (v *VlessProvider) buildVlessLink(clientUUID, remark string) string {
	// vless://UUID@IP:PORT?params#REMARK
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("flow", v.cfg.Flow)
	params.Set("security", "reality")
	params.Set("sni", v.cfg.SNI)
	params.Set("fp", "chrome")
	params.Set("pbk", v.cfg.PublicKey)
	params.Set("sid", v.cfg.ShortID)
	params.Set("type", "tcp")
	params.Set("headerType", "none")

	return fmt.Sprintf("vless://%s@%s:%s?%s#%s",
		clientUUID,
		v.cfg.ServerIP,
		v.cfg.ServerPort,
		params.Encode(),
		url.PathEscape("XPLR-VPN-"+remark),
	)
}

// ── Helpers ──

// DeleteClient removes a client from the inbound (for refunds/expiry cleanup).
func (v *VlessProvider) DeleteClient(clientUUID string) error {
	path := fmt.Sprintf("/panel/api/inbounds/%d/delClient/%s", v.cfg.InboundID, clientUUID)
	resp, err := v.doAPIRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	json.Unmarshal(body, &result)
	if !result.Success {
		return fmt.Errorf("deleteClient failed: %s", result.Msg)
	}
	return nil
}

// ResetClientTraffic resets traffic counters for a client.
func (v *VlessProvider) ResetClientTraffic(email string) error {
	path := fmt.Sprintf("/panel/api/inbounds/%d/resetClientTraffic/%s", v.cfg.InboundID, email)
	resp, err := v.doAPIRequest("POST", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Compile-time check: VlessProvider implements ProductProvider.
var _ shop.ProductProvider = (*VlessProvider)(nil)

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetActiveClients returns the number of active (enabled) clients.
// Useful for admin dashboard monitoring.
func (v *VlessProvider) GetActiveClients() (int, error) {
	resp, err := v.doAPIRequest("POST", "/panel/api/inbounds/list", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool `json:"success"`
		Obj     []struct {
			ID          int `json:"id"`
			ClientStats []struct {
				Email  string `json:"email"`
				Enable bool   `json:"enable"`
			} `json:"clientStats"`
		} `json:"obj"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	count := 0
	for _, inbound := range result.Obj {
		if inbound.ID == v.cfg.InboundID {
			for _, cs := range inbound.ClientStats {
				if cs.Enable {
					count++
				}
			}
			break
		}
	}
	return count, nil
}
