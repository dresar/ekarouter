package usage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
)

func TestUsageRecorderAndSummary(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "usage_test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migrationsDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	rec := NewRecorder(database.DB, 100)

	rec.Record(UsageRecord{
		RequestID:    "req-1",
		ProviderID:   "p1",
		AccountID:    "acc1",
		ModelID:      "gpt-4o",
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
		LatencyMs:    200,
		Status:       200,
	})

	rec.Record(UsageRecord{
		RequestID:    "req-2",
		ProviderID:   "p1",
		AccountID:    "acc1",
		ModelID:      "gpt-4o",
		InputTokens:  200,
		OutputTokens: 100,
		TotalTokens:  300,
		LatencyMs:    400,
		Status:       200,
	})

	rec.Close()

	ctx := context.Background()
	summary, err := rec.GetSummary(ctx)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}

	if summary.TotalRequests != 2 {
		t.Errorf("expected 2 requests, got %d", summary.TotalRequests)
	}
	if summary.TotalTokens != 450 {
		t.Errorf("expected 450 total tokens, got %d", summary.TotalTokens)
	}
	if summary.AvgLatencyMs != 300 {
		t.Errorf("expected 300 avg latency, got %d", summary.AvgLatencyMs)
	}

	err = rec.PruneLogs(ctx, 0)
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
}
