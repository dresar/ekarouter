package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CredentialResponse struct {
	ID             string `json:"id"`
	AccountID      string `json:"account_id"`
	AccountName    string `json:"account_name"`
	ProviderID     string `json:"provider_id"`
	KeyFingerprint string `json:"key_fingerprint"`
	HasAccess      bool   `json:"has_access"`
	HasRefresh     bool   `json:"has_refresh"`
	HasSecret      bool   `json:"has_secret"`
	UpdatedAt      string `json:"updated_at"`
}

func (a *AdminHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	accFilter := r.URL.Query().Get("account_id")
	query := `
SELECT c.id, c.account_id, a.name, a.provider_id, COALESCE(c.key_fingerprint, ''),
       c.encrypted_access, c.encrypted_refresh, c.encrypted_secret, c.updated_at
FROM credentials c
JOIN accounts a ON a.id = c.account_id`
	var args []any
	if accFilter != "" {
		query += " WHERE c.account_id = ?"
		args = append(args, accFilter)
	}
	query += " ORDER BY c.updated_at DESC"

	rows, err := a.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, `{"error":"failed to query credentials"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []CredentialResponse
	for rows.Next() {
		var id, accID, accName, provID, fp, updated string
		var encAccess, encRefresh, encSecret sql.NullString
		if err := rows.Scan(&id, &accID, &accName, &provID, &fp, &encAccess, &encRefresh, &encSecret, &updated); err == nil {
			list = append(list, CredentialResponse{
				ID:             id,
				AccountID:      accID,
				AccountName:    accName,
				ProviderID:     provID,
				KeyFingerprint: fp,
				HasAccess:      encAccess.Valid && encAccess.String != "",
				HasRefresh:     encRefresh.Valid && encRefresh.String != "",
				HasSecret:      encSecret.Valid && encSecret.String != "",
				UpdatedAt:      updated,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a *AdminHandler) GetCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	query := `
SELECT c.id, c.account_id, a.name, a.provider_id, COALESCE(c.key_fingerprint, ''),
       c.encrypted_access, c.encrypted_refresh, c.encrypted_secret, c.updated_at
FROM credentials c
JOIN accounts a ON a.id = c.account_id
WHERE c.id = ? OR c.account_id = ?`

	var credID, accID, accName, provID, fp, updated string
	var encAccess, encRefresh, encSecret sql.NullString
	err := a.db.QueryRowContext(r.Context(), query, id, id).Scan(
		&credID, &accID, &accName, &provID, &fp, &encAccess, &encRefresh, &encSecret, &updated)
	if err != nil {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CredentialResponse{
		ID:             credID,
		AccountID:      accID,
		AccountName:    accName,
		ProviderID:     provID,
		KeyFingerprint: fp,
		HasAccess:      encAccess.Valid && encAccess.String != "",
		HasRefresh:     encRefresh.Valid && encRefresh.String != "",
		HasSecret:      encSecret.Valid && encSecret.String != "",
		UpdatedAt:      updated,
	})
}
