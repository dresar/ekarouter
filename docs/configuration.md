# Configuration Reference

EkaRouter is configured via environment variables or `.env` files.

| Variable | Default | Description |
|---|---|---|
| `EKAROUTER_ENV` | `development` | Runtime environment (`development` or `production`). |
| `EKAROUTER_HOST` | `0.0.0.0` | IP address interface to bind. |
| `EKAROUTER_PORT` | `8080` | Port number to listen on. |
| `EKAROUTER_DB_PATH` | `data/ekarouter.db` | File path for SQLite database file. |
| `EKAROUTER_SECRET_KEY` | *(Random 32 hex)* | Master encryption key used for AES-256-GCM vault encryption. Must be at least 16 characters. |
| `EKAROUTER_SESSION_SECRET` | *(Derived)* | Secret used for session token signing and verification. |
| `EKAROUTER_ADMIN_USER` | `admin` | Default administrator username. |
| `EKAROUTER_ADMIN_PASSWORD` | `admin12345` | Default administrator password. |
| `EKAROUTER_ALLOW_LOCAL_PROVIDERS` | `false` | When true, permits SSRF bypass to local loopback IPs for testing. |
| `EKAROUTER_TOKEN_SAVER_MODE` | `safe` | Heuristic prompt compaction mode (`off`, `safe`, `balanced`). |
| `EKAROUTER_CORS_ORIGINS` | `*` | Comma-separated list of allowed CORS origins. |
| `EKAROUTER_MAX_BODY_BYTES` | `10485760` (10MB) | Maximum allowed HTTP request body size in bytes. |
| `EKAROUTER_REQUEST_TIMEOUT` | `30s` | Default timeout for external HTTP provider requests. |
| `EKAROUTER_LOG_RETENTION_DAYS` | `30` | Number of days to retain audit logs, tool executions, and webhook deliveries before automatic pruning. |
| `EKAROUTER_ENABLE_SCHEDULER` | `true` | Enables background periodic maintenance worker. |
