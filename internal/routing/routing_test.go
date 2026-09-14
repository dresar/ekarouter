package routing

import (
	"testing"
	"time"
)

func TestModelAliasResolution(t *testing.T) {
	router := NewRouter(nil)
	router.SetAlias("claude-code", "claude-3-7-sonnet")

	res := router.ResolveModel("claude-code")
	if res != "claude-3-7-sonnet" {
		t.Fatalf("expected claude-3-7-sonnet, got %s", res)
	}

	unmapped := router.ResolveModel("gpt-4o")
	if unmapped != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", unmapped)
	}
}

func TestPriorityRoutingAndCooldown(t *testing.T) {
	cd := NewCooldownManager()
	router := NewRouter(cd)

	router.SetAccount(&Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	router.SetAccount(&Account{ID: "acc-2", ProviderID: "p2", State: "active", Enabled: true})

	route := &Route{
		Name:     "fast",
		Strategy: StrategyPriority,
		Enabled:  true,
		Items: []RouteItem{
			{ProviderID: "p1", AccountID: "acc-1", ModelName: "fast-1", Priority: 10, Enabled: true},
			{ProviderID: "p2", AccountID: "acc-2", ModelName: "fast-2", Priority: 20, Enabled: true},
		},
	}
	router.SetRoute(route)

	targets, err := router.SelectTargets("fast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].AccountID != "acc-1" {
		t.Errorf("expected acc-1 first, got %s", targets[0].AccountID)
	}

	cd.MarkFailure("acc-1", 10*time.Second)

	targetsAfterCooldown, err := router.SelectTargets("fast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsAfterCooldown) != 1 {
		t.Fatalf("expected 1 eligible target, got %d", len(targetsAfterCooldown))
	}
	if targetsAfterCooldown[0].AccountID != "acc-2" {
		t.Errorf("expected acc-2 when acc-1 is cooling down, got %s", targetsAfterCooldown[0].AccountID)
	}

	cd.MarkSuccess("acc-1")
	targetsRestored, err := router.SelectTargets("fast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetsRestored) != 2 || targetsRestored[0].AccountID != "acc-1" {
		t.Errorf("expected acc-1 restored after success")
	}
}

func TestRoundRobinRouting(t *testing.T) {
	router := NewRouter(nil)
	router.SetAccount(&Account{ID: "acc-1", ProviderID: "p1", State: "active", Enabled: true})
	router.SetAccount(&Account{ID: "acc-2", ProviderID: "p2", State: "active", Enabled: true})

	route := &Route{
		Name:     "rr",
		Strategy: StrategyRoundRobin,
		Enabled:  true,
		Items: []RouteItem{
			{ProviderID: "p1", AccountID: "acc-1", ModelName: "m1", Priority: 10, Enabled: true},
			{ProviderID: "p2", AccountID: "acc-2", ModelName: "m2", Priority: 10, Enabled: true},
		},
	}
	router.SetRoute(route)

	targets1, _ := router.SelectTargets("rr")
	targets2, _ := router.SelectTargets("rr")

	if targets1[0].AccountID == targets2[0].AccountID {
		t.Errorf("expected round robin to rotate first target; got %s and %s", targets1[0].AccountID, targets2[0].AccountID)
	}
}
