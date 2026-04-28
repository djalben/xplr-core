package providers

import (
	"testing"
)

// ── Test: Celitech provider implements ESIMProvider ──
func TestCelitechProvider_ImplementsInterface(t *testing.T) {
	var _ ESIMProvider = (*CelitechProvider)(nil)
}

// ── Test: New provider without env vars falls back to demo ──
func TestCelitechProvider_NotConfigured(t *testing.T) {
	p := &CelitechProvider{
		pkgCache: make(map[string]*celitechPackage),
	}

	if p.isConfigured() {
		t.Error("Provider should not be configured without credentials")
	}
	if p.Name() != "celitech" {
		t.Errorf("Expected name 'celitech', got %q", p.Name())
	}

	// GetDestinations should return demo data
	dests, err := p.GetDestinations()
	if err != nil {
		t.Fatalf("GetDestinations error: %v", err)
	}
	if len(dests) == 0 {
		t.Error("GetDestinations returned empty (expected demo data)")
	}

	// GetPlans should return demo data
	plans, err := p.GetPlans("TR")
	if err != nil {
		t.Fatalf("GetPlans error: %v", err)
	}
	if len(plans) == 0 {
		t.Error("GetPlans returned empty (expected demo data)")
	}

	// OrderESIM should return demo order
	result, err := p.OrderESIM("demo-plan")
	if err != nil {
		t.Fatalf("OrderESIM error: %v", err)
	}
	if result == nil {
		t.Fatal("OrderESIM returned nil result")
	}
	if result.QRData == "" {
		t.Error("Demo order should have QR data")
	}

	// CheckAvailability — unconfigured always returns true
	avail, _ := p.CheckAvailability("any-plan")
	if !avail {
		t.Error("Unconfigured provider should return available=true")
	}
}

// ── Test: LPA string parsing ──
func TestSplitLPA(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			"LPA:1$CELITECH.IDEMIA.IO$AAAAA-BBBBB-CCCCC-DDDDD",
			[]string{"CELITECH.IDEMIA.IO", "AAAAA-BBBBB-CCCCC-DDDDD"},
		},
		{
			"LPA:1$smdp.example.com$X123456",
			[]string{"smdp.example.com", "X123456"},
		},
		{
			"plain-string-no-lpa",
			[]string{"plain-string-no-lpa"},
		},
		{
			"",
			nil,
		},
	}

	for _, tt := range tests {
		parts := splitLPA(tt.input)
		if len(parts) != len(tt.expected) {
			t.Errorf("splitLPA(%q) = %v (len=%d), want len=%d", tt.input, parts, len(parts), len(tt.expected))
			continue
		}
		for i := range parts {
			if parts[i] != tt.expected[i] {
				t.Errorf("splitLPA(%q)[%d] = %q, want %q", tt.input, i, parts[i], tt.expected[i])
			}
		}
	}
}

// ── Test: Package cache for ordering ──
func TestCelitechProvider_PackageCache(t *testing.T) {
	p := &CelitechProvider{
		pkgCache: make(map[string]*celitechPackage),
	}

	// Simulate caching a package
	p.pkgCache["pkg-123"] = &celitechPackage{
		ID:            "pkg-123",
		Destination:   "FRA",
		DataLimitInGB: 5,
		Duration:      30,
		Price:         8.50,
	}

	// CheckAvailability should find cached package (even unconfigured)
	// Note: unconfigured always returns true, but let's test the cache path
	p.clientID = "test"
	p.clientSecret = "test"
	avail, _ := p.CheckAvailability("pkg-123")
	if !avail {
		t.Error("Cached package should be available")
	}
	avail, _ = p.CheckAvailability("nonexistent")
	if avail {
		t.Error("Non-cached package should not be available")
	}
}

// ── Test: Singleton prefers Celitech when configured ──
func TestGetESIMProvider_PrefersConfigured(t *testing.T) {
	// Without env vars set, should get either MobiMatter or demo fallback
	// This test verifies the function doesn't panic
	// Note: singleton is already initialized in the process, so we just test NewCelitechProvider
	p := NewCelitechProvider()
	if p == nil {
		t.Fatal("NewCelitechProvider returned nil")
	}
	if p.apiURL == "" {
		t.Error("Default apiURL should not be empty")
	}
	if p.authURL == "" {
		t.Error("Default authURL should not be empty")
	}
}

// ── Test: Demo destinations have valid data ──
func TestDemoDestinations_ValidData(t *testing.T) {
	dests := getDemoDestinations()
	if len(dests) == 0 {
		t.Fatal("Demo destinations should not be empty")
	}
	for _, d := range dests {
		if d.CountryCode == "" {
			t.Error("Demo destination has empty country code")
		}
		if d.CountryName == "" {
			t.Error("Demo destination has empty country name")
		}
	}
}

// ── Test: Demo plans have valid structure ──
func TestDemoPlans_ValidStructure(t *testing.T) {
	plans := getDemoPlans("TR")
	if len(plans) == 0 {
		t.Fatal("Demo plans should not be empty")
	}
	for _, p := range plans {
		if p.PlanID == "" {
			t.Error("Demo plan has empty plan_id")
		}
		if p.PriceUSD <= 0 {
			t.Errorf("Demo plan %s has invalid price: %f", p.PlanID, p.PriceUSD)
		}
		if p.ValidityDays <= 0 {
			t.Errorf("Demo plan %s has invalid validity: %d", p.PlanID, p.ValidityDays)
		}
	}
}
