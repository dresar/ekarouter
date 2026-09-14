package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/app"
	"github.com/dresar/ekarouter/internal/config"
)

func setupTestApp(t *testing.T) (*app.Application, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_plat_app.db")
	migPath := filepath.Join("..", "..", "migrations")

	cfg := &config.Config{
		Host:                "127.0.0.1",
		Port:                8999,
		DatabasePath:        dbPath,
		SecretKey:           "very-strong-secret-key-32-chars-long",
		AdminUser:           "admin",
		AdminPassword:       "admin12345",
		AllowLocalProviders: true,
		TokenSaverMode:      "off",
		MaxRequestBodyBytes: 10 * 1024 * 1024,
		CORSOrigins:         []string{"*"},
		ReadTimeout:         5 * time.Second,
		WriteTimeout:        5 * time.Second,
		IdleTimeout:         5 * time.Second,
		LogRetentionDays:    30,
	}

	appl, err := app.Setup(cfg, migPath)
	if err != nil {
		t.Fatalf("app.Setup: %v", err)
	}

	_, _ = appl.DB.Exec("INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES ('key_1', 'Dev Key', 'eka_live_123', 'test_token_hash', '*', 1)")
	_, _ = appl.DB.Exec("INSERT INTO client_tokens (id, name, token_hash, scopes, created_at) VALUES ('tok_1', 'CI Token', 'test_client_token_hash', '*', datetime('now'))")

	return appl, "eka_test_token"
}

func TestLiveEndpoint(t *testing.T) {
	appl, _ := setupTestApp(t)
	defer appl.DB.Close()

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rr := httptest.NewRecorder()
	appl.Server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /live, got %d", rr.Code)
	}
}

func TestPlatformAPIEndpoints(t *testing.T) {
	appl, _ := setupTestApp(t)
	defer appl.DB.Close()

	var sessionToken string
	loginPayload := `{"username":"admin","password":"admin12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	appl.Server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Login failed: code=%d body=%s", rr.Code, rr.Body.String())
	}

	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" {
			sessionToken = c.Value
			break
		}
	}
	if sessionToken == "" {
		t.Fatal("Session token not found in cookies")
	}

	authGet := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer "+sessionToken)
		w := httptest.NewRecorder()
		appl.Server.Handler.ServeHTTP(w, r)
		return w
	}

	authPost := func(path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+sessionToken)
		w := httptest.NewRecorder()
		appl.Server.Handler.ServeHTTP(w, r)
		return w
	}

	authPatch := func(path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+sessionToken)
		w := httptest.NewRecorder()
		appl.Server.Handler.ServeHTTP(w, r)
		return w
	}

	authDelete := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		r.Header.Set("Authorization", "Bearer "+sessionToken)
		w := httptest.NewRecorder()
		appl.Server.Handler.ServeHTTP(w, r)
		return w
	}

	// 1. Providers
	res := authGet("/api/v1/providers")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/providers failed: %d", res.Code)
	}

	res = authGet("/api/v1/providers/cloudflare")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/providers/cloudflare failed: %d", res.Code)
	}

	res = authGet("/api/v1/providers/cloudflare/capabilities")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/providers/cloudflare/capabilities failed: %d", res.Code)
	}

	res = authGet("/api/v1/providers/cloudflare/docs")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/providers/cloudflare/docs failed: %d", res.Code)
	}

	res = authPost("/api/v1/providers/cloudflare/validate", map[string]string{"secret": "test-cf-token"})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/providers/cloudflare/validate failed: %d", res.Code)
	}

	// 2. Projects & Environments
	res = authPost("/api/v1/projects", map[string]string{
		"name":        "Test Project",
		"environment": "production",
		"description": "Integration testing project",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/projects failed: %d %s", res.Code, res.Body.String())
	}
	var projCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &projCreated)
	projID := projCreated.Data.ID

	res = authGet("/api/v1/projects")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects failed: %d", res.Code)
	}

	res = authGet("/api/v1/projects/" + projID)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects/:id failed: %d", res.Code)
	}

	res = authPatch("/api/v1/projects/"+projID, map[string]string{
		"name": "Updated Test Project",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/projects/:id failed: %d", res.Code)
	}

	res = authPost("/api/v1/environments", map[string]string{
		"name":        "staging",
		"description": "Staging environment",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/environments failed: %d", res.Code)
	}
	var envCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &envCreated)
	envID := envCreated.Data.ID

	res = authGet("/api/v1/environments")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/environments failed: %d", res.Code)
	}

	res = authPatch("/api/v1/environments/"+envID, map[string]string{
		"name": "staging-updated",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/environments/:id failed: %d", res.Code)
	}

	// 3. Credentials
	credReq := map[string]any{
		"name":            "Cloudflare Prod Key",
		"credential_type": "api_key",
		"provider_id":     "cloudflare",
		"environment":     "production",
		"secret_value":    "cf_api_secret_token_value_xyz",
		"priority":        1,
		"tags":            "dns,workers",
	}
	res = authPost("/api/v1/credentials", credReq)
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/credentials failed: %d %s", res.Code, res.Body.String())
	}

	var credResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID          string `json:"id"`
			MaskedValue string `json:"masked_value"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &credResp)
	credID := credResp.Data.ID
	if credID == "" {
		t.Fatalf("Expected credential ID, got empty")
	}

	res = authGet("/api/v1/credentials/" + credID)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/credentials/:id failed: %d", res.Code)
	}

	res = authPatch("/api/v1/credentials/"+credID, map[string]any{
		"name":     "Renamed Cloudflare Key",
		"priority": 2,
		"tags":     "dns,workers,v2",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/credentials/:id failed: %d", res.Code)
	}

	res = authPost("/api/v1/credentials/"+credID+"/rotate", map[string]string{
		"new_secret": "cf_api_secret_token_new_value_456",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/credentials/:id/rotate failed: %d", res.Code)
	}

	res = authGet("/api/v1/credentials/" + credID + "/usage")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/credentials/:id/usage failed: %d", res.Code)
	}

	res = authGet("/api/v1/credentials/" + credID + "/health")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/credentials/:id/health failed: %d", res.Code)
	}

	// 4. Tools
	toolReq := map[string]any{
		"provider_id":  "cloudflare",
		"name":         "List DNS Zones",
		"description":  "List all zones on Cloudflare account",
		"category":     "developer",
		"method":       "GET",
		"url_template": "https://api.cloudflare.com/client/v4/zones",
		"timeout_ms":   10000,
	}
	res = authPost("/api/v1/tools", toolReq)
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/tools failed: %d %s", res.Code, res.Body.String())
	}

	var toolCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &toolCreated)
	toolID := toolCreated.Data.ID

	res = authGet("/api/v1/tools")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/tools failed: %d", res.Code)
	}

	res = authGet("/api/v1/tools/" + toolID + "/schema")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/tools/:id/schema failed: %d", res.Code)
	}

	res = authPost("/api/v1/request-templates", map[string]any{
		"name":   "Get Cloudflare Zones",
		"method": "GET",
		"path":   "https://api.cloudflare.com/client/v4/zones",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/request-templates failed: %d %s", res.Code, res.Body.String())
	}
	var tmplCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &tmplCreated)
	tmplID := tmplCreated.Data.ID

	res = authGet("/api/v1/request-templates")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/request-templates failed: %d", res.Code)
	}

	res = authGet("/api/v1/request-templates/" + tmplID)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/request-templates/:id failed: %d", res.Code)
	}

	res = authPatch("/api/v1/request-templates/"+tmplID, map[string]string{
		"name": "Updated Get Cloudflare Zones",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/request-templates/:id failed: %d", res.Code)
	}

	res = authPost("/api/v1/request-templates/"+tmplID+"/execute", map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/request-templates/:id/execute failed: %d %s", res.Code, res.Body.String())
	}

	res = authDelete("/api/v1/request-templates/" + tmplID)
	if res.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/request-templates/:id failed: %d", res.Code)
	}

	res = authGet("/api/v1/usage")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/usage failed: %d", res.Code)
	}

	res = authGet("/api/v1/usage/summary")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/usage/summary failed: %d", res.Code)
	}

	res = authGet("/api/v1/usage/providers")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/usage/providers failed: %d", res.Code)
	}

	res = authGet("/api/v1/usage/credentials")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/usage/credentials failed: %d", res.Code)
	}

	res = authGet("/api/v1/usage/projects")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/usage/projects failed: %d", res.Code)
	}

	res = authGet("/api/v1/health")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/health failed: %d", res.Code)
	}

	res = authGet("/api/v1/health/providers")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/health/providers failed: %d", res.Code)
	}

	res = authGet("/api/v1/health/credentials")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/health/credentials failed: %d", res.Code)
	}

	res = authGet("/api/v1/audit-logs")
	if res.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/audit-logs failed: %d", res.Code)
	}

	res = authDelete("/api/v1/credentials/" + credID)
	if res.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/credentials/:id failed: %d", res.Code)
	}

	res = authDelete("/api/v1/environments/" + envID)
	if res.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/environments/:id failed: %d", res.Code)
	}

	res = authDelete("/api/v1/projects/" + projID)
	if res.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/projects/:id failed: %d", res.Code)
	}
}
