package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/limits"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/vault"
	"github.com/go-chi/chi/v5"
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
) *PlatformHandler {
	return &PlatformHandler{
		db:          db,
		registry:    registry,
		vaultStore:  vaultStore,
		limitsEng:   limitsEng,
		rotator:     rotator,
		executor:    executor,
		rbacService: rbacService,
		auditLogger: auditLogger,
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
