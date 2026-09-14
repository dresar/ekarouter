CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'developer',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team_members (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tm_team_user ON team_members(team_id, user_id);

CREATE TABLE IF NOT EXISTS environments (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_env_project ON environments(project_id);

CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general'
);
CREATE INDEX IF NOT EXISTS idx_perm_code ON permissions(code);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_code TEXT NOT NULL REFERENCES permissions(code) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_code)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS client_tokens (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    token_hash TEXT UNIQUE NOT NULL,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    scopes TEXT NOT NULL DEFAULT '*',
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_ct_hash ON client_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_ct_user ON client_tokens(user_id);

CREATE TABLE IF NOT EXISTS provider_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS provider_capabilities (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    capability TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pc_provider ON provider_capabilities(provider_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pc_prov_cap ON provider_capabilities(provider_id, capability);

CREATE TABLE IF NOT EXISTS provider_configs (
    id TEXT PRIMARY KEY,
    provider_id TEXT UNIQUE NOT NULL,
    config_json TEXT NOT NULL DEFAULT '{}',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vault_credentials (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    credential_type TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    environment TEXT NOT NULL DEFAULT 'production',
    encrypted_value TEXT NOT NULL,
    masked_value TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 10,
    tags TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    health_state TEXT NOT NULL DEFAULT 'unknown',
    cooldown_until DATETIME,
    expires_at DATETIME,
    notes TEXT NOT NULL DEFAULT '',
    last_used_at DATETIME,
    last_validated_at DATETIME,
    last_error TEXT NOT NULL DEFAULT '',
    request_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_vc_provider ON vault_credentials(provider_id);
CREATE INDEX IF NOT EXISTS idx_vc_project ON vault_credentials(project_id);
CREATE INDEX IF NOT EXISTS idx_vc_status ON vault_credentials(status);
CREATE INDEX IF NOT EXISTS idx_vc_health ON vault_credentials(health_state);

CREATE TABLE IF NOT EXISTS credential_versions (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    version_num INTEGER NOT NULL,
    encrypted_value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cv_credential ON credential_versions(credential_id);

CREATE TABLE IF NOT EXISTS credential_cooldowns (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    retry_after_seconds INTEGER NOT NULL DEFAULT 60,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ends_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ccd_credential ON credential_cooldowns(credential_id);
CREATE INDEX IF NOT EXISTS idx_ccd_ends ON credential_cooldowns(ends_at);

CREATE TABLE IF NOT EXISTS oauth_connections (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT,
    scopes TEXT NOT NULL DEFAULT '',
    expires_at DATETIME,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_oc_provider ON oauth_connections(provider_id);
CREATE INDEX IF NOT EXISTS idx_oc_user ON oauth_connections(user_id);

CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    target_url TEXT NOT NULL,
    secret_encrypted TEXT NOT NULL,
    events TEXT NOT NULL DEFAULT '*',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_wh_enabled ON webhooks(enabled);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id TEXT PRIMARY KEY,
    webhook_id TEXT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    response_status INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_wd_created ON webhook_deliveries(created_at);

CREATE TABLE IF NOT EXISTS tool_definitions (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    method TEXT NOT NULL DEFAULT 'POST',
    url_template TEXT NOT NULL,
    headers_template TEXT NOT NULL DEFAULT '{}',
    query_template TEXT NOT NULL DEFAULT '{}',
    body_schema TEXT NOT NULL DEFAULT '',
    timeout_ms INTEGER NOT NULL DEFAULT 30000,
    retry_policy TEXT NOT NULL DEFAULT '{"max_retries": 2}',
    rate_limit_policy TEXT NOT NULL DEFAULT '{}',
    required_permissions TEXT NOT NULL DEFAULT 'tools.execute',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_td_provider ON tool_definitions(provider_id);
CREATE INDEX IF NOT EXISTS idx_td_enabled ON tool_definitions(enabled);

CREATE TABLE IF NOT EXISTS tool_executions (
    id TEXT PRIMARY KEY,
    tool_id TEXT NOT NULL REFERENCES tool_definitions(id) ON DELETE CASCADE,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    environment TEXT NOT NULL DEFAULT 'production',
    status TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    response_status INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_te_tool ON tool_executions(tool_id);
CREATE INDEX IF NOT EXISTS idx_te_executed ON tool_executions(executed_at);

CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id TEXT PRIMARY KEY,
    task_type TEXT NOT NULL,
    target_id TEXT NOT NULL DEFAULT '',
    cron_expr TEXT NOT NULL,
    next_run_at DATETIME,
    status TEXT NOT NULL DEFAULT 'active',
    last_run_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_st_status ON scheduled_tasks(status);
CREATE INDEX IF NOT EXISTS idx_st_next ON scheduled_tasks(next_run_at);

CREATE TABLE IF NOT EXISTS task_runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES scheduled_tasks(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_tr_task ON task_runs(task_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_id TEXT NOT NULL,
    actor_type TEXT NOT NULL DEFAULT 'user',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    project_id TEXT,
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL DEFAULT '',
    result TEXT NOT NULL DEFAULT 'success',
    error_category TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_al_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_al_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_al_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_al_created ON audit_logs(created_at);

CREATE TABLE IF NOT EXISTS quota_records (
    id TEXT PRIMARY KEY,
    reference_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    metric TEXT NOT NULL DEFAULT 'requests',
    used_value INTEGER NOT NULL DEFAULT 0,
    max_value INTEGER NOT NULL DEFAULT 0,
    period TEXT NOT NULL DEFAULT 'monthly',
    reset_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_qr_ref ON quota_records(reference_type, reference_id);

CREATE TABLE IF NOT EXISTS rate_limit_records (
    id TEXT PRIMARY KEY,
    reference_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    window_start DATETIME NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_rlr_ref ON rate_limit_records(reference_type, reference_id, window_start);
