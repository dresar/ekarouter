package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/webhooks"
	"github.com/go-chi/chi/v5"
)

func (h *PlatformHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	list, err := h.webhooksMgr.ListWebhooks(r.Context())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	if list == nil {
		list = []*webhooks.Webhook{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string `json:"name"`
		TargetURL string `json:"target_url"`
		Events    string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON body", nil)
		return
	}

	if body.Name == "" || body.TargetURL == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing_fields", "Name and target_url are required", nil)
		return
	}

	secret, err := webhooks.GenerateSecret()
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "secret_generation_error", err.Error(), nil)
		return
	}

	encSecret, err := h.vaultStore.Vault().Encrypt(secret)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "encryption_error", err.Error(), nil)
		return
	}

	wh, err := h.webhooksMgr.CreateWebhook(r.Context(), body.Name, body.TargetURL, body.Events, encSecret)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "webhook_creation_failed", err.Error(), nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "webhook.create",
		ResourceType: "webhook",
		ResourceID:   wh.ID,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, map[string]any{
		"webhook": wh,
		"secret":  secret,
	})
}

func (h *PlatformHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var wh webhooks.Webhook
	var enInt int
	err := h.db.QueryRowContext(r.Context(), "SELECT id, name, target_url, events, enabled, created_at, updated_at FROM webhooks WHERE id = ?", id).Scan(
		&wh.ID, &wh.Name, &wh.TargetURL, &wh.Events, &enInt, &wh.CreatedAt, &wh.UpdatedAt,
	)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Webhook not found", nil)
		return
	}
	wh.Enabled = enInt == 1
	h.writeSuccess(w, r, wh)
}

func (h *PlatformHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.db.ExecContext(r.Context(), "DELETE FROM webhooks WHERE id = ?", id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Webhook not found", nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "webhook.delete",
		ResourceType: "webhook",
		ResourceID:   id,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var encSecret string
	err := h.db.QueryRowContext(r.Context(), "SELECT secret_encrypted FROM webhooks WHERE id = ?", id).Scan(&encSecret)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Webhook not found", nil)
		return
	}

	rawSecret, err := h.vaultStore.Vault().Decrypt(encSecret)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "decryption_error", err.Error(), nil)
		return
	}

	reqID := GetRequestID(r.Context())
	payload := []byte(`{"event":"ping","message":"EkaRouter webhook test delivery","request_id":"` + reqID + `"}`)
	delivery, err := h.webhooksMgr.Dispatch(r.Context(), id, rawSecret, "ping", payload)
	if err != nil && delivery == nil {
		h.writeError(w, r, http.StatusBadGateway, "dispatch_failed", err.Error(), nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "webhook.test",
		ResourceType: "webhook",
		ResourceID:   id,
		RequestID:    reqID,
		Result:       "success",
	})

	h.writeSuccess(w, r, delivery)
}

func (h *PlatformHandler) GetWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	list, err := h.webhooksMgr.ListDeliveries(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	if list == nil {
		list = []*webhooks.Delivery{}
	}
	h.writeSuccess(w, r, list)
}
