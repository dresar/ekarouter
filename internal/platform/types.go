package platform

import (
	"context"
	"time"
)

type Category string

const (
	CategoryDeveloper    Category = "developer"
	CategoryComm         Category = "communication"
	CategoryMonitoring   Category = "monitoring"
	CategoryAutomation   Category = "automation"
	CategoryScrapingData Category = "scraping_and_data"
	CategoryStorage      Category = "storage"
	CategoryPayments     Category = "payments"
	CategoryAnalytics    Category = "analytics"
	CategoryMaps         Category = "maps"
	CategorySecurity     Category = "security"
	CategoryCustom       Category = "custom"
)

type AuthType string

const (
	AuthTypeBearerToken   AuthType = "bearer_auth"
	AuthTypeAPIKeyHeader  AuthType = "api_key_auth"
	AuthTypeBasicAuth     AuthType = "basic_auth"
	AuthTypeOAuth2        AuthType = "oauth2"
	AuthTypePersonalToken AuthType = "personal_access_token"
	AuthTypeServiceAcct   AuthType = "service_account"
	AuthTypeCustomHeader  AuthType = "custom_header"
)

type Capability string

const (
	CapAPIKeyAuth      Capability = "api_key_auth"
	CapBearerAuth      Capability = "bearer_auth"
	CapBasicAuth       Capability = "basic_auth"
	CapOAuth2          Capability = "oauth2"
	CapPersonalToken   Capability = "personal_access_token"
	CapServiceAccount  Capability = "service_account"
	CapSignedRequest   Capability = "signed_request"
	CapWebhook         Capability = "webhook"
	CapUsageAPI        Capability = "usage_api"
	CapQuotaAPI        Capability = "quota_api"
	CapHealthCheck     Capability = "health_check"
	CapResourceListing Capability = "resource_listing"
	CapRequestProxy    Capability = "request_proxy"
	CapOpenAPISchema   Capability = "openapi_schema"
	CapStreaming       Capability = "streaming"
	CapPagination      Capability = "pagination"
	CapFileUpload      Capability = "file_upload"
	CapFileDownload    Capability = "file_download"
)

type HealthStatus struct {
	Healthy   bool          `json:"healthy"`
	Latency   time.Duration `json:"latency"`
	Message   string        `json:"message"`
	CheckedAt time.Time     `json:"checked_at"`
}

type ProviderMetadata struct {
	ID                  string       `json:"id"`
	Name                string       `json:"name"`
	Category            Category     `json:"category"`
	Description         string       `json:"description"`
	BaseURL             string       `json:"base_url"`
	AuthType            AuthType     `json:"auth_type"`
	AuthHeaderName      string       `json:"auth_header_name,omitempty"`
	AuthHeaderPrefix    string       `json:"auth_header_prefix,omitempty"`
	RequiredCredentials []string     `json:"required_credentials"`
	SupportedOperations []string     `json:"supported_operations"`
	Capabilities        []Capability `json:"capabilities"`
	WebsiteURL          string       `json:"website_url"`
	DocsURL             string       `json:"docs_url"`
	APIReferenceURL     string       `json:"api_reference_url"`
	FreeTierStatus      string       `json:"free_tier_status"`
	FreeTierNotes       string       `json:"free_tier_notes,omitempty"`
	Enabled             bool         `json:"enabled"`
}

type ExecutionRequest struct {
	Path             string            `json:"path"`
	Method           string            `json:"method"`
	Headers          map[string]string `json:"headers"`
	QueryParams      map[string]string `json:"query_params"`
	Body             []byte            `json:"body"`
	CredentialSecret string            `json:"-"`
}

type ExecutionResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Latency    time.Duration     `json:"latency"`
}

type ProviderAdapter interface {
	Metadata() ProviderMetadata
	ValidateCredential(ctx context.Context, secret string) (bool, string, error)
	HealthCheck(ctx context.Context, secret string) (HealthStatus, error)
	Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
}
