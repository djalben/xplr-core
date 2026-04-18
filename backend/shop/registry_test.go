package shop

import (
	"testing"
)

// TestRegistryRegisterAndGet verifies that providers can be registered and retrieved.
func TestRegistryRegisterAndGet(t *testing.T) {
	r := &Registry{providers: make(map[string]ProductProvider)}

	// Initially empty
	if p := r.Get("vless"); p != nil {
		t.Fatal("expected nil for unregistered provider")
	}

	// Register demo
	r.Register(NewDemoProvider())
	if p := r.Get("demo"); p == nil {
		t.Fatal("demo provider should be registered")
	}

	// MustGet falls back to demo
	p, err := r.MustGet("vless")
	if err != nil {
		t.Fatalf("MustGet should fall back to demo, got error: %v", err)
	}
	if p.Name() != "demo" {
		t.Fatalf("expected demo fallback, got %s", p.Name())
	}

	// All returns registered providers
	all := r.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(all))
	}
}
