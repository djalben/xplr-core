package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ══════════════════════════════════════════════════════════════
// Celitech eSIM Provider — implements ESIMProvider interface
//
// API docs: https://docs.celitech.com
// Auth: OAuth2 client_credentials → Bearer token
// Env vars:
//   CELITECH_CLIENT_ID     — OAuth2 client ID
//   CELITECH_CLIENT_SECRET — OAuth2 client secret
//   CELITECH_API_URL       — (optional) base API URL, default https://api.celitech.net/v1
//   CELITECH_AUTH_URL      — (optional) token endpoint, default https://auth.celitech.com/oauth2/token
// ══════════════════════════════════════════════════════════════

type CelitechProvider struct {
	clientID     string
	clientSecret string
	apiURL       string
	authURL      string
	client       *http.Client

	// OAuth2 token cache
	tokenMu    sync.Mutex
	token      string
	tokenExpAt time.Time

	// Package cache (planID → package details for ordering)
	pkgMu    sync.RWMutex
	pkgCache map[string]*celitechPackage
}

// celitechPackage — cached package details needed for purchase API
type celitechPackage struct {
	ID            string  `json:"id"`
	Destination   string  `json:"destination"`
	DataLimitInGB float64 `json:"dataLimitInGB"`
	Duration      int     `json:"duration"` // days
	Price         float64 `json:"price"`
}

func NewCelitechProvider() *CelitechProvider {
	clientID := os.Getenv("CELITECH_CLIENT_ID")
	clientSecret := os.Getenv("CELITECH_CLIENT_SECRET")
	apiURL := os.Getenv("CELITECH_API_URL")
	if apiURL == "" {
		apiURL = "https://api.celitech.net/v1"
	}
	authURL := os.Getenv("CELITECH_AUTH_URL")
	if authURL == "" {
		authURL = "https://auth.celitech.com/oauth2/token"
	}
	return &CelitechProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		apiURL:       apiURL,
		authURL:      authURL,
		client:       &http.Client{Timeout: 20 * time.Second},
		pkgCache:     make(map[string]*celitechPackage),
	}
}

func (c *CelitechProvider) Name() string { return "celitech" }

func (c *CelitechProvider) isConfigured() bool {
	return c.clientID != "" && c.clientSecret != ""
}

// ── OAuth2 Token Management ──

func (c *CelitechProvider) getToken() (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Return cached token if still valid (with 60s buffer)
	if c.token != "" && time.Now().Before(c.tokenExpAt.Add(-60*time.Second)) {
		return c.token, nil
	}

	// Request new token via client_credentials grant
	body := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		c.clientID, c.clientSecret)

	req, err := http.NewRequest("POST", c.authURL, bytes.NewBufferString(body))
	if err != nil {
		return "", fmt.Errorf("celitech auth request build: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("celitech auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("celitech auth failed: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"` // seconds
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("celitech auth decode: %w", err)
	}

	c.token = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		c.tokenExpAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	} else {
		c.tokenExpAt = time.Now().Add(55 * time.Minute) // default 55 min
	}

	log.Printf("[CELITECH] ✅ OAuth2 token obtained (expires in %ds)", tokenResp.ExpiresIn)
	return c.token, nil
}

// doRequest — authenticated GET/POST to Celitech API
func (c *CelitechProvider) doRequest(method, path string, reqBody interface{}) (*http.Response, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("celitech auth: %w", err)
	}

	url := c.apiURL + path
	var bodyReader *bytes.Reader
	if reqBody != nil {
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	var req *http.Request
	if bodyReader != nil {
		req, err = http.NewRequest(method, url, bodyReader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

// ── ESIMProvider Interface ──

// GetDestinations — GET /destinations
func (c *CelitechProvider) GetDestinations() ([]ESIMDestination, error) {
	if !c.isConfigured() {
		log.Println("[CELITECH] Not configured, using demo destinations")
		return getDemoDestinations(), nil
	}

	resp, err := c.doRequest("GET", "/destinations", nil)
	if err != nil {
		log.Printf("[CELITECH] ❌ GetDestinations error: %v", err)
		return getDemoDestinations(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[CELITECH] ❌ GetDestinations status=%d", resp.StatusCode)
		return getDemoDestinations(), nil
	}

	var apiResp struct {
		Destinations []struct {
			Name        string `json:"name"`
			Destination string `json:"destination"` // ISO country code
		} `json:"destinations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[CELITECH] ❌ GetDestinations decode: %v", err)
		return getDemoDestinations(), nil
	}

	var dests []ESIMDestination
	for _, d := range apiResp.Destinations {
		code := d.Destination
		if len(code) > 2 {
			// Celitech may return ISO3 codes; we store both but display ISO2 flag
		}
		dests = append(dests, ESIMDestination{
			CountryCode: code,
			CountryName: d.Name,
			FlagEmoji:   countryFlag(code),
			PlanCount:   1,
		})
	}
	if len(dests) == 0 {
		return getDemoDestinations(), nil
	}

	log.Printf("[CELITECH] ✅ Fetched %d destinations", len(dests))
	return dests, nil
}

// GetPlans — GET /packages?destination=XX
func (c *CelitechProvider) GetPlans(countryCode string) ([]ESIMPlan, error) {
	if !c.isConfigured() {
		log.Println("[CELITECH] Not configured, using demo plans")
		return getDemoPlans(countryCode), nil
	}

	path := fmt.Sprintf("/packages?destination=%s", countryCode)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		log.Printf("[CELITECH] ❌ GetPlans error for %s: %v", countryCode, err)
		return getDemoPlans(countryCode), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[CELITECH] ❌ GetPlans status=%d for %s", resp.StatusCode, countryCode)
		return getDemoPlans(countryCode), nil
	}

	var apiResp struct {
		Packages []struct {
			ID            string  `json:"id"`
			Destination   string  `json:"destination"`
			DataLimitInGB float64 `json:"dataLimitInGB"`
			Duration      int     `json:"duration"` // days
			Price         float64 `json:"price"`
		} `json:"packages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[CELITECH] ❌ GetPlans decode for %s: %v", countryCode, err)
		return getDemoPlans(countryCode), nil
	}

	// Cache packages + build response
	c.pkgMu.Lock()
	var plans []ESIMPlan
	for _, p := range apiResp.Packages {
		// Cache for later purchase
		c.pkgCache[p.ID] = &celitechPackage{
			ID:            p.ID,
			Destination:   p.Destination,
			DataLimitInGB: p.DataLimitInGB,
			Duration:      p.Duration,
			Price:         p.Price,
		}

		name := fmt.Sprintf("%s %.0f ГБ", countryCode, p.DataLimitInGB)
		plans = append(plans, ESIMPlan{
			PlanID:       p.ID,
			Provider:     "celitech",
			Name:         name,
			Country:      countryCode,
			CountryCode:  countryCode,
			DataGB:       fmt.Sprintf("%.0f", p.DataLimitInGB),
			ValidityDays: p.Duration,
			PriceUSD:     p.Price,
			Description:  fmt.Sprintf("%.0f ГБ на %d дней", p.DataLimitInGB, p.Duration),
			InStock:      true,
		})
	}
	c.pkgMu.Unlock()

	if len(plans) == 0 {
		return getDemoPlans(countryCode), nil
	}

	log.Printf("[CELITECH] ✅ Fetched %d plans for %s", len(plans), countryCode)
	return plans, nil
}

// OrderESIM — POST /purchases (createPurchaseV2)
func (c *CelitechProvider) OrderESIM(planID string) (*ESIMOrderResult, error) {
	if !c.isConfigured() {
		log.Println("[CELITECH] Not configured, using demo order")
		return getDemoOrder(planID)
	}

	// Look up cached package details
	c.pkgMu.RLock()
	pkg, ok := c.pkgCache[planID]
	c.pkgMu.RUnlock()

	if !ok {
		log.Printf("[CELITECH] ⚠️ Package %s not in cache, falling back to demo", planID)
		return getDemoOrder(planID)
	}

	// Build purchase request
	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, pkg.Duration).Format("2006-01-02")

	purchaseReq := map[string]interface{}{
		"destination":   pkg.Destination,
		"dataLimitInGb": pkg.DataLimitInGB,
		"startDate":     startDate,
		"endDate":       endDate,
		"quantity":      1,
	}

	log.Printf("[CELITECH] 🛒 Creating purchase: dest=%s data=%.0fGB start=%s end=%s",
		pkg.Destination, pkg.DataLimitInGB, startDate, endDate)

	resp, err := c.doRequest("POST", "/purchases", purchaseReq)
	if err != nil {
		return nil, fmt.Errorf("celitech purchase request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("celitech purchase failed: status %d", resp.StatusCode)
	}

	// Parse response — array of purchase+profile objects
	var purchaseResp []struct {
		Purchase struct {
			ID        string `json:"id"`
			PackageID string `json:"packageId"`
		} `json:"purchase"`
		Profile struct {
			ICCID                 string `json:"iccid"`
			ActivationCode        string `json:"activationCode"`
			ManualActivationCode  string `json:"manualActivationCode"`
			IOSActivationLink     string `json:"iosActivationLink"`
			AndroidActivationLink string `json:"androidActivationLink"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&purchaseResp); err != nil {
		return nil, fmt.Errorf("celitech purchase decode: %w", err)
	}

	if len(purchaseResp) == 0 {
		return nil, fmt.Errorf("celitech returned empty purchase array")
	}

	p := purchaseResp[0]

	// activationCode is the LPA string (e.g. "LPA:1$CELITECH.IDEMIA.IO$...")
	// which serves as both QR data and manual activation code
	qrData := p.Profile.ActivationCode
	if qrData == "" {
		qrData = p.Profile.ManualActivationCode
	}

	// Parse SMDP and matching ID from LPA string
	smdp, matchingID := "", ""
	if len(qrData) > 5 {
		parts := splitLPA(qrData)
		if len(parts) >= 2 {
			smdp = parts[0]
		}
		if len(parts) >= 3 {
			matchingID = parts[1]
		}
	}

	log.Printf("[CELITECH] ✅ Purchase OK: id=%s iccid=%s", p.Purchase.ID, p.Profile.ICCID)

	return &ESIMOrderResult{
		OrderID:     p.Purchase.ID,
		QRData:      qrData,
		LPA:         qrData,
		SMDP:        smdp,
		MatchingID:  matchingID,
		ICCID:       p.Profile.ICCID,
		ProviderRef: p.Purchase.ID,
	}, nil
}

// CheckAvailability — Celitech doesn't have a dedicated availability endpoint.
// If the package is in the catalog, it's available.
func (c *CelitechProvider) CheckAvailability(planID string) (bool, error) {
	if !c.isConfigured() {
		return true, nil
	}
	c.pkgMu.RLock()
	_, ok := c.pkgCache[planID]
	c.pkgMu.RUnlock()
	return ok, nil
}

// ── Helpers ──

// splitLPA parses "LPA:1$smdp.server.com$ACTIVATION-CODE" → ["smdp.server.com", "ACTIVATION-CODE"]
func splitLPA(lpa string) []string {
	// Strip "LPA:1$" prefix if present
	prefix := "LPA:1$"
	s := lpa
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		s = s[len(prefix):]
	}
	// Split by $
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '$' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
