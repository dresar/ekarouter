package devtools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/dresar/ekarouter/internal/platform"
	"github.com/google/uuid"
)

type APIOperation struct {
	ID                 string `json:"id"`
	ProviderTemplateID string `json:"provider_template_id"`
	Method             string `json:"method"`
	Path               string `json:"path"`
	Summary            string `json:"summary"`
	IsDestructive      bool   `json:"is_destructive"`
	Confirmed          bool   `json:"confirmed"`
	AuthSchemes        string `json:"auth_schemes"`
	Parameters         string `json:"parameters"`
	RequestBodySchema  string `json:"request_body_schema"`
	ResponseSchema     string `json:"response_schema"`
}

type OpenAPIImportResult struct {
	ProviderTemplate *ProviderTemplate `json:"provider_template"`
	Operations       []APIOperation    `json:"operations"`
	AuthSchemes      []string          `json:"auth_schemes"`
	DestructiveCount int               `json:"destructive_count"`
	TotalCount       int               `json:"total_count"`
}

func ValidateRemoteURL(rawURL string) error {
	return platform.ValidateSSRF(rawURL)
}

func isBlockedHost(host string) bool {
	trimmedHost := strings.TrimSuffix(strings.ToLower(host), ".")
	blocked := []string{
		"localhost", "127.0.0.1", "::1", "0.0.0.0",
		"169.254.169.254", "metadata.google.internal",
		"metadata.internal", "100.100.100.200", "168.63.129.16",
	}
	for _, b := range blocked {
		if trimmedHost == b {
			return true
		}
	}
	if strings.HasSuffix(trimmedHost, ".local") || strings.HasSuffix(trimmedHost, ".internal") ||
		strings.HasSuffix(trimmedHost, ".localhost") || strings.HasSuffix(trimmedHost, ".arpa") {
		return true
	}
	return false
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return platform.IsPrivateIP(parsed)
}

func ParseOpenAPIDocument(doc []byte) (*OpenAPIImportResult, error) {
	var spec map[string]any
	if err := json.Unmarshal(doc, &spec); err != nil {
		return nil, fmt.Errorf("invalid openapi json: %w", err)
	}

	result := &OpenAPIImportResult{}

	info, _ := spec["info"].(map[string]any)
	title := ""
	if info != nil {
		title, _ = info["title"].(string)
	}

	servers, _ := spec["servers"].([]any)
	baseURL := ""
	if len(servers) > 0 {
		if s, ok := servers[0].(map[string]any); ok {
			baseURL, _ = s["url"].(string)
		}
	}

	result.ProviderTemplate = &ProviderTemplate{
		ID:             "pt_" + uuid.NewString()[:8],
		ProviderName:   title,
		BaseURL:        baseURL,
		APIStyle:       "rest",
		AuthLocation:   "header",
		AuthHeaderName: "Authorization",
	}

	secSchemes := extractSecuritySchemes(spec)
	result.AuthSchemes = secSchemes

	paths, _ := spec["paths"].(map[string]any)
	for path, methods := range paths {
		methodMap, ok := methods.(map[string]any)
		if !ok {
			continue
		}
		for method, opData := range methodMap {
			method = strings.ToUpper(method)
			if method == "PARAMETERS" || method == "SUMMARY" || method == "DESCRIPTION" {
				continue
			}

			op, ok := opData.(map[string]any)
			if !ok {
				continue
			}

			summary, _ := op["summary"].(string)
			isDestructive := method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"

			apiOp := APIOperation{
				ID:            "ao_" + uuid.NewString()[:8],
				Method:        method,
				Path:          path,
				Summary:       summary,
				IsDestructive: isDestructive,
				Confirmed:     false,
			}

			if params, ok := op["parameters"].([]any); ok {
				paramBytes, _ := json.Marshal(params)
				apiOp.Parameters = string(paramBytes)
			}

			if reqBody, ok := op["requestBody"].(map[string]any); ok {
				bodyBytes, _ := json.Marshal(reqBody)
				apiOp.RequestBodySchema = string(bodyBytes)
			}

			result.Operations = append(result.Operations, apiOp)
			result.TotalCount++
			if isDestructive {
				result.DestructiveCount++
			}
		}
	}

	return result, nil
}

func extractSecuritySchemes(spec map[string]any) []string {
	var schemes []string
	components, _ := spec["components"].(map[string]any)
	if components == nil {
		return schemes
	}
	secSchemes, _ := components["securitySchemes"].(map[string]any)
	for name := range secSchemes {
		schemes = append(schemes, name)
	}
	return schemes
}

func SaveOpenAPIImport(ctx context.Context, db *sql.DB, result *OpenAPIImportResult) error {
	ts := NewTemplateStore(db)

	if err := ts.CreateProviderTemplate(ctx, result.ProviderTemplate); err != nil {
		return fmt.Errorf("save provider template: %w", err)
	}

	for _, op := range result.Operations {
		op.ProviderTemplateID = result.ProviderTemplate.ID
		_, err := db.ExecContext(ctx, `
INSERT INTO api_operations (id, provider_template_id, method, path, summary,
    is_destructive, confirmed, auth_schemes, parameters, request_body_schema)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			op.ID, op.ProviderTemplateID, op.Method, op.Path, op.Summary,
			boolToInt(op.IsDestructive), boolToInt(op.Confirmed),
			op.AuthSchemes, op.Parameters, op.RequestBodySchema)
		if err != nil {
			continue
		}
	}

	return nil
}
