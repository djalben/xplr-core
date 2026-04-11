package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewMobiMatterProvider() *MobiMatterProvider {
	apiKey := os.Getenv("MOBIMATTER_API_KEY")
	baseURL := os.Getenv("MOBIMATTER_API_URL")
	if baseURL == "" {
		baseURL = "https://api.mobimatter.com/mobimatter/api/v2"
	}
	return &MobiMatterProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *MobiMatterProvider) Name() string { return "mobimatter" }

func (m *MobiMatterProvider) isConfigured() bool {
	return m.apiKey != ""
}

func (m *MobiMatterProvider) doRequest(method, path string) (*http.Response, error) {
	url := m.baseURL + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-key", m.apiKey)
	return m.client.Do(req)
}

// GetDestinations — fetches available countries from MobiMatter
func (m *MobiMatterProvider) GetDestinations() ([]ESIMDestination, error) {
	if !m.isConfigured() {
		log.Println("[MOBIMATTER] API key not set, using demo destinations")
		return getDemoDestinations(), nil
	}

	resp, err := m.doRequest("GET", "/products?type=esim")
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

	resp, err := m.doRequest("GET", fmt.Sprintf("/products?type=esim&countryCode=%s", countryCode))
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

	// TODO: Real MobiMatter order API
	// POST /orders  { "productId": planID, "quantity": 1 }
	// Response contains QR code data, SMDP address, matching ID
	log.Printf("[MOBIMATTER] 🔧 OrderESIM called for plan %s (real API pending)", planID)
	return getDemoOrder(planID)
}

// CheckAvailability — checks if a specific plan is available
func (m *MobiMatterProvider) CheckAvailability(planID string) (bool, error) {
	if !m.isConfigured() {
		return true, nil
	}

	resp, err := m.doRequest("GET", fmt.Sprintf("/products/%s", planID))
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

var (
	esimProvider     ESIMProvider
	esimProviderOnce sync.Once
)

func GetESIMProvider() ESIMProvider {
	esimProviderOnce.Do(func() {
		esimProvider = NewMobiMatterProvider()
		log.Printf("[ESIM-PROVIDER] Initialized: %s (configured=%v)",
			esimProvider.Name(), esimProvider.(*MobiMatterProvider).isConfigured())
	})
	return esimProvider
}

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
