package credpool

import (
	"context"
	"errors"
	"testing"

	"github.com/dresar/ekarouter/internal/providers"
)

func setupEngineTest(t *testing.T) (*Engine, *Store, string) {
	t.Helper()
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "engine-test", OwnerID: "admin"}
	if err := store.CreatePool(ctx, pool); err != nil {
		t.Fatal(err)
	}

	for i, acc := range []string{"acc_a", "acc_b", "acc_c"} {
		m := &Member{PoolID: pool.ID, AccountID: acc, Priority: (i + 1) * 10, Weight: 3 - i}
		if err := store.AddMember(ctx, m); err != nil {
			t.Fatal(err)
		}
	}

	engine := NewEngine(store)
	return engine, store, pool.ID
}

func TestSelectPriorityFallback(t *testing.T) {
	engine, _, poolID := setupEngineTest(t)
	ctx := context.Background()

	member, err := engine.Select(ctx, &SelectRequest{PoolID: poolID, RequestID: "req1"})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if member.AccountID != "acc_a" {
		t.Errorf("expected acc_a (priority 10), got %s", member.AccountID)
	}
}

func TestSelectRoundRobin(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	policy, _ := store.GetPolicy(ctx, poolID)
	policy.Strategy = StrategyRoundRobin
	_ = store.SetPolicy(ctx, policy)

	seen := map[string]int{}
	for i := 0; i < 9; i++ {
		member, err := engine.Select(ctx, &SelectRequest{PoolID: poolID, RequestID: "req"})
		if err != nil {
			t.Fatalf("select %d: %v", i, err)
		}
		seen[member.AccountID]++
	}

	if len(seen) < 2 {
		t.Errorf("round robin should distribute across members, got %v", seen)
	}
}

func TestSelectLRU(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	policy, _ := store.GetPolicy(ctx, poolID)
	policy.Strategy = StrategyLRU
	_ = store.SetPolicy(ctx, policy)

	member, err := engine.Select(ctx, &SelectRequest{PoolID: poolID})
	if err != nil {
		t.Fatalf("select LRU: %v", err)
	}
	if member == nil {
		t.Fatal("expected non-nil member")
	}
}

func TestSelectLowestFailureRate(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	members, _ := store.ListMembers(ctx, poolID)
	for i := 0; i < 5; i++ {
		_ = store.MarkMemberUsed(ctx, members[0].ID, false)
	}
	_ = store.MarkMemberUsed(ctx, members[0].ID, true)

	policy, _ := store.GetPolicy(ctx, poolID)
	policy.Strategy = StrategyLowestFailureRate
	_ = store.SetPolicy(ctx, policy)

	member, err := engine.Select(ctx, &SelectRequest{PoolID: poolID})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if member.AccountID == members[0].AccountID {
		t.Errorf("should not select member with highest failure rate")
	}
}

func TestSelectManual(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	policy, _ := store.GetPolicy(ctx, poolID)
	policy.Strategy = StrategyManual
	_ = store.SetPolicy(ctx, policy)

	members, _ := store.ListMembers(ctx, poolID)
	target := members[1]

	member, err := engine.Select(ctx, &SelectRequest{PoolID: poolID, MemberID: target.ID})
	if err != nil {
		t.Fatalf("manual select: %v", err)
	}
	if member.ID != target.ID {
		t.Errorf("expected %s, got %s", target.ID, member.ID)
	}

	_, err = engine.Select(ctx, &SelectRequest{PoolID: poolID})
	if !errors.Is(err, ErrManualMemberRequired) {
		t.Errorf("expected ErrManualMemberRequired, got %v", err)
	}
}

func TestSelectPoolPaused(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	_ = store.UpdatePoolStatus(ctx, poolID, PoolPaused)

	_, err := engine.Select(ctx, &SelectRequest{PoolID: poolID})
	if !errors.Is(err, ErrPoolPaused) {
		t.Errorf("expected ErrPoolPaused, got %v", err)
	}
}

func TestSelectAllExhausted(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "exhausted-test", OwnerID: "admin"}
	_ = store.CreatePool(ctx, pool)

	m := &Member{PoolID: pool.ID, AccountID: "acc_x", Status: StatusDisabled}
	_ = store.AddMember(ctx, m)

	engine := NewEngine(store)
	_, err := engine.Select(ctx, &SelectRequest{PoolID: pool.ID})
	if !errors.Is(err, ErrAllCredentialsExhausted) {
		t.Errorf("expected ErrAllCredentialsExhausted, got %v", err)
	}
}

func TestSelectEmptyPool(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "empty-test", OwnerID: "admin"}
	_ = store.CreatePool(ctx, pool)

	engine := NewEngine(store)
	_, err := engine.Select(ctx, &SelectRequest{PoolID: pool.ID})
	if !errors.Is(err, ErrNoMembers) {
		t.Errorf("expected ErrNoMembers, got %v", err)
	}
}

func TestFailureClassification(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantRotate    bool
		wantRetry     bool
		wantPermanent bool
	}{
		{
			name:       "rate_limit",
			err:        &providers.ProviderError{StatusCode: 429, Class: providers.ErrorClassRateLimit},
			wantRotate: true, wantRetry: true,
		},
		{
			name:       "quota",
			err:        &providers.ProviderError{StatusCode: 402, Class: providers.ErrorClassQuota},
			wantRotate: true, wantRetry: false,
		},
		{
			name:       "auth_failure",
			err:        &providers.ProviderError{StatusCode: 401, Class: providers.ErrorClassAuth},
			wantRotate: false, wantPermanent: true,
		},
		{
			name:       "invalid_request",
			err:        &providers.ProviderError{StatusCode: 400, Class: providers.ErrorClassInvalidRequest},
			wantRotate: false, wantPermanent: true,
		},
		{
			name:       "server_error",
			err:        &providers.ProviderError{StatusCode: 500, Class: providers.ErrorClassUpstream5xx},
			wantRotate: true, wantRetry: true,
		},
		{
			name:       "timeout",
			err:        &providers.ProviderError{StatusCode: 0, Class: providers.ErrorClassTimeout},
			wantRotate: true, wantRetry: true,
		},
		{
			name:       "nil_error",
			err:        nil,
			wantRotate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := ClassifyForRotation(tt.err)
			if action.ShouldRotate != tt.wantRotate {
				t.Errorf("ShouldRotate: got %v, want %v", action.ShouldRotate, tt.wantRotate)
			}
			if action.ShouldRetry != tt.wantRetry {
				t.Errorf("ShouldRetry: got %v, want %v", action.ShouldRetry, tt.wantRetry)
			}
			if action.IsPermanent != tt.wantPermanent {
				t.Errorf("IsPermanent: got %v, want %v", action.IsPermanent, tt.wantPermanent)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		input       error
		contains    string
		notContains string
	}{
		{nil, "", ""},
		{errors.New("Bearer sk-abc123 failed"), "redacted", "sk-abc123"},
		{errors.New("key_test123 unauthorized"), "redacted", "key_test123"},
		{&providers.ProviderError{Class: providers.ErrorClassAuth, Message: "auth failed"}, "auth", ""},
	}

	for _, tt := range tests {
		result := SanitizeError(tt.input)
		if tt.notContains != "" {
			if containsStr(result, tt.notContains) {
				t.Errorf("sanitized error should not contain %q, got %q", tt.notContains, result)
			}
		}
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || findSubstr(s, substr))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestMarkSuccessFailure(t *testing.T) {
	engine, store, poolID := setupEngineTest(t)
	ctx := context.Background()

	members, _ := store.ListMembers(ctx, poolID)
	m := members[0]

	if err := engine.MarkSuccess(ctx, m.ID); err != nil {
		t.Fatalf("mark success: %v", err)
	}
	updated, _ := store.GetMember(ctx, m.ID)
	if updated.SuccessCount != 1 {
		t.Errorf("success count should be 1, got %d", updated.SuccessCount)
	}

	policy, _ := store.GetPolicy(ctx, poolID)
	if err := engine.MarkFailure(ctx, m.ID, policy); err != nil {
		t.Fatalf("mark failure: %v", err)
	}
	updated2, _ := store.GetMember(ctx, m.ID)
	if updated2.FailureCount != 1 {
		t.Errorf("failure count should be 1, got %d", updated2.FailureCount)
	}
	if updated2.Status != StatusCoolingDown {
		t.Errorf("expected cooling_down after failure, got %s", updated2.Status)
	}
}
