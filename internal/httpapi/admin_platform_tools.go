package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

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
		var ca, ua any
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

	var ptID any
	if body.ProviderTemplateID != "" {
		ptID = body.ProviderTemplateID
	}

	_, err := h.db.ExecContext(r.Context(), `
INSERT INTO request_templates (id, name, provider_template_id, method, path, headers,
    query_params, body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
    created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, body.Name, ptID, body.Method, body.Path, body.Headers,
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
	var ca, ua any

	err := h.db.QueryRowContext(r.Context(), `
SELECT name, COALESCE(provider_template_id, ''), method, path, headers, query_params,
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

	var name, ptID, method, pathTpl, hdrsStr, qryStr, credRef string
	var timeoutMs int
	err := h.db.QueryRowContext(r.Context(), `
SELECT name, COALESCE(provider_template_id, ''), method, path, headers, query_params, credential_ref, timeout_ms 
FROM request_templates WHERE id = ?`, id).Scan(&name, &ptID, &method, &pathTpl, &hdrsStr, &qryStr, &credRef, &timeoutMs)
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "not_found", "Request template not found", nil)
		return
	}

	finalURL := executor.InterpolateString(pathTpl, execParams.Variables)
	if !h.allowLocal {
		if err := platform.ValidateSSRF(finalURL); err != nil {
			h.writeError(w, r, http.StatusBadRequest, "ssrf_violation", err.Error(), nil)
			return
		}
	}

	u, err := url.Parse(finalURL)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid_url", err.Error(), nil)
		return
	}

	var headersTpl map[string]string
	_ = json.Unmarshal([]byte(hdrsStr), &headersTpl)
	finalHeaders, err := executor.InterpolateHeaders(headersTpl, execParams.Variables)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "header_error", err.Error(), nil)
		return
	}

	var queryTpl map[string]string
	_ = json.Unmarshal([]byte(qryStr), &queryTpl)
	q := u.Query()
	for k, v := range queryTpl {
		q.Set(k, executor.InterpolateString(v, execParams.Variables))
	}
	for k, v := range execParams.QueryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	var secret string
	if credRef != "" {
		if sec, decErr := h.vaultStore.GetDecryptedSecret(r.Context(), credRef); decErr == nil {
			secret = sec
		}
	} else if ptID != "" {
		creds, cErr := h.vaultStore.ListCredentials(r.Context(), ptID, execParams.ProjectID, execParams.Environment)
		if cErr == nil && len(creds) == 0 && execParams.ProjectID != "" {
			creds, cErr = h.vaultStore.ListCredentials(r.Context(), ptID, "", execParams.Environment)
		}
		if cErr == nil && len(creds) > 0 {
			sel, selErr := h.rotator.Select(creds, rotator.StrategyPriority)
			if selErr == nil && sel != nil {
				if sec, decErr := h.vaultStore.GetDecryptedSecret(r.Context(), sel.ID); decErr == nil {
					secret = sec
					_ = h.vaultStore.RecordUsage(r.Context(), sel.ID, false)
				}
			}
		}
	}

	var bodyReader io.Reader
	if len(execParams.Body) > 0 {
		bodyReader = bytes.NewReader(execParams.Body)
	}

	req, err := http.NewRequestWithContext(r.Context(), method, u.String(), bodyReader)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "request_creation_failed", err.Error(), nil)
		return
	}

	for k, v := range finalHeaders {
		req.Header.Set(k, v)
	}
	if secret != "" && req.Header.Get("Authorization") == "" {
		if strings.HasPrefix(strings.ToLower(secret), "bearer ") || strings.HasPrefix(strings.ToLower(secret), "basic ") {
			req.Header.Set("Authorization", secret)
		} else {
			req.Header.Set("Authorization", "Bearer "+secret)
		}
	}
	if len(execParams.Body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	timeout := 30 * time.Second
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	client := platform.NewSafeHTTPClient(timeout, h.allowLocal)
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		h.auditLogger.Log(&audit.Record{
			ActorID:      "user",
			Action:       "request_template.execute",
			ResourceType: "request_template",
			ResourceID:   id,
			ProjectID:    execParams.ProjectID,
			RequestID:    GetRequestID(r.Context()),
			Result:       "error",
		})
		h.writeError(w, r, http.StatusBadGateway, "execution_error", err.Error(), map[string]any{
			"template_id":  id,
			"resolved_url": u.String(),
			"latency_ms":   latency.Milliseconds(),
		})
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	respHeaders := make(map[string]string)
	for k := range resp.Header {
		respHeaders[k] = resp.Header.Get(k)
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
		"resolved_url": u.String(),
		"method":       method,
		"status_code":  resp.StatusCode,
		"latency_ms":   latency.Milliseconds(),
		"headers":      respHeaders,
		"body":         string(respBytes),
	})
}
