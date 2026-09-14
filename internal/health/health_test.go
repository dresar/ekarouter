package health

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
)

func TestHealthAndReadyHandlers(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "health_test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	checker := NewChecker(database.DB)

	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	recHealth := httptest.NewRecorder()
	checker.HealthHandler(recHealth, reqHealth)

	if recHealth.Code != http.StatusOK {
		t.Errorf("expected 200 for health, got %d", recHealth.Code)
	}

	reqReady := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recReady := httptest.NewRecorder()
	checker.ReadyHandler(recReady, reqReady)

	if recReady.Code != http.StatusOK {
		t.Errorf("expected 200 for ready, got %d", recReady.Code)
	}
}
