package app

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/config"
)

func TestAppSetupAndGracefulShutdown(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = 18088
	cfg.DatabasePath = filepath.Join(tempDir, "app_test.db")

	migrationsDir := filepath.Join("..", "..", "migrations")
	application, err := Setup(cfg, migrationsDir)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		errChan <- application.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18088/health")
	if err != nil {
		t.Fatalf("failed to call /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("expected clean shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown timed out")
	}
}
