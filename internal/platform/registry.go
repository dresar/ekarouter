package platform

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderAdapter
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ProviderAdapter),
	}
}

func (r *Registry) Register(adapter ProviderAdapter) error {
	if adapter == nil {
		return errors.New("nil provider adapter")
	}
	meta := adapter.Metadata()
	if strings.TrimSpace(meta.ID) == "" {
		return errors.New("provider id is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToLower(meta.ID)] = adapter
	return nil
}

func (r *Registry) Get(id string) (ProviderAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.providers[strings.ToLower(id)]
	return adapter, ok
}

func (r *Registry) List() []ProviderMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]ProviderMetadata, 0, len(r.providers))
	for _, a := range r.providers {
		res = append(res, a.Metadata())
	}
	return res
}

func (r *Registry) ListByCategory(cat Category) []ProviderMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []ProviderMetadata
	for _, a := range r.providers {
		m := a.Metadata()
		if m.Category == cat {
			res = append(res, m)
		}
	}
	return res
}

func (r *Registry) Search(query string) []ProviderMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	var res []ProviderMetadata
	for _, a := range r.providers {
		m := a.Metadata()
		if q == "" ||
			strings.Contains(strings.ToLower(m.ID), q) ||
			strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Description), q) ||
			strings.Contains(string(m.Category), q) {
			res = append(res, m)
		}
	}
	return res
}

func (r *Registry) CheckHealth(ctx context.Context, id string, secret string) (HealthStatus, error) {
	a, ok := r.Get(id)
	if !ok {
		return HealthStatus{}, fmt.Errorf("provider %q not found", id)
	}
	return a.HealthCheck(ctx, secret)
}

func (r *Registry) ValidateCredential(ctx context.Context, id string, secret string) (bool, string, error) {
	a, ok := r.Get(id)
	if !ok {
		return false, "", fmt.Errorf("provider %q not found", id)
	}
	return a.ValidateCredential(ctx, secret)
}
