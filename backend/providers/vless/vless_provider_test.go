package vless

import (
	"os"
	"strings"
	"testing"
)

// TestVlessProviderIntegration runs against the live 3X-UI panel.
// Requires XPANEL_URL and other env vars to be set.
// Skip if not configured (CI-safe).
func TestVlessProviderIntegration(t *testing.T) {
	if os.Getenv("XPANEL_URL") == "" {
		t.Skip("XPANEL_URL not set — skipping live integration test")
	}

	p := NewVlessProvider()
	if p == nil {
		t.Fatal("NewVlessProvider returned nil despite XPANEL_URL being set")
	}

	// 1. Test Name
	if p.Name() != "vless" {
		t.Errorf("Name() = %q, want %q", p.Name(), "vless")
	}

	// 2. Test GetCatalog
	catalog, err := p.GetCatalog()
	if err != nil {
		t.Fatalf("GetCatalog() error: %v", err)
	}
	if len(catalog) == 0 {
		t.Fatal("GetCatalog() returned empty catalog")
	}
	t.Logf("GetCatalog: %d products", len(catalog))
	for _, cp := range catalog {
		t.Logf("  - %s (%s)", cp.Name, cp.ExternalID)
	}

	// 3. Test CreateOrder (creates a real client on the panel)
	result, err := p.CreateOrder("vless-stockholm-30d")
	if err != nil {
		t.Fatalf("CreateOrder() error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("CreateOrder status = %q, want %q", result.Status, "completed")
	}
	if result.ActivationKey == "" {
		t.Fatal("CreateOrder returned empty ActivationKey (vless:// link)")
	}
	if !strings.HasPrefix(result.ActivationKey, "vless://") {
		t.Errorf("ActivationKey doesn't start with vless:// : %s", result.ActivationKey[:50])
	}
	if result.ProviderRef == "" {
		t.Fatal("CreateOrder returned empty ProviderRef (email tag)")
	}
	t.Logf("CreateOrder OK: ref=%s link=%s...", result.ProviderRef, result.ActivationKey[:60])

	// 4. Test CheckStatus on the just-created client
	status, err := p.CheckStatus(result.ProviderRef)
	if err != nil {
		t.Fatalf("CheckStatus() error: %v", err)
	}
	if status.Status != "completed" {
		t.Errorf("CheckStatus status = %q, want %q", status.Status, "completed")
	}
	t.Logf("CheckStatus OK: ref=%s status=%s", status.ProviderRef, status.Status)

	// 5. Test GetBalance (should return nil for self-hosted)
	balance, err := p.GetBalance()
	if err != nil {
		t.Fatalf("GetBalance() error: %v", err)
	}
	if balance != nil {
		t.Errorf("GetBalance() should return nil for self-hosted, got %+v", balance)
	}

	// 6. Cleanup: delete the test client
	// Extract UUID from the vless:// link
	link := result.ActivationKey
	uuidStart := strings.Index(link, "vless://") + 8
	uuidEnd := strings.Index(link[uuidStart:], "@") + uuidStart
	clientUUID := link[uuidStart:uuidEnd]

	if err := p.DeleteClient(clientUUID); err != nil {
		t.Logf("⚠️ Cleanup: failed to delete test client %s: %v", clientUUID, err)
	} else {
		t.Logf("Cleanup: deleted test client %s", clientUUID)
	}
}
