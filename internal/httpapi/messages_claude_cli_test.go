package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	_ "modernc.org/sqlite"
)

type claudeMockAdapter struct {
	lastReq *providers.Request
}

func (m *claudeMockAdapter) Kind() string {
	return "mock"
}

func (m *claudeMockAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return nil, nil
}

func (m *claudeMockAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	m.lastReq = req

	if len(req.Tools) > 0 {
		return &providers.Response{
			ID:           "msg_tool_call_123",
			Model:        req.Model,
			Role:         "assistant",
			Content:      "I will run the command to check the files.",
			FinishReason: "tool_use",
			ToolCalls: []providers.ToolCall{
				{
					ID:   "toolu_bash_01",
					Type: "function",
					Function: providers.FunctionCall{
						Name:      "Bash",
						Arguments: `{"command":"ls -la"}`,
					},
				},
			},
			Usage: providers.Usage{
				PromptTokens:     25,
				CompletionTokens: 15,
				TotalTokens:      40,
			},
		}, nil
	}

	return &providers.Response{
		ID:           "msg_text_123",
		Model:        req.Model,
		Role:         "assistant",
		Content:      "Hello from Claude Code compatibility test!",
		FinishReason: "end_turn",
		Usage: providers.Usage{
			PromptTokens:     10,
			CompletionTokens: 12,
			TotalTokens:      22,
		},
	}, nil
}

func (m *claudeMockAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	ch := make(chan providers.StreamEvent, 2)
	ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "Hello Claude"}
	ch <- providers.StreamEvent{Type: providers.StreamEventDone}
	close(ch)
	return ch, nil
}

func TestClaudeCodeCLIMessagesEndpoint(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	schema := `
	CREATE TABLE api_keys (
		id TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		name TEXT,
		user_id TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	);
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		revoked_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	testAPIKey := "ek-claude-code-test-key"
	keyHash := auth.HashToken(testAPIKey)
	if _, err := db.Exec("INSERT INTO api_keys (id, hash, enabled) VALUES (?, ?, 1)", "key_1", keyHash); err != nil {
		t.Fatalf("failed to insert api key: %v", err)
	}

	mockAdp := &claudeMockAdapter{}
	reg := providers.NewRegistry()
	reg.Register("mock", mockAdp)

	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "prov-1", State: "active", Enabled: true})
	router.SetRoute(&routing.Route{
		Name:     "claude-3-7-sonnet",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "prov-1", ProviderKind: "mock", AccountID: "acc-1", ModelName: "claude-3-7-sonnet", Priority: 10, Enabled: true},
		},
	})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "upstream-key"}, nil
	}

	ts := tokensaver.New("safe")
	usageRec := usage.NewRecorder(db, 100)
	gw := gateway.NewGateway(router, reg, ts, cd, usageRec, credResolver)
	gwHandler := NewGatewayHandler(gw, db)

	t.Run("Auth via x-api-key header and Text Response", func(t *testing.T) {
		reqBody := `{
			"model": "claude-3-7-sonnet",
			"max_tokens": 1024,
			"system": [
				{"type": "text", "text": "You are Claude Code CLI assistant."}
			],
			"messages": [
				{
					"role": "user",
					"content": [
						{"type": "text", "text": "Hello assistant"}
					]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", testAPIKey)

		middleware := GatewayAuthMiddleware(db)
		rec := httptest.NewRecorder()
		middleware(http.HandlerFunc(gwHandler.Messages)).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var res map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}

		if res["type"] != "message" || res["role"] != "assistant" {
			t.Errorf("unexpected message response format: %v", res)
		}
		if res["stop_reason"] != "end_turn" {
			t.Errorf("expected stop_reason 'end_turn', got %v", res["stop_reason"])
		}

		content, ok := res["content"].([]any)
		if !ok || len(content) == 0 {
			t.Fatalf("expected non-empty content blocks, got: %v", res["content"])
		}
		firstBlock := content[0].(map[string]any)
		if firstBlock["type"] != "text" || firstBlock["text"] != "Hello from Claude Code compatibility test!" {
			t.Errorf("unexpected content block: %v", firstBlock)
		}
	})

	t.Run("Anthropic Tool Calling with Bash Tool", func(t *testing.T) {
		reqBody := `{
			"model": "claude-3-7-sonnet",
			"max_tokens": 2048,
			"tools": [
				{
					"name": "Bash",
					"description": "Execute bash command in terminal",
					"input_schema": {
						"type": "object",
						"properties": {
							"command": {"type": "string"}
						},
						"required": ["command"]
					}
				}
			],
			"messages": [
				{
					"role": "user",
					"content": "List files in the project directory"
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", testAPIKey)

		middleware := GatewayAuthMiddleware(db)
		rec := httptest.NewRecorder()
		middleware(http.HandlerFunc(gwHandler.Messages)).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var res map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}

		if res["stop_reason"] != "tool_use" {
			t.Errorf("expected stop_reason 'tool_use', got %v", res["stop_reason"])
		}

		content, ok := res["content"].([]any)
		if !ok || len(content) < 2 {
			t.Fatalf("expected at least 2 blocks (text + tool_use), got: %v", res["content"])
		}

		toolBlock := content[1].(map[string]any)
		if toolBlock["type"] != "tool_use" {
			t.Errorf("expected block type 'tool_use', got %v", toolBlock["type"])
		}
		if toolBlock["name"] != "Bash" {
			t.Errorf("expected tool name 'Bash', got %v", toolBlock["name"])
		}
		inputMap, ok := toolBlock["input"].(map[string]any)
		if !ok || inputMap["command"] != "ls -la" {
			t.Errorf("expected tool input command 'ls -la', got %v", toolBlock["input"])
		}
	})
}
