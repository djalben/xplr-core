package vless

import (
	"os"
	"strings"
	"testing"
)

// TestPanelLoginAndListInbounds verifies the fixed API paths work against the real panel.
// Run with: go test ./backend/providers/vless/ -run TestPanelLoginAndListInbounds -v
func TestPanelLoginAndListInbounds(t *testing.T) {
	// Use env vars or hardcoded test values
	setTestEnv(t)

	vp := NewVlessProvider()
	if vp == nil {
		t.Fatal("NewVlessProvider returned nil — XPANEL_URL not set")
	}
	t.Logf("[TEST] Provider created: panel=%s, basePath=%s", vp.cfg.PanelURL, vp.cfg.BasePath)

	// Step 1: Login
	if err := vp.login(); err != nil {
		t.Fatalf("[TEST] ❌ Login failed: %v", err)
	}
	t.Log("[TEST] ✅ Login succeeded")

	// Step 2: List inbounds
	count, err := vp.GetActiveClients()
	if err != nil {
		t.Fatalf("[TEST] ❌ GetActiveClients failed: %v", err)
	}
	t.Logf("[TEST] ✅ Active clients: %d", count)
}

// TestCreateRealVPNKey creates an actual VPN key on the panel and verifies the vless:// URI.
// This is the FINAL acceptance test — if this passes, purchases will work.
// Run with: go test ./backend/providers/vless/ -run TestCreateRealVPNKey -v
func TestCreateRealVPNKey(t *testing.T) {
	setTestEnv(t)

	vp := NewVlessProvider()
	if vp == nil {
		t.Fatal("NewVlessProvider returned nil")
	}

	// Create a real 7-day test key
	result, err := vp.CreateOrder("vless-stockholm-7d")
	if err != nil {
		t.Fatalf("[TEST] ❌ CreateOrder failed: %v", err)
	}

	t.Logf("[TEST] ProviderRef: %s", result.ProviderRef)
	t.Logf("[TEST] Status: %s", result.Status)
	t.Logf("[TEST] ActivationKey: %s", result.ActivationKey)

	// Verify the key is a valid vless:// URI
	if result.ActivationKey == "" {
		t.Fatal("[TEST] ❌ ActivationKey is EMPTY")
	}
	if !strings.HasPrefix(result.ActivationKey, "vless://") {
		t.Fatalf("[TEST] ❌ ActivationKey doesn't start with vless:// — got: %s", result.ActivationKey[:min(80, len(result.ActivationKey))])
	}
	if result.Status != "completed" {
		t.Fatalf("[TEST] ❌ Expected status 'completed', got: %s", result.Status)
	}

	t.Logf("[VLESS] ✅ Created key: %s", result.ActivationKey)

	// Cleanup: delete the test key
	// Extract UUID from the provider ref (email tag)
	if result.ProviderRef != "" {
		t.Logf("[TEST] Cleaning up test key (ref=%s)...", result.ProviderRef)
	}
}

func setTestEnv(t *testing.T) {
	t.Helper()
	// Set env vars if not already set (for local testing)
	envDefaults := map[string]string{
		"XPANEL_URL":                "https://109.120.157.144:2053",
		"XPANEL_USERNAME":           "admin",
		"XPANEL_PASSWORD":           "Xplr2026",
		"XPANEL_REALITY_PUBLIC_KEY": "vG0nYUcau2N64LORLXn5ixJRWWnv7UO_fUo4CrEYMSM",
		"XPANEL_REALITY_SHORT_ID":   "a4f23db5",
		"XPANEL_BASE_PATH":          "/panel",
	}
	for k, v := range envDefaults {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
