# EkaRouter Project Identity & Architecture Manifest

```text
  ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗   ██╗████████╗███████╗██████╗ 
  ██╔════╝██║ ██╔╝██╔══██╗██╔══██╗██╔═══██╗██║   ██║╚══██╔══╝██╔════╝██╔══██╗
  █████╗  █████═╝ ███████║██████╔╝██║   ██║██║   ██║   ██║   █████╗  ██████╔╝
  ██╔══╝  ██╔═██╗ ██╔══██║██╔══██╗██║   ██║██║   ██║   ██║   ██╔══╝  ██╔══██╗
  ███████╗██║ ╚██╗██║  ██║██║  ██║╚██████╔╝╚██████╔╝   ██║   ███████╗██║  ██║
  ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚═╝  ╚═╝
```

---

## 1. Project Overview & Identity

- **Project Name**: EkaRouter
- **Repository**: [https://github.com/dresar/ekarouter](https://github.com/dresar/ekarouter)
- **Primary Owner & Architect**: Eka Syarif Maulana ([@dresar](https://github.com/dresar))
- **Primary Remote Pattern**: `https://github.com/dresar/ekarouter.git`
- **Core Technology**: Go (Golang 1.24+ / Pure Go, 100% CGO-Free)
- **Database Engine**: Embedded SQLite via `modernc.org/sqlite` (Pure Go runtime, Zero external C dependencies)
- **Classification**: Universal AI Gateway & Enterprise Developer API Management Platform

---

## 2. Core Mission & Vision

EkaRouter is engineered as a single-binary, lightweight, hyper-resilient API router and developer platform designed to unify:
1. **Universal AI Model Routing**: Drop-in OpenAI-compatible endpoints (`/v1/chat/completions`, `/v1/models`, `/v1/responses`) with multi-provider failover, auto-rotation, and prompt token compaction.
2. **Unified Developer Platform**: API key vault with AES-256-GCM encryption, rate limiting, quota tracking, SSRF-guarded proxying, generic HTTP tool calling, and HMAC-SHA256 webhooks.

---

## 3. Engineering Tenets & Golden Rules

### I. Pure Go & Zero CGO Dependencies
- EkaRouter compiles into a single, standalone executable on Windows (`ekarouter.exe`), Linux, and macOS without requiring GCC, Clang, or external C dynamic libraries.
- The embedded database uses `modernc.org/sqlite` with WAL (Write-Ahead Logging) mode and connection-level busy timeouts for high-throughput concurrent access.

### II. Strict Nokomen Standard (Zero Code Comments)
- All Go source files (`*.go`) strictly contain **0 comments** (`//` or `/* */`).
- The code is self-documenting through precise domain naming, single-responsibility functions, and idiomatic Go architecture.
- All high-level rationale, architectural schemas, API references, and system documentation live exclusively in Markdown (`docs/`, `.github/`, and `README.md`).

### III. High-Availability & Self-Healing Rotation
- **Failover Loop**: Network errors, 429 Quotas, and 401/403 Auth errors trigger automatic account cooldown and proceed immediately to subsequent healthy targets in the route.
- **Multi-Strategy Rotator**: Isolated per-pool cursors supporting Priority Fallback, Round-Robin, Least Recently Used (LRU), Lowest Error Rate, Health-Based, and Environment-Aware policies.
- **Deterministic Tie-Breaking**: Equal-priority candidate sets resolve ties deterministically by ID before applying round-robin, preventing map-iteration skew.

### IV. Defense-in-Depth Security
- **Vault Encryption**: AES-256-GCM symmetric authenticated encryption with versioned headers (`v1:<base64>`) and dynamic secret masking.
- **SSRF Defense**: Outbound proxy requests strictly block loopback, link-local, private RFC 1918 subnets, and cloud metadata IPs (`169.254.169.254`).
- **Constant-Time Verification**: Webhook HMAC signatures and API key hashes use `crypto/subtle` constant-time comparisons to prevent timing side-channel attacks.

---

## 4. Architectural Domain Map

| Package | Layer | Architectural Role |
| :--- | :--- | :--- |
| `cmd/ekarouter` | Entrypoint | CLI command parsing, server lifecycle, signal traps |
| `internal/app` | Core | Dependency injection, server bootstrap, graceful shutdown |
| `internal/gateway` | AI Gateway | `/v1/*` OpenAI compatibility, fallback loop, SSE streaming |
| `internal/routing` | Routing | Model alias resolution, route evaluation, cooldown registry |
| `internal/rotator` | Selection | Concurrency-safe multi-strategy credential selector |
| `internal/credpool` | Credential Pool | Multi-tenant credential pools, rotation policies, events |
| `internal/vault` | Vault | AES-256-GCM encryption, secret masking, credential storage |
| `internal/httpapi` | HTTP API | Chi HTTP router, CORS/Auth middleware, admin endpoints |
| `internal/platform` | Developer Platform | 11 provider categories, generic tool calling, SSRF guard |
| `internal/limits` | Rate & Quota | Sliding-window rate limiting, quota tracking, cooldowns |
| `internal/rbac` | Access Control | Role-based permissions, PBKDF2 hashing, client tokens |
| `internal/audit` | Audit | Asynchronous immutable audit logger |
| `internal/webhooks` | Webhooks | HMAC-SHA256 verification and delivery dispatcher |
| `internal/tokensaver` | Optimization | Heuristic prompt compaction (Diff, logs, stack traces) |
| `internal/db` | Database | SQLite connection pool, WAL mode, schema migrations |
| `internal/scheduler` | Background | Cooldown auto-recovery, log retention pruning |
| `internal/providers/*` | Adapters | 50+ specialized AI provider adapters and protocol translators |

---

## 5. Governance & Repository Links

- **Repository**: [https://github.com/dresar/ekarouter](https://github.com/dresar/ekarouter)
- **Issues & Tracking**: [https://github.com/dresar/ekarouter/issues](https://github.com/dresar/ekarouter/issues)
- **Pull Requests**: [https://github.com/dresar/ekarouter/pulls](https://github.com/dresar/ekarouter/pulls)
- **CI/CD Pipeline**: GitHub Actions (`.github/workflows/ci.yml`)
- **Documentation**: Located in [`docs/`](../docs)
