package esim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/internal/config"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type MobiMatterProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewMobiMatterProvider(cfg config.ENV) *MobiMatterProvider {
	return &MobiMatterProvider{
		apiKey:  cfg.MobiMatterAPIKey,
		baseURL: cfg.MobiMatterAPIURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *MobiMatterProvider) Name() string { return "mobimatter" }

func (m *MobiMatterProvider) GetDestinations(ctx context.Context) ([]domain.ESIMDestination, error) {
	if !m.isConfigured() {
		return demoDestinations(), nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products?type=esim")
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
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
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

	if !m.isConfigured() {
		return demoPlans(countryCode), nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products?type=esim&countryCode="+countryCode)
	if err != nil {
		return demoPlans(countryCode), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return demoPlans(countryCode), nil
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
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return demoPlans(countryCode), nil
	}

	out := make([]domain.ESIMPlan, 0, len(apiResp))
	for _, p := range apiResp {
		out = append(out, domain.ESIMPlan{
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
		})
	}
	if len(out) == 0 {
		return demoPlans(countryCode), nil
	}

	return out, nil
}

func (m *MobiMatterProvider) OrderESIM(_ context.Context, planID string) (*domain.ESIMOrderResult, error) {
	// Real API integration pending — keep demo fallback for now.
	return demoOrder(planID)
}

func (m *MobiMatterProvider) CheckAvailability(ctx context.Context, planID string) (bool, error) {
	if planID == "" {
		return false, wrapper.Wrap(domain.NewInvalidInput("planId is required"))
	}

	if !m.isConfigured() {
		return true, nil
	}

	resp, err := m.doRequest(ctx, http.MethodGet, "/products/"+planID)
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
	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		return true, nil
	}

	return product.Available, nil
}

func (m *MobiMatterProvider) isConfigured() bool { return m.apiKey != "" }

func (m *MobiMatterProvider) doRequest(ctx context.Context, method string, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, m.baseURL+path, nil)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Key", m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return resp, nil
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
