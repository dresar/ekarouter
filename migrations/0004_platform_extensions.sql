-- 0004_platform_extensions.sql: Complete developer platform schema extensions

CREATE TABLE IF NOT EXISTS credential_assignments (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
    environment TEXT NOT NULL DEFAULT 'production',
    assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ca_cred ON credential_assignments(credential_id);
CREATE INDEX IF NOT EXISTS idx_ca_proj ON credential_assignments(project_id);

CREATE TABLE IF NOT EXISTS credential_permissions (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    role_or_user_id TEXT NOT NULL,
    permission TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cp_cred ON credential_permissions(credential_id);

CREATE TABLE IF NOT EXISTS credential_health (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    checked_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ch_cred ON credential_health(credential_id);

CREATE TABLE IF NOT EXISTS credential_usage (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES vault_credentials(id) ON DELETE CASCADE,
    request_count INTEGER NOT NULL DEFAULT 0,
    token_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    period_start DATETIME NOT NULL,
    period_end DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cu_cred ON credential_usage(credential_id);

CREATE TABLE IF NOT EXISTS api_requests (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    credential_id TEXT REFERENCES vault_credentials(id) ON DELETE SET NULL,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ar_prov ON api_requests(provider_id);
CREATE INDEX IF NOT EXISTS idx_ar_cred ON api_requests(credential_id);
CREATE INDEX IF NOT EXISTS idx_ar_created ON api_requests(created_at);

CREATE TABLE IF NOT EXISTS api_request_attempts (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL REFERENCES api_requests(id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL DEFAULT 1,
    credential_id TEXT REFERENCES vault_credentials(id) ON DELETE SET NULL,
    status_code INTEGER NOT NULL,
    error_message TEXT NOT NULL DEFAULT '',
    latency_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ara_req ON api_request_attempts(request_id);

CREATE TABLE IF NOT EXISTS api_request_logs (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL,
    headers_json TEXT NOT NULL DEFAULT '{}',
    body_snippet TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_arl_req ON api_request_logs(request_id);

CREATE TABLE IF NOT EXISTS usage_records (
    id TEXT PRIMARY KEY,
    reference_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    metric TEXT NOT NULL DEFAULT 'requests',
    value INTEGER NOT NULL DEFAULT 0,
    period_date TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ur_ref_metric_period ON usage_records(reference_type, reference_id, metric, period_date);

CREATE TABLE IF NOT EXISTS routing_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    condition_expr TEXT NOT NULL DEFAULT '',
    priority INTEGER NOT NULL DEFAULT 10,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rr_prov ON routing_rules(provider_id);

CREATE TABLE IF NOT EXISTS tool_actions (
    id TEXT PRIMARY KEY,
    tool_id TEXT NOT NULL REFERENCES tool_definitions(id) ON DELETE CASCADE,
    action_name TEXT NOT NULL,
    parameters_schema TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ta_tool ON tool_actions(tool_id);

CREATE TABLE IF NOT EXISTS openapi_documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    raw_content TEXT NOT NULL,
    source_url TEXT NOT NULL DEFAULT '',
    parsed_operations INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_od_title ON openapi_documents(title);

CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'info',
    read INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notif_read ON notifications(read);

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qr_unique_ref ON quota_records(reference_type, reference_id, metric);
