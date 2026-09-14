package credpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/providers"
)

type HealthChecker struct {
	store        *Store
	registry     *providers.Registry
	credResolver func(ctx context.Context, accountID string) (*providers.Credentials, error)
	mu           sync.Mutex
	running      map[string]bool
	maxConcurrent int
	sem          chan struct{}
}

func NewHealthChecker(
	store *Store,
	registry *providers.Registry,
	credResolver func(ctx context.Context, accountID string) (*providers.Credentials, error),
	maxConcurrent int,
) *HealthChecker {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	return &HealthChecker{
		store:         store,
		registry:      registry,
		credResolver:  credResolver,
		running:       make(map[string]bool),
		maxConcurrent: maxConcurrent,
		sem:           make(chan struct{}, maxConcurrent),
	}
}

func (hc *HealthChecker) CheckMember(ctx context.Context, member *Member, poolID string) *HealthResult {
	hc.mu.Lock()
	if hc.running[member.ID] {
		hc.mu.Unlock()
		return &HealthResult{
			MemberID:  member.ID,
			Status:    HealthUnknown,
			Message:   "check already in progress",
			CheckedAt: time.Now(),
		}
	}
	hc.running[member.ID] = true
	hc.mu.Unlock()

	defer func() {
		hc.mu.Lock()
		delete(hc.running, member.ID)
		hc.mu.Unlock()
	}()

	select {
	case hc.sem <- struct{}{}:
		defer func() { <-hc.sem }()
	case <-ctx.Done():
		return &HealthResult{
			MemberID:  member.ID,
			Status:    HealthUnknown,
			Message:   "context cancelled before check started",
			CheckedAt: time.Now(),
		}
	}

	creds, err := hc.credResolver(ctx, member.AccountID)
	if err != nil {
		result := &HealthResult{
			MemberID:  member.ID,
			Status:    HealthInvalidConfig,
			Message:   "credential resolution failed",
			CheckedAt: time.Now(),
		}
		_ = hc.store.RecordHealthCheck(ctx, result, poolID)
		return result
	}

	start := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	providerKind := ""
	pool, err := hc.store.GetPool(ctx, poolID)
	if err == nil && pool.ProviderID != "" {
		providerKind = pool.ProviderID
	}

	var status HealthStatus
	var message string

	if providerKind != "" {
		adapter, err := hc.registry.Get(providerKind)
		if err == nil {
			_, modelsErr := adapter.Models(checkCtx, creds)
			if modelsErr == nil {
				status = HealthValid
				message = "models endpoint responded successfully"
			} else {
				status, message = classifyHealthError(modelsErr)
			}
		} else {
			status = HealthUnsupportedValidation
			message = fmt.Sprintf("provider kind %s not in registry", providerKind)
		}
	} else {
		status = HealthUnsupportedValidation
		message = "no provider kind configured for validation"
	}

	latency := time.Since(start)
	result := &HealthResult{
		MemberID:  member.ID,
		Status:    status,
		LatencyMs: int(latency.Milliseconds()),
		Message:   message,
		CheckedAt: time.Now(),
	}

	_ = hc.store.RecordHealthCheck(ctx, result, poolID)
	return result
}

func (hc *HealthChecker) CheckPool(ctx context.Context, poolID string) []HealthResult {
	members, err := hc.store.ListMembers(ctx, poolID)
	if err != nil {
		return nil
	}

	var results []HealthResult
	for i := range members {
		select {
		case <-ctx.Done():
			return results
		default:
		}
		r := hc.CheckMember(ctx, &members[i], poolID)
		results = append(results, *r)
	}
	return results
}

func classifyHealthError(err error) (HealthStatus, string) {
	action := ClassifyForRotation(err)

	switch action.Reason {
	case "authentication_failed":
		return HealthRevoked, "credential authentication failed"
	case "rate_limited":
		return HealthRateLimited, "provider rate limited"
	case "quota_exhausted":
		return HealthRateLimited, "quota exhausted"
	case "upstream_server_error", "upstream_network_error":
		return HealthProviderUnavailable, "provider temporarily unavailable"
	case "upstream_timeout":
		return HealthProviderUnavailable, "provider timeout"
	case "invalid_request":
		return HealthInvalidConfig, "invalid configuration"
	default:
		return HealthUnknown, SanitizeError(err)
	}
}
