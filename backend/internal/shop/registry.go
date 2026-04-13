package shop

import (
	"fmt"
	"log"
	"sync"
)

// Registry holds all registered ProductProviders keyed by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProductProvider
}

var (
	globalRegistry     *Registry
	globalRegistryOnce sync.Once
)

// GetRegistry returns the singleton provider registry.
func GetRegistry() *Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = &Registry{
			providers: make(map[string]ProductProvider),
		}
		// Always register demo provider as fallback
		globalRegistry.Register(NewDemoProvider())
		log.Println("[SHOP-REGISTRY] Initialized with demo provider")
	})
	return globalRegistry
}

// Register adds a provider to the registry.
func (r *Registry) Register(p ProductProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
	log.Printf("[SHOP-REGISTRY] Registered provider: %s", p.Name())
}

// Get returns a provider by name. Returns nil if not found.
func (r *Registry) Get(name string) ProductProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// MustGet returns a provider by name or falls back to "demo".
func (r *Registry) MustGet(name string) (ProductProvider, error) {
	p := r.Get(name)
	if p != nil {
		return p, nil
	}
	p = r.Get("demo")
	if p != nil {
		log.Printf("[SHOP-REGISTRY] ⚠️ Provider %q not found, falling back to demo", name)
		return p, nil
	}
	return nil, fmt.Errorf("no provider registered for %q and no demo fallback", name)
}

// All returns all registered providers.
func (r *Registry) All() []ProductProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ProductProvider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}
