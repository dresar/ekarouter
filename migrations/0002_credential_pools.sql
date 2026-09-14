CREATE TABLE IF NOT EXISTS free_tier_catalog (
    id TEXT PRIMARY KEY,
    provider_name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    official_website TEXT NOT NULL DEFAULT '',
    docs_url TEXT NOT NULL DEFAULT '',
    pricing_url TEXT NOT NULL DEFAULT '',
    free_tier_status TEXT NOT NULL DEFAULT 'unverified',
    free_quota TEXT NOT NULL DEFAULT '',
    reset_interval TEXT NOT NULL DEFAULT '',
    supported_regions TEXT NOT NULL DEFAULT '',
    signup_steps TEXT NOT NULL DEFAULT '',
    auth_method TEXT NOT NULL DEFAULT 'api_key',
    required_scopes TEXT NOT NULL DEFAULT '',
    expiration_behavior TEXT NOT NULL DEFAULT '',
    restrictions TEXT NOT NULL DEFAULT '',
    payment_required INTEGER NOT NULL DEFAULT 0,
    personal_key_required INTEGER NOT NULL DEFAULT 1,
    sandbox_support INTEGER NOT NULL DEFAULT 0,
    verified_at DATETIME,
    confidence TEXT NOT NULL DEFAULT 'low',
    status TEXT NOT NULL DEFAULT 'unverified',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ftc_category ON free_tier_catalog(category);
CREATE INDEX IF NOT EXISTS idx_ftc_status ON free_tier_catalog(status);
CREATE INDEX IF NOT EXISTS idx_ftc_auth_method ON free_tier_catalog(auth_method);

CREATE TABLE IF NOT EXISTS free_tier_sources (
    id TEXT PRIMARY KEY,
    catalog_id TEXT NOT NULL REFERENCES free_tier_catalog(id) ON DELETE CASCADE,
    source_url TEXT NOT NULL,
    source_type TEXT NOT NULL DEFAULT 'documentation',
    verified_at DATETIME,
    notes TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_fts_catalog ON free_tier_sources(catalog_id);

CREATE TABLE IF NOT EXISTS credential_pools (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL DEFAULT 'admin',
    provider_id TEXT REFERENCES providers(id) ON DELETE SET NULL,
    environment TEXT NOT NULL DEFAULT 'production',
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cp_owner ON credential_pools(owner_id);
CREATE INDEX IF NOT EXISTS idx_cp_provider ON credential_pools(provider_id);
CREATE INDEX IF NOT EXISTS idx_cp_environment ON credential_pools(environment);
CREATE INDEX IF NOT EXISTS idx_cp_status ON credential_pools(status);

CREATE TABLE IF NOT EXISTS pool_members (
    id TEXT PRIMARY KEY,
    pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    priority INTEGER NOT NULL DEFAULT 10,
    weight INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'active',
    cooldown_until DATETIME,
    last_used_at DATETIME,
    last_success_at DATETIME,
    last_failure_at DATETIME,
    expires_at DATETIME,
    failure_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    total_requests INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_pm_pool ON pool_members(pool_id);
CREATE INDEX IF NOT EXISTS idx_pm_account ON pool_members(account_id);
CREATE INDEX IF NOT EXISTS idx_pm_status ON pool_members(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pm_pool_account ON pool_members(pool_id, account_id);

CREATE TABLE IF NOT EXISTS rotation_policies (
    id TEXT PRIMARY KEY,
    pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE,
    strategy TEXT NOT NULL DEFAULT 'priority_fallback',
    disable_auto_rotation INTEGER NOT NULL DEFAULT 0,
    max_concurrent INTEGER NOT NULL DEFAULT 1,
    cooldown_seconds INTEGER NOT NULL DEFAULT 30,
    quota_aware INTEGER NOT NULL DEFAULT 0,
    retry_transient_only INTEGER NOT NULL DEFAULT 1,
    max_retries INTEGER NOT NULL DEFAULT 3,
    fallback_on_permanent INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rp_pool ON rotation_policies(pool_id);

CREATE TABLE IF NOT EXISTS credential_health_checks (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES pool_members(id) ON DELETE CASCADE,
    pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'unknown',
    latency_ms INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    checked_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chc_member ON credential_health_checks(member_id);
CREATE INDEX IF NOT EXISTS idx_chc_pool ON credential_health_checks(pool_id);
CREATE INDEX IF NOT EXISTS idx_chc_checked ON credential_health_checks(checked_at);

CREATE TABLE IF NOT EXISTS credential_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pool_id TEXT REFERENCES credential_pools(id) ON DELETE SET NULL,
    member_id TEXT REFERENCES pool_members(id) ON DELETE SET NULL,
    account_id TEXT,
    event_type TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ce_pool ON credential_events(pool_id);
CREATE INDEX IF NOT EXISTS idx_ce_member ON credential_events(member_id);
CREATE INDEX IF NOT EXISTS idx_ce_type ON credential_events(event_type);
CREATE INDEX IF NOT EXISTS idx_ce_created ON credential_events(created_at);

CREATE TABLE IF NOT EXISTS provider_templates (
    id TEXT PRIMARY KEY,
    provider_name TEXT NOT NULL,
    base_url TEXT NOT NULL DEFAULT '',
    api_style TEXT NOT NULL DEFAULT 'rest',
    auth_location TEXT NOT NULL DEFAULT 'header',
    auth_header_name TEXT NOT NULL DEFAULT 'Authorization',
    auth_query_param TEXT NOT NULL DEFAULT '',
    oauth_support INTEGER NOT NULL DEFAULT 0,
    validation_endpoint TEXT NOT NULL DEFAULT '',
    rate_limit_headers TEXT NOT NULL DEFAULT '',
    quota_headers TEXT NOT NULL DEFAULT '',
    streaming_support INTEGER NOT NULL DEFAULT 0,
    retryable_statuses TEXT NOT NULL DEFAULT '429,500,502,503,504',
    credential_schema TEXT NOT NULL DEFAULT '',
    pagination_style TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_pt_name ON provider_templates(provider_name);
CREATE INDEX IF NOT EXISTS idx_pt_style ON provider_templates(api_style);

CREATE TABLE IF NOT EXISTS request_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider_template_id TEXT REFERENCES provider_templates(id) ON DELETE SET NULL,
    method TEXT NOT NULL DEFAULT 'GET',
    path TEXT NOT NULL DEFAULT '/',
    headers TEXT NOT NULL DEFAULT '{}',
    query_params TEXT NOT NULL DEFAULT '{}',
    body_schema TEXT NOT NULL DEFAULT '',
    credential_ref TEXT NOT NULL DEFAULT '',
    timeout_ms INTEGER NOT NULL DEFAULT 30000,
    retry_count INTEGER NOT NULL DEFAULT 0,
    redaction_rules TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rt_provider ON request_templates(provider_template_id);

CREATE TABLE IF NOT EXISTS api_operations (
    id TEXT PRIMARY KEY,
    provider_template_id TEXT NOT NULL REFERENCES provider_templates(id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    is_destructive INTEGER NOT NULL DEFAULT 0,
    confirmed INTEGER NOT NULL DEFAULT 0,
    auth_schemes TEXT NOT NULL DEFAULT '',
    parameters TEXT NOT NULL DEFAULT '[]',
    request_body_schema TEXT NOT NULL DEFAULT '',
    response_schema TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ao_provider ON api_operations(provider_template_id);
CREATE INDEX IF NOT EXISTS idx_ao_destructive ON api_operations(is_destructive);

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL DEFAULT 'admin',
    environment TEXT NOT NULL DEFAULT 'development',
    description TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_proj_owner ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_proj_env ON projects(environment);

CREATE TABLE IF NOT EXISTS project_credentials (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pool_id TEXT REFERENCES credential_pools(id) ON DELETE CASCADE,
    account_id TEXT REFERENCES accounts(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_pc_project ON project_credentials(project_id);
CREATE INDEX IF NOT EXISTS idx_pc_pool ON project_credentials(pool_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pc_proj_pool ON project_credentials(project_id, pool_id);

CREATE TABLE IF NOT EXISTS credential_aliases (
    id TEXT PRIMARY KEY,
    alias TEXT UNIQUE NOT NULL,
    pool_id TEXT REFERENCES credential_pools(id) ON DELETE CASCADE,
    account_id TEXT REFERENCES accounts(id) ON DELETE CASCADE,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    environment TEXT NOT NULL DEFAULT 'production',
    injection_type TEXT NOT NULL DEFAULT 'header',
    injection_key TEXT NOT NULL DEFAULT 'Authorization',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ca_alias ON credential_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_ca_pool ON credential_aliases(pool_id);
CREATE INDEX IF NOT EXISTS idx_ca_project ON credential_aliases(project_id);

CREATE TABLE IF NOT EXISTS quota_limits (
    id TEXT PRIMARY KEY,
    reference_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    limit_type TEXT NOT NULL DEFAULT 'requests',
    max_value INTEGER NOT NULL DEFAULT 0,
    current_value INTEGER NOT NULL DEFAULT 0,
    reset_interval TEXT NOT NULL DEFAULT 'daily',
    last_reset_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    source_url TEXT NOT NULL DEFAULT '',
    verified_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ql_ref ON quota_limits(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_ql_type ON quota_limits(limit_type);

CREATE TABLE IF NOT EXISTS rotation_decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pool_id TEXT NOT NULL REFERENCES credential_pools(id) ON DELETE CASCADE,
    selected_member_id TEXT REFERENCES pool_members(id) ON DELETE SET NULL,
    strategy TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    candidates_count INTEGER NOT NULL DEFAULT 0,
    request_id TEXT NOT NULL DEFAULT '',
    latency_us INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rd_pool ON rotation_decisions(pool_id);
CREATE INDEX IF NOT EXISTS idx_rd_created ON rotation_decisions(created_at);
