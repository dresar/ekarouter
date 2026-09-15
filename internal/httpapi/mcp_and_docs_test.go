package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS models (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL,
			external_name TEXT NOT NULL,
			display_name TEXT NOT NULL,
			context_limit INTEGER NOT NULL DEFAULT 8192,
			input_capability TEXT NOT NULL DEFAULT 'text',
			output_capability TEXT NOT NULL DEFAULT 'text',
			streaming INTEGER NOT NULL DEFAULT 1,
			enabled INTEGER NOT NULL DEFAULT 1,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS providers (
			id TEXT PRIMARY KEY,
			key TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			base_url TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS routes (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			strategy TEXT NOT NULL DEFAULT 'priority',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			prefix TEXT NOT NULL,
			hash TEXT UNIQUE NOT NULL,
			scopes TEXT NOT NULL DEFAULT '*',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME
		);`,
		`INSERT INTO providers (id, key, name, kind, base_url, enabled) VALUES ('p1', 'openai', 'OpenAI', 'openai', 'https://api.openai.com', 1);`,
		`INSERT INTO models (id, provider_id, external_name, display_name, enabled) VALUES ('m1', 'p1', 'gpt-4o', 'GPT-4o', 1);`,
		`INSERT INTO routes (id, name, strategy, enabled) VALUES ('r1', 'fast', 'priority', 1);`,
		`INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES ('k1', 'Default Key', 'er-test1234', 'hash1234', '*', 1);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("failed to execute schema setup query: %v", err)
		}
	}

	return db
}

func TestAIDocsEndpoints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := &config.Config{SecretKey: "01234567890123456789012345678901"}
	cd := routing.NewCooldownManager()
	rt := routing.NewRouter(cd)
	handler := NewAIDocsHandler(db, cfg, rt)

	t.Run("GetDocsMarkdown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/docs", nil)
		rec := httptest.NewRecorder()
		handler.GetDocs(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte("EkaRouter")) {
			t.Fatalf("expected body to contain EkaRouter")
		}
	})

	t.Run("GetDocsJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/docs?format=json", nil)
		rec := httptest.NewRecorder()
		handler.GetDocs(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if res["project"] != "EkaRouter" {
			t.Fatalf("unexpected project name: %v", res["project"])
		}
	})

	t.Run("GetPrompt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/prompt?agent=cursor&format=json", nil)
		rec := httptest.NewRecorder()
		handler.GetPrompt(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to parse prompt JSON: %v", err)
		}
		if res["agent"] != "cursor" {
			t.Fatalf("expected agent cursor, got %v", res["agent"])
		}
	})

	t.Run("GetSkills", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/skills", nil)
		rec := httptest.NewRecorder()
		handler.GetSkills(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to parse skills: %v", err)
		}
		if res["count"].(float64) < 1 {
			t.Fatalf("expected at least 1 skill")
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/status", nil)
		rec := httptest.NewRecorder()
		handler.GetStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to parse status: %v", err)
		}
		if res["status"] != "healthy" {
			t.Fatalf("expected healthy status, got %v", res["status"])
		}
	})

	t.Run("GetKeys", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/keys", nil)
		rec := httptest.NewRecorder()
		handler.GetKeys(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to parse keys: %v", err)
		}
		if res["total"].(float64) != 1 {
			t.Fatalf("expected 1 key, got %v", res["total"])
		}
	})
}

func TestMCPHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cd := routing.NewCooldownManager()
	rt := routing.NewRouter(cd)
	reg := providers.NewRegistry()
	ts := tokensaver.New("conservative")
	cfg := &config.Config{SecretKey: "01234567890123456789012345678901"}
	gw := gateway.NewGateway(rt, reg, ts, cd, nil, nil)
	handler := NewMCPHandler(db, gw, ts, rt, cfg)

	t.Run("Initialize", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode JSON-RPC: %v", err)
		}
		if res.Error != nil {
			t.Fatalf("unexpected error: %v", res.Error)
		}
	})

	t.Run("ToolsList", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode JSON-RPC: %v", err)
		}
		resultMap, ok := res.Result.(map[string]any)
		if !ok {
			t.Fatalf("expected map result")
		}
		tools, ok := resultMap["tools"].([]any)
		if !ok || len(tools) == 0 {
			t.Fatalf("expected list of tools, got %v", resultMap)
		}
	})

	t.Run("CallToolGetDocs", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_project_docs","arguments":{"section":"all"}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		resultMap := res.Result.(map[string]any)
		content := resultMap["content"].([]any)
		firstContent := content[0].(map[string]any)
		text := firstContent["text"].(string)
		if !strings.Contains(text, "EkaRouter") {
			t.Fatalf("expected documentation content to mention EkaRouter")
		}
	})

	t.Run("CallToolCompactTokens", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"compact_tokens","arguments":{"text":"hello    world\n\n\n\nfoo"}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		resultMap := res.Result.(map[string]any)
		if resultMap["isError"] == true {
			t.Fatalf("compact_tokens returned error: %v", resultMap)
		}
	})

	t.Run("ResourcesList", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":5,"method":"resources/list"}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		_ = json.NewDecoder(rec.Body).Decode(&res)
		resultMap := res.Result.(map[string]any)
		resList := resultMap["resources"].([]any)
		if len(resList) == 0 {
			t.Fatalf("expected non-empty resources list")
		}
	})

	t.Run("GetMCPInfo", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var mcpInfo map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&mcpInfo); err != nil {
			t.Fatalf("failed to decode MCP info: %v", err)
		}
		if mcpInfo["service"] != "EkaRouter MCP Server" {
			t.Fatalf("unexpected service name: %v", mcpInfo["service"])
		}
	})

	t.Run("CallToolGenerateApiKey", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"generate_api_key","arguments":{"name":"Test Subagent"}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		resultMap := res.Result.(map[string]any)
		if resultMap["isError"] == true {
			t.Fatalf("generate_api_key returned error: %v", resultMap)
		}
		content := resultMap["content"].([]any)[0].(map[string]any)["text"].(string)
		if !strings.Contains(content, "eka_live_") && !strings.Contains(content, "secret_key") {
			t.Fatalf("expected generated secret key in response, got: %s", content)
		}
	})

	t.Run("CallToolGetApiKeysCreateNew", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_api_keys","arguments":{"create_new":true,"name":"Cursor Assistant"}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		_ = json.NewDecoder(rec.Body).Decode(&res)
		resultMap := res.Result.(map[string]any)
		if resultMap["isError"] == true {
			t.Fatalf("get_api_keys returned error: %v", resultMap)
		}
		content := resultMap["content"].([]any)[0].(map[string]any)["text"].(string)
		if !strings.Contains(content, "secret_key") {
			t.Fatalf("expected secret_key when create_new=true, got: %s", content)
		}
	})

	t.Run("CallToolListModels", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"list_available_models","arguments":{}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()
		handler.HandleJSONRPC(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res JSONRPCResponse
		_ = json.NewDecoder(rec.Body).Decode(&res)
		resultMap := res.Result.(map[string]any)
		content := resultMap["content"].([]any)[0].(map[string]any)["text"].(string)
		if !strings.Contains(content, "gpt-4o") || !strings.Contains(content, "fast") {
			t.Fatalf("expected model list to include gpt-4o and combo fast, got: %s", content)
		}
	})
}

func TestAIDocsModelsAndKeyGeneration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := &config.Config{SecretKey: "01234567890123456789012345678901"}
	cd := routing.NewCooldownManager()
	rt := routing.NewRouter(cd)
	handler := NewAIDocsHandler(db, cfg, rt)

	t.Run("GetModels", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ai/models", nil)
		rec := httptest.NewRecorder()
		handler.GetModels(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if res["total_models"].(float64) < 1 {
			t.Fatalf("expected at least 1 model, got %v", res["total_models"])
		}
		if res["total_combos"].(float64) < 1 {
			t.Fatalf("expected at least 1 combo, got %v", res["total_combos"])
		}
	})

	t.Run("CreateKey", func(t *testing.T) {
		body := `{"name":"Automated AI Agent"}`
		req := httptest.NewRequest(http.MethodPost, "/api/ai/keys", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		handler.CreateKey(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		var res map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if res["success"] != true {
			t.Fatalf("expected success: true, got %v", res)
		}
		key, ok := res["key"].(string)
		if !ok || !strings.HasPrefix(key, "eka_live_") {
			t.Fatalf("expected valid eka_live_ key, got: %v", key)
		}
	})
}
