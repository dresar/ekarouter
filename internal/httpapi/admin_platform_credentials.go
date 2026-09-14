package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/vault"
	"github.com/go-chi/chi/v5"
)

func (h *PlatformHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider_id")
	project := r.URL.Query().Get("project_id")
	env := r.URL.Query().Get("environment")

	list, err := h.vaultStore.ListCredentials(r.Context(), provider, project, env)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "database_error", err.Error(), nil)
		return
	}
	if list == nil {
		list = []*vault.Credential{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name           string               `json:"name"`
		CredentialType vault.CredentialType `json:"credential_type"`
		ProviderID     string               `json:"provider_id"`
		ProjectID      string               `json:"project_id"`
		TeamID         string               `json:"team_id"`
		Environment    string               `json:"environment"`
		SecretValue    string               `json:"secret_value"`
		Priority       int                  `json:"priority"`
		Tags           string               `json:"tags"`
		Notes          string               `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON body", nil)
		return
	}

	if body.Name == "" || body.ProviderID == "" || body.SecretValue == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing_fields", "Name, provider_id, and secret_value are required", nil)
		return
	}

	cred := &vault.Credential{
		Name:           body.Name,
		CredentialType: body.CredentialType,
		ProviderID:     body.ProviderID,
		ProjectID:      body.ProjectID,
		TeamID:         body.TeamID,
		Environment:    body.Environment,
		Priority:       body.Priority,
		Tags:           body.Tags,
		Notes:          body.Notes,
	}

	if err := h.vaultStore.CreateCredential(r.Context(), cred, body.SecretValue); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "vault_error", err.Error(), nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "credential.create",
		ResourceType: "credential",
		ResourceID:   cred.ID,
		ProjectID:    cred.ProjectID,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, cred)
}

func (h *PlatformHandler) GetCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cred, err := h.vaultStore.GetCredential(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "credential_not_found", "Credential not found", nil)
		return
	}
	h.writeSuccess(w, r, cred)
}

func (h *PlatformHandler) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name     string `json:"name"`
		Priority int    `json:"priority"`
		Tags     string `json:"tags"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}

	if err := h.vaultStore.UpdateCredential(r.Context(), id, body.Name, body.Priority, body.Tags, body.Notes); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.vaultStore.DeleteCredential(r.Context(), id); err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) RotateCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		NewSecret string `json:"new_secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.NewSecret == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "new_secret is required", nil)
		return
	}

	if err := h.vaultStore.RotateSecret(r.Context(), id, body.NewSecret); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "rotation_error", err.Error(), nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "credential.rotate",
		ResourceType: "credential",
		ResourceID:   id,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, map[string]any{"rotated": true, "id": id})
}

func (h *PlatformHandler) EnableCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = h.vaultStore.SetStatus(r.Context(), id, "active")
	h.writeSuccess(w, r, map[string]any{"status": "active", "id": id})
}

func (h *PlatformHandler) DisableCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = h.vaultStore.SetStatus(r.Context(), id, "disabled")
	h.writeSuccess(w, r, map[string]any{"status": "disabled", "id": id})
}

func (h *PlatformHandler) TestCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cred, err := h.vaultStore.GetCredential(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Credential not found", nil)
		return
	}

	secret, err := h.vaultStore.GetDecryptedSecret(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "decryption_error", err.Error(), nil)
		return
	}

	adapter, ok := h.registry.Get(cred.ProviderID)
	if !ok {
		h.writeError(w, r, http.StatusBadRequest, "no_adapter", "Provider adapter not found", nil)
		return
	}

	status, err := adapter.HealthCheck(r.Context(), secret)
	if err != nil {
		_ = h.vaultStore.SetHealthState(r.Context(), id, vault.HealthUnhealthy, err.Error())
		h.writeError(w, r, http.StatusBadGateway, "test_failed", err.Error(), status)
		return
	}

	_ = h.vaultStore.SetHealthState(r.Context(), id, vault.HealthHealthy, "")
	h.writeSuccess(w, r, map[string]any{
		"tested":  true,
		"healthy": status.Healthy,
		"latency": status.Latency.Milliseconds(),
		"message": status.Message,
	})
}

func (h *PlatformHandler) GetCredentialUsage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cred, err := h.vaultStore.GetCredential(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Credential not found", nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{
		"id":            id,
		"request_count": cred.RequestCount,
		"error_count":   cred.ErrorCount,
		"last_used_at":  cred.LastUsedAt,
	})
}

func (h *PlatformHandler) GetCredentialHealth(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cred, err := h.vaultStore.GetCredential(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Credential not found", nil)
		return
	}
	inCd, rem, _ := h.limitsEng.IsInCooldown(r.Context(), id)
	h.writeSuccess(w, r, map[string]any{
		"id":                id,
		"health_state":      cred.HealthState,
		"last_error":        cred.LastError,
		"last_validated_at": cred.LastValidated,
		"in_cooldown":       inCd,
		"cooldown_rem_sec":  int(rem.Seconds()),
	})
}
