package integration_test

import (
	"sync"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/vault"
)

func TestRotatorStrategies(t *testing.T) {
	rot := rotator.NewRotator()

	futureCooldown := time.Now().Add(time.Hour)
	candidates := []*vault.Credential{
		{ID: "cred-A", ProviderID: "openai", Status: "active", Priority: 10, RequestCount: 100, ErrorCount: 2, HealthState: vault.HealthHealthy},
		{ID: "cred-B", ProviderID: "openai", Status: "active", Priority: 5, RequestCount: 10, ErrorCount: 0, HealthState: vault.HealthHealthy},
		{ID: "cred-C", ProviderID: "openai", Status: "active", Priority: 20, RequestCount: 5, ErrorCount: 0, HealthState: vault.HealthHealthy},
		{ID: "cred-D", ProviderID: "openai", Status: "active", Priority: 1, RequestCount: 2, ErrorCount: 0, HealthState: vault.HealthUnhealthy, CooldownUntil: &futureCooldown},
	}

	t.Run("Priority strategy selects highest priority eligible credential", func(t *testing.T) {
		selected, err := rot.SelectForPool("pool-1", candidates, rotator.StrategyPriority)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if selected.ID != "cred-B" {
			t.Fatalf("Expected cred-B (priority 5), got %s", selected.ID)
		}
	})

	t.Run("Least used strategy selects candidate with lowest request count", func(t *testing.T) {
		selected, err := rot.SelectForPool("pool-2", candidates, rotator.StrategyLeastUsed)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if selected.ID != "cred-C" {
			t.Fatalf("Expected cred-C (5 requests), got %s", selected.ID)
		}
	})

	t.Run("Round robin rotates sequentially between eligible credentials", func(t *testing.T) {
		seen := make(map[string]int)
		for i := 0; i < 6; i++ {
			sel, err := rot.SelectForPool("pool-rr", candidates, rotator.StrategyRoundRobin)
			if err != nil {
				t.Fatalf("RR failed at %d: %v", i, err)
			}
			seen[sel.ID]++
		}
		if seen["cred-A"] == 0 || seen["cred-B"] == 0 || seen["cred-C"] == 0 {
			t.Fatalf("Round robin did not distribute across eligible credentials: %v", seen)
		}
	})

	t.Run("All candidates in cooldown returns no eligible credentials", func(t *testing.T) {
		inCooldown := []*vault.Credential{
			{ID: "c1", ProviderID: "p1", Status: "active", CooldownUntil: &futureCooldown},
			{ID: "c2", ProviderID: "p1", Status: "active", CooldownUntil: &futureCooldown},
		}
		_, err := rot.SelectForPool("pool-exhausted", inCooldown, rotator.StrategyPriority)
		if err == nil {
			t.Fatalf("Expected error when all credentials are in cooldown")
		}
	})
}

func TestCooldownManager(t *testing.T) {
	cd := routing.NewCooldownManager()

	accID := "acc-ratelimited-1"
	if cd.IsCoolingDown(accID) {
		t.Fatalf("Expected account to not be in cooldown initially")
	}

	cd.MarkFailure(accID, 5*time.Second)

	if !cd.IsCoolingDown(accID) {
		t.Fatalf("Expected account to be cooling down after failure")
	}

	cd.Reset(accID)
	if cd.IsCoolingDown(accID) {
		t.Fatalf("Expected account to be active after reset")
	}
}

func TestConcurrentRotatorSafety(t *testing.T) {
	rot := rotator.NewRotator()
	candidates := []*vault.Credential{
		{ID: "c1", ProviderID: "concurrent-p", Status: "active", Priority: 10, RequestCount: 0},
		{ID: "c2", ProviderID: "concurrent-p", Status: "active", Priority: 10, RequestCount: 0},
		{ID: "c3", ProviderID: "concurrent-p", Status: "active", Priority: 10, RequestCount: 0},
	}

	var wg sync.WaitGroup
	workers := 20
	iterations := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, err := rot.SelectForPool("pool-concurrent", candidates, rotator.StrategyRoundRobin)
				if err != nil {
					t.Errorf("Worker %d iteration %d failed: %v", workerID, j, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestProviderFallbackExecution(t *testing.T) {
	cd := routing.NewCooldownManager()
	r := routing.NewRouter(cd)

	r.SetAccount(&routing.Account{ID: "acc-primary", ProviderID: "prov-1", State: "active", Enabled: true})
	r.SetAccount(&routing.Account{ID: "acc-secondary", ProviderID: "prov-2", State: "active", Enabled: true})

	r.SetRoute(&routing.Route{
		Name:     "fallback-model",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "prov-1", ProviderKind: "openai", AccountID: "acc-primary", ModelName: "gpt-4o", Priority: 1, Enabled: true, MaxRetries: 1},
			{ProviderID: "prov-2", ProviderKind: "anthropic", AccountID: "acc-secondary", ModelName: "claude-3-5-sonnet", Priority: 10, Enabled: true, MaxRetries: 1},
		},
	})

	targets, err := r.SelectTargets("fallback-model")
	if err != nil {
		t.Fatalf("SelectTargets failed: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets in fallback chain, got %d", len(targets))
	}

	if targets[0].AccountID != "acc-primary" || targets[1].AccountID != "acc-secondary" {
		t.Fatalf("Fallback order incorrect: [0]=%s, [1]=%s", targets[0].AccountID, targets[1].AccountID)
	}

	cd.MarkFailure("acc-primary", time.Hour)

	targetsAfterCooldown, err := r.SelectTargets("fallback-model")
	if err != nil {
		t.Fatalf("SelectTargets after cooldown failed: %v", err)
	}

	if len(targetsAfterCooldown) != 1 || targetsAfterCooldown[0].AccountID != "acc-secondary" {
		t.Fatalf("Expected only secondary target after primary cooldown, got %v", targetsAfterCooldown)
	}
}
