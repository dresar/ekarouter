package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/limits"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/vault"
	"github.com/dresar/ekarouter/internal/webhooks"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlatformHandler struct {
	db          *sql.DB
	registry    *platform.Registry
	vaultStore  *vault.Store
	limitsEng   *limits.Engine
	rotator     *rotator.Rotator
	executor    *executor.Executor
	rbacService *rbac.Service
	auditLogger *audit.Logger
	webhooksMgr *webhooks.Manager
	allowLocal  bool
}

func NewPlatformHandler(
	db *sql.DB,
	registry *platform.Registry,
	vaultStore *vault.Store,
	limitsEng *limits.Engine,
	rotator *rotator.Rotator,
	executor *executor.Executor,
	rbacService *rbac.Service,
	auditLogger *audit.Logger,
	opts ...any,
) *PlatformHandler {
	var allowLocal bool
	var wm *webhooks.Manager
	for _, opt := range opts {
		if b, ok := opt.(bool); ok {
			allowLocal = b
		}
		if m, ok := opt.(*webhooks.Manager); ok {
			wm = m
		}
	}
	if wm == nil {
		wm = webhooks.NewManager(db, allowLocal)
	}

	rows, err := db.Query("SELECT config_json FROM provider_configs")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cfgStr string
			if err := rows.Scan(&cfgStr); err == nil {
				var meta platform.ProviderMetadata
				if err := json.Unmarshal([]byte(cfgStr), &meta); err == nil {
					_ = registry.Register(platform.NewBaseAdapter(meta, 30*time.Second))
				}
			}
		}
	}

	return &PlatformHandler{
		db:          db,
		registry:    registry,
		vaultStore:  vaultStore,
		limitsEng:   limitsEng,
		rotator:     rotator,
		executor:    executor,
		rbacService: rbacService,
		auditLogger: auditLogger,
		webhooksMgr: wm,
		allowLocal:  allowLocal,
	}
}

func (h *PlatformHandler) writeSuccess(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	reqID := GetRequestID(r.Context())
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
		"meta": map[string]any{
			"request_id": reqID,
		},
	})
}

func (h *PlatformHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	reqID := GetRequestID(r.Context())
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": details,
		},
		"meta": map[string]any{
			"request_id": reqID,
		},
	})
}

func (h *PlatformHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	q := r.URL.Query().Get("q")

	var list []platform.ProviderMetadata
	if q != "" {
		list = h.registry.Search(q)
	} else if cat != "" {
		list = h.registry.ListByCategory(platform.Category(cat))
	} else {
		list = h.registry.List()
	}

	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adapter, ok := h.registry.Get(id)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}
	h.writeSuccess(w, r, adapter.Metadata())
}

func (h *PlatformHandler) ValidateProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adapter, ok := h.registry.Get(id)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}

	var req struct {
		Secret string `json:"secret"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	valid, msg, err := adapter.ValidateCredential(r.Context(), req.Secret)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "validation_failed", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{
		"valid":   valid,
		"message": msg,
	})
}

func (h *PlatformHandler) HealthProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adapter, ok := h.registry.Get(id)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}

	status, err := adapter.HealthCheck(r.Context(), "")
	if err != nil {
		h.writeError(w, r, http.StatusServiceUnavailable, "unhealthy", err.Error(), status)
		return
	}

	h.writeSuccess(w, r, status)
}

func (h *PlatformHandler) GetProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adapter, ok := h.registry.Get(id)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{
		"provider_id":  id,
		"capabilities": adapter.Metadata().Capabilities,
	})
}

func (h *PlatformHandler) GetProviderDocs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	adapter, ok := h.registry.Get(id)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}
	m := adapter.Metadata()
	h.writeSuccess(w, r, map[string]any{
		"provider_id":       id,
		"name":              m.Name,
		"website_url":       m.WebsiteURL,
		"docs_url":          m.DocsURL,
		"api_reference_url": m.APIReferenceURL,
		"free_tier_status":  m.FreeTierStatus,
		"free_tier_notes":   m.FreeTierNotes,
		"auth_type":         m.AuthType,
		"required_creds":    m.RequiredCredentials,
		"supported_ops":     m.SupportedOperations,
	})
}

func (h *PlatformHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID                  string                `json:"id"`
		Name                string                `json:"name"`
		Category            platform.Category     `json:"category"`
		Description         string                `json:"description"`
		BaseURL             string                `json:"base_url"`
		AuthType            platform.AuthType     `json:"auth_type"`
		AuthHeaderName      string                `json:"auth_header_name"`
		AuthHeaderPrefix    string                `json:"auth_header_prefix"`
		RequiredCredentials []string              `json:"required_credentials"`
		SupportedOperations []string              `json:"supported_operations"`
		Capabilities        []platform.Capability `json:"capabilities"`
		WebsiteURL          string                `json:"website_url"`
		DocsURL             string                `json:"docs_url"`
		APIReferenceURL     string                `json:"api_reference_url"`
		TimeoutMS           int                   `json:"timeout_ms"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON body", nil)
		return
	}

	if body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing_fields", "Provider name is required", nil)
		return
	}

	if body.ID == "" {
		body.ID = strings.ToLower(strings.ReplaceAll(body.Name, " ", "-"))
	}
	if body.Category == "" {
		body.Category = platform.CategoryCustom
	}
	if body.AuthType == "" {
		body.AuthType = platform.AuthTypeAPIKeyHeader
	}

	if body.BaseURL != "" && !h.allowLocal {
		if err := platform.ValidateSSRF(body.BaseURL); err != nil {
			h.writeError(w, r, http.StatusBadRequest, "ssrf_violation", err.Error(), nil)
			return
		}
	}

	timeout := 30 * time.Second
	if body.TimeoutMS > 0 {
		timeout = time.Duration(body.TimeoutMS) * time.Millisecond
	}

	meta := platform.ProviderMetadata{
		ID:                  body.ID,
		Name:                body.Name,
		Category:            body.Category,
		Description:         body.Description,
		BaseURL:             body.BaseURL,
		AuthType:            body.AuthType,
		AuthHeaderName:      body.AuthHeaderName,
		AuthHeaderPrefix:    body.AuthHeaderPrefix,
		RequiredCredentials: body.RequiredCredentials,
		SupportedOperations: body.SupportedOperations,
		Capabilities:        body.Capabilities,
		WebsiteURL:          body.WebsiteURL,
		DocsURL:             body.DocsURL,
		APIReferenceURL:     body.APIReferenceURL,
		FreeTierStatus:      "available",
		Enabled:             true,
	}

	adapter := platform.NewBaseAdapter(meta, timeout)
	if err := h.registry.Register(adapter); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "registration_failed", err.Error(), nil)
		return
	}

	metaBytes, _ := json.Marshal(meta)
	_, _ = h.db.ExecContext(r.Context(), `
INSERT INTO provider_configs (id, provider_id, config_json, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(provider_id) DO UPDATE SET config_json = excluded.config_json, updated_at = CURRENT_TIMESTAMP`,
		uuid.New().String(), body.ID, string(metaBytes))

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "provider.create",
		ResourceType: "provider",
		ResourceID:   body.ID,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, meta)
}
