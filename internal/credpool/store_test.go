package credpool

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "credpool_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.Close()

	db, err := sql.Open("sqlite", "file:"+tmpFile.Name()+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
CREATE TABLE IF NOT EXISTS providers (id TEXT PRIMARY KEY, key TEXT UNIQUE NOT NULL, name TEXT NOT NULL, kind TEXT NOT NULL, base_url TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS accounts (id TEXT PRIMARY KEY, provider_id TEXT NOT NULL REFERENCES providers(id), name TEXT NOT NULL, auth_type TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'active', priority INTEGER NOT NULL DEFAULT 10, enabled INTEGER NOT NULL DEFAULT 1, expires_at DATETIME, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS credential_pools (id TEXT PRIMARY KEY, name TEXT NOT NULL, owner_id TEXT NOT NULL DEFAULT 'admin', provider_id TEXT, environment TEXT NOT NULL DEFAULT 'production', status TEXT NOT NULL DEFAULT 'active', created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS pool_members (id TEXT PRIMARY KEY, pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE, account_id TEXT NOT NULL, priority INTEGER NOT NULL DEFAULT 10, weight INTEGER NOT NULL DEFAULT 1, status TEXT NOT NULL DEFAULT 'active', cooldown_until DATETIME, last_used_at DATETIME, last_success_at DATETIME, last_failure_at DATETIME, expires_at DATETIME, failure_count INTEGER NOT NULL DEFAULT 0, success_count INTEGER NOT NULL DEFAULT 0, total_requests INTEGER NOT NULL DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pm_pool_account ON pool_members(pool_id, account_id);
CREATE TABLE IF NOT EXISTS rotation_policies (id TEXT PRIMARY KEY, pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE, strategy TEXT NOT NULL DEFAULT 'priority_fallback', disable_auto_rotation INTEGER NOT NULL DEFAULT 0, max_concurrent INTEGER NOT NULL DEFAULT 1, cooldown_seconds INTEGER NOT NULL DEFAULT 30, quota_aware INTEGER NOT NULL DEFAULT 0, retry_transient_only INTEGER NOT NULL DEFAULT 1, max_retries INTEGER NOT NULL DEFAULT 3, fallback_on_permanent INTEGER NOT NULL DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rp_pool ON rotation_policies(pool_id);
CREATE TABLE IF NOT EXISTS credential_health_checks (id TEXT PRIMARY KEY, member_id TEXT NOT NULL, pool_id TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'unknown', latency_ms INTEGER NOT NULL DEFAULT 0, message TEXT NOT NULL DEFAULT '', checked_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS credential_events (id INTEGER PRIMARY KEY AUTOINCREMENT, pool_id TEXT, member_id TEXT, account_id TEXT, event_type TEXT NOT NULL, details TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS rotation_decisions (id INTEGER PRIMARY KEY AUTOINCREMENT, pool_id TEXT NOT NULL, selected_member_id TEXT, strategy TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', candidates_count INTEGER NOT NULL DEFAULT 0, request_id TEXT NOT NULL DEFAULT '', latency_us INTEGER NOT NULL DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPoolCRUD(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "test-pool", OwnerID: "admin", Environment: "development"}
	if err := store.CreatePool(ctx, pool); err != nil {
		t.Fatalf("create pool: %v", err)
	}
	if pool.ID == "" {
		t.Fatal("pool ID empty after create")
	}

	got, err := store.GetPool(ctx, pool.ID)
	if err != nil {
		t.Fatalf("get pool: %v", err)
	}
	if got.Name != "test-pool" || got.Environment != "development" {
		t.Errorf("pool mismatch: %+v", got)
	}

	pools, err := store.ListPools(ctx, "admin")
	if err != nil || len(pools) != 1 {
		t.Fatalf("list pools: err=%v, count=%d", err, len(pools))
	}

	if err := store.UpdatePoolStatus(ctx, pool.ID, PoolPaused); err != nil {
		t.Fatalf("pause pool: %v", err)
	}
	paused, _ := store.GetPool(ctx, pool.ID)
	if paused.Status != PoolPaused {
		t.Errorf("expected paused, got %s", paused.Status)
	}

	if err := store.DeletePool(ctx, pool.ID); err != nil {
		t.Fatalf("delete pool: %v", err)
	}
	_, err = store.GetPool(ctx, pool.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestMemberCRUD(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "member-test", OwnerID: "admin"}
	_ = store.CreatePool(ctx, pool)

	m := &Member{PoolID: pool.ID, AccountID: "acc1", Priority: 5, Weight: 2}
	if err := store.AddMember(ctx, m); err != nil {
		t.Fatalf("add member: %v", err)
	}

	members, err := store.ListMembers(ctx, pool.ID)
	if err != nil || len(members) != 1 {
		t.Fatalf("list members: err=%v, count=%d", err, len(members))
	}
	if members[0].Priority != 5 || members[0].Weight != 2 {
		t.Errorf("member mismatch: %+v", members[0])
	}

	if err := store.MarkMemberUsed(ctx, m.ID, true); err != nil {
		t.Fatalf("mark success: %v", err)
	}
	updated, _ := store.GetMember(ctx, m.ID)
	if updated.SuccessCount != 1 || updated.TotalRequests != 1 {
		t.Errorf("counters wrong: success=%d, total=%d", updated.SuccessCount, updated.TotalRequests)
	}

	if err := store.MarkMemberUsed(ctx, m.ID, false); err != nil {
		t.Fatalf("mark failure: %v", err)
	}
	updated2, _ := store.GetMember(ctx, m.ID)
	if updated2.FailureCount != 1 || updated2.TotalRequests != 2 {
		t.Errorf("counters wrong: failure=%d, total=%d", updated2.FailureCount, updated2.TotalRequests)
	}

	cooldownTime := time.Now().Add(1 * time.Hour)
	if err := store.SetMemberCooldown(ctx, m.ID, cooldownTime); err != nil {
		t.Fatalf("set cooldown: %v", err)
	}
	cooled, _ := store.GetMember(ctx, m.ID)
	if cooled.Status != StatusCoolingDown {
		t.Errorf("expected cooling_down, got %s", cooled.Status)
	}

	if err := store.ClearMemberCooldown(ctx, m.ID); err != nil {
		t.Fatalf("clear cooldown: %v", err)
	}
	cleared, _ := store.GetMember(ctx, m.ID)
	if cleared.Status != StatusActive {
		t.Errorf("expected active, got %s", cleared.Status)
	}

	if err := store.RemoveMember(ctx, m.ID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	members2, _ := store.ListMembers(ctx, pool.ID)
	if len(members2) != 0 {
		t.Errorf("expected 0 members after remove, got %d", len(members2))
	}
}

func TestMemberAvailability(t *testing.T) {
	active := Member{Status: StatusActive}
	if !active.IsAvailable() {
		t.Error("active member should be available")
	}

	disabled := Member{Status: StatusDisabled}
	if disabled.IsAvailable() {
		t.Error("disabled member should not be available")
	}

	past := time.Now().Add(-1 * time.Hour)
	expired := Member{Status: StatusActive, ExpiresAt: &past}
	if expired.IsAvailable() {
		t.Error("expired member should not be available")
	}

	future := time.Now().Add(1 * time.Hour)
	cooled := Member{Status: StatusActive, CooldownUntil: &future}
	if cooled.IsAvailable() {
		t.Error("cooling down member should not be available")
	}

	valid := Member{Status: StatusValid}
	if !valid.IsAvailable() {
		t.Error("valid member should be available")
	}
}

func TestRotationPolicyCRUD(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "policy-test", OwnerID: "admin"}
	_ = store.CreatePool(ctx, pool)

	policy, err := store.GetPolicy(ctx, pool.ID)
	if err != nil {
		t.Fatalf("get default policy: %v", err)
	}
	if policy.Strategy != StrategyPriorityFallback {
		t.Errorf("default strategy should be priority_fallback, got %s", policy.Strategy)
	}

	policy.Strategy = StrategyRoundRobin
	policy.CooldownSeconds = 60
	if err := store.SetPolicy(ctx, policy); err != nil {
		t.Fatalf("update policy: %v", err)
	}

	updated, _ := store.GetPolicy(ctx, pool.ID)
	if updated.Strategy != StrategyRoundRobin || updated.CooldownSeconds != 60 {
		t.Errorf("policy not updated: %+v", updated)
	}
}

func TestEventRecording(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	pool := &Pool{Name: "event-test", OwnerID: "admin"}
	_ = store.CreatePool(ctx, pool)

	_ = store.RecordEvent(ctx, &Event{
		PoolID:  pool.ID,
		Type:    EventCreated,
		Details: "pool created",
	})

	events, err := store.ListEvents(ctx, pool.ID, 10)
	if err != nil || len(events) != 1 {
		t.Fatalf("list events: err=%v, count=%d", err, len(events))
	}
	if events[0].Type != EventCreated {
		t.Errorf("expected created event, got %s", events[0].Type)
	}
}
