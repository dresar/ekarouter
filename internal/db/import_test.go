package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
)

func TestImport9RouterBackupComprehensive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ekarouter-import-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	migDir := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	crypto, err := auth.NewCryptoService("test-secret-key-32-bytes-long-now!!")
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	backupPath := filepath.Join("..", "..", "9router-backup-2026-09-14T19-54-33-161Z.json")
	stats, err := Import9RouterBackup(database, crypto, backupPath)
	if err != nil {
		t.Fatalf("failed to import backup: %v", err)
	}

	if stats.Providers < 40 {
		t.Errorf("expected at least 40 providers, got %d", stats.Providers)
	}
	if stats.Accounts < 182 {
		t.Errorf("expected at least 182 accounts, got %d", stats.Accounts)
	}
	if stats.Credentials < 182 {
		t.Errorf("expected at least 182 credentials, got %d", stats.Credentials)
	}
	if stats.ProxyPools != 16 {
		t.Errorf("expected 16 proxy pools, got %d", stats.ProxyPools)
	}
	if stats.APIKeys != 1 {
		t.Errorf("expected 1 API key, got %d", stats.APIKeys)
	}
	if stats.Routes != 2 {
		t.Errorf("expected 2 routes, got %d", stats.Routes)
	}

	var oaBaseURL string
	err = database.QueryRow("SELECT base_url FROM providers WHERE id = 'node_openagentic_id'").Scan(&oaBaseURL)
	if err != nil {
		t.Fatalf("failed to find node_openagentic_id: %v", err)
	}
	if oaBaseURL != "https://openagentic.id/api/v1" {
		t.Errorf("expected https://openagentic.id/api/v1, got %s", oaBaseURL)
	}

	var proxyCount int
	_ = database.QueryRow("SELECT COUNT(*) FROM proxy_profiles WHERE enabled = 1").Scan(&proxyCount)
	if proxyCount != 16 {
		t.Errorf("expected 16 enabled proxies, got %d", proxyCount)
	}

	var routeItemCount int
	_ = database.QueryRow("SELECT COUNT(*) FROM route_items WHERE enabled = 1").Scan(&routeItemCount)
	if routeItemCount < 60 {
		t.Errorf("expected at least 60 route items, got %d", routeItemCount)
	}
}
