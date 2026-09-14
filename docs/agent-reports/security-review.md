# Security Review Report: EkaRouter

## 1. Threat Modeling & Scope

The security posture of EkaRouter was audited across network perimeters, data persistence at rest, credentials in memory, and authentication tokens.

## 2. Findings & Hardening Applied

### 2.1 Cryptographic Storage of Provider Secrets
- **Observation in Reference**: 9Router stored provider API keys and refresh tokens in plaintext JSON inside database rows.
- **EkaRouter Implementation**: Stored provider keys and tokens are encrypted at rest using AES-256-GCM authenticated encryption (`internal/auth/auth.go`). The master encryption key is derived via SHA-256 from a 256-bit entropy secret (`EKAROUTER_SECRET_KEY`). Decrypted credentials are only present in memory during upstream request execution.

### 2.2 API Key and Session Protection
- Gateway API keys (`eka_live_...`) and management session tokens are never stored in raw plaintext.
- The system stores only cryptographic one-way SHA-256 hashes (`HashToken`).
- For user identification, only the first 13 characters (e.g. `eka_live_650a...`) are preserved as a non-secret prefix.
- Management sessions feature expiration timestamps (`expires_at`) and immediate revocation capabilities (`revoked_at`).

### 2.3 SSRF (Server-Side Request Forgery) Protection
- Configurable base URLs and outbound provider transports are validated against private, link-local, loopback, and multicast IP ranges (`internal/proxy/proxy.go`).
- Calls to `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `::1`, and `fc00::/7` are blocked by default.
- Local provider access (e.g. locally hosted Ollama on `127.0.0.1:11434`) requires explicit operator opt-in via `EKAROUTER_ALLOW_LOCAL_PROVIDERS=true`.

### 2.4 Payload and Memory Boundaries
- Max body limit middleware (`http.MaxBytesReader`) restricts incoming request bodies to 10 MB (configurable) to prevent memory exhaustion DoS attacks.
- CORS policy restricts allowed origins to explicitly configured dashboard origins.
- Sensitive fields (passwords, tokens, API keys) are omitted from JSON outputs across all `/api/*` endpoints.
