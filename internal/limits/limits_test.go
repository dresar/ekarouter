package limits_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/limits"
)

func TestEngineRateLimit(t *testing.T) {
	eng := limits.NewEngine(nil)

	key := "user:123"
	if !eng.AllowRate(key, 2, time.Second) {
		t.Fatal("First request should be allowed")
	}
	if !eng.AllowRate(key, 2, time.Second) {
		t.Fatal("Second request should be allowed")
	}
	if eng.AllowRate(key, 2, time.Second) {
		t.Fatal("Third request should be blocked")
	}
}

func TestEngineQuotaAndCooldown(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_limits.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	eng := limits.NewEngine(database.DB)
	ctx := context.Background()

	refType := "project"
	refID := "proj-alpha"
	metric := "requests"

	if err := eng.SetQuota(ctx, refType, refID, metric, "monthly", 100, nil); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}

	st, used, max, err := eng.CheckQuota(ctx, refType, refID, metric)
	if err != nil || st != limits.StateAvailable || used != 0 || max != 100 {
		t.Fatalf("CheckQuota: state=%s, used=%d, max=%d, err=%v", st, used, max, err)
	}

	if err := eng.IncrementQuota(ctx, refType, refID, metric, 95); err != nil {
		t.Fatalf("IncrementQuota: %v", err)
	}

	st, _, _, _ = eng.CheckQuota(ctx, refType, refID, metric)
	if st != limits.StateLimited {
		t.Fatalf("Expected StateLimited at 95%%, got %s", st)
	}

	if err := eng.IncrementQuota(ctx, refType, refID, metric, 10); err != nil {
		t.Fatalf("IncrementQuota: %v", err)
	}

	st, _, _, _ = eng.CheckQuota(ctx, refType, refID, metric)
	if st != limits.StateExhausted {
		t.Fatalf("Expected StateExhausted at 105, got %s", st)
	}

	_, _ = database.ExecContext(ctx, `
INSERT INTO vault_credentials (
    id, name, credential_type, provider_id, environment, encrypted_value,
    masked_value, status, health_state, notes, created_at, updated_at
) VALUES ('cred-xyz', 'Test Key', 'api_key', 'stripe', 'production', 'v1:xyz', 'sk-****abcd', 'active', 'healthy', '', datetime('now'), datetime('now'))`)

	if err := eng.SetCooldown(ctx, "cred-xyz", 5*time.Second, "rate_limited_429"); err != nil {
		t.Fatalf("SetCooldown: %v", err)
	}

	inCd, rem, err := eng.IsInCooldown(ctx, "cred-xyz")
	if err != nil || !inCd || rem <= 0 {
		t.Fatalf("Expected credential to be in cooldown: inCd=%v, rem=%v, err=%v", inCd, rem, err)
	}
}
