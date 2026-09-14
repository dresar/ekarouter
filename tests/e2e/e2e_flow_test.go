package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/app"
	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/routing"
)

func TestEndToEndPlatformAndGatewayWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "e2e.db")
	migPath := filepath.Join("..", "..", "migrations")

	cfg := &config.Config{
		Host:                "127.0.0.1",
		Port:                8998,
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

	application, err := app.Setup(cfg, migPath)
	if err != nil {
		t.Fatalf("Failed to setup e2e application: %v", err)
	}
	defer application.DB.Close()

	loginBody := `{"username":"admin","password":"admin12345"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	rrLogin := httptest.NewRecorder()
	application.Server.Handler.ServeHTTP(rrLogin, reqLogin)
	if rrLogin.Code != http.StatusOK {
		t.Fatalf("Step 1 failed - login: %d %s", rrLogin.Code, rrLogin.Body.String())
	}

	var sessionToken string
	for _, c := range rrLogin.Result().Cookies() {
		if c.Name == "session_token" {
			sessionToken = c.Value
			break
		}
	}
	if sessionToken == "" {
		t.Fatalf("Step 1 failed - no session token cookie")
	}

	adminReq := func(method, path string, body any) *httptest.ResponseRecorder {
		var rdr *bytes.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rdr = bytes.NewReader(b)
		} else {
			rdr = bytes.NewReader([]byte{})
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+sessionToken)
		rr := httptest.NewRecorder()
		application.Server.Handler.ServeHTTP(rr, req)
		return rr
	}

	resProj := adminReq(http.MethodPost, "/api/v1/projects", map[string]string{
		"name":        "E2E Production Project",
		"description": "Production routing test",
	})
	if resProj.Code != http.StatusOK && resProj.Code != http.StatusCreated {
		t.Fatalf("Step 2 failed - create project: %d %s", resProj.Code, resProj.Body.String())
	}
	var projResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(resProj.Body.Bytes(), &projResp)
	projID := projResp.Data.ID

	resEnv := adminReq(http.MethodPost, "/api/v1/environments", map[string]string{
		"name":        "production",
		"project_id":  projID,
		"description": "Live production environment",
	})
	if resEnv.Code != http.StatusOK && resEnv.Code != http.StatusCreated {
		t.Fatalf("Step 3 failed - create environment: %d %s", resEnv.Code, resEnv.Body.String())
	}

	resCred := adminReq(http.MethodPost, "/api/v1/credentials", map[string]any{
		"name":         "OpenAI Production Key",
		"provider_id":  "openai",
		"secret_value": "sk-live-mock-api-key-value-1234567890",
		"project_id":   projID,
		"environment":  "production",
	})
	if resCred.Code != http.StatusOK && resCred.Code != http.StatusCreated {
		t.Fatalf("Step 4 failed - create credential: %d %s", resCred.Code, resCred.Body.String())
	}
	var credResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(resCred.Body.Bytes(), &credResp)
	credID := credResp.Data.ID

	_, err = application.DB.Exec(`
INSERT INTO providers (id, key, name, kind, base_url, enabled)
VALUES ('prov_openai', 'openai', 'OpenAI', 'openai', 'https://api.openai.com/v1', 1)
ON CONFLICT(id) DO UPDATE SET enabled=1`)
	if err != nil {
		t.Fatalf("Step 5 failed - insert provider: %v", err)
	}

	_, err = application.DB.Exec(`
INSERT INTO accounts (id, provider_id, name, auth_type, state, priority, enabled)
VALUES ('acc_openai_prod', 'prov_openai', 'OpenAI Main Account', 'api_key', 'active', 10, 1)
ON CONFLICT(id) DO UPDATE SET enabled=1`)
	if err != nil {
		t.Fatalf("Step 5 failed - insert account: %v", err)
	}

	crypto, _ := auth.NewCryptoService(cfg.SecretKey)
	encAccess, _ := crypto.Encrypt("sk-test-openai-credential-key")
	_, err = application.DB.Exec(`
INSERT INTO credentials (id, account_id, encrypted_access)
VALUES ('cred_openai_prod', 'acc_openai_prod', ?)
ON CONFLICT(account_id) DO UPDATE SET encrypted_access = excluded.encrypted_access`, encAccess)
	if err != nil {
		t.Fatalf("Step 5 failed - insert credentials: %v", err)
	}

	_, err = application.DB.Exec(`
INSERT INTO models (id, provider_id, external_name, display_name, enabled)
VALUES ('mod_gpt4o', 'prov_openai', 'gpt-4o', 'GPT-4o', 1)
ON CONFLICT(id) DO UPDATE SET enabled=1`)
	if err != nil {
		t.Fatalf("Step 5 failed - insert model: %v", err)
	}

	_, err = application.DB.Exec(`
INSERT INTO routes (id, name, strategy, enabled)
VALUES ('rt_gpt4o', 'gpt-4o', 'priority', 1)
ON CONFLICT(id) DO UPDATE SET enabled=1`)
	if err != nil {
		t.Fatalf("Step 5 failed - insert route: %v", err)
	}

	_, err = application.DB.Exec(`
INSERT INTO route_items (id, route_id, provider_id, account_id, model_id, priority, enabled)
VALUES ('ri_gpt4o', 'rt_gpt4o', 'prov_openai', 'acc_openai_prod', 'mod_gpt4o', 10, 1)
ON CONFLICT(id) DO UPDATE SET enabled=1`)
	if err != nil {
		t.Fatalf("Step 5 failed - insert route item: %v", err)
	}

	_ = routing.NewRouter(routing.NewCooldownManager()).LoadFromDB(context.Background(), application.DB.DB)

	rawDevKey, prefix, keyHash, _ := auth.GenerateApiKey()
	_, err = application.DB.Exec("INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES ('dev_key_e2e', 'E2E Dev Key', ?, ?, '*', 1)", prefix, keyHash)
	if err != nil {
		t.Fatalf("Step 6 failed - insert api key: %v", err)
	}

	reqModels := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	reqModels.Header.Set("Authorization", "Bearer "+rawDevKey)
	rrModels := httptest.NewRecorder()
	application.Server.Handler.ServeHTTP(rrModels, reqModels)
	if rrModels.Code != http.StatusOK {
		t.Fatalf("Step 7 failed - GET /v1/models: %d %s", rrModels.Code, rrModels.Body.String())
	}
	if !strings.Contains(rrModels.Body.String(), "gpt-4o") {
		t.Fatalf("Step 7 failed - models list does not contain gpt-4o: %s", rrModels.Body.String())
	}

	resRotate := adminReq(http.MethodPost, "/api/v1/credentials/"+credID+"/rotate", map[string]string{
		"new_secret": "sk-live-mock-rotated-key-9999999999",
	})
	if resRotate.Code != http.StatusOK {
		t.Fatalf("Step 8 failed - rotate credential: %d %s", resRotate.Code, resRotate.Body.String())
	}

	resHealth := adminReq(http.MethodGet, "/api/v1/health", nil)
	if resHealth.Code != http.StatusOK {
		t.Fatalf("Step 9 failed - GET /api/v1/health: %d", resHealth.Code)
	}

	resUsage := adminReq(http.MethodGet, "/api/v1/usage/summary", nil)
	if resUsage.Code != http.StatusOK {
		t.Fatalf("Step 9 failed - GET /api/v1/usage/summary: %d", resUsage.Code)
	}

	resAudit := adminReq(http.MethodGet, "/api/v1/audit-logs", nil)
	if resAudit.Code != http.StatusOK {
		t.Fatalf("Step 10 failed - GET /api/v1/audit-logs: %d", resAudit.Code)
	}

	resDelCred := adminReq(http.MethodDelete, "/api/v1/credentials/"+credID, nil)
	if resDelCred.Code != http.StatusOK {
		t.Fatalf("Step 11 failed - delete credential: %d", resDelCred.Code)
	}
}
