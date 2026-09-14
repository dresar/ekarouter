package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dresar/ekarouter/internal/proxy"
	"github.com/go-chi/chi/v5"
)

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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, strategy, enabled FROM routes ORDER BY name ASC")
	if err != nil {
		http.Error(w, `{"error":"failed to query routes"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RouteItemDTO struct {
		ID         string `json:"id"`
		ProviderID string `json:"provider_id"`
		AccountID  string `json:"account_id"`
		ModelID    string `json:"model_id"`
		Priority   int    `json:"priority"`
		Weight     int    `json:"weight"`
		TimeoutMs  int    `json:"timeout_ms"`
		MaxRetries int    `json:"max_retries"`
		Enabled    bool   `json:"enabled"`
	}

	type RouteDTO struct {
		ID        string         `json:"id"`
		Name      string         `json:"name"`
		Strategy  string         `json:"strategy"`
		Enabled   bool           `json:"enabled"`
		ItemCount int            `json:"item_count"`
		Items     []RouteItemDTO `json:"items"`
	}

	var list []RouteDTO
	for rows.Next() {
		var rd RouteDTO
		var enabledInt int
		if err := rows.Scan(&rd.ID, &rd.Name, &rd.Strategy, &enabledInt); err == nil {
			rd.Enabled = enabledInt == 1
			itemRows, err := a.db.QueryContext(r.Context(), "SELECT id, provider_id, COALESCE(account_id, ''), COALESCE(model_id, ''), priority, weight, timeout_ms, max_retries, enabled FROM route_items WHERE route_id = ? ORDER BY priority ASC", rd.ID)
			if err == nil {
				for itemRows.Next() {
					var item RouteItemDTO
					var itEnabled int
					if err := itemRows.Scan(&item.ID, &item.ProviderID, &item.AccountID, &item.ModelID, &item.Priority, &item.Weight, &item.TimeoutMs, &item.MaxRetries, &itEnabled); err == nil {
						item.Enabled = itEnabled == 1
						rd.Items = append(rd.Items, item)
					}
				}
				itemRows.Close()
			}
			if rd.Items == nil {
				rd.Items = make([]RouteItemDTO, 0)
			}
			rd.ItemCount = len(rd.Items)
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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
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

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListModelsAdmin(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, provider_id, external_name, display_name, context_limit, input_capability, output_capability, streaming, enabled FROM models")
	if err != nil {
		http.Error(w, `{"error":"failed to query models"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ModelRow struct {
		ID               string `json:"id"`
		ProviderID       string `json:"provider_id"`
		ExternalName     string `json:"external_name"`
		DisplayName      string `json:"display_name"`
		ContextLimit     int    `json:"context_limit"`
		InputCapability  string `json:"input_capability"`
		OutputCapability string `json:"output_capability"`
		Streaming        bool   `json:"streaming"`
		Enabled          bool   `json:"enabled"`
	}

	var list []ModelRow
	for rows.Next() {
		var m ModelRow
		var streamInt, enabledInt int
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ExternalName, &m.DisplayName, &m.ContextLimit, &m.InputCapability, &m.OutputCapability, &streamInt, &enabledInt); err == nil {
			m.Streaming = streamInt == 1
			m.Enabled = enabledInt == 1
			list = append(list, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID               string `json:"id"`
		ProviderID       string `json:"provider_id"`
		ExternalName     string `json:"external_name"`
		DisplayName      string `json:"display_name"`
		ContextLimit     int    `json:"context_limit"`
		InputCapability  string `json:"input_capability"`
		OutputCapability string `json:"output_capability"`
		Streaming        bool   `json:"streaming"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.ContextLimit <= 0 {
		body.ContextLimit = 8192
	}
	streamInt := 1
	if !body.Streaming {
		streamInt = 0
	}

	_, err := a.db.ExecContext(r.Context(), `
INSERT INTO models (id, provider_id, external_name, display_name, context_limit, input_capability, output_capability, streaming, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
ON CONFLICT(id) DO UPDATE SET
external_name = excluded.external_name,
display_name = excluded.display_name,
context_limit = excluded.context_limit,
streaming = excluded.streaming`,
		body.ID, body.ProviderID, body.ExternalName, body.DisplayName, body.ContextLimit, body.InputCapability, body.OutputCapability, streamInt)

	if err != nil {
		http.Error(w, `{"error":"failed to save model"}`, http.StatusInternalServerError)
		return
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM models WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete model"}`, http.StatusInternalServerError)
		return
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) ListProxyProfiles(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, scheme, host, port, username, enabled FROM proxy_profiles")
	if err != nil {
		http.Error(w, `{"error":"failed to query proxy profiles"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ProfileDTO struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Scheme   string `json:"scheme"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Enabled  bool   `json:"enabled"`
	}

	var list []ProfileDTO
	for rows.Next() {
		var p ProfileDTO
		var enabledInt int
		var user sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Scheme, &p.Host, &p.Port, &user, &enabledInt); err == nil {
			p.Enabled = enabledInt == 1
			p.Username = user.String
			list = append(list, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateProxyProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Scheme   string `json:"scheme"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	var encPass string
	if body.Password != "" {
		encPass, _ = a.crypto.Encrypt(body.Password)
	}

	_, err := a.db.ExecContext(r.Context(), `
INSERT INTO proxy_profiles (id, name, scheme, host, port, username, encrypted_password, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		body.ID, body.Name, body.Scheme, body.Host, body.Port, body.Username, encPass)

	if err != nil {
		http.Error(w, `{"error":"failed to save proxy profile"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := a.db.ExecContext(r.Context(), "DELETE FROM proxy_profiles WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete proxy profile"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) TestProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pName, pScheme, pHost, pUser, pPass sql.NullString
	var pPort, pEnabled int
	err := a.db.QueryRowContext(r.Context(), "SELECT name, scheme, host, port, username, encrypted_password, enabled FROM proxy_profiles WHERE id = ?", id).Scan(
		&pName, &pScheme, &pHost, &pPort, &pUser, &pPass, &pEnabled)
	if err != nil {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}

	pass, _ := a.crypto.Decrypt(pPass.String)
	prof := &proxy.Profile{
		ID:       id,
		Name:     pName.String,
		Scheme:   pScheme.String,
		Host:     pHost.String,
		Port:     pPort,
		Username: pUser.String,
		Password: pass,
	}

	ok, statusCode, latency, testErr := proxy.TestProfile(r.Context(), prof, 10*time.Second)
	errStr := ""
	if testErr != nil {
		errStr = testErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":         ok,
		"status":     statusCode,
		"latency_ms": latency,
		"error":      errStr,
	})
}

func (a *AdminHandler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProviderID string `json:"provider_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProviderID == "" {
		http.Error(w, `{"error":"provider_id is required"}`, http.StatusBadRequest)
		return
	}

	rawState, codeChallenge, err := a.oauthMgr.GenerateState(body.ProviderID, 10*time.Minute)
	if err != nil {
		http.Error(w, `{"error":"failed to generate oauth state"}`, http.StatusInternalServerError)
		return
	}

	authURL := "/oauth/authorize?provider=" + body.ProviderID + "&state=" + rawState + "&code_challenge=" + codeChallenge

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"transaction_id": rawState,
		"auth_url":       authURL,
		"state":          rawState,
		"code_challenge": codeChallenge,
	})
}

func (a *AdminHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		State       string `json:"state"`
		Code        string `json:"code"`
		AccountName string `json:"account_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.State == "" || body.Code == "" {
		http.Error(w, `{"error":"state and code are required"}`, http.StatusBadRequest)
		return
	}

	stateRec, err := a.oauthMgr.ConsumeState(body.State)
	if err != nil {
		http.Error(w, `{"error":"invalid or expired oauth state"}`, http.StatusBadRequest)
		return
	}

	accID := "acc_oauth_" + body.State[:8]
	accName := body.AccountName
	if accName == "" {
		accName = stateRec.ProviderID + "-oauth"
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		"INSERT INTO accounts (id, provider_id, name, auth_type, priority, state, enabled) VALUES (?, ?, ?, 'oauth', 10, 'active', 1)",
		accID, stateRec.ProviderID, accName)
	if err != nil {
		http.Error(w, `{"error":"failed to insert account"}`, http.StatusInternalServerError)
		return
	}

	encAccess, _ := a.crypto.Encrypt(body.Code)
	_, err = tx.ExecContext(r.Context(),
		"INSERT INTO credentials (id, account_id, encrypted_access) VALUES (?, ?, ?)",
		"cred_"+accID, accID, encAccess)
	if err != nil {
		http.Error(w, `{"error":"failed to insert credentials"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error":"commit failed"}`, http.StatusInternalServerError)
		return
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":     "connected",
		"account_id": accID,
	})
}
