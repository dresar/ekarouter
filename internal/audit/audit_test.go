package audit_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/db"
)

func TestAuditLogger(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_audit.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	logger := audit.NewLogger(database.DB, 100)
	defer logger.Close()

	logger.Log(&audit.Record{
		ActorID:      "user-1",
		ActorType:    "user",
		Action:       "credential.create",
		ResourceType: "credential",
		ResourceID:   "cred-100",
		IPAddress:    "203.0.113.1",
		UserAgent:    "Mozilla/5.0",
		RequestID:    "req-abc",
		Result:       "success",
	})

	time.Sleep(50 * time.Millisecond)

	records, err := logger.Query(context.Background(), "credential.create", "", "", 10)
	if err != nil || len(records) != 1 {
		t.Fatalf("Query audit failed: len=%d, err=%v", len(records), err)
	}

	if records[0].ActorID != "user-1" || records[0].ResourceID != "cred-100" {
		t.Fatalf("Unexpected record data: %+v", records[0])
	}
}
