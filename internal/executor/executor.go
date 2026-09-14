package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/vault"
	"github.com/google/uuid"
)

type ToolDefinition struct {
	ID                  string            `json:"id"`
	ProviderID          string            `json:"provider_id"`
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	Category            string            `json:"category"`
	Method              string            `json:"method"`
	URLTemplate         string            `json:"url_template"`
	HeadersTemplate     map[string]string `json:"headers_template"`
	QueryTemplate       map[string]string `json:"query_template"`
	BodySchema          string            `json:"body_schema,omitempty"`
	TimeoutMS           int               `json:"timeout_ms"`
	MaxRetries          int               `json:"max_retries"`
	RequiredPermissions string            `json:"required_permissions"`
	Enabled             bool              `json:"enabled"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type ExecutionParams struct {
	ToolID      string            `json:"tool_id"`
	Variables   map[string]string `json:"variables"`
	QueryParams map[string]string `json:"query_params"`
	Body        []byte            `json:"body"`
	ProjectID   string            `json:"project_id,omitempty"`
	Environment string            `json:"environment,omitempty"`
}

type ExecutionResult struct {
	ID         string            `json:"id"`
	ToolID     string            `json:"tool_id"`
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	LatencyMS  int64             `json:"latency_ms"`
	Error      string            `json:"error,omitempty"`
	ExecutedAt time.Time         `json:"executed_at"`
}

type Executor struct {
	db         *sql.DB
	registry   *platform.Registry
	vault      *vault.Store
	allowLocal bool
	rotator    *rotator.Rotator
}

func NewExecutor(db *sql.DB, registry *platform.Registry, vaultStore *vault.Store, allowLocal bool, rot ...*rotator.Rotator) *Executor {
	var r *rotator.Rotator
	if len(rot) > 0 && rot[0] != nil {
		r = rot[0]
	} else {
		r = rotator.NewRotator()
	}
	return &Executor{
		db:         db,
		registry:   registry,
		vault:      vaultStore,
		allowLocal: allowLocal,
		rotator:    r,
	}
}

func (e *Executor) CreateTool(ctx context.Context, tool *ToolDefinition) error {
	if tool.Name == "" || tool.ProviderID == "" || tool.URLTemplate == "" {
		return errors.New("name, provider_id, and url_template are required")
	}
	if tool.ID == "" {
		tool.ID = uuid.New().String()
	}
	if tool.Method == "" {
		tool.Method = "GET"
	}
	if tool.TimeoutMS <= 0 {
		tool.TimeoutMS = 30000
	}
	now := time.Now().UTC()
	tool.CreatedAt = now
	tool.UpdatedAt = now

	hdrs, _ := json.Marshal(tool.HeadersTemplate)
	qry, _ := json.Marshal(tool.QueryTemplate)

	query := `
INSERT INTO tool_definitions (
    id, provider_id, name, description, category, method, url_template,
    headers_template, query_template, body_schema, timeout_ms, retry_policy,
    rate_limit_policy, required_permissions, enabled, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{"max_retries": 2}', '{}', ?, ?, ?, ?)`

	_, err := e.db.ExecContext(ctx, query,
		tool.ID, tool.ProviderID, tool.Name, tool.Description, tool.Category,
		tool.Method, tool.URLTemplate, string(hdrs), string(qry), tool.BodySchema,
		tool.TimeoutMS, tool.RequiredPermissions, 1, now, now,
	)
	return err
}

func (e *Executor) GetTool(ctx context.Context, id string) (*ToolDefinition, error) {
	query := `
SELECT id, provider_id, name, description, category, method, url_template,
       headers_template, query_template, body_schema, timeout_ms,
       required_permissions, enabled, created_at, updated_at
FROM tool_definitions WHERE id = ?`

	var t ToolDefinition
	var hdrs, qry string
	var enabledInt int

	err := e.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.ProviderID, &t.Name, &t.Description, &t.Category, &t.Method,
		&t.URLTemplate, &hdrs, &qry, &t.BodySchema, &t.TimeoutMS,
		&t.RequiredPermissions, &enabledInt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.Enabled = enabledInt == 1
	_ = json.Unmarshal([]byte(hdrs), &t.HeadersTemplate)
	_ = json.Unmarshal([]byte(qry), &t.QueryTemplate)
	return &t, nil
}

func (e *Executor) ListTools(ctx context.Context, providerID string) ([]*ToolDefinition, error) {
	query := `
SELECT id, provider_id, name, description, category, method, url_template,
       headers_template, query_template, body_schema, timeout_ms,
       required_permissions, enabled, created_at, updated_at
FROM tool_definitions`
	var args []any
	if providerID != "" {
		query += " WHERE provider_id = ?"
		args = append(args, providerID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ToolDefinition
	for rows.Next() {
		var t ToolDefinition
		var hdrs, qry string
		var enabledInt int
		if err := rows.Scan(
			&t.ID, &t.ProviderID, &t.Name, &t.Description, &t.Category, &t.Method,
			&t.URLTemplate, &hdrs, &qry, &t.BodySchema, &t.TimeoutMS,
			&t.RequiredPermissions, &enabledInt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		t.Enabled = enabledInt == 1
		_ = json.Unmarshal([]byte(hdrs), &t.HeadersTemplate)
		_ = json.Unmarshal([]byte(qry), &t.QueryTemplate)
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (e *Executor) ExecuteTool(ctx context.Context, params *ExecutionParams) (*ExecutionResult, error) {
	tool, err := e.GetTool(ctx, params.ToolID)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %w", err)
	}
	if !tool.Enabled {
		return nil, errors.New("tool is disabled")
	}

	adapter, ok := e.registry.Get(tool.ProviderID)
	if !ok {
		return nil, fmt.Errorf("provider adapter %q not found", tool.ProviderID)
	}

	env := params.Environment
	if env == "" {
		env = "production"
	}

	var secret string
	var selectedCredID string
	if e.vault != nil {
		creds, err := e.vault.ListCredentials(ctx, tool.ProviderID, params.ProjectID, env)
		if err == nil && len(creds) == 0 && params.ProjectID != "" {
			creds, err = e.vault.ListCredentials(ctx, tool.ProviderID, "", env)
		}
		if err == nil && len(creds) > 0 {
			rot := e.rotator
			if rot == nil {
				rot = rotator.NewRotator()
			}
			sel, selErr := rot.Select(creds, rotator.StrategyPriority)
			if selErr == nil && sel != nil {
				sec, decErr := e.vault.GetDecryptedSecret(ctx, sel.ID)
				if decErr == nil {
					secret = sec
					selectedCredID = sel.ID
				}
			}
		}
	}

	finalURL := InterpolateString(tool.URLTemplate, params.Variables)
	if !e.allowLocal {
		if err := platform.ValidateSSRF(finalURL); err != nil {
			return nil, fmt.Errorf("SSRF rejection: %w", err)
		}
	}

	finalHeaders, err := InterpolateHeaders(tool.HeadersTemplate, params.Variables)
	if err != nil {
		return nil, fmt.Errorf("invalid headers: %w", err)
	}

	mergedQuery := make(map[string]string)
	for k, v := range tool.QueryTemplate {
		mergedQuery[k] = InterpolateString(v, params.Variables)
	}
	for k, v := range params.QueryParams {
		mergedQuery[k] = v
	}

	execReq := &platform.ExecutionRequest{
		Path:             finalURL,
		Method:           tool.Method,
		Headers:          finalHeaders,
		QueryParams:      mergedQuery,
		Body:             params.Body,
		CredentialSecret: secret,
	}

	timeout := time.Duration(tool.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	resp, err := adapter.Execute(execCtx, execReq)
	latency := time.Since(start)

	if selectedCredID != "" && e.vault != nil {
		hasError := err != nil || (resp != nil && resp.StatusCode >= 500)
		_ = e.vault.RecordUsage(ctx, selectedCredID, hasError)
	}

	execID := uuid.New().String()
	res := &ExecutionResult{
		ID:         execID,
		ToolID:     tool.ID,
		LatencyMS:  latency.Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}

	var projID any
	if params.ProjectID != "" {
		projID = params.ProjectID
	}

	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		_, _ = e.db.ExecContext(ctx, `
INSERT INTO tool_executions (id, tool_id, project_id, environment, status, latency_ms, response_status, error_message, executed_at)
VALUES (?, ?, ?, ?, 'error', ?, 0, ?, ?)`, execID, tool.ID, projID, env, res.LatencyMS, res.Error, res.ExecutedAt)
		return res, err
	}

	res.Status = "success"
	res.StatusCode = resp.StatusCode
	res.Headers = resp.Headers
	res.Body = resp.Body

	_, _ = e.db.ExecContext(ctx, `
INSERT INTO tool_executions (id, tool_id, project_id, environment, status, latency_ms, response_status, error_message, executed_at)
VALUES (?, ?, ?, ?, 'success', ?, ?, '', ?)`, execID, tool.ID, projID, env, res.LatencyMS, res.StatusCode, res.ExecutedAt)

	return res, nil
}
