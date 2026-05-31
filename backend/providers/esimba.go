package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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

// ── Keepgo (Esimba) API DTOs ──
// Schema confirmed live against https://myaccount.keepgo.com/api/v2/bundles:
//   {"ack":"success","bundles":[{"id":1,"bundle_type":"regional","name":"Orion",
//     "img":".../flags/3x2/vn.svg","coverage":["Vietnam"],
//     "refills":[{"title":"3 GB / 30 days","amount_mb":3072,"amount_days":30,"price_usd":5.53}]}]}

// esimbaRefill is a single data package attached to a bundle.
type esimbaRefill struct {
	Title      string  `json:"title"`
	AmountMB   int     `json:"amount_mb"`
	AmountDays *int    `json:"amount_days"` // nullable (regional volume bundles have no validity)
	PriceUSD   float64 `json:"price_usd"`
}

// days returns the validity in days (0 when the API sends null).
func (r esimbaRefill) days() int {
	if r.AmountDays == nil {
		return 0
	}
	return *r.AmountDays
}

// esimbaBundle is a country/regional offer containing one or more refills.
type esimbaBundle struct {
	ID          flexID         `json:"id"`
	Name        string         `json:"name"`
	BundleType  string         `json:"bundle_type"` // "country" | "regional"
	ProductType string         `json:"product_type"`
	Img         string         `json:"img"`
	Description string         `json:"description"`
	Coverage    []string       `json:"coverage"`
	Refills     []esimbaRefill `json:"refills"`
}

// esimbaBundlesResponse is the envelope returned by GET /bundles.
type esimbaBundlesResponse struct {
	Ack     string         `json:"ack"`
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

	// In-memory bundle cache (guarded by cacheMu) to avoid hitting the slow
	// Keepgo API + re-parsing on every GetDestinations/GetPlans call.
	cacheMu       sync.RWMutex
	cachedBundles []esimbaBundle
	cacheTime     time.Time
}

// bundleCacheTTL is how long a cached catalog stays fresh.
const bundleCacheTTL = 15 * time.Minute

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

// fetchBundles returns the bundle catalog, served from a 15-minute in-memory
// cache when fresh. Uses RWMutex: a read-lock fast path for cache hits and a
// write-lock slow path (with double-check) that performs the single API call.
func (e *EsimbaProvider) fetchBundles() ([]esimbaBundle, error) {
	// Fast path — read lock, return cached data if still fresh.
	e.cacheMu.RLock()
	if e.cachedBundles != nil && time.Since(e.cacheTime) < bundleCacheTTL {
		bundles, age := e.cachedBundles, time.Since(e.cacheTime)
		e.cacheMu.RUnlock()
		log.Printf("[ESIMBA] cache HIT — %d bundles (age %s)", len(bundles), age.Round(time.Second))
		return bundles, nil
	}
	e.cacheMu.RUnlock()

	// Slow path — write lock, re-check (another goroutine may have filled it),
	// then make the real network call exactly once.
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()
	if e.cachedBundles != nil && time.Since(e.cacheTime) < bundleCacheTTL {
		return e.cachedBundles, nil
	}

	log.Printf("[ESIMBA] cache MISS — fetching fresh catalog from Keepgo")
	bundles, err := e.fetchBundlesFromAPI()
	if err != nil {
		return nil, err
	}
	e.cachedBundles = bundles
	e.cacheTime = time.Now()
	return bundles, nil
}

// fetchBundlesFromAPI performs the raw GET /bundles?with_async=false request.
func (e *EsimbaProvider) fetchBundlesFromAPI() ([]esimbaBundle, error) {
	const path = "/bundles?with_async=false"
	fullURL := e.baseURL + path

	log.Printf("[ESIMBA-DEBUG] → GET %s (apiKey set=%v, accessToken set=%v)", fullURL, e.apiKey != "", e.accessToken != "")

	resp, err := e.doRequest("GET", path, nil)
	if err != nil {
		log.Printf("[ESIMBA-DEBUG] ❌ request error for %s: %v", fullURL, err)
		return nil, fmt.Errorf("esimba: GET /bundles: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[ESIMBA-DEBUG] ❌ %s status=%d, body read error: %v", fullURL, resp.StatusCode, readErr)
		return nil, fmt.Errorf("esimba: read bundles body: %w", readErr)
	}

	// Raw debug output — visible in Vercel logs to diagnose 401/JSON errors.
	log.Printf("[ESIMBA-DEBUG] ← %s | HTTP %d", fullURL, resp.StatusCode)
	log.Printf("[ESIMBA-DEBUG] Raw response body: %s", string(raw))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("esimba: GET /bundles status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed esimbaBundlesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		log.Printf("[ESIMBA-DEBUG] ❌ JSON parse error: %v", err)
		return nil, fmt.Errorf("esimba: decode bundles: %w", err)
	}

	log.Printf("[ESIMBA-DEBUG] ✅ parsed %d bundles", len(parsed.Bundles))
	return parsed.Bundles, nil
}

// retailMarkup is the approved unit-economics multiplier (RetailPrice = wholesale × 5).
const retailMarkup = 5.0

// round2 rounds a monetary value to 2 decimal places.
func round2(v float64) float64 { return math.Round(v*100) / 100 }

// applySmartMarkup converts a wholesale USD price into a polished retail price:
//   - base markup ×5
//   - $2.00–$5.00  → rounded to the nearest $0.50  (e.g. 2.44 → 2.50)
//   - above $5.00  → rounded up to the next ".90"  (e.g. 11.55 → 11.90)
//   - below $2.00  → rounded to 2 decimals
func applySmartMarkup(wholesale float64) float64 {
	retail := wholesale * retailMarkup
	switch {
	case retail >= 2.0 && retail <= 5.0:
		return math.Round(retail*2) / 2
	case retail > 5.0:
		whole := math.Floor(retail)
		if retail <= whole+0.90 {
			return whole + 0.90
		}
		return whole + 1.90
	default:
		return round2(retail)
	}
}

// ── ESIMProvider interface ──

// GetDestinations derives the unique list of countries from the bundle catalog.
func (e *EsimbaProvider) GetDestinations() ([]ESIMDestination, error) {
	log.Printf("[ESIMBA] GetDestinations: fetching catalog from %s", e.baseURL)
	bundles, err := e.fetchBundles()
	if err != nil {
		log.Printf("[ESIMBA] GetDestinations failed: %v", err)
		return nil, err
	}

	counts := map[string]int{}
	names := map[string]string{}
	for _, b := range bundles {
		for code, name := range b.coveredCountries() {
			counts[code] += len(b.Refills)
			if names[code] == "" {
				names[code] = name
			}
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
	log.Printf("[ESIMBA] GetDestinations: %d countries derived from %d bundles", len(dests), len(bundles))
	return dests, nil
}

// GetPlans flattens every refill of every bundle for a country into ESIMPlan.
// Retail price is computed here as wholesale × retailMarkup, rounded to 2dp.
func (e *EsimbaProvider) GetPlans(countryCode string) ([]ESIMPlan, error) {
	log.Printf("[ESIMBA] GetPlans(%s): fetching catalog from %s", countryCode, e.baseURL)
	bundles, err := e.fetchBundles()
	if err != nil {
		log.Printf("[ESIMBA] GetPlans(%s) failed: %v", countryCode, err)
		return nil, err
	}

	cc := strings.ToUpper(strings.TrimSpace(countryCode))
	var plans []ESIMPlan
	for _, b := range bundles {
		covered := b.coveredCountries()
		countryName, ok := covered[cc]
		if cc != "" && !ok {
			continue
		}
		for _, r := range b.Refills {
			retail := applySmartMarkup(r.PriceUSD)
			plans = append(plans, ESIMPlan{
				PlanID:       encodeEsimbaPlanID(b.ID.String(), r.AmountMB, r.days()),
				Provider:     "esimba",
				Name:         bundlePlanName(r),
				Country:      countryName,
				CountryCode:  cc,
				DataGB:       formatMBasGB(r.AmountMB),
				ValidityDays: r.days(),
				PriceUSD:     retail,               // retail = wholesale × 5
				OldPrice:     round2(retail * 1.2), // strike-through reference price
				CostPrice:    r.PriceUSD,           // original wholesale price
				Description:  b.Description,
				InStock:      true,
			})
		}
	}
	log.Printf("[ESIMBA] GetPlans(%s): %d plans (markup ×%.0f applied)", countryCode, len(plans), retailMarkup)
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
		// Pick a representative country code/name for the catalog entry.
		var code, name string
		for c, n := range b.coveredCountries() {
			code, name = c, n
			break
		}
		for _, r := range b.Refills {
			catalog = append(catalog, shop.CatalogProduct{
				ExternalID:  encodeEsimbaPlanID(b.ID.String(), r.AmountMB, r.days()),
				Name:        bundlePlanName(r),
				Description: b.Description,
				Category:    "esim",
				Country:     name,
				CountryCode: code,
				CostPrice:   decimal.NewFromFloat(r.PriceUSD),
				Currency:    "USD",
				InStock:     true,
				Meta: map[string]any{
					"data_gb":       formatMBasGB(r.AmountMB),
					"validity_days": r.days(),
					"amount_mb":     r.AmountMB,
					"bundle_id":     b.ID.String(),
					"bundle_type":   b.BundleType,
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

// bundlePlanName builds a clean, user-facing tariff name purely from refill
// data — e.g. "1 GB / 30 дней". The Keepgo bundle name (Andromeda/Orion/
// Eridanus, etc.) is an internal system ID and must NEVER reach the frontend.
func bundlePlanName(r esimbaRefill) string {
	data := dataLabel(r.AmountMB)
	if r.days() > 0 {
		return fmt.Sprintf("%s / %d %s", data, r.days(), pluralDays(r.days()))
	}
	return fmt.Sprintf("%s / Бессрочно", data)
}

// dataLabel renders a traffic volume as a human string: "1 GB", "1.5 GB", "512 MB".
func dataLabel(mb int) string {
	if mb <= 0 {
		return "0 MB"
	}
	if mb%1024 == 0 {
		return fmt.Sprintf("%d GB", mb/1024)
	}
	if mb >= 1024 {
		gb := float64(mb) / 1024.0
		return strconv.FormatFloat(gb, 'f', 1, 64) + " GB"
	}
	return fmt.Sprintf("%d MB", mb)
}

// pluralDays returns the correct Russian plural for the number of days.
func pluralDays(n int) string {
	mod100 := n % 100
	if mod100 >= 11 && mod100 <= 14 {
		return "дней"
	}
	switch n % 10 {
	case 1:
		return "день"
	case 2, 3, 4:
		return "дня"
	default:
		return "дней"
	}
}

// ── Country resolution ──
// Keepgo bundles carry no ISO code: country bundles encode it in the flag image
// URL (".../flags/3x2/XX.svg"), while regional bundles only list `coverage`
// country names. coveredCountries maps a bundle to {ISO-2: display name}.

func (b esimbaBundle) coveredCountries() map[string]string {
	out := map[string]string{}

	// Country-type bundles: authoritative ISO-2 from the flag image.
	if code := isoFromFlagImg(b.Img); code != "" {
		name := b.Name
		if len(b.Coverage) == 1 {
			name = b.Coverage[0]
		}
		out[code] = name
	}

	// Expand coverage names (handles regional bundles and adds missing codes).
	for _, name := range b.Coverage {
		if code := isoFromName(name); code != "" {
			if _, exists := out[code]; !exists {
				out[code] = name
			}
		}
	}
	return out
}

// isoFromFlagImg extracts the ISO-2 code from a Keepgo flag URL such as
// "https://myaccount.keepgo.com/img/flags/3x2/vn.svg" → "VN".
func isoFromFlagImg(img string) string {
	const marker = "/flags/"
	i := strings.Index(img, marker)
	if i < 0 {
		return ""
	}
	rest := img[i+len(marker):]
	// rest is like "3x2/vn.svg" or "vn.svg"
	if slash := strings.LastIndex(rest, "/"); slash >= 0 {
		rest = rest[slash+1:]
	}
	code := strings.TrimSuffix(rest, ".svg")
	if len(code) == 2 {
		return strings.ToUpper(code)
	}
	return ""
}

// isoFromName resolves a Keepgo coverage country name to its ISO-2 code.
func isoFromName(name string) string {
	return countryNameToISO[strings.TrimSpace(name)]
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
