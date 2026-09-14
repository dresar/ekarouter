package scheduler_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/scheduler"
)

func TestSchedulerLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sched.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	reg := platform.NewRegistry()
	sched := scheduler.NewScheduler(database.DB, reg, 30, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)
	time.Sleep(250 * time.Millisecond)
	sched.Stop()

	var runCount int
	err = database.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_runs").Scan(&runCount)
	if err != nil || runCount == 0 {
		t.Fatalf("Expected recorded task runs, got %d, err=%v", runCount, err)
	}
}
