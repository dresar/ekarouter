# EkaRouter Provider Architecture & PRD Gap Analysis Report

## Executive Summary

This report documents the architectural status of the EkaRouter backend relative to the specifications in `EkaRouter_Backend_PRD/`, the complete modularization of `internal/providers/` into decoupled provider "brains", the resolution of the `/supermemory` boundary, and the remaining implementation gaps.

---

## 1. Provider Modularization Architecture ("Otak Sendiri")

The provider subsystem in `internal/providers/` has been completely restructured from monolithic files into dedicated domain packages. Each provider package contains its own protocols, transformers, signature stores, and capability profiles:

```
internal/providers/
├── registry.go              # Decoupled provider factory & HTTP error classification
├── types.go                 # Neutral internal request/response representations
├── README.md                # Subsystem architecture & integration guide
├── gemini/                  # Google Gemini Brain
│   ├── adapter.go           # Core Provider implementation & error classification
│   ├── api.go               # Google AI Studio REST v1beta payload transformer
│   ├── cli.go               # Gemini CLI/Companion protocol URL & headers
│   ├── agy.go               # Antigravity (AGY) protocol, system prompt handling & ThoughtSignatureStore
│   ├── parser.go            # SSE candidate stream parser & token accounting
│   ├── capabilities.go      # Model registry, vision/multimodal capabilities
│   ├── gemini_test.go       # Comprehensive test suite (8 tests)
│   └── README.md
├── openai/                  # OpenAI & Codex Brain
│   ├── adapter.go           # OpenAI adapter supporting Chat Completions & Responses API
│   ├── reasoning.go         # Reasoning effort parser (o1/o3/DeepSeek-R1) & Codex ID cleaner
│   ├── parser.go            # SSE chunk stream parser & usage delta extractor
│   ├── capabilities.go      # Vision, tools, json_schema capabilities
│   ├── openai_test.go       # Comprehensive test suite (8 tests)
│   └── README.md
├── anthropic/               # Anthropic Claude Brain
│   ├── adapter.go           # Claude Messages API adapter & header configuration
│   ├── thinking.go          # Thinking budget allocator & tool cloaking (_ide suffix)
│   ├── parser.go            # Claude SSE event processor (content_block_delta, message_delta)
│   ├── capabilities.go      # Claude model registry & token limits
│   ├── anthropic_test.go    # Comprehensive test suite (5 tests)
│   └── README.md
└── custom/                  # Self-Hosted & Alternative Engines Brain
    ├── adapter.go           # OpenAI-compatible engine adapter
    ├── backends.go          # Header injection (HTTP-Referer, X-Title for OpenRouter, Ollama, vLLM)
    ├── capabilities.go      # Dynamic backend capability profile
    ├── custom_test.go       # Integration & header forwarding tests (4 tests)
    └── README.md
```

### Key Architectural Highlights
1. **Gemini AGY Thinking Signature Cache**: Bounded in-memory store (`ThoughtSignatureStore`) with a 2,000-entry ceiling to prevent memory leaks, with automatic backfill of `DefaultThinkingAGSignature` for Gemini 3+ models.
2. **Anthropic Tool Cloaking**: Automatic recursive sanitization of tool names (appending `_ide`) when operating under OAuth tokens to bypass strict third-party tool restrictions.
3. **OpenAI Reasoning Normalization**: Universal mapping of reasoning effort levels across o1, o3, and DeepSeek-R1 variants, alongside Codex server ID stripping.
4. **Custom Backends**: Seamless routing to self-hosted Ollama, vLLM, DeepSeek, LocalAI, and OpenRouter with appropriate attribution headers.

---

## 2. Supermemory Boundary & Governance Directive

### Clarification & Policy
- **What `/supermemory` is**:
  - `/supermemory` is an **AI Agent Harness Skill** located in the operator's global Antigravity/Gemini configuration (`~/.gemini/config/skills/supermemory/SKILL.md`).
  - Its sole purpose is **cross-session memory and context persistence for AI coding assistants** (Google Antigravity, Claude Code, Cursor) so that architectural decisions, user preferences, and workspace context are preserved across turns.
- **What `/supermemory` is NOT**:
  - It is **NOT** a backend feature of EkaRouter.
  - It must **NEVER** be coded in Go, added to `go.mod`, or implemented as an interactive terminal CLI menu inside the `ekarouter` source tree.
- **Remediation**:
  - Any prior attempt to clone 9Router's interactive terminal CLI menus or import supermemory into the Go codebase was completely purged. EkaRouter remains a lightweight, single-process, headless AI Gateway.

---

## 3. Comprehensive PRD Gap Analysis (What Remains to be Built)

Based on a thorough review of `EkaRouter_Backend_PRD/` against the current implementation, the following components remain to be implemented in subsequent phases:

### A. Advanced Binary & CLI Provider Protocols (PRD Section 04)
1. **Cursor IDE (`cursor.go`)**:
   - Implementation of ConnectRPC framed binary Protobuf over HTTP/2 (`/agent.v1.AgentService/Run`).
   - Necessary for direct upstream connectivity to Cursor backends.
2. **AWS CodeWhisperer / Kiro (`kiro.go`)**:
   - AWS EventStream binary protocol framing with CRC32 checksums, payload signing, and auto-repair.
3. **Devin & Grok CLI (`devin.go`, `grok_cli.go`)**:
   - Devin CLI session state negotiation and Grok CLI rotating token lifecycle.

### B. Decoupled Translator Layer (PRD Layer 2)
- **Specification**: PRD Section 03 & 04 specify a dedicated, stateless `internal/translator/` package:
  - `request/`: Protocol converter from internal standard to provider format.
  - `response/`: Protocol converter from provider response/SSE to OpenAI standard.
  - `concerns/`: Cross-cutting concerns (thinking extraction, tool cloaking, context continuity).
- **Current State**: Protocol transformation is currently encapsulated inside each provider package (`gemini/api.go`, `anthropic/thinking.go`, `openai/reasoning.go`). While functional and modular, extracting pure translators into `internal/translator/` will allow zero-dependency unit testing.

### C. Gateway Capacity & Quota Management (PRD Section 05 & 12)
1. **Modality Capacity Adapter (`internal/gateway/capacity_adapter.go`)**:
   - Deep inspection of incoming request payloads to detect modalities (`vision`, `pdf`, `audio`, `function_calling`).
   - Automatically filter out account pools or provider routes that lack the required multimodal capacity before dispatching.
2. **Antigravity Quota Circuit Breaker (`internal/gateway/antigravity_quota.go`)**:
   - In-memory lazy quota cache with `singleflight` deduplication.
   - 3-strike circuit breaker pattern: 3 consecutive 429/quota-exceeded responses trigger an automatic 15-minute cooldown for that specific account.

### D. Extended Public Ingress Endpoints (PRD Section 07)
1. **Native Claude Messages Endpoint (`/v1/messages`)**:
   - Direct ingress for clients sending native Anthropic payload formats, translated into internal standard and dispatched across any configured provider.
2. **Media Endpoints**:
   - `/v1/embeddings`: Vector embedding proxy.
   - `/v1/images/generations`: DALL-E / Imagen image generation proxy.
   - `/v1/audio/speech` & `/v1/audio/transcriptions`: TTS and STT endpoints.

---

## 4. Verification & Status Summary

| Check | Command | Result |
|---|---|---|
| Go Test Suite | `go test -count=1 ./...` | 100% PASS (20 packages) |
| Go Vet | `go vet ./...` | 0 warnings |
| Go Formatting | `gofmt -l cmd internal scripts` | 0 files unformatted |
| Zero Comments Rule | Regex audit `^\s*(//|/\*)` | 0 comments (`/nokomen`) |
| Max File Length | `< 1000 lines` | 100% compliant (max 569 lines) |
| Documentation | `README.md` in all packages | 100% compliant |
| Git Status | Clean & pushed to `origin/main` | Commit `b942573` |
