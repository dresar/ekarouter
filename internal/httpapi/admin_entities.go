package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dresar/ekarouter/internal/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
		Enabled *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.ID == "" {
		body.ID = "prov_" + uuid.NewString()[:8]
	}
	if body.Key == "" {
		body.Key = body.ID
	}
	if body.Name == "" {
		body.Name = body.Key
	}

	enabledInt := 1
	if body.Enabled != nil && !*body.Enabled {
		enabledInt = 0
	}

	_, err := a.db.ExecContext(r.Context(),
		`INSERT INTO providers (id, key, name, kind, base_url, enabled)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			key = excluded.key,
			name = excluded.name,
			kind = excluded.kind,
			base_url = excluded.base_url,
			enabled = excluded.enabled,
			updated_at = CURRENT_TIMESTAMP`,
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
	proxyMap := make(map[string]ProxyProfileDTO)
	if pRows, err := a.db.QueryContext(r.Context(), "SELECT id, name, scheme, host, port, enabled FROM proxy_profiles"); err == nil {
		for pRows.Next() {
			var p ProxyProfileDTO
			var en int
			if err := pRows.Scan(&p.ID, &p.Name, &p.Scheme, &p.Host, &p.Port, &en); err == nil {
				p.Enabled = en == 1
				proxyMap[p.ID] = p
			}
		}
		pRows.Close()
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT a.id, a.provider_id, a.name, a.auth_type, a.state, a.priority, a.enabled,
		       COALESCE(c.encrypted_access, ''), COALESCE(c.encrypted_secret, '')
		FROM accounts a
		LEFT JOIN credentials c ON a.id = c.account_id
	`)
	if err != nil {
		http.Error(w, `{"error":"failed to query accounts"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Acc struct {
		ID           string `json:"id"`
		ProviderID   string `json:"provider_id"`
		Name         string `json:"name"`
		AuthType     string `json:"auth_type"`
		State        string `json:"state"`
		Priority     int    `json:"priority"`
		Enabled      bool   `json:"enabled"`
		ProxyPoolID  string `json:"proxy_pool_id,omitempty"`
		ProxyName    string `json:"proxy_name,omitempty"`
		ProxyURL     string `json:"proxy_url,omitempty"`
		MaskedSecret string `json:"masked_secret,omitempty"`
		LastError    string `json:"last_error,omitempty"`
	}

	var list []Acc
	for rows.Next() {
		var acc Acc
		var enabledInt int
		var encAccess, encSecret string
		if err := rows.Scan(&acc.ID, &acc.ProviderID, &acc.Name, &acc.AuthType, &acc.State, &acc.Priority, &enabledInt, &encAccess, &encSecret); err == nil {
			acc.Enabled = enabledInt == 1

			if a.crypto != nil {
				if encAccess != "" {
					if decAccess, err := a.crypto.Decrypt(encAccess); err == nil && decAccess != "" {
						if len(decAccess) > 10 {
							acc.MaskedSecret = decAccess[:6] + "..." + decAccess[len(decAccess)-4:]
						} else {
							acc.MaskedSecret = "••••••••"
						}
					}
				}

				if encSecret != "" {
					if decSecret, err := a.crypto.Decrypt(encSecret); err == nil && decSecret != "" {
						var meta map[string]any
						if err := json.Unmarshal([]byte(decSecret), &meta); err == nil {
							poolID, _ := meta["proxyPoolId"].(string)
							if poolID == "" {
								if psd, ok := meta["providerSpecificData"].(map[string]any); ok {
									poolID, _ = psd["proxyPoolId"].(string)
								}
							}
							if poolID != "" {
								acc.ProxyPoolID = poolID
								if prof, ok := proxyMap[poolID]; ok {
									acc.ProxyName = prof.Name
									if prof.Port > 0 && prof.Port != 80 && prof.Port != 443 {
										acc.ProxyURL = fmt.Sprintf("%s://%s:%d", prof.Scheme, prof.Host, prof.Port)
									} else {
										acc.ProxyURL = fmt.Sprintf("%s://%s", prof.Scheme, prof.Host)
									}
								}
							}
							if lastErr, ok := meta["lastError"].(string); ok && lastErr != "" {
								acc.LastError = lastErr
							}
						}
					}
				}
			}

			list = append(list, acc)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID          string `json:"id"`
		ProviderID  string `json:"provider_id"`
		Name        string `json:"name"`
		AuthType    string `json:"auth_type"`
		Priority    int    `json:"priority"`
		APIKey      string `json:"api_key"`
		SecretKey   string `json:"secret_key"`
		ProxyPoolID string `json:"proxy_pool_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.ID == "" {
		body.ID = "acc_" + uuid.NewString()[:8]
	}
	if body.Priority <= 0 {
		body.Priority = 10
	}
	if body.Name == "" {
		body.Name = body.ID
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO accounts (id, provider_id, name, auth_type, priority, state, enabled)
		VALUES (?, ?, ?, ?, ?, 'active', 1)
		ON CONFLICT(id) DO UPDATE SET
			provider_id = excluded.provider_id,
			name = excluded.name,
			auth_type = excluded.auth_type,
			priority = excluded.priority,
			state = 'active',
			enabled = 1,
			updated_at = CURRENT_TIMESTAMP`,
		body.ID, body.ProviderID, body.Name, body.AuthType, body.Priority)
	if err != nil {
		http.Error(w, `{"error":"failed to insert account"}`, http.StatusInternalServerError)
		return
	}

	encKey, _ := a.crypto.Encrypt(body.APIKey)

	meta := make(map[string]any)
	if body.SecretKey != "" {
		_ = json.Unmarshal([]byte(body.SecretKey), &meta)
		if meta == nil {
			meta = map[string]any{"secret": body.SecretKey}
		}
	}
	if body.ProxyPoolID != "" {
		meta["proxyPoolId"] = body.ProxyPoolID
	}
	metaBytes, _ := json.Marshal(meta)
	encSecret, _ := a.crypto.Encrypt(string(metaBytes))

	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(account_id) DO UPDATE SET
			encrypted_access = excluded.encrypted_access,
			encrypted_secret = excluded.encrypted_secret,
			updated_at = CURRENT_TIMESTAMP`,
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

func (a *AdminHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        *string `json:"name"`
		Priority    *int    `json:"priority"`
		Enabled     *bool   `json:"enabled"`
		State       *string `json:"state"`
		ProxyPoolID *string `json:"proxy_pool_id"`
		APIKey      *string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.Name != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *body.Name, id)
	}
	if body.Priority != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET priority = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *body.Priority, id)
	}
	if body.Enabled != nil {
		en := 0
		if *body.Enabled {
			en = 1
		}
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", en, id)
	}
	if body.State != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET state = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *body.State, id)
	}
	if body.ProxyPoolID != nil && a.crypto != nil {
		var encSec string
		_ = a.db.QueryRowContext(r.Context(), "SELECT encrypted_secret FROM credentials WHERE account_id = ?", id).Scan(&encSec)
		var meta map[string]any
		if encSec != "" {
			if sec, err := a.crypto.Decrypt(encSec); err == nil {
				_ = json.Unmarshal([]byte(sec), &meta)
			}
		}
		if meta == nil {
			meta = make(map[string]any)
		}
		meta["proxyPoolId"] = *body.ProxyPoolID
		bytes, _ := json.Marshal(meta)
		newEnc, _ := a.crypto.Encrypt(string(bytes))
		_, _ = a.db.ExecContext(r.Context(), `
			INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret)
			VALUES (?, ?, '', ?)
			ON CONFLICT(account_id) DO UPDATE SET
				encrypted_secret = excluded.encrypted_secret,
				updated_at = CURRENT_TIMESTAMP`,
			"cred_"+id, id, newEnc)
	}
	if body.APIKey != nil && a.crypto != nil {
		newKey, _ := a.crypto.Encrypt(*body.APIKey)
		_, _ = a.db.ExecContext(r.Context(), `
			INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret)
			VALUES (?, ?, ?, '')
			ON CONFLICT(account_id) DO UPDATE SET
				encrypted_access = excluded.encrypted_access,
				updated_at = CURRENT_TIMESTAMP`,
			"cred_"+id, id, newKey)
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "updated", "id": id})
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

func (a *AdminHandler) ApplyBatchProxy(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProviderID  string `json:"provider_id"`
		Mode        string `json:"mode"`
		ProxyPoolID string `json:"proxy_pool_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProviderID == "" {
		http.Error(w, `{"error":"provider_id and mode required"}`, http.StatusBadRequest)
		return
	}

	accRows, err := a.db.QueryContext(r.Context(), "SELECT a.id, COALESCE(c.encrypted_secret, '') FROM accounts a LEFT JOIN credentials c ON c.account_id = a.id WHERE a.provider_id = ? ORDER BY a.priority DESC, a.name ASC", body.ProviderID)
	if err != nil {
		http.Error(w, `{"error":"failed to query accounts"}`, http.StatusInternalServerError)
		return
	}
	defer accRows.Close()

	type accItem struct {
		id        string
		encSecret string
	}
	var accList []accItem
	for accRows.Next() {
		var item accItem
		if err := accRows.Scan(&item.id, &item.encSecret); err == nil {
			accList = append(accList, item)
		}
	}

	if len(accList) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"updated": 0, "status": "no_accounts"})
		return
	}

	var proxyList []string
	if body.Mode == "rotate" {
		pRows, err := a.db.QueryContext(r.Context(), "SELECT id FROM proxy_profiles WHERE enabled = 1 ORDER BY id ASC")
		if err == nil {
			for pRows.Next() {
				var pid string
				if err := pRows.Scan(&pid); err == nil {
					proxyList = append(proxyList, pid)
				}
			}
			pRows.Close()
		}
		if len(proxyList) == 0 {
			pRows2, err := a.db.QueryContext(r.Context(), "SELECT id FROM proxy_profiles ORDER BY id ASC")
			if err == nil {
				for pRows2.Next() {
					var pid string
					if err := pRows2.Scan(&pid); err == nil {
						proxyList = append(proxyList, pid)
					}
				}
				pRows2.Close()
			}
		}
		if len(proxyList) == 0 {
			http.Error(w, `{"error":"No proxies available to rotate"}`, http.StatusBadRequest)
			return
		}
	}

	count := 0
	for i, acc := range accList {
		targetProxy := ""
		if body.Mode == "rotate" {
			targetProxy = proxyList[i%len(proxyList)]
		} else if body.Mode == "single" {
			targetProxy = body.ProxyPoolID
		} else if body.Mode == "none" {
			targetProxy = ""
		}

		var meta map[string]any
		if acc.encSecret != "" && a.crypto != nil {
			if sec, err := a.crypto.Decrypt(acc.encSecret); err == nil {
				_ = json.Unmarshal([]byte(sec), &meta)
			}
		}
		if meta == nil {
			meta = make(map[string]any)
		}

		if targetProxy == "" {
			delete(meta, "proxyPoolId")
		} else {
			meta["proxyPoolId"] = targetProxy
		}

		newEnc := ""
		if a.crypto != nil {
			bytes, _ := json.Marshal(meta)
			newEnc, _ = a.crypto.Encrypt(string(bytes))
		}

		_, _ = a.db.ExecContext(r.Context(), `
			INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret)
			VALUES (?, ?, '', ?)
			ON CONFLICT(account_id) DO UPDATE SET
				encrypted_secret = excluded.encrypted_secret,
				updated_at = CURRENT_TIMESTAMP`,
			"cred_"+acc.id, acc.id, newEnc)
		count++
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"mode":    body.Mode,
		"updated": count,
	})
}

func (a *AdminHandler) TestAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var providerID string
	var enabledInt int
	var encAccess, encSecret string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT a.provider_id, a.enabled, COALESCE(c.encrypted_access, ''), COALESCE(c.encrypted_secret, '')
		FROM accounts a
		LEFT JOIN credentials c ON c.account_id = a.id
		WHERE a.id = ?`, id).Scan(&providerID, &enabledInt, &encAccess, &encSecret)

	if err != nil {
		http.Error(w, `{"error":"account not found"}`, http.StatusNotFound)
		return
	}

	apiKey := ""
	if encAccess != "" && a.crypto != nil {
		apiKey, _ = a.crypto.Decrypt(encAccess)
	}

	var meta map[string]any
	if encSecret != "" && a.crypto != nil {
		if sec, err := a.crypto.Decrypt(encSecret); err == nil {
			_ = json.Unmarshal([]byte(sec), &meta)
		}
	}
	if meta == nil {
		meta = make(map[string]any)
	}

	proxyPoolID, _ := meta["proxyPoolId"].(string)
	var proxyURL string
	if proxyPoolID != "" {
		var scheme, host string
		var port int
		var user sql.NullString
		if err := a.db.QueryRowContext(r.Context(), "SELECT scheme, host, port, username FROM proxy_profiles WHERE id = ?", proxyPoolID).Scan(&scheme, &host, &port, &user); err == nil {
			if port > 0 && port != 80 && port != 443 {
				proxyURL = fmt.Sprintf("%s://%s:%d", scheme, host, port)
			} else {
				proxyURL = fmt.Sprintf("%s://%s", scheme, host)
			}
		}
	}

	start := time.Now()
	httpClient := &http.Client{Timeout: 7 * time.Second}
	if proxyURL != "" {
		if parsedURL, err := url.Parse(proxyURL); err == nil {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(parsedURL),
			}
		}
	}

	testURL := ""
	var testReq *http.Request

	switch providerID {
	case "gemini", "gemini-cli":
		testURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
		testReq, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, testURL, nil)
	case "groq":
		testURL = "https://api.groq.com/openai/v1/models"
		testReq, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, testURL, nil)
		testReq.Header.Set("Authorization", "Bearer "+apiKey)
	case "openrouter":
		testURL = "https://openrouter.ai/api/v1/models"
		testReq, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, testURL, nil)
		testReq.Header.Set("Authorization", "Bearer "+apiKey)
	case "openai":
		testURL = "https://api.openai.com/v1/models"
		testReq, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, testURL, nil)
		testReq.Header.Set("Authorization", "Bearer "+apiKey)
	default:
		testURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
		testReq, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, testURL, nil)
	}

	resp, err := httpClient.Do(testReq)
	latency := time.Since(start).Milliseconds()

	healthy := false
	statusMsg := ""

	if err != nil {
		statusMsg = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			healthy = true
			statusMsg = "OK"
		} else {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			statusMsg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	if healthy {
		delete(meta, "lastError")
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET state = 'active', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	} else {
		meta["lastError"] = statusMsg
		_, _ = a.db.ExecContext(r.Context(), "UPDATE accounts SET state = 'cooling_down', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	}

	if a.crypto != nil {
		bytes, _ := json.Marshal(meta)
		newEnc, _ := a.crypto.Encrypt(string(bytes))
		_, _ = a.db.ExecContext(r.Context(), `
			INSERT INTO credentials (id, account_id, encrypted_access, encrypted_secret)
			VALUES (?, ?, '', ?)
			ON CONFLICT(account_id) DO UPDATE SET
				encrypted_secret = excluded.encrypted_secret,
				updated_at = CURRENT_TIMESTAMP`,
			"cred_"+id, id, newEnc)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"healthy": healthy,
		"latency": latency,
		"message": statusMsg,
	})
}


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

func (a *AdminHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, strategy, enabled FROM routes ORDER BY name ASC")
	if err != nil {
		http.Error(w, `{"error":"failed to query routes"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	if body.ID == "" {
		body.ID = "route_" + uuid.NewString()[:8]
	}
	if body.Strategy == "" {
		body.Strategy = "priority"
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO routes (id, name, strategy, enabled)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			strategy = excluded.strategy,
			enabled = 1,
			updated_at = CURRENT_TIMESTAMP`,
		body.ID, body.Name, body.Strategy)
	if err != nil {
		http.Error(w, `{"error":"failed to insert route"}`, http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(r.Context(), "DELETE FROM route_items WHERE route_id = ?", body.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to clean old route items"}`, http.StatusInternalServerError)
		return
	}

	for idx, item := range body.Items {
		timeout := item.TimeoutMs
		if timeout <= 0 {
			timeout = 60000
		}
		retries := item.MaxRetries
		if retries <= 0 {
			retries = 2
		}
		weight := item.Weight
		if weight <= 0 {
			weight = 1
		}
		itemID := item.ID
		if itemID == "" {
			itemID = fmt.Sprintf("%s_item_%d", body.ID, idx+1)
		}
		var accID any = nil
		if item.AccountID != "" {
			accID = item.AccountID
		}
		var modelID any = nil
		if item.ModelID != "" {
			modelID = item.ModelID
		}
		_, err = tx.ExecContext(r.Context(),
			"INSERT INTO route_items (id, route_id, provider_id, account_id, model_id, priority, weight, enabled, timeout_ms, max_retries) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)",
			itemID, body.ID, item.ProviderID, accID, modelID, item.Priority, weight, timeout, retries)
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

func (a *AdminHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rd RouteDTO
	var enabledInt int
	err := a.db.QueryRowContext(r.Context(), "SELECT id, name, strategy, enabled FROM routes WHERE id = ?", id).Scan(
		&rd.ID, &rd.Name, &rd.Strategy, &enabledInt)
	if err != nil {
		http.Error(w, `{"error":"route not found"}`, http.StatusNotFound)
		return
	}
	rd.Enabled = enabledInt == 1
	rd.Items = make([]RouteItemDTO, 0)
	itemRows, err := a.db.QueryContext(r.Context(), "SELECT id, provider_id, COALESCE(account_id, ''), COALESCE(model_id, ''), priority, weight, timeout_ms, max_retries, enabled FROM route_items WHERE route_id = ? ORDER BY priority ASC", rd.ID)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var item RouteItemDTO
			var itEnabled int
			if err := itemRows.Scan(&item.ID, &item.ProviderID, &item.AccountID, &item.ModelID, &item.Priority, &item.Weight, &item.TimeoutMs, &item.MaxRetries, &itEnabled); err == nil {
				item.Enabled = itEnabled == 1
				rd.Items = append(rd.Items, item)
			}
		}
	}
	rd.ItemCount = len(rd.Items)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rd)
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

func (a *AdminHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Enabled     *bool   `json:"enabled"`
		DisplayName *string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.Enabled != nil {
		en := 0
		if *body.Enabled {
			en = 1
		}
		_, _ = a.db.ExecContext(r.Context(), "UPDATE models SET enabled = ? WHERE id = ?", en, id)
	}
	if body.DisplayName != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE models SET display_name = ? WHERE id = ?", *body.DisplayName, id)
	}

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "updated", "id": id})
}

func (a *AdminHandler) BatchToggleModels(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProviderID string `json:"provider_id"`
		Enabled    bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProviderID == "" {
		http.Error(w, `{"error":"provider_id and enabled required"}`, http.StatusBadRequest)
		return
	}

	en := 0
	if body.Enabled {
		en = 1
	}

	res, err := a.db.ExecContext(r.Context(), "UPDATE models SET enabled = ? WHERE provider_id = ?", en, body.ProviderID)
	if err != nil {
		http.Error(w, `{"error":"failed to update models"}`, http.StatusInternalServerError)
		return
	}
	count, _ := res.RowsAffected()

	if a.router != nil {
		_ = a.router.LoadFromDB(r.Context(), a.db)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"provider_id": body.ProviderID,
		"enabled":     body.Enabled,
		"count":       count,
	})
}


type ProxyProfileDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Scheme         string `json:"scheme"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	MaskedPassword string `json:"masked_password,omitempty"`
	Enabled        bool   `json:"enabled"`
}

func (a *AdminHandler) ListProxyProfiles(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, scheme, host, port, username, enabled FROM proxy_profiles")
	if err != nil {
		http.Error(w, `{"error":"failed to query proxy profiles"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []ProxyProfileDTO
	for rows.Next() {
		var p ProxyProfileDTO
		var enabledInt int
		var user sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Scheme, &p.Host, &p.Port, &user, &enabledInt); err == nil {
			p.Enabled = enabledInt == 1
			p.Username = user.String
			list = append(list, p)
		}
	}
	if list == nil {
		list = make([]ProxyProfileDTO, 0)
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
		Enabled  *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if body.ID == "" {
		body.ID = "proxy_" + uuid.NewString()[:8]
	}

	enabled := 1
	if body.Enabled != nil && !*body.Enabled {
		enabled = 0
	}

	var encPass string
	if body.Password != "" {
		encPass, _ = a.crypto.Encrypt(body.Password)
	}

	_, err := a.db.ExecContext(r.Context(), `
INSERT INTO proxy_profiles (id, name, scheme, host, port, username, encrypted_password, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
name = excluded.name,
scheme = excluded.scheme,
host = excluded.host,
port = excluded.port,
username = excluded.username,
encrypted_password = CASE WHEN excluded.encrypted_password != '' THEN excluded.encrypted_password ELSE proxy_profiles.encrypted_password END,
enabled = excluded.enabled`,
		body.ID, body.Name, body.Scheme, body.Host, body.Port, body.Username, encPass, enabled)

	if err != nil {
		http.Error(w, `{"error":"failed to save proxy profile"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": body.ID})
}

func (a *AdminHandler) DeleteProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := a.db.ExecContext(r.Context(), "DELETE FROM proxy_profiles WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete proxy profile"}`, http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (a *AdminHandler) GetProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p ProxyProfileDTO
	var enabledInt int
	var user, pass sql.NullString
	err := a.db.QueryRowContext(r.Context(), "SELECT id, name, scheme, host, port, username, encrypted_password, enabled FROM proxy_profiles WHERE id = ?", id).Scan(
		&p.ID, &p.Name, &p.Scheme, &p.Host, &p.Port, &user, &pass, &enabledInt)
	if err != nil {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}
	p.Enabled = enabledInt == 1
	p.Username = user.String
	if pass.Valid && pass.String != "" {
		p.MaskedPassword = "••••••••"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (a *AdminHandler) UpdateProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name     *string `json:"name"`
		Scheme   *string `json:"scheme"`
		Host     *string `json:"host"`
		Port     *int    `json:"port"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Enabled  *bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	var exists int
	if err := a.db.QueryRowContext(r.Context(), "SELECT 1 FROM proxy_profiles WHERE id = ?", id).Scan(&exists); err != nil {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}

	if body.Name != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET name = ? WHERE id = ?", *body.Name, id)
	}
	if body.Scheme != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET scheme = ? WHERE id = ?", *body.Scheme, id)
	}
	if body.Host != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET host = ? WHERE id = ?", *body.Host, id)
	}
	if body.Port != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET port = ? WHERE id = ?", *body.Port, id)
	}
	if body.Username != nil {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET username = ? WHERE id = ?", *body.Username, id)
	}
	if body.Password != nil && *body.Password != "" {
		encPass, _ := a.crypto.Encrypt(*body.Password)
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET encrypted_password = ? WHERE id = ?", encPass, id)
	}
	if body.Enabled != nil {
		en := 0
		if *body.Enabled {
			en = 1
		}
		_, _ = a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET enabled = ? WHERE id = ?", en, id)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated", "id": id})
}

func (a *AdminHandler) EnableProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET enabled = 1 WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to enable proxy profile"}`, http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "enabled", "id": id})
}

func (a *AdminHandler) DisableProxyProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := a.db.ExecContext(r.Context(), "UPDATE proxy_profiles SET enabled = 0 WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"failed to disable proxy profile"}`, http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, `{"error":"proxy profile not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "id": id})
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
