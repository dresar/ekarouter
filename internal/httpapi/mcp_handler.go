package httpapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type MCPHandler struct {
	db       *sql.DB
	gw       *gateway.Gateway
	ts       *tokensaver.TokenSaver
	router   *routing.Router
	cfg      *config.Config
	mu       sync.RWMutex
	sessions map[string]chan []byte
}

func NewMCPHandler(
	db *sql.DB,
	gw *gateway.Gateway,
	ts *tokensaver.TokenSaver,
	router *routing.Router,
	cfg *config.Config,
) *MCPHandler {
	return &MCPHandler{
		db:       db,
		gw:       gw,
		ts:       ts,
		router:   router,
		cfg:      cfg,
		sessions: make(map[string]chan []byte),
	}
}

func (h *MCPHandler) HandleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &JSONRPCError{Code: -32700, Message: "Parse error"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := h.dispatch(r.Context(), &req)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *MCPHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionBytes := make([]byte, 16)
	_, _ = rand.Read(sessionBytes)
	sessionID := hex.EncodeToString(sessionBytes)

	msgChan := make(chan []byte, 32)
	h.mu.Lock()
	h.sessions[sessionID] = msgChan
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
		close(msgChan)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	endpointURL := fmt.Sprintf("/mcp/messages?sessionId=%s", sessionID)
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURL)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(msg))
			flusher.Flush()
		}
	}
}

func (h *MCPHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
		return
	}

	resp := h.dispatch(r.Context(), &req)
	respBytes, _ := json.Marshal(resp)

	if sessionID != "" {
		h.mu.RLock()
		ch, exists := h.sessions[sessionID]
		h.mu.RUnlock()
		if exists {
			select {
			case ch <- respBytes:
			default:
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}

func (h *MCPHandler) dispatch(ctx context.Context, req *JSONRPCRequest) JSONRPCResponse {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "ekarouter-mcp",
				"version": "1.0.0",
			},
		}

	case "notifications/initialized":
		resp.Result = map[string]any{}

	case "ping":
		resp.Result = map[string]any{}

	case "tools/list":
		resp.Result = map[string]any{
			"tools": h.listTools(),
		}

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &JSONRPCError{Code: -32602, Message: "Invalid params for tools/call"}
			return resp
		}
		result, isErr := h.callTool(ctx, params.Name, params.Arguments)
		resp.Result = map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": result,
				},
			},
			"isError": isErr,
		}

	case "resources/list":
		resp.Result = map[string]any{
			"resources": []map[string]any{
				{
					"uri":         "ekarouter://docs/official",
					"name":        "Official Architecture & Technical Manual",
					"description": "Comprehensive guide to EkaRouter architecture, fallbacks, and API endpoints",
					"mimeType":    "text/markdown",
				},
				{
					"uri":         "ekarouter://docs/skills",
					"name":        "AI Agent Skills Suite",
					"description": "Official agent skills including ui-ux-text and ekarouter gateway skills",
					"mimeType":    "text/markdown",
				},
				{
					"uri":         "ekarouter://config/mcp",
					"name":        "MCP Client Configuration",
					"description": "Ready-to-use configuration for Claude Desktop and Cursor",
					"mimeType":    "application/json",
				},
			},
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		_ = json.Unmarshal(req.Params, &params)
		content := ""
		mimeType := "text/plain"
		switch params.URI {
		case "ekarouter://docs/official":
			content = ProjectArchitectureDoc
			mimeType = "text/markdown"
		case "ekarouter://docs/skills":
			content = MasterAIPromptDoc
			mimeType = "text/markdown"
		case "ekarouter://config/mcp":
			content = `{\n  "mcpServers": {\n    "ekarouter": {\n      "url": "http://localhost:8080/mcp/sse"\n    }\n  }\n}`
			mimeType = "application/json"
		default:
			resp.Error = &JSONRPCError{Code: -32602, Message: fmt.Sprintf("Unknown resource URI: %s", params.URI)}
			return resp
		}
		resp.Result = map[string]any{
			"contents": []map[string]any{
				{
					"uri":      params.URI,
					"mimeType": mimeType,
					"text":     content,
				},
			},
		}

	case "prompts/list":
		resp.Result = map[string]any{
			"prompts": []map[string]any{
				{
					"name":        "ekarouter-master",
					"description": "Master operational prompt for AI coding assistants using EkaRouter",
				},
				{
					"name":        "tokensaver-compact",
					"description": "Compaction prompt for optimizing tool outputs and large context diffs",
				},
			},
		}

	case "prompts/get":
		var params struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(req.Params, &params)
		promptText := MasterAIPromptDoc
		if params.Name == "tokensaver-compact" {
			promptText = "Compress the following text aggressively while preserving functional meaning, removing boilerplate, whitespace, and repetitive lines."
		}
		resp.Result = map[string]any{
			"description": params.Name,
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": promptText,
					},
				},
			},
		}

	default:
		resp.Error = &JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return resp
}

func (h *MCPHandler) listTools() []map[string]any {
	return []map[string]any{
		{
			"name":        "get_project_docs",
			"description": "Retrieve comprehensive official documentation, architecture, and API endpoints for EkaRouter.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"section": map[string]any{
						"type":        "string",
						"description": "Optional doc section (all, architecture, prompt, endpoints, skills)",
					},
				},
			},
		},
		{
			"name":        "get_ai_prompt",
			"description": "Retrieve recommended master prompt and system instructions for AI coding agents to operate EkaRouter.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_type": map[string]any{
						"type":        "string",
						"description": "Target agent type (cursor, claude-code, antigravity, opencode, generic)",
					},
				},
			},
		},
		{
			"name":        "list_available_models",
			"description": "List all active AI models, routing combos, and fallback providers configured in EkaRouter.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "get_api_keys",
			"description": "Get or generate active EkaRouter API keys and gateway integration snippets for external AI tools.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"auto_create": map[string]any{
						"type":        "boolean",
						"description": "If true and no active key exists, automatically creates a new gateway API key",
					},
				},
			},
		},
		{
			"name":        "chat_completion",
			"description": "Execute an OpenAI-compatible chat completion directly through EkaRouter across configured fallback providers.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"model": map[string]any{
						"type":        "string",
						"description": "Model name or routing combo name (e.g., fast, smart, code, gpt-4o, claude-3-7-sonnet)",
					},
					"messages": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"role":    map[string]any{"type": "string"},
								"content": map[string]any{"type": "string"},
							},
							"required": []string{"role", "content"},
						},
						"description": "Array of role and content message objects",
					},
					"temperature": map[string]any{
						"type":        "number",
						"description": "Sampling temperature between 0.0 and 1.0",
					},
					"max_tokens": map[string]any{
						"type":        "integer",
						"description": "Max tokens in response",
					},
				},
				"required": []string{"model", "messages"},
			},
		},
		{
			"name":        "check_gateway_health",
			"description": "Check gateway health status, active counts, database connectivity, and uptime stats.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "compact_tokens",
			"description": "Compress text or prompt payload using EkaRouter TokenSaver engine to reduce token usage.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": "The raw text or prompt content to compress",
					},
					"mode": map[string]any{
						"type":        "string",
						"enum":        []string{"conservative", "aggressive"},
						"description": "Compression aggressiveness level",
					},
				},
				"required": []string{"text"},
			},
		},
	}
}

func (h *MCPHandler) callTool(ctx context.Context, name string, args json.RawMessage) (string, bool) {
	switch name {
	case "get_project_docs":
		var params struct {
			Section string `json:"section"`
		}
		_ = json.Unmarshal(args, &params)
		if params.Section == "prompt" {
			return MasterAIPromptDoc, false
		}
		return ProjectArchitectureDoc, false

	case "get_ai_prompt":
		var params struct {
			AgentType string `json:"agent_type"`
		}
		_ = json.Unmarshal(args, &params)
		if params.AgentType != "" {
			return fmt.Sprintf("# Target Agent: %s\n\n%s", strings.ToUpper(params.AgentType), MasterAIPromptDoc), false
		}
		return MasterAIPromptDoc, false

	case "list_available_models":
		rows, err := h.db.QueryContext(ctx, `
			SELECT 'model' as kind, external_name, display_name FROM models WHERE enabled = 1
			UNION
			SELECT 'combo' as kind, name, strategy FROM routes WHERE enabled = 1
		`)
		if err != nil {
			return fmt.Sprintf("Error querying models: %v", err), true
		}
		defer rows.Close()

		type Item struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
			Info string `json:"info"`
		}
		var items []Item
		for rows.Next() {
			var it Item
			if err := rows.Scan(&it.Kind, &it.Name, &it.Info); err == nil {
				items = append(items, it)
			}
		}
		resBytes, _ := json.MarshalIndent(items, "", "  ")
		return string(resBytes), false

	case "get_api_keys":
		var params struct {
			AutoCreate bool `json:"auto_create"`
		}
		_ = json.Unmarshal(args, &params)

		rows, err := h.db.QueryContext(ctx, "SELECT id, name, prefix, scopes, enabled, created_at FROM api_keys WHERE enabled = 1")
		if err != nil {
			return fmt.Sprintf("Database query failed: %v", err), true
		}
		defer rows.Close()

		type KeyRow struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Prefix    string `json:"prefix"`
			Scopes    string `json:"scopes"`
			CreatedAt string `json:"created_at"`
		}
		var keys []KeyRow
		for rows.Next() {
			var k KeyRow
			var enabledInt int
			if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scopes, &enabledInt, &k.CreatedAt); err == nil {
				keys = append(keys, k)
			}
		}

		if len(keys) == 0 && params.AutoCreate {
			rawKey, prefix, hash, genErr := auth.GenerateApiKey()
			if genErr != nil {
				return fmt.Sprintf("Failed to generate API key: %v", genErr), true
			}
			id := "key_" + prefix[9:]
			_, insErr := h.db.ExecContext(ctx, "INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES (?, ?, ?, ?, ?, ?)",
				id, "MCP Generated Key", prefix, hash, "*", 1)
			if insErr != nil {
				return fmt.Sprintf("Failed to store generated API key: %v", insErr), true
			}
			return fmt.Sprintf("Generated new gateway key successfully:\nKey: %s\nPrefix: %s\nScope: *\nUse with Authorization: Bearer %s", rawKey, prefix, rawKey), false
		}

		resBytes, _ := json.MarshalIndent(map[string]any{
			"active_keys": keys,
			"hint":        "Use full API key with Authorization: Bearer <key> against /v1 endpoints.",
		}, "", "  ")
		return string(resBytes), false

	case "chat_completion":
		var params struct {
			Model       string              `json:"model"`
			Messages    []providers.Message `json:"messages"`
			Temperature *float64            `json:"temperature,omitempty"`
			MaxTokens   *int                `json:"max_tokens,omitempty"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return fmt.Sprintf("Invalid chat parameters: %v", err), true
		}
		if params.Model == "" || len(params.Messages) == 0 {
			return "Model and messages are required", true
		}

		req := &providers.Request{
			ID:          fmt.Sprintf("mcp-%d", time.Now().UnixNano()),
			Model:       params.Model,
			Messages:    params.Messages,
			Temperature: params.Temperature,
			MaxTokens:   params.MaxTokens,
		}

		resp, err := h.gw.Execute(ctx, req)
		if err != nil {
			return fmt.Sprintf("Chat completion failed: %v", err), true
		}

		return resp.Content, false

	case "check_gateway_health":
		var modelCount, providerCount, routeCount int
		_ = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM models WHERE enabled = 1").Scan(&modelCount)
		_ = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM providers WHERE enabled = 1").Scan(&providerCount)
		_ = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM routes WHERE enabled = 1").Scan(&routeCount)

		info := map[string]any{
			"status":          "healthy",
			"database":        "connected",
			"active_models":   modelCount,
			"active_provider": providerCount,
			"active_routes":   routeCount,
			"mcp_support":     "v2024-11-05",
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		}
		resBytes, _ := json.MarshalIndent(info, "", "  ")
		return string(resBytes), false

	case "compact_tokens":
		var params struct {
			Text string `json:"text"`
			Mode string `json:"mode"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return fmt.Sprintf("Invalid parameters: %v", err), true
		}
		if params.Text == "" {
			return "Text cannot be empty", true
		}

		compacted := h.ts.Compact(params.Text)
		origLen := len(params.Text)
		compLen := len(compacted)
		savedBytes := origLen - compLen
		savedPct := 0.0
		if origLen > 0 {
			savedPct = float64(savedBytes) / float64(origLen) * 100.0
		}

		resBytes, _ := json.MarshalIndent(map[string]any{
			"original_bytes":  origLen,
			"compacted_bytes": compLen,
			"saved_bytes":     savedBytes,
			"saved_percent":   fmt.Sprintf("%.1f%%", savedPct),
			"compacted_text":  compacted,
		}, "", "  ")
		return string(resBytes), false

	default:
		return fmt.Sprintf("Tool %s not found", name), true
	}
}
