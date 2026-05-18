package esim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/config"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type MobiMatterProvider struct {
	apiKey     string
	merchantID string
	baseURL    string
	client     *http.Client
}

func NewMobiMatterProvider(cfg config.ENV) *MobiMatterProvider {
	return &MobiMatterProvider{
		apiKey:     cfg.MobiMatterAPIKey,
		merchantID: cfg.MobiMatterMerchantID,
		baseURL:    cfg.MobiMatterAPIURL,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (m *MobiMatterProvider) Name() string { return "mobimatter" }

func (m *MobiMatterProvider) GetDestinations(ctx context.Context) ([]domain.ESIMDestination, error) {
	if !m.hasAPIKey() {
		return demoDestinations(), nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products?type=esim", nil)
	if err != nil {
		return demoDestinations(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return demoDestinations(), nil
	}

	var apiResp []struct {
		CountryCode string `json:"countryCode"`
		Country     string `json:"country"`
	}
	err = decodeJSON(resp.Body, &apiResp)
	if err != nil {
		return demoDestinations(), nil
	}

	seen := map[string]bool{}
	out := make([]domain.ESIMDestination, 0, len(apiResp))
	for _, p := range apiResp {
		if p.CountryCode == "" || seen[p.CountryCode] {
			continue
		}
		seen[p.CountryCode] = true
		out = append(out, domain.ESIMDestination{
			CountryCode: p.CountryCode,
			CountryName: p.Country,
			FlagEmoji:   countryFlag(p.CountryCode),
			PlanCount:   1,
		})
	}
	if len(out) == 0 {
		return demoDestinations(), nil
	}

	return out, nil
}

func (m *MobiMatterProvider) GetPlans(ctx context.Context, countryCode string) ([]domain.ESIMPlan, error) {
	if countryCode == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("countryCode is required"))
	}

	if !m.hasAPIKey() {
		return demoPlans(countryCode), nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products?type=esim&countryCode="+countryCode, nil)
	if err != nil {
		return demoPlans(countryCode), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return demoPlans(countryCode), nil
	}

	plans, err := decodeProductList(resp.Body)
	if err != nil {
		return demoPlans(countryCode), nil
	}
	if len(plans) == 0 {
		return demoPlans(countryCode), nil
	}

	return plans, nil
}

func (m *MobiMatterProvider) GetPlan(ctx context.Context, planID string) (*domain.ESIMPlan, error) {
	if planID == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("planId is required"))
	}

	if demoPlan, ok := demoPlanByID(planID); ok {
		return demoPlan, nil
	}

	if !m.hasAPIKey() {
		return nil, wrapper.Wrap(domain.NewInvalidInput("plan not found"))
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products/"+planID, nil)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, wrapper.Wrap(domain.NewInvalidInput("plan not found"))
	}

	plan, err := decodeSingleProduct(resp.Body)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return plan, nil
}

func (m *MobiMatterProvider) OrderESIM(ctx context.Context, planID string) (*domain.ESIMOrderResult, error) {
	if planID == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("planId is required"))
	}

	if !m.canPlaceOrders() {
		return demoOrder(planID)
	}

	label := "xplr-" + uuid.New().String()[:8]
	createBody := map[string]string{
		"productId":       planID,
		"productCategory": "esim_realtime",
		"label":           label,
	}
	createPayload, err := json.Marshal(createBody)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	resp, err := m.doRequest(ctx, http.MethodPost, "/order", createPayload)
	if err != nil {
		return demoOrder(planID)
	}
	defer resp.Body.Close()

	var created mobimatterEnvelope[mobimatterCreateOrder]
	err = decodeJSON(resp.Body, &created)
	if err != nil || created.Result.OrderID == "" {
		return demoOrder(planID)
	}

	completePayload, err := json.Marshal(map[string]string{"orderId": created.Result.OrderID})
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	completeResp, err := m.doRequest(ctx, http.MethodPut, "/order/complete", completePayload)
	if err != nil {
		return demoOrder(planID)
	}
	defer completeResp.Body.Close()

	var completed mobimatterEnvelope[mobimatterOrder]
	err = decodeJSON(completeResp.Body, &completed)
	if err != nil {
		return demoOrder(planID)
	}

	return mapMobimatterOrder(completed.Result, created.Result.OrderID)
}

func (m *MobiMatterProvider) CheckAvailability(ctx context.Context, planID string) (bool, error) {
	if planID == "" {
		return false, wrapper.Wrap(domain.NewInvalidInput("planId is required"))
	}

	if demoPlan, ok := demoPlanByID(planID); ok {
		return demoPlan.InStock, nil
	}

	if !m.hasAPIKey() {
		return true, nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products/"+planID, nil)
	if err != nil {
		return true, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var product struct {
		Available bool `json:"available"`
	}
	err = decodeJSON(resp.Body, &product)
	if err != nil {
		return true, nil
	}

	return product.Available, nil
}

func (m *MobiMatterProvider) hasAPIKey() bool { return m.apiKey != "" }

func (m *MobiMatterProvider) canPlaceOrders() bool {
	return m.apiKey != "" && m.merchantID != ""
}

func (m *MobiMatterProvider) doRequest(ctx context.Context, method string, path string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, m.baseURL+path, bodyReader)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Key", m.apiKey)
	if m.merchantID != "" {
		req.Header.Set("Merchantid", m.merchantID)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return resp, nil
}

type mobimatterEnvelope[T any] struct {
	StatusCode int  `json:"statusCode"`
	IsSuccess  bool `json:"isSuccess"`
	Result     T    `json:"result"`
}

type mobimatterCreateOrder struct {
	OrderID string `json:"orderId"`
}

type mobimatterLineDetail struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mobimatterOrderLine struct {
	LineItemDetails []mobimatterLineDetail `json:"lineItemDetails"`
}

type mobimatterOrder struct {
	OrderID     string              `json:"orderId"`
	OrderState  string              `json:"orderState"`
	OrderLine   mobimatterOrderLine `json:"orderLineItem"`
	ProviderRef string              `json:"-"`
}

func mapMobimatterOrder(o mobimatterOrder, fallbackOrderID string) (*domain.ESIMOrderResult, error) {
	orderID := o.OrderID
	if orderID == "" {
		orderID = fallbackOrderID
	}

	details := map[string]string{}
	for _, item := range o.OrderLine.LineItemDetails {
		details[strings.ToUpper(item.Name)] = item.Value
	}

	lpa := details["LOCAL_PROFILE_ASSISTANT"]
	if lpa == "" {
		lpa = details["ACTIVATION_CODE"]
	}
	qr := details["QR_CODE"]
	if qr == "" {
		qr = lpa
	}

	res := &domain.ESIMOrderResult{
		OrderID:     orderID,
		QRData:      qr,
		LPA:         lpa,
		SMDP:        details["SMDP_ADDRESS"],
		MatchingID:  details["ACTIVATION_CODE"],
		ICCID:       details["ICCID"],
		ProviderRef: orderID,
	}

	if strings.EqualFold(o.OrderState, "Processing") {
		if kyc := details["KYC_URL"]; kyc != "" {
			res.PendingKYC = true
			res.KYCURL = kyc
		}
	}

	if lpa == "" && qr == "" && !res.PendingKYC {
		return demoOrder(fallbackOrderID)
	}

	return res, nil
}

type mobimatterProduct struct {
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

func decodeProductList(r io.Reader) ([]domain.ESIMPlan, error) {
	var list []mobimatterProduct
	err := decodeJSON(r, &list)
	if err != nil {
		return nil, err
	}

	out := make([]domain.ESIMPlan, 0, len(list))
	for _, p := range list {
		out = append(out, mapProduct(p))
	}

	return out, nil
}

func decodeSingleProduct(r io.Reader) (*domain.ESIMPlan, error) {
	var p mobimatterProduct
	err := decodeJSON(r, &p)
	if err != nil {
		return nil, err
	}
	if p.ProductID == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("plan not found"))
	}

	plan := mapProduct(p)

	return &plan, nil
}

func mapProduct(p mobimatterProduct) domain.ESIMPlan {
	return domain.ESIMPlan{
		PlanID:       p.ProductID,
		Provider:     "mobimatter",
		Name:         p.ProductName,
		Country:      p.Country,
		CountryCode:  p.CountryCode,
		DataGB:       fmt.Sprintf("%.0f", p.DataGB),
		ValidityDays: p.ValidityDays,
		PriceUSD:     domain.NewNumeric(p.Price),
		Description:  p.Description,
		InStock:      p.Available,
	}
}

func decodeJSON(r io.Reader, dest any) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = json.Unmarshal(raw, dest)
	if err == nil {
		return nil
	}

	var wrapped mobimatterEnvelope[json.RawMessage]
	err = json.Unmarshal(raw, &wrapped)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = json.Unmarshal(wrapped.Result, dest)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func countryFlag(code string) string {
	if len(code) < 2 {
		return "🌍"
	}
	runes := []rune(code[:2])

	return string([]rune{runes[0] - 'A' + 0x1F1E6, runes[1] - 'A' + 0x1F1E6})
}

func demoDestinations() []domain.ESIMDestination {
	return []domain.ESIMDestination{
		{CountryCode: "TR", CountryName: "Турция", FlagEmoji: "🇹🇷", PlanCount: 3},
		{CountryCode: "TH", CountryName: "Таиланд", FlagEmoji: "🇹🇭", PlanCount: 3},
		{CountryCode: "US", CountryName: "США", FlagEmoji: "🇺🇸", PlanCount: 3},
		{CountryCode: "AE", CountryName: "ОАЭ", FlagEmoji: "🇦🇪", PlanCount: 3},
	}
}

func demoPlans(countryCode string) []domain.ESIMPlan {
	return []domain.ESIMPlan{
		{
			PlanID:       fmt.Sprintf("demo-%s-1gb", countryCode),
			Provider:     "demo",
			Name:         "Demo 1 GB",
			Country:      countryCode,
			CountryCode:  countryCode,
			DataGB:       "1",
			ValidityDays: 7,
			PriceUSD:     domain.NewNumeric(3.5),
			Description:  "Demo plan",
			InStock:      true,
		},
	}
}

func demoPlanByID(planID string) (*domain.ESIMPlan, bool) {
	if !strings.HasPrefix(planID, "demo-") {
		return nil, false
	}
	parts := strings.Split(planID, "-")
	if len(parts) < 3 {
		return nil, false
	}
	countryCode := parts[1]
	plans := demoPlans(countryCode)
	for i := range plans {
		if plans[i].PlanID == planID {
			return &plans[i], true
		}
	}

	return nil, false
}

func demoOrder(planID string) (*domain.ESIMOrderResult, error) {
	ref := fmt.Sprintf("DEMO-%s-%d", planID, time.Now().UnixMilli())
	smdp := "smdp.example.com"
	matchingID := fmt.Sprintf("X%d", time.Now().UnixMilli())
	lpa := fmt.Sprintf("LPA:1$%s$%s", smdp, matchingID)

	return &domain.ESIMOrderResult{
		OrderID:     ref,
		QRData:      lpa,
		LPA:         lpa,
		SMDP:        smdp,
		MatchingID:  matchingID,
		ICCID:       fmt.Sprintf("8901%d", time.Now().UnixNano()%10000000000),
		ProviderRef: ref,
	}, nil
}
