package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/routing"
)

type AIDocsHandler struct {
	db     *sql.DB
	cfg    *config.Config
	router *routing.Router
}

func NewAIDocsHandler(db *sql.DB, cfg *config.Config, router *routing.Router) *AIDocsHandler {
	return &AIDocsHandler{
		db:     db,
		cfg:    cfg,
		router: router,
	}
}

const ProjectArchitectureDoc = `# EkaRouter — Official Technical Manual & Architecture

EkaRouter is an enterprise AI Model Gateway, Intelligent Router, and Token Optimizer in Go 1.24 with SQLite and Chi router.

## 1. System Overview & Core Subsystems

1. **Ingress AI Gateway (/v1)**
   - OpenAI-compatible chat completions: POST /v1/chat/completions
   - Model catalog discovery: GET /v1/models
   - Server-Sent Events (SSE) streaming and tool calls across all upstream providers.
   - Built-in token compaction pipeline via TokenSaver.

2. **Smart Router & Dynamic Fallback Engine**
   - Combos & Priority Chains: routes requests through weighted, priority, or latency-optimized lists of provider accounts.
   - Automatic Failover: switches instantly to backup providers upon HTTP 429 (rate-limit) or 5xx errors.
   - Adaptive Cooldowns: temporarily isolates failing providers without breaking caller sessions.

3. **TokenSaver Engine**
   - Context compaction engine reducing billable token usage by 20% to 45%.
   - Strips redundant whitespace, system boilerplate, repetitive diffs, and formatting artifacts safely.
   - Controlled via header: X-Token-Saver: conservative | aggressive | off

4. **Credential Vault & Credential Pools**
   - AES-256-GCM encrypted storage for upstream API keys and bearer tokens.
   - Credential Pools with Round-Robin, Least-Used, and Weighted selection policies.
   - Health checking and automatic key retirement.

5. **Outbound Proxy Egress**
   - HTTP, HTTPS, and SOCKS5 proxy profiles assigned per provider or account.
   - Bypasses IP-based geographic or regional rate limits.

6. **Model Context Protocol (MCP) Server (/mcp & /mcp/sse)**
   - Exposes tools, resources, and prompts to external AI assistants (Cursor, Claude Desktop, Antigravity, OpenCode).
   - Allows external AI agents to query models, fetch active API keys, and route completions.

7. **Web Console & Dashboard (/overview, /boost, /learn)**
   - Clean React 19 + Tailwind CSS single-page management console.
   - Real-time observability, playground, audit logs, and developer documentation.
`

const MasterAIPromptDoc = `# EkaRouter AI Agent Master Prompt & Execution Context

Use this master prompt when configuring AI coding agents (Claude Code, Cursor, Antigravity, OpenCode, ChatGPT, Windsurf):

---
You are an autonomous AI engineer connected to an active instance of EkaRouter (Enterprise AI Gateway & Smart Model Router).

## Connection Details
- Gateway Base URL: http://localhost:8080 (or process port)
- OpenAI-compatible Endpoint: http://localhost:8080/v1/chat/completions
- Model Discovery: http://localhost:8080/v1/models
- MCP Server Endpoint: http://localhost:8080/mcp (or http://localhost:8080/mcp/sse)
- Official Documentation: http://localhost:8080/api/ai/docs

## Operational Guidelines
1. Model Selection: Always query GET /v1/models or use configured combo routes (e.g., fast, smart, code). EkaRouter handles automatic upstream failover across OpenAI, Anthropic, Gemini, Groq, Cerebras, and Ollama.
2. Token Economy: Keep X-Token-Saver: conservative active for large contexts and code diffs to automatically reduce billable tokens.
3. MCP Tool Usage: If connecting via MCP, use list_available_models to view models, get_api_keys to obtain gateway access, and get_project_docs to read API specs.
4. UI/UX Text Standard: When modifying frontend code, strictly adhere to the ui-ux-text skill:
   - Form Placeholders: Strictly 1 word.
   - Buttons & Action Labels: Max 1-2 words.
   - Titles & Headings: Short hierarchy (1-3 words).
   - Subtitles: Max 1 short sentence.
   - Error messages & toasts: Concise (2-3 words).
   - Zero repetition and progressive disclosure via [i] tooltips.
---
`

func (h *AIDocsHandler) GetDocs(w http.ResponseWriter, r *http.Request) {
	section := strings.ToLower(r.URL.Query().Get("section"))
	format := strings.ToLower(r.URL.Query().Get("format"))

	doc := ProjectArchitectureDoc
	if section == "prompt" {
		doc = MasterAIPromptDoc
	}

	if format == "json" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"project":     "EkaRouter",
			"version":     "1.0.0",
			"description": "Enterprise-grade AI Model Router and Gateway",
			"section":     section,
			"content":     doc,
			"endpoints": map[string]string{
				"chat_completions": "/v1/chat/completions",
				"models":           "/v1/models",
				"ai_docs":          "/api/ai/docs",
				"ai_prompt":        "/api/ai/prompt",
				"ai_status":        "/api/ai/status",
				"ai_skills":        "/api/ai/skills",
				"mcp_jsonrpc":      "/mcp",
				"mcp_sse":          "/mcp/sse",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write([]byte(doc))
}

func (h *AIDocsHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	agent := strings.ToLower(r.URL.Query().Get("agent"))
	format := strings.ToLower(r.URL.Query().Get("format"))

	prompt := MasterAIPromptDoc
	if agent != "" {
		prompt = fmt.Sprintf("# Target Agent: %s\n\n%s", strings.ToUpper(agent), MasterAIPromptDoc)
	}

	if format == "json" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"agent":  agent,
			"prompt": prompt,
			"role":   "system",
		})
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write([]byte(prompt))
}

func (h *AIDocsHandler) GetSkills(w http.ResponseWriter, r *http.Request) {
	skills := []map[string]any{
		{
			"name":        "ui-ux-text",
			"description": "Standar wajib UI/UX microcopy dan hierarchy visual: super singkat, konsisten, rapi, mudah dipindai, responsive, accessible, dan zero text bloat.",
			"location":    "skills/ui-ux-text/SKILL.md",
			"rules": []string{
				"Placeholder: strictly 1 word",
				"Button: max 1-2 words",
				"Title: 1-3 words",
				"Subtitle: max 1 short sentence",
				"Helper: max 1 short sentence",
				"Toast: max 2-3 words",
				"Zero UX filler",
			},
		},
		{
			"name":        "ekarouter",
			"description": "Gateway entry point, environment setup, model discovery, and health checks",
			"location":    "skills/ekarouter/SKILL.md",
		},
		{
			"name":        "ekarouter-chat",
			"description": "OpenAI-compatible chat completions, streaming SSE, tool calls, and combos",
			"location":    "skills/ekarouter-chat/SKILL.md",
		},
		{
			"name":        "ekarouter-admin",
			"description": "Management API: keys, accounts, providers, combos, proxy profiles, and backups",
			"location":    "skills/ekarouter-admin/SKILL.md",
		},
		{
			"name":        "ekarouter-tokensaver",
			"description": "Safe context compaction for git diffs, repeated lines, and tool outputs",
			"location":    "skills/ekarouter-tokensaver/SKILL.md",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"skills": skills,
		"count":  len(skills),
	})
}

func (h *AIDocsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	var modelCount int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM models WHERE enabled = 1").Scan(&modelCount)

	var providerCount int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM providers WHERE enabled = 1").Scan(&providerCount)

	var routeCount int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM routes WHERE enabled = 1").Scan(&routeCount)

	var keyCount int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM api_keys WHERE enabled = 1").Scan(&keyCount)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":          "healthy",
		"version":         "1.0.0",
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"models_count":    modelCount,
		"providers_count": providerCount,
		"routes_count":    routeCount,
		"keys_count":      keyCount,
		"features": map[string]bool{
			"chat_completions": true,
			"tokensaver":       true,
			"mcp":              true,
			"vault":            true,
			"credential_pools": true,
			"proxies":          true,
		},
	})
}

func (h *AIDocsHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT id, name, prefix, scopes, enabled, created_at, last_used_at FROM api_keys WHERE enabled = 1")
	if err != nil {
		http.Error(w, `{"error":"failed to query api keys"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type KeySummary struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Prefix     string  `json:"prefix"`
		Scopes     string  `json:"scopes"`
		CreatedAt  string  `json:"created_at"`
		LastUsedAt *string `json:"last_used_at,omitempty"`
	}

	var keys []KeySummary
	for rows.Next() {
		var k KeySummary
		var enabledInt int
		var lastUsed sql.NullString
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scopes, &enabledInt, &k.CreatedAt, &lastUsed); err == nil {
			if lastUsed.Valid {
				k.LastUsedAt = &lastUsed.String
			}
			keys = append(keys, k)
		}
	}
	if keys == nil {
		keys = make([]KeySummary, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"keys":  keys,
		"total": len(keys),
		"integration_snippets": map[string]string{
			"curl": "curl -X POST http://localhost:8080/v1/chat/completions -H \"Authorization: Bearer <YOUR_API_KEY>\" -H \"Content-Type: application/json\" -d '{\"model\": \"fast\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello EkaRouter\"}]}'",
			"python": "from openai import OpenAI\n\nclient = OpenAI(base_url=\"http://localhost:8080/v1\", api_key=\"<YOUR_API_KEY>\")\nresponse = client.chat.completions.create(model=\"fast\", messages=[{\"role\": \"user\", \"content\": \"Hello EkaRouter\"}])\nprint(response.choices[0].message.content)",
			"mcp_config": "{\n  \"mcpServers\": {\n    \"ekarouter\": {\n      \"url\": \"http://localhost:8080/mcp/sse\"\n    }\n  }\n}",
		},
	})
}

func (h *AIDocsHandler) ServeHTMLDocs(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>EkaRouter Documentation</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; max-width: 860px; margin: 40px auto; padding: 0 20px; color: #1e293b; background: #f8fafc; }
    pre { background: #0f172a; color: #f1f5f9; padding: 16px; border-radius: 8px; overflow-x: auto; font-size: 13px; }
    code { font-family: monospace; background: #e2e8f0; padding: 2px 6px; border-radius: 4px; font-size: 13px; }
    pre code { background: none; padding: 0; }
    h1, h2, h3 { color: #0f172a; }
    a { color: #2563eb; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .badge { display: inline-block; padding: 4px 10px; border-radius: 9999px; background: #dbeafe; color: #1e40af; font-size: 12px; font-weight: 600; margin-bottom: 12px; }
    .nav { display: flex; gap: 12px; margin-bottom: 24px; }
    .nav a { background: #ffffff; border: 1px solid #cbd5e1; padding: 6px 14px; border-radius: 6px; font-size: 13px; font-weight: 500; }
  </style>
</head>
<body>
  <div class="badge">EkaRouter v1.0.0</div>
  <div class="nav">
    <a href="/api/ai/docs?format=json">API JSON</a>
    <a href="/api/ai/prompt">Master AI Prompt</a>
    <a href="/api/ai/skills">Skills Registry</a>
    <a href="/mcp">MCP Server</a>
  </div>
  <div id="content">
    <pre>%s</pre>
  </div>
</body>
</html>`, ProjectArchitectureDoc)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
