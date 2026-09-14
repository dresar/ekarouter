# Go Architecture Report

## 1. Architectural Decisions

1. **Pure Go SQLite Driver**: `modernc.org/sqlite` is chosen over `mattn/go-sqlite3` to ensure a completely CGO-free build path. This guarantees that native single-binary executables for Windows (`.exe`) and Linux (`amd64`, `arm64`) can be cross-compiled on any platform without requiring GCC or MinGW toolchains.
2. **Router**: `github.com/go-chi/chi/v5` is selected as the lightweight HTTP routing framework because of its minimal overhead, complete compatibility with `net/http`, and standard middleware ecosystem.
3. **Neutral Internal Gateway Pipeline**: The core gateway communicates only through domain types (`Request`, `Response`, `StreamEvent`). Provider-specific payloads are translated at the adapter boundary in `internal/providers`.
4. **Resilience & Fallback**: Routing errors are classified into transient (429, quota exhausted, timeout, connection failure, 5xx) vs non-retryable (400, 401, 403, 404). Only transient errors trigger route fallback.
5. **Security Enclave**: AES-256-GCM symmetric encryption secures all secrets at rest. API keys are hashed with SHA-256; raw keys are never stored in the database. Outbound network requests validate target IP addresses against private network ranges to prevent SSRF vulnerabilities.
