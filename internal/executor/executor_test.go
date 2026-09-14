package executor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/vault"
)

func TestInterpolationAndSecurity(t *testing.T) {
	tmpl := "https://api.example.com/v1/users/{{user_id}}/records/{{record_id}}"
	vars := map[string]string{
		"user_id":   "usr_123",
		"record_id": "rec_999",
	}

	res := executor.InterpolateString(tmpl, vars)
	expected := "https://api.example.com/v1/users/usr_123/records/rec_999"
	if res != expected {
		t.Fatalf("Interpolation failed: got %s, want %s", res, expected)
	}

	_, err := executor.SanitizeHeaderValue("clean-value")
	if err != nil {
		t.Fatalf("Unexpected error for clean header: %v", err)
	}

	_, err = executor.SanitizeHeaderValue("bad-value\r\nInjected-Header: evil")
	if err == nil {
		t.Fatal("Expected error for CRLF injection in header value, got nil")
	}

	_, err = executor.SanitizeHeaderKey("Invalid:Header")
	if err == nil {
		t.Fatal("Expected error for colon in header key, got nil")
	}
}

func TestToolExecutor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-test-token") != "secret-token-123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"id":"test_result"}`))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_exec.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	v, _ := vault.NewVault("test-secret-key-at-least-16-bytes")
	vStore := vault.NewStore(database.DB, v)

	reg := platform.NewRegistry()
	mockAdapter := platform.NewBaseAdapterWithOptions(platform.ProviderMetadata{
		ID:       "mock-tool-provider",
		Name:     "Mock Provider",
		Category: platform.CategoryDeveloper,
		BaseURL:  server.URL,
		AuthType: platform.AuthTypeCustomHeader,
		Enabled:  true,
	}, 5*time.Second, true)
	_ = reg.Register(mockAdapter)

	exec := executor.NewExecutor(database.DB, reg, vStore, true)
	ctx := context.Background()

	cred := &vault.Credential{
		Name:           "Mock Credential",
		CredentialType: vault.TypeAPIKey,
		ProviderID:     "mock-tool-provider",
		Environment:    "production",
		Priority:       1,
	}
	if err := vStore.CreateCredential(ctx, cred, "test-secret"); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	tool := &executor.ToolDefinition{
		ProviderID:  "mock-tool-provider",
		Name:        "Test Tool",
		Description: "Executes a test operation",
		Category:    "developer",
		Method:      "POST",
		URLTemplate: server.URL + "/action/{{action_id}}",
		HeadersTemplate: map[string]string{
			"x-test-token": "secret-token-123",
		},
		QueryTemplate: map[string]string{
			"mode": "{{mode}}",
		},
		TimeoutMS: 5000,
	}

	if err := exec.CreateTool(ctx, tool); err != nil {
		t.Fatalf("CreateTool: %v", err)
	}

	params := &executor.ExecutionParams{
		ToolID: tool.ID,
		Variables: map[string]string{
			"action_id": "act_42",
			"mode":      "fast",
		},
		Body:        []byte(`{"param":"val"}`),
		Environment: "production",
	}

	res, err := exec.ExecuteTool(ctx, params)
	if err != nil {
		t.Fatalf("ExecuteTool: %v", err)
	}

	if res.StatusCode != http.StatusOK || res.Status != "success" {
		t.Fatalf("Execution returned invalid result: %+v", res)
	}

	tools, err := exec.ListTools(ctx, "mock-tool-provider")
	if err != nil || len(tools) != 1 {
		t.Fatalf("ListTools: len=%d, err=%v", len(tools), err)
	}
}
