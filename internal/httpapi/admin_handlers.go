package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
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
}

func NewAdminHandler(
	db *sql.DB,
	cfg *config.Config,
	crypto *auth.CryptoService,
	usageRec *usage.Recorder,
	ts *tokensaver.TokenSaver,
) *AdminHandler {
	return &AdminHandler{
		db:         db,
		cfg:        cfg,
		crypto:     crypto,
		usageRec:   usageRec,
		tokenSaver: ts,
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

func (a *AdminHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, key, name, kind, base_url, enabled, created_at, updated_at FROM providers")
	if err != nil {
		http.Error(w, `{"error":"failed to query providers"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Prov struct {
		ID        string `json:"id"`
		Key       string `json:"key"`
		Name      string `json:"name"`
		Kind      string `json:"kind"`
		BaseURL   string `json:"base_url"`
		Enabled   bool   `json:"enabled"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	var list []Prov
	for rows.Next() {
		var p Prov
		var enabledInt int
		if err := rows.Scan(&p.ID, &p.Key, &p.Name, &p.Kind, &p.BaseURL, &enabledInt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			p.Enabled = enabledInt == 1
			list = append(list, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID      string `json:"id"`
		Key     string `json:"key"`
		Name    string `json:"name"`
		Kind    string `json:"kind"`
		BaseURL string `json:"base_url"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	enabledInt := 1
	if !body.Enabled {
		enabledInt = 0
	}

	_, err := a.db.ExecContext(r.Context(),
		"INSERT INTO providers (id, key, name, kind, base_url, enabled) VALUES (?, ?, ?, ?, ?, ?)",
		body.ID, body.Key, body.Name, body.Kind, body.BaseURL, enabledInt)
	if err != nil {
		http.Error(w, `{"error":"failed to create provider"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM providers WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete provider"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, provider_id, name, auth_type, state, priority, enabled FROM accounts")
	if err != nil {
		http.Error(w, `{"error":"failed to query accounts"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Acc struct {
		ID         string `json:"id"`
		ProviderID string `json:"provider_id"`
		Name       string `json:"name"`
		AuthType   string `json:"auth_type"`
		State      string `json:"state"`
		Priority   int    `json:"priority"`
		Enabled    bool   `json:"enabled"`
	}

	var list []Acc
	for rows.Next() {
		var acc Acc
		var enabledInt int
		if err := rows.Scan(&acc.ID, &acc.ProviderID, &acc.Name, &acc.AuthType, &acc.State, &acc.Priority, &enabledInt); err == nil {
			acc.Enabled = enabledInt == 1
			list = append(list, acc)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID         string `json:"id"`
		ProviderID string `json:"provider_id"`
		Name       string `json:"name"`
		AuthType   string `json:"auth_type"`
		Priority   int    `json:"priority"`
		APIKey     string `json:"api_key"`
		SecretKey  string `json:"secret_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		"INSERT INTO accounts (id, provider_id, name, auth_type, priority, state, enabled) VALUES (?, ?, ?, ?, ?, 'active', 1)",
		body.ID, body.ProviderID, body.Name, body.AuthType, body.Priority)
	if err != nil {
		http.Error(w, `{"error":"failed to insert account"}`, http.StatusInternalServerError)
		return
	}

	encKey, _ := a.crypto.Encrypt(body.APIKey)
	encSecret, _ := a.crypto.Encrypt(body.SecretKey)

	_, err = tx.ExecContext(r.Context(),
		"INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret) VALUES (?, ?, ?, ?)",
		"cred_"+body.ID, body.ID, encKey, encSecret)
	if err != nil {
		http.Error(w, `{"error":"failed to insert credentials"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error":"commit failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete account"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, strategy, enabled FROM routes")
	if err != nil {
		http.Error(w, `{"error":"failed to query routes"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RouteDTO struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Strategy string `json:"strategy"`
		Enabled  bool   `json:"enabled"`
	}

	var list []RouteDTO
	for rows.Next() {
		var rd RouteDTO
		var enabledInt int
		if err := rows.Scan(&rd.ID, &rd.Name, &rd.Strategy, &enabledInt); err == nil {
			rd.Enabled = enabledInt == 1
			list = append(list, rd)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Strategy string `json:"strategy"`
		Items    []struct {
			ID         string `json:"id"`
			ProviderID string `json:"provider_id"`
			AccountID  string `json:"account_id"`
			ModelID    string `json:"model_id"`
			Priority   int    `json:"priority"`
			Weight     int    `json:"weight"`
			TimeoutMs  int    `json:"timeout_ms"`
			MaxRetries int    `json:"max_retries"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		"INSERT INTO routes (id, name, strategy, enabled) VALUES (?, ?, ?, 1)",
		body.ID, body.Name, body.Strategy)
	if err != nil {
		http.Error(w, `{"error":"failed to insert route"}`, http.StatusInternalServerError)
		return
	}

	for _, item := range body.Items {
		timeout := item.TimeoutMs
		if timeout <= 0 {
			timeout = 60000
		}
		retries := item.MaxRetries
		if retries <= 0 {
			retries = 2
		}
		_, err = tx.ExecContext(r.Context(),
			"INSERT INTO route_items (id, route_id, provider_id, account_id, model_id, priority, weight, enabled, timeout_ms, max_retries) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)",
			item.ID, body.ID, item.ProviderID, item.AccountID, item.ModelID, item.Priority, item.Weight, timeout, retries)
		if err != nil {
			http.Error(w, `{"error":"failed to insert route item"}`, http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error":"commit failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM routes WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete route"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
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
