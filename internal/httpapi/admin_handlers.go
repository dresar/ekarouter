package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	db         *sql.DB
	cfg        *config.Config
	crypto     *auth.CryptoService
	usageRec   *usage.Recorder
	tokenSaver *tokensaver.TokenSaver
	router     *routing.Router
	oauthMgr   *oauth.Manager
}

func NewAdminHandler(
	db *sql.DB,
	cfg *config.Config,
	crypto *auth.CryptoService,
	usageRec *usage.Recorder,
	ts *tokensaver.TokenSaver,
	router *routing.Router,
	oauthMgr *oauth.Manager,
) *AdminHandler {
	if oauthMgr == nil {
		oauthMgr = oauth.NewManager()
	}
	return &AdminHandler{
		db:         db,
		cfg:        cfg,
		crypto:     crypto,
		usageRec:   usageRec,
		tokenSaver: ts,
		router:     router,
		oauthMgr:   oauthMgr,
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

	if body.Username != a.cfg.AdminUser || body.Password != a.cfg.AdminPassword {
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
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": userID,
	})
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
