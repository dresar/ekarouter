package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/app"
	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
)

type TestEnv struct {
	App         *app.Application
	AdminToken  string
	ApiKeyToken string
	ClientToken string
	DB          *sql.DB
}

func SetupTestEnv(t *testing.T) *TestEnv {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "qa_test.db")
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

	application, err := app.Setup(cfg, migPath)
	if err != nil {
		t.Fatalf("Failed to setup test application: %v", err)
	}

	loginPayload := `{"username":"admin","password":"admin12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	application.Server.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Admin login failed: %d %s", rr.Code, rr.Body.String())
	}

	var sessionToken string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" {
			sessionToken = c.Value
			break
		}
	}

	rawApiKey, prefix, keyHash, err := auth.GenerateApiKey()
	if err != nil {
		t.Fatalf("GenerateApiKey: %v", err)
	}
	_, err = application.DB.Exec("INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES ('qa_key_1', 'QA Key', ?, ?, '*', 1)", prefix, keyHash)
	if err != nil {
		t.Fatalf("Insert api_key: %v", err)
	}

	rawClientTok, _, ctHash, err := auth.GenerateApiKey()
	if err != nil {
		t.Fatalf("Generate client token: %v", err)
	}
	_, err = application.DB.Exec("INSERT INTO client_tokens (id, name, token_hash, scopes, created_at) VALUES ('qa_ct_1', 'QA Client Token', ?, '*', datetime('now'))", ctHash)
	if err != nil {
		t.Fatalf("Insert client_token: %v", err)
	}

	return &TestEnv{
		App:         application,
		AdminToken:  sessionToken,
		ApiKeyToken: rawApiKey,
		ClientToken: rawClientTok,
		DB:          application.DB.DB,
	}
}

func (e *TestEnv) Teardown() {
	if e.App != nil && e.App.DB != nil {
		_ = e.App.DB.Close()
	}
}

func (e *TestEnv) Request(method, path string, body any, authHeader string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			b, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(b)
		}
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rr := httptest.NewRecorder()
	e.App.Server.Handler.ServeHTTP(rr, req)
	return rr
}

func (e *TestEnv) AdminReq(method, path string, body any) *httptest.ResponseRecorder {
	return e.Request(method, path, body, "Bearer "+e.AdminToken)
}

func (e *TestEnv) GatewayReq(method, path string, body any) *httptest.ResponseRecorder {
	return e.Request(method, path, body, "Bearer "+e.ApiKeyToken)
}

func (e *TestEnv) PlatformClientReq(method, path string, body any) *httptest.ResponseRecorder {
	return e.Request(method, path, body, "Bearer "+e.ClientToken)
}

func (e *TestEnv) SeedDummyRoute(t *testing.T, modelName string) {
	_, _ = e.DB.Exec("INSERT OR IGNORE INTO providers (id, key, name, kind, base_url, enabled) VALUES ('prov_mock', 'mock', 'Mock Provider', 'mock', 'https://mock.api', 1)")
	_, _ = e.DB.Exec("INSERT OR IGNORE INTO accounts (id, provider_id, name, auth_type, state, enabled) VALUES ('acc_mock', 'prov_mock', 'Mock Account', 'api_key', 'active', 1)")
	_, _ = e.DB.Exec("INSERT OR IGNORE INTO models (id, provider_id, external_name, display_name, enabled) VALUES ('mod_mock', 'prov_mock', ?, ?, 1)", modelName, modelName)
	_, _ = e.DB.Exec("INSERT OR IGNORE INTO routes (id, name, strategy, enabled) VALUES ('rt_mock', ?, 'priority', 1)", modelName)
	_, _ = e.DB.Exec("INSERT OR IGNORE INTO route_items (id, route_id, provider_id, account_id, model_id, priority, enabled) VALUES ('ri_mock', 'rt_mock', 'prov_mock', 'acc_mock', 'mod_mock', 10, 1)")

	crypto, _ := auth.NewCryptoService("very-strong-secret-key-32-chars-long")
	encAccess, _ := crypto.Encrypt("sk-mock-secret-access")
	_, _ = e.DB.Exec("INSERT OR REPLACE INTO credentials (id, account_id, encrypted_access) VALUES ('cred_mock', 'acc_mock', ?)", encAccess)

	if e.App.Router != nil {
		_ = e.App.Router.LoadFromDB(context.Background(), e.DB)
	}
}

func (e *TestEnv) SeedLiveMockRoute(t *testing.T, modelName, mockBaseURL string) {
	provID := "prov_" + modelName
	accID := "acc_" + modelName
	modID := "mod_" + modelName
	rtID := "rt_" + modelName
	riID := "ri_" + modelName
	credID := "cred_" + modelName

	_, _ = e.DB.Exec("INSERT INTO providers (id, key, name, kind, base_url, enabled) VALUES (?, ?, 'Live Mock', 'openai', ?, 1) ON CONFLICT(id) DO UPDATE SET base_url = excluded.base_url, enabled = 1", provID, modelName, mockBaseURL)
	_, _ = e.DB.Exec("INSERT INTO accounts (id, provider_id, name, auth_type, state, priority, enabled) VALUES (?, ?, 'Live Mock Account', 'api_key', 'active', 1, 1) ON CONFLICT(id) DO UPDATE SET enabled = 1", accID, provID)
	_, _ = e.DB.Exec("INSERT INTO models (id, provider_id, external_name, display_name, enabled) VALUES (?, ?, ?, ?, 1) ON CONFLICT(id) DO UPDATE SET enabled = 1", modID, provID, modelName, modelName)
	_, _ = e.DB.Exec("INSERT INTO routes (id, name, strategy, enabled) VALUES (?, ?, 'priority', 1) ON CONFLICT(id) DO UPDATE SET name = excluded.name, enabled = 1", rtID, modelName)
	_, _ = e.DB.Exec("INSERT INTO route_items (id, route_id, provider_id, account_id, model_id, priority, enabled) VALUES (?, ?, ?, ?, ?, 1, 1) ON CONFLICT(id) DO UPDATE SET enabled = 1", riID, rtID, provID, accID, modID)

	crypto, _ := auth.NewCryptoService("very-strong-secret-key-32-chars-long")
	encAccess, _ := crypto.Encrypt("sk-mock-secret-access")
	_, _ = e.DB.Exec("INSERT INTO credentials (id, account_id, encrypted_access) VALUES (?, ?, ?) ON CONFLICT(account_id) DO UPDATE SET encrypted_access = excluded.encrypted_access", credID, accID, encAccess)

	if e.App.Router != nil {
		_ = e.App.Router.LoadFromDB(context.Background(), e.DB)
	}
}
