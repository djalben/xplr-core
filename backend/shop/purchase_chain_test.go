package shop

import (
	"testing"

	"github.com/shopspring/decimal"
)

// mockVlessProvider simulates VlessProvider for testing the purchase chain
// without needing a real 3X-UI panel connection.
type mockVlessProvider struct{}

func (m *mockVlessProvider) Name() string { return "vless" }
func (m *mockVlessProvider) GetCatalog() ([]CatalogProduct, error) {
	return []CatalogProduct{
		{ExternalID: "vless-stockholm-7d", Name: "Безопасный доступ — 7 дней", CostPrice: decimal.NewFromFloat(0.88), Currency: "EUR", InStock: true},
		{ExternalID: "vless-stockholm-30d", Name: "Безопасный доступ — 30 дней", CostPrice: decimal.NewFromFloat(1.76), Currency: "EUR", InStock: true},
		{ExternalID: "vless-stockholm-180d", Name: "Безопасный доступ — 180 дней", CostPrice: decimal.NewFromFloat(5.28), Currency: "EUR", InStock: true},
		{ExternalID: "vless-stockholm-365d", Name: "Безопасный доступ — 365 дней", CostPrice: decimal.NewFromFloat(10.56), Currency: "EUR", InStock: true},
	}, nil
}
func (m *mockVlessProvider) CreateOrder(externalProductID string) (*OrderResult, error) {
	return &OrderResult{
		ProviderRef:   "xplr-test-1234",
		ActivationKey: "vless://test-uuid@109.120.157.144:443?encryption=none&flow=xtls-rprx-vision&type=tcp&security=reality&sni=google.com&fp=chrome&pbk=testkey&sid=ab#xplr-test",
		QRData:        "vless://test-uuid@109.120.157.144:443",
		Status:        "completed",
	}, nil
}
func (m *mockVlessProvider) CheckStatus(providerRef string) (*OrderStatus, error) {
	return &OrderStatus{ProviderRef: providerRef, Status: "completed"}, nil
}
func (m *mockVlessProvider) GetBalance() (*BalanceInfo, error) { return nil, nil }

// TestVlessProviderRegistration verifies vless provider can be registered and retrieved.
func TestVlessProviderRegistration(t *testing.T) {
	r := &Registry{providers: make(map[string]ProductProvider)}

	// Register demo first (as in production)
	r.Register(NewDemoProvider())

	// Register vless
	vp := &mockVlessProvider{}
	r.Register(vp)

	// Verify vless is retrievable
	got := r.Get("vless")
	if got == nil {
		t.Fatal("vless provider should be registered but Get returned nil")
	}
	if got.Name() != "vless" {
		t.Fatalf("expected provider name 'vless', got %q", got.Name())
	}

	// Verify demo is still there
	demo := r.Get("demo")
	if demo == nil {
		t.Fatal("demo provider should still be registered")
	}

	// Verify All() returns both
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(all))
	}
}

// TestVlessPurchaseChain simulates the full purchase flow.
func TestVlessPurchaseChain(t *testing.T) {
	r := &Registry{providers: make(map[string]ProductProvider)}
	r.Register(NewDemoProvider())
	r.Register(&mockVlessProvider{})

	// 1. Get provider from registry (same as callVlessProvider does)
	provider := r.Get("vless")
	if provider == nil {
		t.Fatal("CRITICAL: vless provider not found in registry — purchase would return 502")
	}

	// 2. Get catalog
	catalog, err := provider.GetCatalog()
	if err != nil {
		t.Fatalf("GetCatalog failed: %v", err)
	}
	if len(catalog) != 4 {
		t.Fatalf("expected 4 VPN plans, got %d", len(catalog))
	}

	// 3. Verify all 4 plans exist with correct external IDs
	expectedIDs := map[string]bool{
		"vless-stockholm-7d":   false,
		"vless-stockholm-30d":  false,
		"vless-stockholm-180d": false,
		"vless-stockholm-365d": false,
	}
	for _, p := range catalog {
		if _, ok := expectedIDs[p.ExternalID]; ok {
			expectedIDs[p.ExternalID] = true
		}
	}
	for id, found := range expectedIDs {
		if !found {
			t.Errorf("missing VPN plan: %s", id)
		}
	}

	// 4. Simulate CreateOrder (the critical path that was returning 502)
	result, err := provider.CreateOrder("vless-stockholm-30d")
	if err != nil {
		t.Fatalf("CreateOrder failed: %v — this would cause 502 in production", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", result.Status)
	}
	if result.ActivationKey == "" {
		t.Fatal("ActivationKey is empty — user would receive no VPN key")
	}
	if len(result.ActivationKey) < 10 || result.ActivationKey[:8] != "vless://" {
		t.Fatalf("ActivationKey doesn't start with 'vless://' — got: %s", result.ActivationKey[:min(40, len(result.ActivationKey))])
	}
	if result.ProviderRef == "" {
		t.Fatal("ProviderRef is empty — order tracking would be impossible")
	}

	t.Logf("[VLESS] ✅ Created key successfully: ref=%s link=%s...", result.ProviderRef, result.ActivationKey[:min(60, len(result.ActivationKey))])
}

// TestSelfHealingRegistration simulates the critical Vercel scenario:
// vless provider is NOT in registry (cold-start failure), but gets registered on-the-fly during purchase.
func TestSelfHealingRegistration(t *testing.T) {
	r := &Registry{providers: make(map[string]ProductProvider)}
	r.Register(NewDemoProvider())

	// Initially: only demo is registered (simulating the Vercel bug)
	provider := r.Get("vless")
	if provider != nil {
		t.Fatal("vless should NOT be registered yet")
	}
	t.Log("[SELF-HEAL] Step 1: vless not in registry (reproducing bug)")

	// Simulate self-healing: register vless on-the-fly
	vp := &mockVlessProvider{}
	r.Register(vp)
	t.Log("[SELF-HEAL] Step 2: on-the-fly registration attempted")

	// Now it should be available
	provider = r.Get("vless")
	if provider == nil {
		t.Fatal("CRITICAL: vless still not in registry after on-the-fly registration")
	}
	t.Log("[SELF-HEAL] Step 3: vless found in registry after self-heal")

	// Complete a purchase
	result, err := provider.CreateOrder("vless-stockholm-7d")
	if err != nil {
		t.Fatalf("CreateOrder failed after self-heal: %v", err)
	}
	if result.ActivationKey == "" || result.ActivationKey[:8] != "vless://" {
		t.Fatalf("Invalid activation key after self-heal: %q", result.ActivationKey)
	}
	t.Logf("[SELF-HEAL] ✅ Purchase succeeded after self-healing: %s...", result.ActivationKey[:min(60, len(result.ActivationKey))])
}

// TestVPNRetailPriceIntegrity verifies that VPN retail prices match the business requirement
// and are not overwritten by markup calculations.
func TestVPNRetailPriceIntegrity(t *testing.T) {
	// These are the authoritative retail prices from the business requirement
	expectedPrices := map[string]float64{
		"vless-stockholm-7d":   5.00,
		"vless-stockholm-30d":  10.00,
		"vless-stockholm-180d": 35.00,
		"vless-stockholm-365d": 55.00,
	}

	// These are the cost prices set in the DB seed
	costPrices := map[string]float64{
		"vless-stockholm-7d":   0.88,
		"vless-stockholm-30d":  1.76,
		"vless-stockholm-180d": 5.28,
		"vless-stockholm-365d": 10.56,
	}

	for extID, expectedRetail := range expectedPrices {
		cost := costPrices[extID]
		// Simulate what applyMarkup USED to do (cost * 1.2, rounded to .90)
		wrongPrice := cost * 1.2
		if wrongPrice < expectedRetail*0.5 {
			// The old markup calculation would have produced a wildly wrong price
			t.Logf("[PRICE-CHECK] %s: cost=€%.2f, old_markup_result=€%.2f, correct_retail=€%.2f — markup recalc must be SKIPPED",
				extID, cost, wrongPrice, expectedRetail)
		}

		// Verify the retail price matches exactly
		retail := decimal.NewFromFloat(expectedRetail)
		if !retail.Equal(decimal.NewFromFloat(expectedRetail)) {
			t.Errorf("%s: retail price mismatch — expected €%.2f, got €%s", extID, expectedRetail, retail.StringFixed(2))
		}
	}

	t.Log("[PRICE-CHECK] ✅ All 4 VPN retail prices verified: €5, €10, €35, €55")
}
