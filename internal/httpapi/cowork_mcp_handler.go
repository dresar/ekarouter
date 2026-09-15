package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type CoworkMCPHandler struct {
	client  *http.Client
	cacheMu sync.RWMutex
	cached  []MCPServerInfo
	lastGen time.Time
}

type MCPServerInfo struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Transport   string   `json:"transport"`
	OAuth       bool     `json:"oauth"`
	ToolNames   []string `json:"toolNames"`
	ToolCount   int      `json:"toolCount"`
	IconURL     string   `json:"iconUrl,omitempty"`
}

func NewCoworkMCPHandler() *CoworkMCPHandler {
	return &CoworkMCPHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func defaultCuratedMCPServers() []MCPServerInfo {
	return []MCPServerInfo{
		{
			Name:        "exa",
			Title:       "Exa",
			Description: "Real-time web search and code documentation",
			URL:         "https://mcp.exa.ai/mcp",
			Transport:   "http",
			OAuth:       false,
			ToolNames:   []string{"web_search_exa", "web_fetch_exa"},
			ToolCount:   2,
			IconURL:     "https://exa.ai/favicon.ico",
		},
		{
			Name:        "tavily",
			Title:       "Tavily",
			Description: "Real-time web search optimized for LLM agents",
			URL:         "https://mcp.tavily.com/mcp",
			Transport:   "http",
			OAuth:       true,
			ToolNames:   []string{"tavily_search", "tavily_extract", "tavily_crawl", "tavily_map"},
			ToolCount:   4,
			IconURL:     "https://tavily.com/favicon.ico",
		},
		{
			Name:        "browsermcp",
			Title:       "Browser MCP",
			Description: "Control running Chrome browser via DevTools Protocol",
			URL:         "http://localhost:8080/mcp/sse",
			Transport:   "sse",
			OAuth:       false,
			ToolNames:   []string{"browser_navigate", "browser_snapshot", "browser_click", "browser_type", "browser_screenshot"},
			ToolCount:   5,
		},
		{
			Name:        "filesystem",
			Title:       "Local Filesystem",
			Description: "Secure file access, search, and directory operations in the workspace",
			URL:         "http://localhost:8080/mcp/sse",
			Transport:   "sse",
			OAuth:       false,
			ToolNames:   []string{"read_file", "write_file", "list_directory", "search_files"},
			ToolCount:   4,
		},
		{
			Name:        "github",
			Title:       "GitHub MCP",
			Description: "Interact with GitHub repositories, issues, PRs, and code trees",
			URL:         "https://api.githubcopilot.com/mcp",
			Transport:   "http",
			OAuth:       true,
			ToolNames:   []string{"get_file_contents", "create_or_update_file", "push_files", "search_repositories"},
			ToolCount:   4,
			IconURL:     "https://github.githubassets.com/favicons/favicon.png",
		},
		{
			Name:        "neon",
			Title:       "Neon PostgreSQL MCP",
			Description: "Serverless Postgres database management, queries, schemas, and branching",
			URL:         "http://localhost:8080/mcp/sse",
			Transport:   "sse",
			OAuth:       false,
			ToolNames:   []string{"run_sql", "describe_table_schema", "get_database_tables", "list_projects"},
			ToolCount:   4,
			IconURL:     "https://neon.tech/favicon.ico",
		},
		{
			Name:        "ekarouter",
			Title:       "EkaRouter Gateway MCP",
			Description: "Dynamic AI multi-model routing, prompt caching, failover, and token saving",
			URL:         "http://localhost:8080/mcp/sse",
			Transport:   "sse",
			OAuth:       false,
			ToolNames:   []string{"list_available_models", "get_api_keys", "get_project_docs", "test_model_chat"},
			ToolCount:   4,
		},
	}
}

func (h *CoworkMCPHandler) HandleRegistry(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	h.cacheMu.RLock()
	servers := h.cached
	cacheValid := time.Since(h.lastGen) < 30*time.Minute && len(servers) > 0
	h.cacheMu.RUnlock()

	if !cacheValid {
		servers = defaultCuratedMCPServers()
		go h.refreshExternalRegistry()
	}

	var filtered []MCPServerInfo
	for _, s := range servers {
		if q == "" || strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Title), q) ||
			strings.Contains(strings.ToLower(s.Description), q) {
			filtered = append(filtered, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"servers": filtered,
		"count":   len(filtered),
	})
}

func (h *CoworkMCPHandler) refreshExternalRegistry() {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.anthropic.com/mcp-registry/v0/servers?limit=100", nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	var payload struct {
		Servers []struct {
			Server struct {
				Name        string `json:"name"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Remotes     []struct {
					Type string `json:"type"`
					URL  string `json:"url"`
				} `json:"remotes"`
			} `json:"server"`
			Meta map[string]any `json:"_meta"`
		} `json:"servers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return
	}

	results := defaultCuratedMCPServers()
	seen := make(map[string]bool)
	for _, s := range results {
		seen[s.URL] = true
	}

	for _, item := range payload.Servers {
		s := item.Server
		if len(s.Remotes) == 0 {
			continue
		}
		remoteURL := s.Remotes[0].URL
		if remoteURL == "" || seen[remoteURL] || !strings.HasPrefix(remoteURL, "https://") {
			continue
		}
		seen[remoteURL] = true
		transport := s.Remotes[0].Type
		if transport == "" {
			transport = "http"
		}
		title := s.Title
		if title == "" {
			title = s.Name
		}
		results = append(results, MCPServerInfo{
			Name:        s.Name,
			Title:       title,
			Description: s.Description,
			URL:         remoteURL,
			Transport:   transport,
			OAuth:       false,
			ToolNames:   []string{},
			ToolCount:   0,
		})
	}

	h.cacheMu.Lock()
	h.cached = results
	h.lastGen = time.Now()
	h.cacheMu.Unlock()
}

func (h *CoworkMCPHandler) HandleTools(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]string{
				"name":    "ekarouter",
				"version": "1.0",
			},
		},
	}
	initBytes, _ := json.Marshal(initReq)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, body.URL, bytes.NewReader(initBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create request: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"probe failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"requiresAuth": true,
			"tools":        []any{},
		})
		return
	}

	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	listBytes, _ := json.Marshal(listReq)

	req2, err := http.NewRequestWithContext(ctx, http.MethodPost, body.URL, bytes.NewReader(listBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create tools/list request: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := h.client.Do(req2)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"tools/list request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp2.Body.Close()

	bodyBytes, _ := io.ReadAll(resp2.Body)
	var rpcResp struct {
		Result struct {
			Tools []any `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(bodyBytes, &rpcResp); err == nil && len(rpcResp.Result.Tools) > 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tools": rpcResp.Result.Tools,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tools": []any{},
	})
}

func (h *CoworkMCPHandler) HandlePluginSSE(w http.ResponseWriter, r *http.Request) {
	plugin := chi.URLParam(r, "plugin")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("event: endpoint\ndata: /api/mcp/%s/message\n\n", plugin)
	_, _ = w.Write([]byte(msg))
	flusher.Flush()

	<-r.Context().Done()
}

func (h *CoworkMCPHandler) HandlePluginMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
		return
	}

	res := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"status": "received",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
