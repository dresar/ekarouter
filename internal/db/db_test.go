package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDBOpenAndMigrate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	var count int
	err = database.QueryRow("SELECT COUNT(1) FROM schema_migrations WHERE version = 1").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migration version: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 migration recorded, got %d", count)
	}

	backupPath := filepath.Join(tempDir, "backup.db")
	if err := database.Backup(backupPath); err != nil {
		t.Fatalf("failed to backup db: %v", err)
	}

	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("backup file does not exist: %v", err)
	}
}
