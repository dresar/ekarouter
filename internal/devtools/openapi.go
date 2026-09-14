package devtools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

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
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("only http and https schemes allowed, got: %s", scheme)
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("empty hostname")
	}

	if isBlockedHost(host) {
		return fmt.Errorf("blocked host: %s (local/private addresses not allowed)", host)
	}

	ips, err := net.LookupHost(host)
	if err == nil {
		for _, ip := range ips {
			if isPrivateIP(ip) {
				return fmt.Errorf("resolved to private IP: %s", ip)
			}
		}
	}

	return nil
}

func isBlockedHost(host string) bool {
	lower := strings.ToLower(host)
	blocked := []string{
		"localhost", "127.0.0.1", "::1", "0.0.0.0",
		"169.254.169.254", "metadata.google.internal",
		"metadata.internal", "100.100.100.200",
	}
	for _, b := range blocked {
		if lower == b {
			return true
		}
	}
	if strings.HasSuffix(lower, ".local") || strings.HasSuffix(lower, ".internal") {
		return true
	}
	return false
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	private := []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"127.0.0.0/8", "169.254.0.0/16", "::1/128",
		"fc00::/7", "fe80::/10",
	}
	for _, cidr := range private {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(parsed) {
			return true
		}
	}
	return false
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
