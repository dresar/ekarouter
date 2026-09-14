package httpapi

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/limits"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/vault"
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

func (h *PlatformHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	list, err := h.rbacService.ListProjects(r.Context())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "database_error", err.Error(), nil)
		return
	}
	if list == nil {
		list = []*rbac.Project{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Project name is required", nil)
		return
	}

	p, err := h.rbacService.CreateProject(r.Context(), body.Name, "admin", body.Environment, body.Description)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "project_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, p)
}

func (h *PlatformHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.rbacService.GetProject(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found", nil)
		return
	}
	h.writeSuccess(w, r, p)
}

func (h *PlatformHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}
	if err := h.rbacService.UpdateProject(r.Context(), id, body.Name, body.Environment, body.Description); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.rbacService.DeleteProject(r.Context(), id); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "delete_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT id, name, COALESCE(project_id, ''), description, created_at FROM environments ORDER BY created_at DESC")
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, name, pID, desc string
		var ca time.Time
		_ = rows.Scan(&id, &name, &pID, &desc, &ca)
		list = append(list, map[string]any{
			"id":          id,
			"name":        name,
			"project_id":  pID,
			"description": desc,
			"created_at":  ca,
		})
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		ProjectID   string `json:"project_id"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Environment name is required", nil)
		return
	}

	envID := uuid.New().String()
	var pID any
	if body.ProjectID != "" {
		pID = body.ProjectID
	}
	_, err := h.db.ExecContext(r.Context(), `
INSERT INTO environments (id, name, project_id, description, created_at)
VALUES (?, ?, ?, ?, ?)`, envID, body.Name, pID, body.Description, time.Now().UTC())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{
		"id":          envID,
		"name":        body.Name,
		"project_id":  body.ProjectID,
		"description": body.Description,
	})
}

func (h *PlatformHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}
	_, err := h.db.ExecContext(r.Context(), "UPDATE environments SET name = ?, description = ? WHERE id = ?", body.Name, body.Description, id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.ExecContext(r.Context(), "DELETE FROM environments WHERE id = ?", id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "delete_error", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	pID := r.URL.Query().Get("provider_id")
	tools, err := h.executor.ListTools(r.Context(), pID)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "tool_error", err.Error(), nil)
		return
	}
	if tools == nil {
		tools = []*executor.ToolDefinition{}
	}
	h.writeSuccess(w, r, tools)
}

func (h *PlatformHandler) CreateTool(w http.ResponseWriter, r *http.Request) {
	var tool executor.ToolDefinition
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	if err := h.executor.CreateTool(r.Context(), &tool); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "create_tool_error", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, tool)
}

func (h *PlatformHandler) GetTool(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tool, err := h.executor.GetTool(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "tool_not_found", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, tool)
}

func (h *PlatformHandler) GetToolSchema(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tool, err := h.executor.GetTool(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "tool_not_found", err.Error(), nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{
		"tool_id":              tool.ID,
		"name":                 tool.Name,
		"provider_id":          tool.ProviderID,
		"method":               tool.Method,
		"url_template":         tool.URLTemplate,
		"headers_template":     tool.HeadersTemplate,
		"query_template":       tool.QueryTemplate,
		"body_schema":          tool.BodySchema,
		"required_permissions": tool.RequiredPermissions,
	})
}

func (h *PlatformHandler) ExecuteTool(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var params executor.ExecutionParams
	_ = json.NewDecoder(r.Body).Decode(&params)
	params.ToolID = id

	res, err := h.executor.ExecuteTool(r.Context(), &params)
	if err != nil {
		h.writeError(w, r, http.StatusBadGateway, "execution_error", err.Error(), res)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "tool.execute",
		ResourceType: "tool",
		ResourceID:   id,
		ProjectID:    params.ProjectID,
		RequestID:    GetRequestID(r.Context()),
		Result:       res.Status,
	})

	h.writeSuccess(w, r, res)
}

func (h *PlatformHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	subPath := chi.URLParam(r, "*")

	adapter, ok := h.registry.Get(providerID)
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "provider_not_found", "Provider not registered", nil)
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)

	var secret string
	creds, err := h.vaultStore.ListCredentials(r.Context(), providerID, "", "production")
	if err == nil && len(creds) > 0 {
		sel, err := h.rotator.Select(creds, rotator.StrategyPriority)
		if err == nil {
			sec, err := h.vaultStore.GetDecryptedSecret(r.Context(), sel.ID)
			if err == nil {
				secret = sec
				_ = h.vaultStore.RecordUsage(r.Context(), sel.ID, false)
			}
		}
	}

	headers := make(map[string]string)
	for k := range r.Header {
		if strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "Host") {
			continue
		}
		headers[k] = r.Header.Get(k)
	}

	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	execReq := &platform.ExecutionRequest{
		Path:             subPath,
		Method:           r.Method,
		Headers:          headers,
		QueryParams:      queryParams,
		Body:             bodyBytes,
		CredentialSecret: secret,
	}

	resp, err := adapter.Execute(r.Context(), execReq)
	if err != nil {
		h.writeError(w, r, http.StatusBadGateway, "proxy_error", err.Error(), nil)
		return
	}

	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(resp.Body)
}

func (h *PlatformHandler) HealthSummary(w http.ResponseWriter, r *http.Request) {
	h.writeSuccess(w, r, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"database": "connected",
			"vault":    "operational",
			"proxy":    "ready",
		},
	})
}

func (h *PlatformHandler) HealthProviders(w http.ResponseWriter, r *http.Request) {
	list := h.registry.List()
	results := make([]map[string]any, 0, len(list))
	for _, p := range list {
		results = append(results, map[string]any{
			"provider_id": p.ID,
			"name":        p.Name,
			"category":    p.Category,
			"status":      "online",
		})
	}
	h.writeSuccess(w, r, results)
}

func (h *PlatformHandler) HealthCredentials(w http.ResponseWriter, r *http.Request) {
	creds, _ := h.vaultStore.ListCredentials(r.Context(), "", "", "")
	res := make([]map[string]any, 0, len(creds))
	for _, c := range creds {
		res = append(res, map[string]any{
			"id":           c.ID,
			"name":         c.Name,
			"provider_id":  c.ProviderID,
			"health_state": c.HealthState,
			"status":       c.Status,
		})
	}
	h.writeSuccess(w, r, res)
}

func (h *PlatformHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	act := r.URL.Query().Get("action")
	actor := r.URL.Query().Get("actor_id")
	resType := r.URL.Query().Get("resource_type")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	logs, err := h.auditLogger.Query(r.Context(), act, actor, resType, limit)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "audit_query_error", err.Error(), nil)
		return
	}
	if logs == nil {
		logs = []*audit.Record{}
	}
	h.writeSuccess(w, r, logs)
}

func (h *PlatformHandler) GetUsageSummary(w http.ResponseWriter, r *http.Request) {
	var totalReqs, totalErrors int64
	_ = h.db.QueryRowContext(r.Context(), "SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(error_count), 0) FROM vault_credentials").Scan(&totalReqs, &totalErrors)

	var activeCreds int
	_ = h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM vault_credentials WHERE status = 'active'").Scan(&activeCreds)

	h.writeSuccess(w, r, map[string]any{
		"total_requests":     totalReqs,
		"total_errors":       totalErrors,
		"active_credentials": activeCreds,
	})
}

func (h *PlatformHandler) GetUsageProviders(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT provider_id, COUNT(*) as cred_count,
       COALESCE(SUM(request_count), 0) as total_requests,
       COALESCE(SUM(error_count), 0) as total_errors
FROM vault_credentials GROUP BY provider_id ORDER BY total_requests DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var pID string
		var credCount int
		var reqs, errs int64
		if err := rows.Scan(&pID, &credCount, &reqs, &errs); err == nil {
			list = append(list, map[string]any{
				"provider_id":        pID,
				"active_credentials": credCount,
				"total_requests":     reqs,
				"total_errors":       errs,
			})
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) GetUsageCredentials(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT id, name, provider_id, COALESCE(project_id, ''), environment,
       request_count, error_count, last_used_at
FROM vault_credentials ORDER BY request_count DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, name, pID, projID, env string
		var reqs, errs int64
		var lu sql.NullTime
		if err := rows.Scan(&id, &name, &pID, &projID, &env, &reqs, &errs, &lu); err == nil {
			item := map[string]any{
				"id":            id,
				"name":          name,
				"provider_id":   pID,
				"project_id":    projID,
				"environment":   env,
				"request_count": reqs,
				"error_count":   errs,
			}
			if lu.Valid {
				item["last_used_at"] = lu.Time
			}
			list = append(list, item)
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) GetUsageProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT COALESCE(project_id, 'default') as proj_id,
       COUNT(*) as cred_count,
       COALESCE(SUM(request_count), 0) as total_requests,
       COALESCE(SUM(error_count), 0) as total_errors
FROM vault_credentials GROUP BY project_id ORDER BY total_requests DESC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var pID string
		var credCount int
		var reqs, errs int64
		if err := rows.Scan(&pID, &credCount, &reqs, &errs); err == nil {
			list = append(list, map[string]any{
				"project_id":         pID,
				"active_credentials": credCount,
				"total_requests":     reqs,
				"total_errors":       errs,
			})
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) ListRequestTemplates(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
SELECT id, name, provider_template_id, method, path, headers, query_params,
       body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
       created_at, updated_at
FROM request_templates ORDER BY name ASC`)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "db_error", err.Error(), nil)
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, name, ptID, method, path, hdrs, qry, bodySch, credRef, redact string
		var timeout, retry int
		var ca, ua time.Time
		if err := rows.Scan(&id, &name, &ptID, &method, &path, &hdrs, &qry, &bodySch, &credRef, &timeout, &retry, &redact, &ca, &ua); err == nil {
			list = append(list, map[string]any{
				"id":                   id,
				"name":                 name,
				"provider_template_id": ptID,
				"method":               method,
				"path":                 path,
				"headers":              hdrs,
				"query_params":         qry,
				"body_schema":          bodySch,
				"credential_ref":       credRef,
				"timeout_ms":           timeout,
				"retry_count":          retry,
				"redaction_rules":      redact,
				"created_at":           ca,
				"updated_at":           ua,
			})
		}
	}
	if list == nil {
		list = []map[string]any{}
	}
	h.writeSuccess(w, r, list)
}

func (h *PlatformHandler) CreateRequestTemplate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name               string `json:"name"`
		ProviderTemplateID string `json:"provider_template_id"`
		Method             string `json:"method"`
		Path               string `json:"path"`
		Headers            string `json:"headers"`
		QueryParams        string `json:"query_params"`
		BodySchema         string `json:"body_schema"`
		CredentialRef      string `json:"credential_ref"`
		TimeoutMs          int    `json:"timeout_ms"`
		RetryCount         int    `json:"retry_count"`
		RedactionRules     string `json:"redaction_rules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		h.writeError(w, r, http.StatusBadRequest, "invalid_request", "Template name is required", nil)
		return
	}

	id := "rt_" + uuid.NewString()[:8]
	if body.Method == "" {
		body.Method = "GET"
	}
	if body.TimeoutMs <= 0 {
		body.TimeoutMs = 30000
	}
	now := time.Now().UTC()

	_, err := h.db.ExecContext(r.Context(), `
INSERT INTO request_templates (id, name, provider_template_id, method, path, headers,
    query_params, body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
    created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, body.Name, body.ProviderTemplateID, body.Method, body.Path, body.Headers,
		body.QueryParams, body.BodySchema, body.CredentialRef, body.TimeoutMs, body.RetryCount,
		body.RedactionRules, now, now)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "insert_error", err.Error(), nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{
		"id":                   id,
		"name":                 body.Name,
		"provider_template_id": body.ProviderTemplateID,
		"method":               body.Method,
		"path":                 body.Path,
	})
}

func (h *PlatformHandler) GetRequestTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name, ptID, method, path, hdrs, qry, bodySch, credRef, redact string
	var timeout, retry int
	var ca, ua time.Time

	err := h.db.QueryRowContext(r.Context(), `
SELECT name, provider_template_id, method, path, headers, query_params,
       body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
       created_at, updated_at
FROM request_templates WHERE id = ?`, id).Scan(&name, &ptID, &method, &path, &hdrs, &qry, &bodySch, &credRef, &timeout, &retry, &redact, &ca, &ua)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Request template not found", nil)
		return
	}

	h.writeSuccess(w, r, map[string]any{
		"id":                   id,
		"name":                 name,
		"provider_template_id": ptID,
		"method":               method,
		"path":                 path,
		"headers":              hdrs,
		"query_params":         qry,
		"body_schema":          bodySch,
		"credential_ref":       credRef,
		"timeout_ms":           timeout,
		"retry_count":          retry,
		"redaction_rules":      redact,
		"created_at":           ca,
		"updated_at":           ua,
	})
}

func (h *PlatformHandler) UpdateRequestTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Method      string `json:"method"`
		Path        string `json:"path"`
		Headers     string `json:"headers"`
		QueryParams string `json:"query_params"`
		BodySchema  string `json:"body_schema"`
		TimeoutMs   int    `json:"timeout_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_json", "Malformed JSON", nil)
		return
	}

	now := time.Now().UTC()
	res, err := h.db.ExecContext(r.Context(), `
UPDATE request_templates
SET name = COALESCE(NULLIF(?, ''), name),
    method = COALESCE(NULLIF(?, ''), method),
    path = COALESCE(NULLIF(?, ''), path),
    headers = COALESCE(NULLIF(?, ''), headers),
    query_params = COALESCE(NULLIF(?, ''), query_params),
    body_schema = COALESCE(NULLIF(?, ''), body_schema),
    updated_at = ?
WHERE id = ?`, body.Name, body.Method, body.Path, body.Headers, body.QueryParams, body.BodySchema, now, id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "update_error", err.Error(), nil)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Request template not found", nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"updated": true, "id": id})
}

func (h *PlatformHandler) DeleteRequestTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.db.ExecContext(r.Context(), "DELETE FROM request_templates WHERE id = ?", id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "delete_error", err.Error(), nil)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Request template not found", nil)
		return
	}
	h.writeSuccess(w, r, map[string]any{"deleted": true, "id": id})
}

func (h *PlatformHandler) ExecuteRequestTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var execParams struct {
		Variables   map[string]string `json:"variables"`
		QueryParams map[string]string `json:"query_params"`
		Body        json.RawMessage   `json:"body"`
		ProjectID   string            `json:"project_id"`
		Environment string            `json:"environment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&execParams)

	var name, ptID, method, pathTpl string
	var timeoutMs int
	err := h.db.QueryRowContext(r.Context(), "SELECT name, provider_template_id, method, path, timeout_ms FROM request_templates WHERE id = ?", id).Scan(&name, &ptID, &method, &pathTpl, &timeoutMs)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Request template not found", nil)
		return
	}

	finalURL := executor.InterpolateString(pathTpl, execParams.Variables)
	if err := platform.ValidateSSRF(finalURL); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "ssrf_violation", err.Error(), nil)
		return
	}

	h.auditLogger.Log(&audit.Record{
		ActorID:      "user",
		Action:       "request_template.execute",
		ResourceType: "request_template",
		ResourceID:   id,
		ProjectID:    execParams.ProjectID,
		RequestID:    GetRequestID(r.Context()),
		Result:       "success",
	})

	h.writeSuccess(w, r, map[string]any{
		"template_id":  id,
		"name":         name,
		"resolved_url": finalURL,
		"method":       method,
		"status":       "executed",
	})
}

