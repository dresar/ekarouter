package rotator_test

import (
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/vault"
)

func TestRotatorStrategies(t *testing.T) {
	rot := rotator.NewRotator()

	now := time.Now().UTC()
	cdTime := now.Add(1 * time.Hour)

	c1 := &vault.Credential{ID: "c1", Status: "active", Priority: 1, RequestCount: 10, ErrorCount: 1, HealthState: vault.HealthHealthy}
	c2 := &vault.Credential{ID: "c2", Status: "active", Priority: 2, RequestCount: 5, ErrorCount: 0, HealthState: vault.HealthHealthy}
	c3 := &vault.Credential{ID: "c3", Status: "active", Priority: 1, RequestCount: 20, ErrorCount: 5, HealthState: vault.HealthDegraded}
	cCooldown := &vault.Credential{ID: "c4", Status: "active", CooldownUntil: &cdTime, Priority: 1, HealthState: vault.HealthHealthy}
	cDisabled := &vault.Credential{ID: "c5", Status: "disabled", Priority: 1, HealthState: vault.HealthHealthy}

	candidates := []*vault.Credential{c1, c2, c3, cCooldown, cDisabled}

	sel, err := rot.Select(candidates, rotator.StrategyPriority)
	if err != nil || sel.ID != "c1" {
		t.Fatalf("Expected c1 for Priority, got %v, err=%v", sel, err)
	}

	sel, err = rot.Select(candidates, rotator.StrategyLeastUsed)
	if err != nil || sel.ID != "c2" {
		t.Fatalf("Expected c2 for LeastUsed, got %v, err=%v", sel, err)
	}

	sel, err = rot.Select(candidates, rotator.StrategyLowestErrorRate)
	if err != nil || sel.ID != "c2" {
		t.Fatalf("Expected c2 for LowestErrorRate, got %v, err=%v", sel, err)
	}

	selA, _ := rot.Select(candidates, rotator.StrategyRoundRobin)
	selB, _ := rot.Select(candidates, rotator.StrategyRoundRobin)
	if selA == nil || selB == nil {
		t.Fatal("RoundRobin returned nil selection")
	}

	sel, err = rot.Select(candidates, rotator.StrategyHealthBased)
	if err != nil || (sel.ID != "c1" && sel.ID != "c2") {
		t.Fatalf("Expected healthy credential (c1 or c2), got %v", sel)
	}
}

func TestRotatorExhausted(t *testing.T) {
	rot := rotator.NewRotator()
	now := time.Now().UTC()
	cd := now.Add(10 * time.Minute)

	candidates := []*vault.Credential{
		{ID: "c1", Status: "disabled"},
		{ID: "c2", Status: "active", CooldownUntil: &cd},
	}

	sel, err := rot.Select(candidates, rotator.StrategyPriority)
	if err == nil || sel != nil {
		t.Fatalf("Expected error for no eligible credentials, got %v", sel)
	}
}
