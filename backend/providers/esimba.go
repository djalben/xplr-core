package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/djalben/xplr-core/backend/shop"
	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// Esimba Provider — official Esimba API v2 integration.
// Implements both ESIMProvider and shop.ProductProvider.
//
// Clean Architecture note: this layer ONLY performs network I/O and
// JSON (de)serialization. No business rules (markup, payments) here.
// ══════════════════════════════════════════════════════════════

// flexID unmarshals a JSON value that may be either a number or a string
// into a Go string (Esimba returns numeric IDs in some endpoints).
type flexID string

func (f *flexID) UnmarshalJSON(b []byte) error {
	*f = flexID(strings.Trim(string(b), `"`))
	return nil
}

func (f flexID) String() string { return string(f) }

// ── Esimba API DTOs ──

// esimbaRefill is a single data package attached to a bundle.
type esimbaRefill struct {
	ID         flexID  `json:"id"`
	RefillMB   int     `json:"refill_mb"`
	RefillDays int     `json:"refill_days"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}

// esimbaBundle is a country/region offer containing one or more refills.
type esimbaBundle struct {
	ID          flexID         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Country     string         `json:"country"`
	CountryCode string         `json:"country_code"`
	Refills     []esimbaRefill `json:"refills"`
}

// esimbaBundlesResponse is the envelope returned by GET /bundles.
type esimbaBundlesResponse struct {
	Bundles []esimbaBundle `json:"bundles"`
}

// esimbaLineResponse is the envelope returned by POST /line/create.
type esimbaLineResponse struct {
	SimCard struct {
		ICCID   string `json:"iccid"`
		LPACode string `json:"lpa_code"`
	} `json:"sim_card"`
}

// ── Provider ──

// EsimbaProvider talks to the Esimba REST API v2.
type EsimbaProvider struct {
	baseURL     string
	apiKey      string
	accessToken string
	client      *http.Client
}

// NewEsimbaProvider builds the provider from environment variables:
//
//	ESIMBA_API_URL      — base URL, e.g. https://panel.esimba.io/api/v2
//	ESIMBA_API_KEY      — apiKey header
//	ESIMBA_ACCESS_TOKEN — accessToken header
func NewEsimbaProvider() *EsimbaProvider {
	baseURL := strings.TrimRight(os.Getenv("ESIMBA_API_URL"), "/")
	apiKey := os.Getenv("ESIMBA_API_KEY")
	accessToken := os.Getenv("ESIMBA_ACCESS_TOKEN")

	if baseURL == "" || apiKey == "" || accessToken == "" {
		log.Println("🚨 [ESIMBA] ESIMBA_API_URL / ESIMBA_API_KEY / ESIMBA_ACCESS_TOKEN missing — eSIM orders will FAIL")
	}

	return &EsimbaProvider{
		baseURL:     baseURL,
		apiKey:      apiKey,
		accessToken: accessToken,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (e *EsimbaProvider) Name() string { return "esimba" }

func (e *EsimbaProvider) isConfigured() bool {
	return e.baseURL != "" && e.apiKey != "" && e.accessToken != ""
}

// doRequest performs an authenticated request against the Esimba API.
func (e *EsimbaProvider) doRequest(method, path string, body interface{}) (*http.Response, error) {
	if !e.isConfigured() {
		return nil, fmt.Errorf("esimba: provider not configured (missing env vars)")
	}

	var reqBody *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("esimba: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, e.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", e.apiKey)
	req.Header.Set("accessToken", e.accessToken)

	return e.client.Do(req)
}

// fetchBundles retrieves the raw catalog from GET /bundles?with_async=false.
func (e *EsimbaProvider) fetchBundles() ([]esimbaBundle, error) {
	resp, err := e.doRequest("GET", "/bundles?with_async=false", nil)
	if err != nil {
		return nil, fmt.Errorf("esimba: GET /bundles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("esimba: GET /bundles status %d", resp.StatusCode)
	}

	var parsed esimbaBundlesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("esimba: decode bundles: %w", err)
	}
	return parsed.Bundles, nil
}

// ── ESIMProvider interface ──

// GetDestinations derives the unique list of countries from the bundle catalog.
func (e *EsimbaProvider) GetDestinations() ([]ESIMDestination, error) {
	bundles, err := e.fetchBundles()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	names := map[string]string{}
	for _, b := range bundles {
		cc := strings.ToUpper(b.CountryCode)
		if cc == "" {
			continue
		}
		counts[cc] += len(b.Refills)
		if names[cc] == "" {
			names[cc] = b.Country
		}
	}

	dests := make([]ESIMDestination, 0, len(counts))
	for cc, n := range counts {
		dests = append(dests, ESIMDestination{
			CountryCode: cc,
			CountryName: names[cc],
			FlagEmoji:   countryFlag(cc),
			PlanCount:   n,
		})
	}
	return dests, nil
}

// GetPlans flattens every refill of every bundle for a country into ESIMPlan.
func (e *EsimbaProvider) GetPlans(countryCode string) ([]ESIMPlan, error) {
	bundles, err := e.fetchBundles()
	if err != nil {
		return nil, err
	}

	cc := strings.ToUpper(strings.TrimSpace(countryCode))
	var plans []ESIMPlan
	for _, b := range bundles {
		if cc != "" && strings.ToUpper(b.CountryCode) != cc {
			continue
		}
		for _, r := range b.Refills {
			plans = append(plans, ESIMPlan{
				PlanID:       encodeEsimbaPlanID(b.ID.String(), r.RefillMB, r.RefillDays),
				Provider:     "esimba",
				Name:         bundlePlanName(b, r),
				Country:      b.Country,
				CountryCode:  strings.ToUpper(b.CountryCode),
				DataGB:       formatMBasGB(r.RefillMB),
				ValidityDays: r.RefillDays,
				PriceUSD:     r.Price,
				CostPrice:    r.Price,
				Description:  b.Name,
				InStock:      true,
			})
		}
	}
	return plans, nil
}

// OrderESIM provisions a new eSIM line via POST /line/create.
func (e *EsimbaProvider) OrderESIM(planID string) (*ESIMOrderResult, error) {
	bundleID, refillMB, refillDays, err := decodeEsimbaPlanID(planID)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"bundle_id": bundleID,
		"refill_mb": refillMB,
	}
	// refill_days is required for Travel-type bundles.
	if refillDays > 0 {
		reqBody["refill_days"] = refillDays
	}

	resp, err := e.doRequest("POST", "/line/create", reqBody)
	if err != nil {
		return nil, fmt.Errorf("esimba: POST /line/create: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("esimba: POST /line/create status %d", resp.StatusCode)
	}

	var parsed esimbaLineResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("esimba: decode line response: %w", err)
	}

	if parsed.SimCard.ICCID == "" || parsed.SimCard.LPACode == "" {
		return nil, fmt.Errorf("esimba: line/create returned empty iccid/lpa_code")
	}

	log.Printf("[ESIMBA] ✅ Line created: iccid=%s", parsed.SimCard.ICCID)

	return &ESIMOrderResult{
		OrderID:     parsed.SimCard.ICCID,
		QRData:      parsed.SimCard.LPACode,
		LPA:         parsed.SimCard.LPACode,
		ICCID:       parsed.SimCard.ICCID,
		ProviderRef: parsed.SimCard.ICCID,
	}, nil
}

// CheckAvailability verifies the bundle referenced by planID still exists.
func (e *EsimbaProvider) CheckAvailability(planID string) (bool, error) {
	bundleID, _, _, err := decodeEsimbaPlanID(planID)
	if err != nil {
		return false, err
	}

	bundles, err := e.fetchBundles()
	if err != nil {
		return false, err
	}
	for _, b := range bundles {
		if b.ID.String() == bundleID {
			return true, nil
		}
	}
	return false, nil
}

// ── shop.ProductProvider interface ──

// GetCatalog returns the full Esimba catalog as shop.CatalogProduct entries.
func (e *EsimbaProvider) GetCatalog() ([]shop.CatalogProduct, error) {
	bundles, err := e.fetchBundles()
	if err != nil {
		return nil, err
	}

	var catalog []shop.CatalogProduct
	for _, b := range bundles {
		for _, r := range b.Refills {
			currency := r.Currency
			if currency == "" {
				currency = "USD"
			}
			catalog = append(catalog, shop.CatalogProduct{
				ExternalID:  encodeEsimbaPlanID(b.ID.String(), r.RefillMB, r.RefillDays),
				Name:        bundlePlanName(b, r),
				Description: b.Name,
				Category:    "esim",
				Country:     b.Country,
				CountryCode: strings.ToUpper(b.CountryCode),
				CostPrice:   decimal.NewFromFloat(r.Price),
				Currency:    currency,
				InStock:     true,
				Meta: map[string]any{
					"data_gb":       formatMBasGB(r.RefillMB),
					"validity_days": r.RefillDays,
					"refill_mb":     r.RefillMB,
					"bundle_id":     b.ID.String(),
					"bundle_type":   b.Type,
					"provider":      "esimba",
				},
			})
		}
	}
	return catalog, nil
}

// CreateOrder places an order and adapts the result to shop.OrderResult.
func (e *EsimbaProvider) CreateOrder(externalProductID string) (*shop.OrderResult, error) {
	result, err := e.OrderESIM(externalProductID)
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

// CheckStatus — Esimba provisioning is synchronous, so a known ICCID means completed.
func (e *EsimbaProvider) CheckStatus(providerRef string) (*shop.OrderStatus, error) {
	return &shop.OrderStatus{
		ProviderRef: providerRef,
		Status:      "completed",
	}, nil
}

// GetBalance — Esimba API does not expose a deposit balance endpoint.
func (e *EsimbaProvider) GetBalance() (*shop.BalanceInfo, error) {
	return nil, nil
}

// ── PlanID codec ──
// A plan is uniquely identified by (bundle_id, refill_mb, refill_days).
// We pack them into a single opaque string so existing handlers that pass
// a single plan_id keep working.

func encodeEsimbaPlanID(bundleID string, refillMB, refillDays int) string {
	return fmt.Sprintf("%s|%d|%d", bundleID, refillMB, refillDays)
}

func decodeEsimbaPlanID(planID string) (bundleID string, refillMB, refillDays int, err error) {
	parts := strings.Split(planID, "|")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("esimba: invalid plan id %q", planID)
	}
	bundleID = parts[0]
	if bundleID == "" {
		return "", 0, 0, fmt.Errorf("esimba: empty bundle id in plan %q", planID)
	}
	if refillMB, err = strconv.Atoi(parts[1]); err != nil {
		return "", 0, 0, fmt.Errorf("esimba: invalid refill_mb in plan %q: %w", planID, err)
	}
	if refillDays, err = strconv.Atoi(parts[2]); err != nil {
		return "", 0, 0, fmt.Errorf("esimba: invalid refill_days in plan %q: %w", planID, err)
	}
	return bundleID, refillMB, refillDays, nil
}

// ── small formatting helpers ──

func formatMBasGB(mb int) string {
	if mb <= 0 {
		return "0"
	}
	gb := float64(mb) / 1024.0
	if gb >= 1 {
		return strconv.FormatFloat(gb, 'f', -1, 64)
	}
	return fmt.Sprintf("%dMB", mb)
}

func bundlePlanName(b esimbaBundle, r esimbaRefill) string {
	country := b.Country
	if country == "" {
		country = strings.ToUpper(b.CountryCode)
	}
	data := formatMBasGB(r.RefillMB)
	if r.RefillDays > 0 {
		return fmt.Sprintf("%s %s · %d дн.", country, data, r.RefillDays)
	}
	return fmt.Sprintf("%s %s", country, data)
}

// Compile-time checks: EsimbaProvider implements both interfaces.
var _ ESIMProvider = (*EsimbaProvider)(nil)
var _ shop.ProductProvider = (*EsimbaProvider)(nil)

// Esimba singleton.
var (
	esimbaProvider     *EsimbaProvider
	esimbaProviderOnce sync.Once
)

// getEsimbaProvider returns the lazily-initialized Esimba singleton.
func getEsimbaProvider() *EsimbaProvider {
	esimbaProviderOnce.Do(func() {
		esimbaProvider = NewEsimbaProvider()
		log.Printf("[ESIM-PROVIDER] Initialized: esimba (configured=%v)", esimbaProvider.isConfigured())
	})
	return esimbaProvider
}
