# Contributing to EkaRouter

Thank you for your interest in contributing! This guide is designed to be crystal-clear for **both human developers and AI coding agents**. Read it fully before opening a pull request.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Before You Start](#before-you-start)
3. [Development Environment Setup](#development-environment-setup)
4. [Project Architecture Overview](#project-architecture-overview)
5. [How to Add a New AI Provider](#how-to-add-a-new-ai-provider)
6. [How to Add a New Platform Provider](#how-to-add-a-new-platform-provider)
7. [Coding Standards](#coding-standards)
8. [Testing Requirements](#testing-requirements)
9. [Pull Request Checklist](#pull-request-checklist)
10. [For AI Agents](#for-ai-agents)

---

## Code of Conduct

- Be respectful and constructive in all communications.
- Only contribute code you have the right to license under MIT.
- Security vulnerabilities must be reported **privately** via GitHub Issues marked [SECURITY], not in public PRs.

---

## Before You Start

1. **Check open issues** first — the issue may already be tracked or in progress.
2. **Open an issue first** for non-trivial changes (new providers, architecture changes, new features). Get a thumbs-up before writing code.
3. **Fork the repo**, work on a feature branch, then open a PR against main.

---

## Development Environment Setup

### Prerequisites

| Tool     | Minimum Version | Notes                                       |
|----------|-----------------|---------------------------------------------|
| Go       | 1.21+           | CGO is **not required** — pure Go SQLite    |
| Node.js  | 18+             | For frontend (React + Vite)                 |
| Git      | 2.x             | Standard                                    |

### Clone and Build

`ash
git clone https://github.com/dresar/ekarouter.git
cd ekarouter

# Build backend
go build -o ekarouter ./cmd/ekarouter

# Install and build frontend
cd frontend && npm install && npm run build && cd ..
`

### Configure Environment

`ash
cp .env.example .env
# Edit .env — at minimum set EKAROUTER_SECRET_KEY and EKAROUTER_ADMIN_PASSWORD
`

### Run the Server

`ash
# Backend listens on http://0.0.0.0:8080
./ekarouter serve
`

### Run Tests

`ash
go test ./...
`

---

## Project Architecture Overview

`
ekarouter/
├── cmd/
│   ├── ekarouter/        ← Main binary entrypoint (flag parsing, signal handling)
│   └── check_keys/       ← DB migration helper / diagnostic tool
├── internal/
│   ├── app/              ← Dependency injection, service wiring, startup
│   ├── audit/            ← Async immutable audit logger
│   ├── auth/             ← AES-256-GCM crypto service, SHA-256 token hashing
│   ├── cli/              ← Interactive terminal menu (management subcommands)
│   ├── combos/           ← Route combos (multi-account pool configurations)
│   ├── config/           ← Env-var configuration loader
│   ├── credpool/         ← AI credential pool and rotation policies
│   ├── db/               ← SQLite connection, WAL mode, migration runner
│   ├── devtools/         ← OpenAPI 3.x document parser and template generator
│   ├── executor/         ← Generic parameterized HTTP tool execution engine
│   ├── freetier/         ← Free-tier provider catalog and capability seeding
│   ├── gateway/          ← Core AI gateway pipeline, fallback, SSE streaming
│   ├── headroom/         ← Context window budget management
│   ├── health/           ← /live /ready /health liveness endpoints
│   ├── httpapi/          ← Chi HTTP router, CORS, auth middleware, all handlers
│   ├── limits/           ← Sliding-window rate limiter and quota engine
│   ├── oauth/            ← Ephemeral PKCE OAuth state and token manager
│   ├── platform/         ← Universal developer platform (11 API categories)
│   ├── pricing/          ← Token cost estimation per provider
│   ├── providers/        ← AI provider adapters (OpenAI, Anthropic, Gemini, etc.)
│   ├── proxy/            ← Outbound HTTP transport, SOCKS5/HTTP proxy chaining
│   ├── quota/            ← Persistent daily/monthly quota tracking
│   ├── rbac/             ← PBKDF2 user auth, PAT tokens (eka_pat_*)
│   ├── rotator/          ← Concurrency-safe multi-strategy credential selector
│   ├── routing/          ← Model alias resolution, route resolver, cooldown manager
│   ├── scheduler/        ← Background periodic maintenance worker
│   ├── tokensaver/       ← Heuristic prompt compaction algorithms
│   ├── usage/            ← Non-blocking per-request usage recorder
│   ├── vault/            ← Encrypted credential vault (AES-GCM, versioned)
│   └── webhooks/         ← HMAC-SHA256 webhook in/out dispatcher
├── frontend/             ← React 18 + Vite + TypeScript admin dashboard
│   └── src/
│       ├── pages/        ← 22 UI pages (providers, routing, proxies, usage, etc.)
│       ├── components/   ← Reusable UI components
│       ├── api/          ← Typed fetch wrappers for all backend API endpoints
│       └── context/      ← React context providers (auth, theme)
├── migrations/           ← Versioned SQL files (0001_init.sql, 0002_*.sql, ...)
├── docs/                 ← Technical documentation (Markdown)
└── scripts/              ← Build helpers and smoke tests
`

### AI Gateway Request Lifecycle

`
Client  POST /v1/chat/completions
          │
          ▼
  Chi Router + Auth Middleware
          │
          ▼
  GatewayHandler (httpapi/)
          │
          ▼
  Gateway.Execute()
     ├── TokenSaver.CompactWithMode()      ← optional prompt compression
     ├── Router.SelectTargets(modelID)     ← resolves ordered target list from DB
     └── for each target:
           ├── registry.Get(providerKind)  ← fetches the provider Adapter
           ├── credResolve(accountID)      ← decrypts credentials from vault
           ├── adapter.Execute(req, creds) ← calls upstream provider API
           │     ├── success → record usage, return response
           │     ├── auth/quota error → cooldown 15-30min, skip to next target
           │     └── transient error → retry with backoff, then skip to next
           └── all targets fail → 503 "all routes exhausted"
`

### Key SQLite Tables

| Table                | Purpose                                                   |
|----------------------|-----------------------------------------------------------|
| providers          | AI provider registry (gemini, openai, groq, ...)          |
| ccounts           | Provider accounts (credential containers)                 |
| credentials        | AES-256-GCM encrypted API keys per account                |
| models             | Model registry per provider                               |
| outes             | Named routes (e.g., "default", "fast", "budget")          |
| oute_items        | Ordered model/account targets within a route              |
| proxy_pools        | Proxy configurations (HTTP, SOCKS5)                       |
| pi_keys           | Developer PAT tokens (eka_pat_*)                        |
| usage_records      | Per-request token counts and latency metrics              |
| udit_logs         | Immutable chronological action log                        |
| ault_credentials  | Platform API keys, AES-GCM encrypted                     |
| 	ool_definitions   | Generic HTTP tool definitions                             |
| quota_records      | Daily/monthly quota counters per account                  |

---

## How to Add a New AI Provider

AI provider adapters live in internal/providers/<name>/. Every adapter must implement the providers.Adapter interface defined in [internal/providers/types.go](internal/providers/types.go):

`go
type Adapter interface {
    Kind() string
    Models(ctx context.Context, creds *Credentials) ([]ModelInfo, error)
    Execute(ctx context.Context, req *Request, creds *Credentials) (*Response, error)
    ExecuteStream(ctx context.Context, req *Request, creds *Credentials) (<-chan StreamEvent, error)
}
`

### Step 1 — Create the package

`
internal/providers/myprovider/
└── adapter.go
`

### Step 2 — Implement the adapter

`go
package myprovider

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/dresar/ekarouter/internal/providers"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Kind() string { return "myprovider" }

func (a *Adapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
    return providers.GetCanonicalModels("myprovider"), nil
}

func (a *Adapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
    // 1. Build the upstream request body (translate providers.Request → provider format)
    // 2. Call the upstream API via creds.HTTPClient (already proxy-aware)
    // 3. On non-2xx: use providers.ClassifyHTTPError(resp.StatusCode, body)
    // 4. Unmarshal the response, translate to providers.Response
    // 5. NEVER return (nil, nil)
    return nil, fmt.Errorf("not implemented")
}

func (a *Adapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
    // 1. Open SSE/chunked stream to upstream
    // 2. Launch a goroutine that reads events and writes to the returned channel
    // 3. Close the channel after StreamEventDone or StreamEventError
    return nil, fmt.Errorf("not implemented")
}
`

### Step 3 — Register canonical models

Add to internal/providers/canonical_models.go:

`go
"myprovider": {
    {
        ID:           "my-model-v1",
        Name:         "My Model V1",
        ContextLimit: 32768,
        Streaming:    true,
        Capabilities: Capabilities{
            Vision:      false,
            ToolCalling: true,
            Reasoning:   false,
            Streaming:   true,
        },
    },
},
`

### Step 4 — Register the adapter

In internal/app/ (the wiring layer), register the new adapter:

`go
registry.Register("myprovider", myprovider.New())
`

### Step 5 — Write tests

internal/providers/myprovider/adapter_test.go must cover at minimum:

- Kind() returns "myprovider"
- Execute() with a mock HTTP client returns a valid *providers.Response
- Execute() with a 401 response returns a *providers.ProviderError with Class == ErrorClassAuth
- Execute() with a 429 response returns ErrorClassRateLimit

---

## How to Add a New Platform Provider

Platform providers (developer tool APIs — GitHub, Stripe, Resend, etc.) live in [internal/platform/catalog.go](internal/platform/catalog.go).

`go
registry.Register(platform.NewBaseAdapter(platform.ProviderMetadata{
    ID:                  "my-service",
    Name:                "My Service",
    Category:            platform.CategoryDeveloper,
    Description:         "Short one-sentence description of what this API does.",
    BaseURL:             "https://api.myservice.com/v1",
    AuthType:            platform.AuthTypeBearerToken,
    RequiredCredentials: []string{"api_key"},
    SupportedOperations: []string{"get_status", "create_resource"},
    Capabilities: []platform.Capability{
        platform.CapBearerAuth,
        platform.CapHealthCheck,
        platform.CapRequestProxy,
    },
    WebsiteURL:     "https://myservice.com",
    DocsURL:        "https://myservice.com/docs",
    FreeTierStatus: "available",
    Enabled:        true,
}, 30*time.Second))
`

Available categories: CategoryDeveloper, CategoryCommunication, CategoryMonitoring, CategoryAutomation, CategoryScrapingData, CategoryStorage, CategoryPayments, CategoryAnalytics, CategoryMaps, CategorySecurity, CategoryCustom.

---

## Coding Standards

### Zero Comments — /nokomen

**All // inline comments and /* */ block comments inside .go files are forbidden.**

Documentation belongs in docs/*.md. Code must be self-documenting through precise naming.

This is enforced on every PR review. Code with comments will be rejected.

### Self-Documenting Names

`go
// WRONG — ambiguous abbreviations
func proc(r *providers.Request, c *providers.Credentials) (*providers.Response, error)

// RIGHT — clear intent from the name alone
func translateAndForwardToUpstream(req *providers.Request, creds *providers.Credentials) (*providers.Response, error)
`

### Error Handling

- Return *providers.ProviderError from all adapter methods — never raw error strings.
- Use providers.ClassifyHTTPError(statusCode, body) to translate HTTP status codes.
- When the upstream returns non-JSON (HTML error page), capture the first 300 bytes and include it in the error message — see internal/providers/openai/parser.go for the pattern.
- Never silently drop errors with _ = err.

### Secrets

- No API keys, tokens, or passwords in any source file.
- All secrets go in .env (already in .gitignore).
- The vault at runtime uses AES-256-GCM — never store plaintext credentials in the DB.

### Database Schema Changes

New tables or columns require a **new numbered migration file**:

`
migrations/0004_add_my_feature.sql
`

Never modify existing migrations/*.sql files — they are immutable once shipped.

### Frontend Standards

- Functional React components with hooks only — no class components.
- All backend API calls go through rontend/src/api/ typed wrappers.
- Style via the existing CSS system in rontend/src/index.css — no inline styles, no Tailwind.
- TypeScript strict mode — no ny casts without a comment explaining why.

---

## Testing Requirements

Every pull request must:

| Requirement | Command |
|-------------|---------|
| Build clean | go build ./... |
| All tests pass | go test ./... |
| No race conditions | go test -race ./... |
| No vet warnings | go vet ./... |

New code must include tests for:
- The happy path (success case)
- At least one error path (failure case)
- Any public function or exported method

---

## Pull Request Checklist

Copy and complete this checklist in your PR description:

`
- [ ] go build ./... compiles without errors
- [ ] go test ./... passes
- [ ] go test -race ./... passes
- [ ] go vet ./... has zero warnings
- [ ] No code comments (// or /* */) inside .go files
- [ ] No hardcoded secrets or credentials anywhere
- [ ] New AI provider registered in canonical_models.go (if applicable)
- [ ] New platform provider registered in catalog.go (if applicable)
- [ ] New DB schema has a migration file in migrations/ (if applicable)
- [ ] PR description explains what changed, why, and how to test it
`

---

## For AI Agents

This section is written specifically for AI coding agents (Claude, Gemini, GPT, Cursor, Copilot, etc.) working on this repository autonomously.

### Files to Read First

Before making any changes, read these files in order:

1. [internal/providers/types.go](internal/providers/types.go) — defines Adapter interface, Request, Response, StreamEvent, Capabilities, Credentials
2. [internal/providers/registry.go](internal/providers/registry.go) — defines ProviderError, ErrorClass, ClassifyHTTPError(), IsTransient()
3. [internal/gateway/gateway.go](internal/gateway/gateway.go) — the core request routing and fallback pipeline
4. [internal/providers/canonical_models.go](internal/providers/canonical_models.go) — all canonical model definitions for every provider

### Hard Rules for AI Agents

| Rule | Consequence of Violation |
|------|--------------------------|
| No // comments inside .go files | PR rejected |
| No hardcoded secrets in any file | Security incident |
| Never modify existing migrations/*.sql | Data corruption on upgrade |
| Never write directly to data/ekarouter.db in source code | Use migration SQL or the check_keys tool |
| Channel returned by ExecuteStream must be closed | Goroutine leak |
| Execute() must never return (nil, nil) | Runtime panic downstream |
| All outbound HTTP must use creds.HTTPClient | Bypasses proxy and SSRF guard |

### Common Task Patterns

#### Add/update a model for an existing provider
1. Edit internal/providers/canonical_models.go — add to the correct provider map key
2. If needed, update the DB with cmd/check_keys/main.go or a new SQL migration
3. Verify: go build ./... passes

#### Fix "invalid character" / "looking for beginning of value" errors
Upstream returned HTML instead of JSON. Pattern fix (see internal/providers/openai/parser.go):
`go
rawBody, _ := io.ReadAll(resp.Body)
if err := json.Unmarshal(rawBody, &result); err != nil {
    snippet := string(rawBody)
    if len(snippet) > 300 {
        snippet = snippet[:300]
    }
    return nil, providers.ClassifyHTTPError(resp.StatusCode, snippet)
}
`

#### Add a new admin API endpoint
1. Handler function → internal/httpapi/admin_handlers.go
2. Route registration → internal/httpapi/router.go
3. Typed fetch wrapper → rontend/src/api/
4. UI component → appropriate rontend/src/pages/ file

#### Update a provider's model list (e.g., Groq changed IDs)
1. Check the provider's official docs for current model IDs
2. Update internal/providers/canonical_models.go
3. Update cmd/check_keys/main.go if DB seeding is needed
4. Run go build ./... and go test ./...

#### Add a new route target in the DB
Use the admin API: POST /api/v1/routes/{routeID}/items with provider_id, ccount_id, model_name, priority.

### Architecture Constraints

| Constraint | Why It Exists |
|------------|---------------|
| No CGO anywhere | Must compile on Windows without GCC/MSVC |
| modernc.org/sqlite only | Pure Go, single static binary |
| sync.RWMutex for shared state | Concurrent request safety |
| AES-256-GCM for vault | Zero plaintext secrets at rest |
| SSRF guard on every outbound call | Defense against metadata service attacks |
| Cooldown manager for failed accounts | Prevents hammering broken credentials |
| Buffered chan providers.StreamEvent (size 16) | Backpressure management for SSE streams |

### DB Schema Quick Reference

`sql
-- Core AI routing tables
SELECT * FROM providers;
SELECT * FROM accounts WHERE provider_id = 'groq';
SELECT * FROM models WHERE provider_id = 'groq' AND enabled = 1;
SELECT * FROM routes;
SELECT * FROM route_items WHERE route_id = 'default';

-- Diagnose a broken account
SELECT a.id, a.alias, c.created_at FROM accounts a
JOIN credentials c ON c.account_id = a.id
WHERE a.provider_id = 'groq';

-- Check usage / errors
SELECT provider_id, model_id, status, error_class, COUNT(*) as cnt
FROM usage_records
WHERE created_at > datetime('now', '-1 hour')
GROUP BY provider_id, model_id, status, error_class;
`

---

## Questions?

Open a [GitHub Issue](https://github.com/dresar/ekarouter/issues) with the question label. For security issues, use [SECURITY] in the title.
