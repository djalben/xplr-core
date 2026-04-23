package vpn

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/djalben/xplr-core/backend/internal/config"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type XPanelConfig struct {
	PanelURL   string
	BasePath   string
	Username   string
	Password   string
	InboundID  int
	ServerIP   string
	ServerPort string
	SNI        string
	PublicKey  string
	ShortID    string
	Flow       string
}

type VlessXPanelProvider struct {
	cfg    XPanelConfig
	client *http.Client

	mu     sync.Mutex
	authed bool
}

func NewVlessXPanelProvider(cfg config.ENV) *VlessXPanelProvider {
	basePath := strings.TrimRight(cfg.XPanelBasePath, "/")
	if basePath == "" {
		basePath = "/panel"
	}

	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		//nolint:gosec // 3X-UI panels are often self-signed; trusted by deployment config.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &VlessXPanelProvider{
		cfg: XPanelConfig{
			PanelURL:   strings.TrimRight(cfg.XPanelURL, "/"),
			BasePath:   basePath,
			Username:   cfg.XPanelUsername,
			Password:   cfg.XPanelPassword,
			InboundID:  cfg.XPanelInboundID,
			ServerIP:   cfg.XPanelServerIP,
			ServerPort: cfg.XPanelServerPort,
			SNI:        cfg.XPanelSNI,
			PublicKey:  cfg.XPanelRealityPublicKey,
			ShortID:    cfg.XPanelRealityShortID,
			Flow:       cfg.XPanelFlow,
		},
		client: &http.Client{
			Timeout:   8 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
	}
}

func (v *VlessXPanelProvider) Name() string { return "vless" }

func (v *VlessXPanelProvider) Enabled() bool { return v.cfg.PanelURL != "" }

func (v *VlessXPanelProvider) CreateOrder(ctx context.Context, externalProductID string) (providerRef string, activationKey string, meta string, err error) {
	if !v.Enabled() {
		return "", "", "", wrapper.Wrap(domain.NewInvalidInput("vpn provider is not configured"))
	}

	durationDays := 30
	switch {
	case strings.Contains(externalProductID, "7d"):
		durationDays = 7
	case strings.Contains(externalProductID, "180d"):
		durationDays = 180
	case strings.Contains(externalProductID, "365d"):
		durationDays = 365
	}

	trafficQuotas := map[int]int64{
		7:   15 * 1024 * 1024 * 1024,
		30:  60 * 1024 * 1024 * 1024,
		180: 300 * 1024 * 1024 * 1024,
		365: 600 * 1024 * 1024 * 1024,
	}
	totalBytes := trafficQuotas[durationDays]
	if totalBytes == 0 {
		totalBytes = 60 * 1024 * 1024 * 1024
	}

	clientUUID := domain.NewUUID().String()
	clientEmail := "xplr-" + clientUUID[:8]
	expiryMs := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour).UnixMilli()

	err = v.addClient(ctx, clientUUID, clientEmail, expiryMs, totalBytes)
	if err != nil {
		return "", "", "", wrapper.Wrap(err)
	}

	connLink := v.buildVlessLink(clientUUID, clientEmail)

	metaBytes, err := json.Marshal(map[string]any{
		"traffic_bytes": totalBytes,
		"expire_ms":     expiryMs,
		"client_email":  clientEmail,
		"duration_days": durationDays,
	})
	if err != nil {
		return "", "", "", wrapper.Wrap(err)
	}

	return clientEmail, connLink, string(metaBytes), nil
}

func (v *VlessXPanelProvider) GetClientTraffic(ctx context.Context, providerRef string) (*domain.VPNClientTraffic, error) {
	if !v.Enabled() {
		return nil, wrapper.Wrap(domain.NewInvalidInput("vpn provider is not configured"))
	}

	resp, err := v.doAPIRequest(ctx, http.MethodGet, v.writeAPI()+"/getClientTraffics/"+providerRef, nil)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool `json:"success"`
		Obj     *struct {
			Email  string `json:"email"`
			Enable bool   `json:"enable"`
			Up     int64  `json:"up"`
			Down   int64  `json:"down"`
		} `json:"obj"`
		Msg string `json:"msg"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if !result.Success || result.Obj == nil {
		return nil, wrapper.Wrap(domain.NewNotFound("vpn client not found"))
	}

	return &domain.VPNClientTraffic{
		ProviderRef: providerRef,
		Enabled:     result.Obj.Enable,
		Upload:      result.Obj.Up,
		Download:    result.Obj.Down,
	}, nil
}

type inboundClient struct {
	ID         string `json:"id"`
	Flow       string `json:"flow"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp"`
	Total      int64  `json:"total"`
	ExpiryTime int64  `json:"expiryTime"`
	Enable     bool   `json:"enable"`
	TGID       string `json:"tgId"`
	SubID      string `json:"subId"`
}

func (v *VlessXPanelProvider) DeleteClientByEmail(ctx context.Context, email string) error {
	if !v.Enabled() {
		return wrapper.Wrap(domain.NewInvalidInput("vpn provider is not configured"))
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return wrapper.Wrap(domain.NewInvalidInput("email is required"))
	}

	// 3x-ui: POST /{id}/delClientByEmail/{email}
	path := fmt.Sprintf("%s/%d/delClientByEmail/%s", v.writeAPI(), v.cfg.InboundID, url.PathEscape(email))
	resp, err := v.doAPIRequest(ctx, http.MethodPost, path, url.Values{})
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if !result.Success {
		return wrapper.Wrap(domain.NewInvalidInput("xpanel delClientByEmail failed"))
	}

	return nil
}

func (v *VlessXPanelProvider) UpdateClientByEmail(ctx context.Context, email string, totalBytes *int64, expiryMs *int64) error {
	if !v.Enabled() {
		return wrapper.Wrap(domain.NewInvalidInput("vpn provider is not configured"))
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return wrapper.Wrap(domain.NewInvalidInput("email is required"))
	}
	if totalBytes == nil && expiryMs == nil {
		return wrapper.Wrap(domain.NewInvalidInput("nothing to update"))
	}

	client, err := v.getClientByEmail(ctx, email)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if client == nil {
		return wrapper.Wrap(domain.NewNotFound("vpn client not found"))
	}

	if totalBytes != nil {
		client.Total = *totalBytes
	}
	if expiryMs != nil {
		client.ExpiryTime = *expiryMs
	}

	settingsJSON, err := json.Marshal([]inboundClient{*client})
	if err != nil {
		return wrapper.Wrap(err)
	}

	form := url.Values{}
	form.Set("id", strconv.Itoa(v.cfg.InboundID))
	form.Set("settings", fmt.Sprintf(`{"clients":%s}`, string(settingsJSON)))

	// 3x-ui: POST /updateClient/{clientId}
	path := v.writeAPI() + "/updateClient/" + url.PathEscape(client.ID)
	resp, err := v.doAPIRequest(ctx, http.MethodPost, path, form)
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if !result.Success {
		return wrapper.Wrap(domain.NewInvalidInput("xpanel updateClient failed"))
	}

	return nil
}

func (v *VlessXPanelProvider) getClientByEmail(ctx context.Context, email string) (*inboundClient, error) {
	type inboundObj struct {
		ID       int    `json:"id"`
		Settings string `json:"settings"`
	}
	var result struct {
		Success bool        `json:"success"`
		Obj     *inboundObj `json:"obj"`
		Msg     string      `json:"msg"`
	}

	path := fmt.Sprintf("%s/get/%d", v.writeAPI(), v.cfg.InboundID)
	resp, err := v.doAPIRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if !result.Success || result.Obj == nil {
		return nil, wrapper.Wrap(domain.NewNotFound("xpanel inbound not found"))
	}

	var settings struct {
		Clients []inboundClient `json:"clients"`
	}
	err = json.Unmarshal([]byte(result.Obj.Settings), &settings)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	for i := range settings.Clients {
		c := settings.Clients[i]
		if strings.EqualFold(c.Email, email) {
			// Fill defaults the panel expects.
			if c.Flow == "" {
				c.Flow = v.cfg.Flow
			}

			return &c, nil
		}
	}

	return nil, wrapper.Wrap(domain.NewNotFound("vpn client not found"))
}

func (v *VlessXPanelProvider) addClient(ctx context.Context, clientUUID string, email string, expiryMs int64, totalBytes int64) error {
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

	settingsJSON, err := json.Marshal(clientSettings)
	if err != nil {
		return wrapper.Wrap(err)
	}

	formData := url.Values{}
	formData.Set("id", strconv.Itoa(v.cfg.InboundID))
	formData.Set("settings", fmt.Sprintf(`{"clients":%s}`, string(settingsJSON)))

	resp, err := v.doAPIRequest(ctx, http.MethodPost, v.writeAPI()+"/addClient", formData)
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if !result.Success {
		return wrapper.Wrap(domain.NewInvalidInput("xpanel addClient failed"))
	}

	return nil
}

func (v *VlessXPanelProvider) writeAPI() string { return v.cfg.BasePath + "/panel/api/inbounds" }

func (v *VlessXPanelProvider) login(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.authed {
		return nil
	}

	payload := fmt.Sprintf(`{"username":"%s","password":"%s"}`, v.cfg.Username, v.cfg.Password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.cfg.PanelURL+v.cfg.BasePath+"/login", strings.NewReader(payload))
	if err != nil {
		return wrapper.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if !result.Success {
		return wrapper.Wrap(domain.NewInvalidInput("xpanel login rejected"))
	}

	v.authed = true

	return nil
}

func (v *VlessXPanelProvider) doAPIRequest(ctx context.Context, method string, path string, formData url.Values) (*http.Response, error) {
	makeReq := func() (*http.Response, error) {
		var body io.Reader
		contentType := "application/json"
		if formData != nil {
			body = strings.NewReader(formData.Encode())
			contentType = "application/x-www-form-urlencoded"
		}

		req, err := http.NewRequestWithContext(ctx, method, v.cfg.PanelURL+path, body)
		if err != nil {
			return nil, wrapper.Wrap(err)
		}
		req.Header.Set("Content-Type", contentType)

		return v.client.Do(req)
	}

	err := v.login(ctx)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	resp, err := makeReq()
	if err != nil {
		v.mu.Lock()
		v.authed = false
		v.mu.Unlock()

		err2 := v.login(ctx)
		if err2 != nil {
			return nil, wrapper.Wrap(err2)
		}
		resp, err = makeReq()
		if err != nil {
			return nil, wrapper.Wrap(err)
		}
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()

		v.mu.Lock()
		v.authed = false
		v.mu.Unlock()

		err := v.login(ctx)
		if err != nil {
			return nil, wrapper.Wrap(err)
		}

		return makeReq()
	}

	return resp, nil
}

func (v *VlessXPanelProvider) buildVlessLink(clientUUID string, remark string) string {
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

	hostport := net.JoinHostPort(v.cfg.ServerIP, v.cfg.ServerPort)
	u := &url.URL{
		Scheme:   "vless",
		User:     url.User(clientUUID),
		Host:     hostport,
		RawQuery: params.Encode(),
		Fragment: url.PathEscape("XPLR-VPN-" + remark),
	}

	return u.String()
}
