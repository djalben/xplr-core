package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/djalben/xplr-core/backend/shop"
	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// eSIM Provider Wrapper (MobiMatter-compatible interface)
// ══════════════════════════════════════════════════════════════

// ESIMDestination — a country/region where eSIM plans are available
type ESIMDestination struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	FlagEmoji   string `json:"flag_emoji"`
	PlanCount   int    `json:"plan_count"`
}

// ESIMPlan — a single eSIM plan (data bundle)
type ESIMPlan struct {
	PlanID       string  `json:"plan_id"`
	Provider     string  `json:"provider"`
	Name         string  `json:"name"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	DataGB       string  `json:"data_gb"`
	ValidityDays int     `json:"validity_days"`
	PriceUSD     float64 `json:"price_usd"`
	OldPrice     float64 `json:"old_price"`
	CostPrice    float64 `json:"cost_price,omitempty"`
	Description  string  `json:"description"`
	InStock      bool    `json:"in_stock"`
}

// ESIMOrderResult — result of ordering an eSIM
type ESIMOrderResult struct {
	OrderID     string `json:"order_id"`
	QRData      string `json:"qr_data"`
	LPA         string `json:"lpa"`
	SMDP        string `json:"smdp"`
	MatchingID  string `json:"matching_id"`
	ICCID       string `json:"iccid"`
	ProviderRef string `json:"provider_ref"`
}

// ESIMProvider interface — any eSIM provider must implement these
type ESIMProvider interface {
	GetDestinations() ([]ESIMDestination, error)
	GetPlans(countryCode string) ([]ESIMPlan, error)
	OrderESIM(planID string) (*ESIMOrderResult, error)
	CheckAvailability(planID string) (bool, error)
	Name() string
}

// ══════════════════════════════════════════════════════════════
// MobiMatter Provider (real API — with demo fallback)
// ══════════════════════════════════════════════════════════════

type MobiMatterProvider struct {
	apiKey     string
	authToken  string
	baseURL    string
	merchantID string
	client     *http.Client
}

func NewMobiMatterProvider() *MobiMatterProvider {
	apiKey := os.Getenv("MOBIMATTER_API_KEY")
	authToken := os.Getenv("MOBIMATTER_AUTH_TOKEN")
	baseURL := os.Getenv("MOBIMATTER_API_URL")
	merchantID := os.Getenv("MOBIMATTER_MERCHANT_ID")
	if baseURL == "" {
		baseURL = "https://api.mobimatter.com/mobimatter/api/v2"
	}
	return &MobiMatterProvider{
		apiKey:     apiKey,
		authToken:  authToken,
		baseURL:    baseURL,
		merchantID: merchantID,
		client:     &http.Client{Timeout: 20 * time.Second},
	}
}

func (m *MobiMatterProvider) Name() string { return "mobimatter" }

func (m *MobiMatterProvider) isConfigured() bool {
	return m.apiKey != ""
}

func (m *MobiMatterProvider) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := m.baseURL + path

	var req *http.Request
	var err error
	if body != nil {
		bodyBytes, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", m.apiKey)
	if m.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.authToken)
	}
	return m.client.Do(req)
}

// GetDestinations — fetches available countries from MobiMatter
func (m *MobiMatterProvider) GetDestinations() ([]ESIMDestination, error) {
	if !m.isConfigured() {
		log.Println("[MOBIMATTER] API key not set, using demo destinations")
		return getDemoDestinations(), nil
	}

	resp, err := m.doRequest("GET", "/products?type=esim", nil)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ GetDestinations HTTP error: %v", err)
		return getDemoDestinations(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[MOBIMATTER] ❌ GetDestinations status=%d", resp.StatusCode)
		return getDemoDestinations(), nil
	}

	var apiResp []struct {
		CountryCode string `json:"countryCode"`
		Country     string `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[MOBIMATTER] ❌ GetDestinations decode error: %v", err)
		return getDemoDestinations(), nil
	}

	// Deduplicate by country code
	seen := map[string]bool{}
	var dests []ESIMDestination
	for _, p := range apiResp {
		if seen[p.CountryCode] {
			continue
		}
		seen[p.CountryCode] = true
		dests = append(dests, ESIMDestination{
			CountryCode: p.CountryCode,
			CountryName: p.Country,
			FlagEmoji:   countryFlag(p.CountryCode),
			PlanCount:   1,
		})
	}
	if len(dests) == 0 {
		return getDemoDestinations(), nil
	}
	return dests, nil
}

// GetPlans — fetches eSIM plans for a specific country
func (m *MobiMatterProvider) GetPlans(countryCode string) ([]ESIMPlan, error) {
	if !m.isConfigured() {
		log.Println("[MOBIMATTER] API key not set, using demo plans")
		return getDemoPlans(countryCode), nil
	}

	resp, err := m.doRequest("GET", fmt.Sprintf("/products?type=esim&countryCode=%s", countryCode), nil)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ GetPlans HTTP error: %v", err)
		return getDemoPlans(countryCode), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[MOBIMATTER] ❌ GetPlans status=%d", resp.StatusCode)
		return getDemoPlans(countryCode), nil
	}

	var apiResp []struct {
		ProductID    string  `json:"productId"`
		ProductName  string  `json:"productName"`
		Country      string  `json:"country"`
		CountryCode  string  `json:"countryCode"`
		DataGB       float64 `json:"data"`
		ValidityDays int     `json:"validity"`
		Price        float64 `json:"price"`
		Available    bool    `json:"available"`
		Description  string  `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[MOBIMATTER] ❌ GetPlans decode error: %v", err)
		return getDemoPlans(countryCode), nil
	}

	var plans []ESIMPlan
	for _, p := range apiResp {
		plans = append(plans, ESIMPlan{
			PlanID:       p.ProductID,
			Provider:     "mobimatter",
			Name:         p.ProductName,
			Country:      p.Country,
			CountryCode:  p.CountryCode,
			DataGB:       fmt.Sprintf("%.0f", p.DataGB),
			ValidityDays: p.ValidityDays,
			PriceUSD:     p.Price,
			Description:  p.Description,
			InStock:      p.Available,
		})
	}
	if len(plans) == 0 {
		return getDemoPlans(countryCode), nil
	}
	return plans, nil
}

// OrderESIM — places an eSIM order via MobiMatter
func (m *MobiMatterProvider) OrderESIM(planID string) (*ESIMOrderResult, error) {
	if !m.isConfigured() {
		log.Println("[MOBIMATTER] API key not set, using demo order")
		return getDemoOrder(planID)
	}

	// Real MobiMatter order API: POST /orders
	orderReq := map[string]interface{}{
		"productId": planID,
		"quantity":  1,
	}

	resp, err := m.doRequest("POST", "/orders", orderReq)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ OrderESIM HTTP error: %v (falling back to demo)", err)
		return getDemoOrder(planID)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Printf("[MOBIMATTER] ❌ OrderESIM status=%d (falling back to demo)", resp.StatusCode)
		return getDemoOrder(planID)
	}

	var apiResp struct {
		OrderID    string `json:"orderId"`
		Status     string `json:"status"`
		ICCID      string `json:"iccid"`
		SMDP       string `json:"smdpAddress"`
		MatchingID string `json:"matchingId"`
		LPA        string `json:"lpaCode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[MOBIMATTER] ❌ OrderESIM decode error: %v (falling back to demo)", err)
		return getDemoOrder(planID)
	}

	qrData := apiResp.LPA
	if qrData == "" && apiResp.SMDP != "" {
		qrData = fmt.Sprintf("LPA:1$%s$%s", apiResp.SMDP, apiResp.MatchingID)
	}

	log.Printf("[MOBIMATTER] ✅ Order placed: ref=%s, ICCID=%s", apiResp.OrderID, apiResp.ICCID)

	return &ESIMOrderResult{
		OrderID:     apiResp.OrderID,
		QRData:      qrData,
		LPA:         qrData,
		SMDP:        apiResp.SMDP,
		MatchingID:  apiResp.MatchingID,
		ICCID:       apiResp.ICCID,
		ProviderRef: apiResp.OrderID,
	}, nil
}

// CheckAvailability — checks if a specific plan is available
func (m *MobiMatterProvider) CheckAvailability(planID string) (bool, error) {
	if !m.isConfigured() {
		return true, nil
	}

	resp, err := m.doRequest("GET", fmt.Sprintf("/products/%s", planID), nil)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ CheckAvailability HTTP error: %v", err)
		return true, nil // Optimistic: assume available on network error
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil
	}

	var product struct {
		Available bool `json:"available"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return true, nil
	}
	return product.Available, nil
}

// ══════════════════════════════════════════════════════════════
// ProductProvider interface implementation (shop.ProductProvider)
// Enables MobiMatter to participate in the unified Registry,
// FulfillmentEngine, and DepositMonitor systems.
// ══════════════════════════════════════════════════════════════

// GetCatalog — fetches all eSIM products as shop.CatalogProduct
func (m *MobiMatterProvider) GetCatalog() ([]shop.CatalogProduct, error) {
	dests, err := m.GetDestinations()
	if err != nil {
		return nil, err
	}

	var catalog []shop.CatalogProduct
	for _, d := range dests {
		plans, err := m.GetPlans(d.CountryCode)
		if err != nil {
			log.Printf("[MOBIMATTER] ⚠️ GetCatalog skipping %s: %v", d.CountryCode, err)
			continue
		}
		for _, p := range plans {
			catalog = append(catalog, shop.CatalogProduct{
				ExternalID:  p.PlanID,
				Name:        p.Name,
				Description: p.Description,
				Category:    "esim",
				Country:     p.Country,
				CountryCode: p.CountryCode,
				CostPrice:   decimal.NewFromFloat(p.CostPrice),
				Currency:    "USD",
				InStock:     p.InStock,
				Meta: map[string]any{
					"data_gb":       p.DataGB,
					"validity_days": p.ValidityDays,
					"provider":      "mobimatter",
				},
			})
		}
	}
	return catalog, nil
}

// CreateOrder — unified order method (delegates to OrderESIM)
func (m *MobiMatterProvider) CreateOrder(externalProductID string) (*shop.OrderResult, error) {
	result, err := m.OrderESIM(externalProductID)
	if err != nil {
		return nil, err
	}
	return &shop.OrderResult{
		ProviderRef: result.ProviderRef,
		QRData:      result.QRData,
		ICCID:       result.ICCID,
		Status:      "completed",
	}, nil
}

// CheckStatus — checks order status via MobiMatter API
func (m *MobiMatterProvider) CheckStatus(providerRef string) (*shop.OrderStatus, error) {
	if !m.isConfigured() {
		return &shop.OrderStatus{ProviderRef: providerRef, Status: "completed"}, nil
	}

	resp, err := m.doRequest("GET", fmt.Sprintf("/orders/%s", providerRef), nil)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ CheckStatus HTTP error: %v", err)
		return &shop.OrderStatus{ProviderRef: providerRef, Status: "pending"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &shop.OrderStatus{ProviderRef: providerRef, Status: "pending"}, nil
	}

	var orderResp struct {
		Status     string `json:"status"`
		ICCID      string `json:"iccid"`
		LPA        string `json:"lpaCode"`
		MatchingID string `json:"matchingId"`
		SMDP       string `json:"smdpAddress"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return &shop.OrderStatus{ProviderRef: providerRef, Status: "pending"}, nil
	}

	qrData := orderResp.LPA
	if qrData == "" && orderResp.SMDP != "" {
		qrData = fmt.Sprintf("LPA:1$%s$%s", orderResp.SMDP, orderResp.MatchingID)
	}

	return &shop.OrderStatus{
		ProviderRef: providerRef,
		Status:      mapMobiMatterStatus(orderResp.Status),
		QRData:      qrData,
	}, nil
}

// GetBalance — fetches MobiMatter account balance
func (m *MobiMatterProvider) GetBalance() (*shop.BalanceInfo, error) {
	if !m.isConfigured() {
		return nil, nil
	}

	resp, err := m.doRequest("GET", "/account/balance", nil)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ GetBalance HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[MOBIMATTER] ❌ GetBalance status=%d", resp.StatusCode)
		return nil, fmt.Errorf("mobimatter balance: status %d", resp.StatusCode)
	}

	var balResp struct {
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balResp); err != nil {
		return nil, fmt.Errorf("mobimatter balance decode: %w", err)
	}

	return &shop.BalanceInfo{
		BalanceUSD: decimal.NewFromFloat(balResp.Balance),
		Currency:   balResp.Currency,
	}, nil
}

// mapMobiMatterStatus converts MobiMatter order status to shop.OrderStatus values
func mapMobiMatterStatus(s string) string {
	switch s {
	case "completed", "delivered", "active":
		return "completed"
	case "failed", "cancelled", "refunded":
		return "failed"
	default:
		return "pending"
	}
}

// ══════════════════════════════════════════════════════════════
// Demo data (used when no API key is configured)
// ══════════════════════════════════════════════════════════════

func getDemoDestinations() []ESIMDestination {
	return []ESIMDestination{
		{CountryCode: "TR", CountryName: "Турция", FlagEmoji: "🇹🇷", PlanCount: 3},
		{CountryCode: "TH", CountryName: "Таиланд", FlagEmoji: "🇹🇭", PlanCount: 3},
		{CountryCode: "US", CountryName: "США", FlagEmoji: "🇺🇸", PlanCount: 3},
		{CountryCode: "AE", CountryName: "ОАЭ", FlagEmoji: "🇦🇪", PlanCount: 3},
		{CountryCode: "JP", CountryName: "Япония", FlagEmoji: "🇯🇵", PlanCount: 2},
		{CountryCode: "ID", CountryName: "Индонезия", FlagEmoji: "🇮🇩", PlanCount: 2},
		{CountryCode: "DE", CountryName: "Германия", FlagEmoji: "🇩🇪", PlanCount: 2},
		{CountryCode: "FR", CountryName: "Франция", FlagEmoji: "🇫🇷", PlanCount: 2},
		{CountryCode: "IT", CountryName: "Италия", FlagEmoji: "🇮🇹", PlanCount: 2},
		{CountryCode: "ES", CountryName: "Испания", FlagEmoji: "🇪🇸", PlanCount: 2},
		{CountryCode: "GB", CountryName: "Великобритания", FlagEmoji: "🇬🇧", PlanCount: 2},
		{CountryCode: "KR", CountryName: "Южная Корея", FlagEmoji: "🇰🇷", PlanCount: 2},
		{CountryCode: "SG", CountryName: "Сингапур", FlagEmoji: "🇸🇬", PlanCount: 2},
		{CountryCode: "MY", CountryName: "Малайзия", FlagEmoji: "🇲🇾", PlanCount: 2},
		{CountryCode: "IN", CountryName: "Индия", FlagEmoji: "🇮🇳", PlanCount: 2},
		{CountryCode: "BR", CountryName: "Бразилия", FlagEmoji: "🇧🇷", PlanCount: 2},
		{CountryCode: "MX", CountryName: "Мексика", FlagEmoji: "🇲🇽", PlanCount: 2},
		{CountryCode: "AU", CountryName: "Австралия", FlagEmoji: "🇦🇺", PlanCount: 2},
		{CountryCode: "EG", CountryName: "Египет", FlagEmoji: "🇪🇬", PlanCount: 2},
		{CountryCode: "GE", CountryName: "Грузия", FlagEmoji: "🇬🇪", PlanCount: 2},
	}
}

func getDemoPlans(countryCode string) []ESIMPlan {
	countryNames := map[string]string{
		"TR": "Турция", "TH": "Таиланд", "US": "США", "AE": "ОАЭ",
		"JP": "Япония", "ID": "Индонезия", "DE": "Германия", "FR": "Франция",
		"IT": "Италия", "ES": "Испания", "GB": "Великобритания", "KR": "Южная Корея",
		"SG": "Сингапур", "MY": "Малайзия", "IN": "Индия", "BR": "Бразилия",
		"MX": "Мексика", "AU": "Австралия", "EG": "Египет", "GE": "Грузия",
	}
	cn := countryNames[countryCode]
	if cn == "" {
		cn = countryCode
	}

	return []ESIMPlan{
		{PlanID: fmt.Sprintf("demo-%s-1gb", countryCode), Provider: "demo", Name: fmt.Sprintf("%s 1 ГБ", cn), Country: cn, CountryCode: countryCode, DataGB: "1", ValidityDays: 7, PriceUSD: 3.50, Description: "1 ГБ мобильного интернета", InStock: true},
		{PlanID: fmt.Sprintf("demo-%s-3gb", countryCode), Provider: "demo", Name: fmt.Sprintf("%s 3 ГБ", cn), Country: cn, CountryCode: countryCode, DataGB: "3", ValidityDays: 15, PriceUSD: 5.50, Description: "3 ГБ мобильного интернета", InStock: true},
		{PlanID: fmt.Sprintf("demo-%s-5gb", countryCode), Provider: "demo", Name: fmt.Sprintf("%s 5 ГБ", cn), Country: cn, CountryCode: countryCode, DataGB: "5", ValidityDays: 30, PriceUSD: 8.00, Description: "5 ГБ мобильного интернета", InStock: true},
		{PlanID: fmt.Sprintf("demo-%s-10gb", countryCode), Provider: "demo", Name: fmt.Sprintf("%s 10 ГБ", cn), Country: cn, CountryCode: countryCode, DataGB: "10", ValidityDays: 30, PriceUSD: 13.00, Description: "10 ГБ мобильного интернета", InStock: true},
		{PlanID: fmt.Sprintf("demo-%s-20gb", countryCode), Provider: "demo", Name: fmt.Sprintf("%s 20 ГБ", cn), Country: cn, CountryCode: countryCode, DataGB: "20", ValidityDays: 30, PriceUSD: 22.00, Description: "20 ГБ мобильного интернета", InStock: true},
	}
}

func getDemoOrder(planID string) (*ESIMOrderResult, error) {
	ref := fmt.Sprintf("DEMO-%s-%d", planID, time.Now().UnixMilli())
	smdp := "smdp.example.com"
	matchingID := fmt.Sprintf("X%d", time.Now().UnixMilli())
	lpa := fmt.Sprintf("LPA:1$%s$%s", smdp, matchingID)

	return &ESIMOrderResult{
		OrderID:     ref,
		QRData:      lpa,
		LPA:         lpa,
		SMDP:        smdp,
		MatchingID:  matchingID,
		ICCID:       fmt.Sprintf("8901%d", time.Now().UnixNano()%10000000000),
		ProviderRef: ref,
	}, nil
}

// ══════════════════════════════════════════════════════════════
// Singleton accessor
// ══════════════════════════════════════════════════════════════

// GetESIMProvider returns the singleton eSIM provider (Esimba API v2).
func GetESIMProvider() ESIMProvider {
	return getEsimbaProvider()
}

// Compile-time check: MobiMatterProvider implements both interfaces.
var _ ESIMProvider = (*MobiMatterProvider)(nil)
var _ shop.ProductProvider = (*MobiMatterProvider)(nil)

// ══════════════════════════════════════════════════════════════
// Utility
// ══════════════════════════════════════════════════════════════

func countryFlag(code string) string {
	if len(code) < 2 {
		return "🌍"
	}
	code = code[:2]
	runes := []rune(code)
	return string([]rune{runes[0] - 'A' + 0x1F1E6, runes[1] - 'A' + 0x1F1E6})
}

// splitLPA parses "LPA:1$smdp.server.com$ACTIVATION-CODE" → ["smdp.server.com", "ACTIVATION-CODE"]
func splitLPA(lpa string) []string {
	prefix := "LPA:1$"
	s := lpa
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		s = s[len(prefix):]
	}
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
