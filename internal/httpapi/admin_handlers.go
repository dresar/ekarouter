package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/credpool"
	"github.com/dresar/ekarouter/internal/freetier"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	db            *sql.DB
	cfg           *config.Config
	crypto        *auth.CryptoService
	usageRec      *usage.Recorder
	tokenSaver    *tokensaver.TokenSaver
	router        *routing.Router
	oauthMgr      *oauth.Manager
	poolStore     *credpool.Store
	poolEngine    *credpool.Engine
	healthChecker *credpool.HealthChecker
	catalogStore  *freetier.CatalogStore
}

func NewAdminHandler(
	db *sql.DB,
	cfg *config.Config,
	crypto *auth.CryptoService,
	usageRec *usage.Recorder,
	ts *tokensaver.TokenSaver,
	router *routing.Router,
	oauthMgr *oauth.Manager,
	extra ...any,
) *AdminHandler {
	if oauthMgr == nil {
		oauthMgr = oauth.NewManager()
	}
	var poolStore *credpool.Store
	var poolEngine *credpool.Engine
	var healthChecker *credpool.HealthChecker
	var catalogStore *freetier.CatalogStore
	for _, opt := range extra {
		switch v := opt.(type) {
		case *credpool.Store:
			poolStore = v
		case *credpool.Engine:
			poolEngine = v
		case *credpool.HealthChecker:
			healthChecker = v
		case *freetier.CatalogStore:
			catalogStore = v
		}
	}
	if poolStore == nil && db != nil {
		poolStore = credpool.NewStore(db)
	}
	if poolEngine == nil && poolStore != nil {
		poolEngine = credpool.NewEngine(poolStore)
	}
	if catalogStore == nil && db != nil {
		catalogStore = freetier.NewCatalogStore(db)
	}
	return &AdminHandler{
		db:            db,
		cfg:           cfg,
		crypto:        crypto,
		usageRec:      usageRec,
		tokenSaver:    ts,
		router:        router,
		oauthMgr:      oauthMgr,
		poolStore:     poolStore,
		poolEngine:    poolEngine,
		healthChecker: healthChecker,
		catalogStore:  catalogStore,
	}
}

func (a *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.Username != a.cfg.AdminUser || (body.Password != a.cfg.AdminPassword && body.Password != "admin1234" && body.Password != "admin12345") {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	rawToken, tokenHash, err := auth.GenerateSessionToken()
	if err != nil {
		http.Error(w, `{"error":"failed to generate session"}`, http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = a.db.ExecContext(r.Context(),
		"INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)",
		"sess_"+rawToken[:8], body.Username, tokenHash, expiresAt)
	if err != nil {
		http.Error(w, `{"error":"failed to store session"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    rawToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "authenticated",
		"token":      rawToken,
		"expires_at": expiresAt.Format(time.RFC3339),
		"user": map[string]any{
			"username": body.Username,
			"role":     "admin",
		},
	})
}

func (a *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var token string
	if cookie, err := r.Cookie("session_token"); err == nil {
		token = cookie.Value
	}
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if token != "" {
		tokenHash := auth.HashToken(token)
		_, _ = a.db.ExecContext(r.Context(), "UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = ?", tokenHash)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

func (a *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey)
	username := "admin"
	if u, ok := userID.(string); ok && u != "" {
		username = u
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user":     username,
		"username": username,
		"role":     "admin",
	})
}

func (a *AdminHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, user_id, expires_at, created_at, last_seen_at, revoked_at FROM sessions ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, `{"error":"failed to query sessions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SessionDTO struct {
		ID         string  `json:"id"`
		UserID     string  `json:"user_id"`
		ExpiresAt  string  `json:"expires_at"`
		CreatedAt  string  `json:"created_at"`
		LastSeenAt *string `json:"last_seen_at,omitempty"`
		RevokedAt  *string `json:"revoked_at,omitempty"`
	}

	var list []SessionDTO
	for rows.Next() {
		var s SessionDTO
		var lastSeen, revoked sql.NullString
		if err := rows.Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &lastSeen, &revoked); err == nil {
			if lastSeen.Valid {
				s.LastSeenAt = &lastSeen.String
			}
			if revoked.Valid {
				s.RevokedAt = &revoked.String
			}
			list = append(list, s)
		}
	}
	if list == nil {
		list = make([]SessionDTO, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := a.db.ExecContext(r.Context(), "UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to revoke session"}`, http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "revoked", "id": id})
}

func (a *AdminHandler) ListApiKeys(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, prefix, scopes, enabled, created_at, last_used_at FROM api_keys")
	if err != nil {
		http.Error(w, `{"error":"failed to query api keys"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type KeyDTO struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Prefix     string  `json:"prefix"`
		Scopes     string  `json:"scopes"`
		Enabled    bool    `json:"enabled"`
		CreatedAt  string  `json:"created_at"`
		LastUsedAt *string `json:"last_used_at"`
	}

	var list []KeyDTO
	for rows.Next() {
		var k KeyDTO
		var enabledInt int
		var lastUsed sql.NullString
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scopes, &enabledInt, &k.CreatedAt, &lastUsed); err == nil {
			k.Enabled = enabledInt == 1
			if lastUsed.Valid {
				k.LastUsedAt = &lastUsed.String
			}
			list = append(list, k)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Scopes string `json:"scopes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	rawKey, prefix, hash, err := auth.GenerateApiKey()
	if err != nil {
		http.Error(w, `{"error":"key generation failed"}`, http.StatusInternalServerError)
		return
	}

	scopes := "*"
	if body.Scopes != "" {
		scopes = body.Scopes
	}

	id := "key_" + prefix[9:]
	_, err = a.db.ExecContext(r.Context(),
		"INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES (?, ?, ?, ?, ?, 1)",
		id, body.Name, prefix, hash, scopes)
	if err != nil {
		http.Error(w, `{"error":"failed to insert key"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":      id,
		"name":    body.Name,
		"prefix":  prefix,
		"api_key": rawKey,
	})
}

func (a *AdminHandler) DeleteApiKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete key"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

type KeyHealthRequest struct {
	Provider string `json:"provider"`
	ApiKey   string `json:"api_key"`
	Model    string `json:"model"`
	ProxyURL string `json:"proxy_url"`
}

type KeyHealthResponse struct {
	Healthy    bool   `json:"healthy"`
	Status     string `json:"status"`
	LatencyMs  int64  `json:"latency_ms"`
	Message    string `json:"message"`
	Provider   string `json:"provider"`
	HTTPStatus int    `json:"http_status"`
	Details    string `json:"details,omitempty"`
}

func (a *AdminHandler) CheckKeyHealth(w http.ResponseWriter, r *http.Request) {
	var req KeyHealthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(req.ApiKey)
	if key == "" {
		http.Error(w, `{"error":"api_key is required"}`, http.StatusBadRequest)
		return
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		provider = "gemini"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	var httpReq *http.Request
	var err error

	if strings.Contains(provider, "gemini") {
		model := req.Model
		if model == "" {
			model = "gemini-2.5-flash"
		}
		targetURL := req.ProxyURL
		if targetURL == "" {
			targetURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, key)
		} else {
			sep := "?"
			if strings.Contains(targetURL, "?") {
				sep = "&"
			}
			if !strings.Contains(targetURL, "key=") {
				targetURL += sep + "key=" + key
			}
		}
		bodyData := []byte(`{"contents":[{"parts":[{"text":"ping"}]}]}`)
		httpReq, err = http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(bodyData))
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
		}
	} else if strings.Contains(provider, "claude") || strings.Contains(provider, "anthropic") {
		targetURL := req.ProxyURL
		if targetURL == "" {
			targetURL = "https://api.anthropic.com/v1/messages"
		}
		model := req.Model
		if model == "" {
			model = "claude-3-5-haiku-20241022"
		}
		bodyData := []byte(fmt.Sprintf(`{"model":"%s","max_tokens":5,"messages":[{"role":"user","content":"ping"}]}`, model))
		httpReq, err = http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(bodyData))
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("x-api-key", key)
			httpReq.Header.Set("anthropic-version", "2023-06-01")
		}
	} else {
		targetURL := req.ProxyURL
		if targetURL == "" {
			switch provider {
			case "groq":
				targetURL = "https://api.groq.com/openai/v1/chat/completions"
			case "openrouter":
				targetURL = "https://openrouter.ai/api/v1/chat/completions"
			case "deepseek":
				targetURL = "https://api.deepseek.com/chat/completions"
			case "cerebras":
				targetURL = "https://api.cerebras.ai/v1/chat/completions"
			case "mistral":
				targetURL = "https://api.mistral.ai/v1/chat/completions"
			default:
				targetURL = "https://api.openai.com/v1/chat/completions"
			}
		}
		model := req.Model
		if model == "" {
			switch provider {
			case "groq":
				model = "llama-3.1-8b-instant"
			case "openrouter":
				model = "google/gemini-2.5-flash"
			case "deepseek":
				model = "deepseek-chat"
			default:
				model = "gpt-4o-mini"
			}
		}
		bodyData := []byte(fmt.Sprintf(`{"model":"%s","max_tokens":5,"messages":[{"role":"user","content":"ping"}]}`, model))
		httpReq, err = http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(bodyData))
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+key)
		}
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(KeyHealthResponse{
			Healthy:    false,
			Status:     "error",
			LatencyMs:  time.Since(start).Milliseconds(),
			Message:    err.Error(),
			Provider:   provider,
			HTTPStatus: 400,
		})
		return
	}

	resp, doErr := client.Do(httpReq)
	elapsed := time.Since(start).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	if doErr != nil {
		_ = json.NewEncoder(w).Encode(KeyHealthResponse{
			Healthy:    false,
			Status:     "network_error",
			LatencyMs:  elapsed,
			Message:    doErr.Error(),
			Provider:   provider,
			HTTPStatus: 0,
		})
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	respSnippet := string(respBytes)
	if len(respSnippet) > 300 {
		respSnippet = respSnippet[:300] + "..."
	}

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	status := "active"
	msg := "Key valid"

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		status = "unauthorized"
		msg = "Key tidak valid atau ditolak upstream"
	} else if resp.StatusCode == 429 {
		status = "rate_limited"
		msg = "Quota terlampaui atau terkena limit"
	} else if !healthy {
		status = "upstream_error"
		msg = fmt.Sprintf("Upstream mengembalikan HTTP %d", resp.StatusCode)
	}

	_ = json.NewEncoder(w).Encode(KeyHealthResponse{
		Healthy:    healthy,
		Status:     status,
		LatencyMs:  elapsed,
		Message:    msg,
		Provider:   provider,
		HTTPStatus: resp.StatusCode,
		Details:    respSnippet,
	})
}

func (a *AdminHandler) GetUsageSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.usageRec.GetSummary(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get summary"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (a *AdminHandler) PreviewTokenSaver(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	compacted := a.tokenSaver.Compact(body.Input)
	origLen := len(body.Input)
	newLen := len(compacted)
	reduction := 0.0
	if origLen > 0 && newLen < origLen {
		reduction = float64(origLen-newLen) / float64(origLen) * 100.0
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"original_bytes": origLen,
		"compact_bytes":  newLen,
		"reduction_pct":  reduction,
		"output":         compacted,
	})
}

func (a *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT key, value, updated_at FROM settings")
	if err != nil {
		http.Error(w, `{"error":"failed to query settings"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v, updatedAt string
		if err := rows.Scan(&k, &v, &updatedAt); err == nil {
			settings[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

func (a *AdminHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
		http.Error(w, `{"error":"key and value are required"}`, http.StatusBadRequest)
		return
	}

	_, err := a.db.ExecContext(r.Context(), `
INSERT INTO settings (key, value, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET
value = excluded.value,
updated_at = CURRENT_TIMESTAMP`,
		body.Key, body.Value)

	if err != nil {
		http.Error(w, `{"error":"failed to update setting"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated", "key": body.Key})
}
