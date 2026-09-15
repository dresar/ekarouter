export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string | ApiErrorDetail
  meta?: {
    request_id?: string
    total?: number
  }
}

export interface ApiErrorDetail {
  code?: string
  message: string
  details?: unknown
}

export interface UserProfile {
  id?: string
  username: string
  role: string
  is_default_password?: boolean
}

export interface AuthResponse {
  token: string
  user: UserProfile
}

export interface SystemHealth {
  status: string
  database?: string
  uptime?: string
  version?: string
}

export interface Provider {
  id: string
  key: string
  name: string
  kind: string
  base_url: string
  enabled: boolean
  category?: string
  created_at?: string
  updated_at?: string
}

export interface PlatformProvider {
  id: string
  name: string
  category: string
  description?: string
  base_url: string
  auth_type?: string
  auth_header_name?: string
  auth_header_prefix?: string
  doc_url?: string
  website_url?: string
  api_reference_url?: string
  free_tier_status?: string
  free_tier_notes?: string
  capabilities?: string[]
  supported_operations?: string[]
  required_credentials?: string[]
  enabled?: boolean
}

export interface Account {
  id: string
  provider_id: string
  name: string
  auth_type: string
  state: 'active' | 'cooling_down' | 'disabled' | 'unavailable' | string
  priority: number
  enabled: boolean
  consecutive_failures?: number
  cooldown_until?: string
  proxy_pool_id?: string
  proxy_name?: string
  proxy_url?: string
  masked_secret?: string
  last_error?: string
}

export interface RouteItem {
  id: string
  route_id?: string
  provider_id: string
  account_id?: string
  model_id?: string
  priority: number
  weight: number
  timeout_ms?: number
  max_retries?: number
  enabled: boolean
}

export interface Route {
  id: string
  name: string
  strategy: 'priority' | 'round_robin' | 'least_used' | 'random' | 'weighted'
  enabled: boolean
  item_count?: number
  items?: RouteItem[]
  timeout_seconds?: number
  retry_count?: number
  created_at?: string
}

export interface ModelCatalogEntry {
  id: string
  provider_id: string
  external_name: string
  display_name: string
  context_limit?: number
  streaming?: boolean
}

export interface VaultCredential {
  id: string
  name: string
  credential_type: 'api_key' | 'bearer_token' | 'basic_auth' | 'oauth2'
  provider_id: string
  project_id?: string
  team_id?: string
  environment: 'production' | 'staging' | 'development'
  masked_value: string
  masked_secret?: string
  preview?: string
  priority: number
  tags: string
  status: 'active' | 'cooling_down' | 'disabled' | 'revoked'
  health_state: 'healthy' | 'degraded' | 'unhealthy' | 'unknown'
  cooldown_until?: string
  request_count: number
  error_count: number
  last_used_at?: string
  last_validated_at?: string
  notes?: string
  created_at: string
}

export interface ProxyProfile {
  id: string
  name: string
  scheme: 'http' | 'https' | 'socks5' | 'relay'
  host: string
  port: number
  username?: string
  password?: string
  enabled: boolean
  is_active?: boolean
  last_latency_ms?: number
}

export interface ApiKey {
  id: string
  name: string
  prefix: string
  scopes: string
  enabled: boolean
  created_at: string
  last_used_at?: string
}

export interface ToolDefinition {
  id: string
  provider_id: string
  name: string
  description: string
  category: 'developer' | 'communication' | 'monitoring' | 'storage' | 'payments' | 'custom'
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  url_template: string
  timeout_ms: number
  retry_count: number
  created_at: string
}

export interface RequestTemplate {
  id: string
  name: string
  method: string
  path: string
  headers?: Record<string, string>
  body_schema?: string
}

export interface UsageSummary {
  total_requests: number
  success_requests?: number
  failed_requests?: number
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  cached_tokens?: number
  estimated_cost?: number
  by_provider?: Record<string, number>
  by_model?: Record<string, number>
  total_errors?: number
  active_credentials?: number
}

export interface UsageAdminSummary {
  total_requests: number
  total_tokens: number
  prompt_tokens: number
  output_tokens: number
  avg_latency_ms: number
}

export interface ProviderUsageItem {
  provider_id: string
  active_credentials: number
  total_requests: number
  total_errors: number
}

export interface CredentialUsageItem {
  id: string
  name: string
  provider_id: string
  project_id?: string
  environment: string
  request_count: number
  error_count: number
  last_used_at?: string
}

export interface ProjectUsageItem {
  project_id: string
  active_credentials: number
  total_requests: number
  total_errors: number
}

export interface AuditRecord {
  id: string
  actor_id: string
  action: string
  resource_type: string
  resource_id: string
  result: string
  timestamp: string
  details?: Record<string, unknown>
}

export interface CredentialPool {
  id: string
  name: string
  owner_id?: string
  provider_id: string
  environment: string
  status: 'active' | 'paused' | 'disabled'
  strategy?: string
  member_count?: number
}

export interface PoolMember {
  id: string
  pool_id: string
  account_id: string
  priority: number
  weight: number
  status: string
  cooldown_until?: string
}

export interface FreeTierItem {
  id: string
  provider_name: string
  category: string
  model_name?: string
  free_quota_description?: string
  free_quota?: string
  free_tier_status?: string
  rpm_limit?: number
  tpm_limit?: number
  requires_credit_card?: boolean
  payment_required?: boolean
  verified?: boolean
  status?: string
  official_website?: string
  docs_url?: string
  documentation_url?: string
  reset_interval?: string
}

export interface BackupItem {
  filename: string
  size_bytes: number
  created_at: string
}

export interface ModelQuota {
  id: string
  name: string
  used: number
  total: number
  remaining_percentage: number
  reset_at?: string
  display_name?: string
}

export interface AccountQuota {
  account_id: string
  account_name: string
  provider_id: string
  provider_name: string
  auth_type: string
  email?: string
  state: string
  plan?: string
  overall_remaining: number
  reset_at?: string
  quotas: ModelQuota[]
  last_checked: string
}
