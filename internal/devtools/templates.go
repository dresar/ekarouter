package devtools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProviderTemplate struct {
	ID                 string    `json:"id"`
	ProviderName       string    `json:"provider_name"`
	BaseURL            string    `json:"base_url"`
	APIStyle           string    `json:"api_style"`
	AuthLocation       string    `json:"auth_location"`
	AuthHeaderName     string    `json:"auth_header_name"`
	AuthQueryParam     string    `json:"auth_query_param"`
	OAuthSupport       bool      `json:"oauth_support"`
	ValidationEndpoint string    `json:"validation_endpoint"`
	RateLimitHeaders   string    `json:"rate_limit_headers"`
	QuotaHeaders       string    `json:"quota_headers"`
	StreamingSupport   bool      `json:"streaming_support"`
	RetryableStatuses  string    `json:"retryable_statuses"`
	CredentialSchema   string    `json:"credential_schema"`
	PaginationStyle    string    `json:"pagination_style"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type RequestTemplate struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ProviderTemplateID string    `json:"provider_template_id"`
	Method             string    `json:"method"`
	Path               string    `json:"path"`
	Headers            string    `json:"headers"`
	QueryParams        string    `json:"query_params"`
	BodySchema         string    `json:"body_schema"`
	CredentialRef      string    `json:"credential_ref"`
	TimeoutMs          int       `json:"timeout_ms"`
	RetryCount         int       `json:"retry_count"`
	RedactionRules     string    `json:"redaction_rules"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TemplateStore struct {
	db *sql.DB
}

func NewTemplateStore(db *sql.DB) *TemplateStore {
	return &TemplateStore{db: db}
}

func (ts *TemplateStore) CreateProviderTemplate(ctx context.Context, pt *ProviderTemplate) error {
	if pt.ID == "" {
		pt.ID = "pt_" + uuid.NewString()[:8]
	}
	now := time.Now()
	pt.CreatedAt = now
	pt.UpdatedAt = now

	_, err := ts.db.ExecContext(ctx, `
INSERT INTO provider_templates (id, provider_name, base_url, api_style, auth_location,
    auth_header_name, auth_query_param, oauth_support, validation_endpoint,
    rate_limit_headers, quota_headers, streaming_support, retryable_statuses,
    credential_schema, pagination_style, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pt.ID, pt.ProviderName, pt.BaseURL, pt.APIStyle, pt.AuthLocation,
		pt.AuthHeaderName, pt.AuthQueryParam, boolToInt(pt.OAuthSupport),
		pt.ValidationEndpoint, pt.RateLimitHeaders, pt.QuotaHeaders,
		boolToInt(pt.StreamingSupport), pt.RetryableStatuses,
		pt.CredentialSchema, pt.PaginationStyle, pt.CreatedAt, pt.UpdatedAt)
	return err
}

func (ts *TemplateStore) ListProviderTemplates(ctx context.Context) ([]ProviderTemplate, error) {
	rows, err := ts.db.QueryContext(ctx, `
SELECT id, provider_name, base_url, api_style, auth_location, auth_header_name,
    auth_query_param, oauth_support, validation_endpoint, rate_limit_headers,
    quota_headers, streaming_support, retryable_statuses, credential_schema,
    pagination_style, created_at, updated_at
FROM provider_templates ORDER BY provider_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []ProviderTemplate
	for rows.Next() {
		var pt ProviderTemplate
		var oauth, streaming int
		if err := rows.Scan(
			&pt.ID, &pt.ProviderName, &pt.BaseURL, &pt.APIStyle, &pt.AuthLocation,
			&pt.AuthHeaderName, &pt.AuthQueryParam, &oauth, &pt.ValidationEndpoint,
			&pt.RateLimitHeaders, &pt.QuotaHeaders, &streaming, &pt.RetryableStatuses,
			&pt.CredentialSchema, &pt.PaginationStyle, &pt.CreatedAt, &pt.UpdatedAt,
		); err != nil {
			continue
		}
		pt.OAuthSupport = oauth == 1
		pt.StreamingSupport = streaming == 1
		templates = append(templates, pt)
	}
	return templates, nil
}

func (ts *TemplateStore) CreateRequestTemplate(ctx context.Context, rt *RequestTemplate) error {
	if rt.ID == "" {
		rt.ID = "rt_" + uuid.NewString()[:8]
	}
	now := time.Now()
	rt.CreatedAt = now
	rt.UpdatedAt = now

	_, err := ts.db.ExecContext(ctx, `
INSERT INTO request_templates (id, name, provider_template_id, method, path, headers,
    query_params, body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
    created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rt.ID, rt.Name, rt.ProviderTemplateID, rt.Method, rt.Path, rt.Headers,
		rt.QueryParams, rt.BodySchema, rt.CredentialRef, rt.TimeoutMs, rt.RetryCount,
		rt.RedactionRules, rt.CreatedAt, rt.UpdatedAt)
	return err
}

func (ts *TemplateStore) ListRequestTemplates(ctx context.Context) ([]RequestTemplate, error) {
	rows, err := ts.db.QueryContext(ctx, `
SELECT id, name, provider_template_id, method, path, headers, query_params,
    body_schema, credential_ref, timeout_ms, retry_count, redaction_rules,
    created_at, updated_at
FROM request_templates ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []RequestTemplate
	for rows.Next() {
		var rt RequestTemplate
		if err := rows.Scan(
			&rt.ID, &rt.Name, &rt.ProviderTemplateID, &rt.Method, &rt.Path,
			&rt.Headers, &rt.QueryParams, &rt.BodySchema, &rt.CredentialRef,
			&rt.TimeoutMs, &rt.RetryCount, &rt.RedactionRules, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			continue
		}
		templates = append(templates, rt)
	}
	return templates, nil
}

func GenerateCurl(rt *RequestTemplate, baseURL string) string {
	var parts []string
	parts = append(parts, "curl")

	if rt.Method != "GET" {
		parts = append(parts, "-X", rt.Method)
	}

	fullURL := strings.TrimRight(baseURL, "/") + rt.Path
	parts = append(parts, fmt.Sprintf("'%s'", fullURL))

	if rt.Headers != "" && rt.Headers != "{}" {
		parts = append(parts, "-H", "'Content-Type: application/json'")
	}

	if rt.CredentialRef != "" {
		parts = append(parts, "-H", fmt.Sprintf("'Authorization: Bearer ${%s}'", rt.CredentialRef))
	}

	if rt.BodySchema != "" {
		parts = append(parts, "-d", fmt.Sprintf("'%s'", rt.BodySchema))
	}

	if rt.TimeoutMs > 0 {
		parts = append(parts, fmt.Sprintf("--max-time %d", rt.TimeoutMs/1000))
	}

	return strings.Join(parts, " \\\n  ")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
