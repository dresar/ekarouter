# EkaRouter Provider Integration Report

**Date:** 2026-09-15  
**Version:** 1.0.0  
**Scope:** Validation of 23 AI Provider Adapters and 27 Developer Platform Adapters

---

## 1. AI Provider Adapters Summary

Every registered AI provider adapter was validated for request construction, error classification, headers, authentication method, and stream handling.

| Provider | Adapter Kind | Auth Type | Base URL / Endpoint | Streaming (SSE) | Tool Calling | Status |
|---|---|---|---|---|---|---|
| **OpenAI** | `openai` | Bearer Token | `https://api.openai.com/v1` | Supported | Supported | **ACTIVE** |
| **Anthropic** | `anthropic` | Header (`x-api-key`) | `https://api.anthropic.com/v1` | Supported | Supported | **ACTIVE** |
| **Google Gemini** | `gemini` | Query Param / Header | `https://generativelanguage.googleapis.com/v1beta` | Supported | Supported | **ACTIVE** |
| **Gemini CLI** | `gemini-cli` | Local CLI Token | Internal Pipe | Supported | Supported | **ACTIVE** |
| **Antigravity** | `antigravity` | Bearer / Internal | Internal Route | Supported | Supported | **ACTIVE** |
| **Groq** | `groq` | Bearer Token | `https://api.groq.com/openai/v1` | Supported | Supported | **ACTIVE** |
| **Cerebras** | `cerebras` | Bearer Token | `https://api.cerebras.ai/v1` | Supported | Supported | **ACTIVE** |
| **OpenRouter** | `openrouter` | Bearer Token | `https://openrouter.ai/api/v1` | Supported | Supported | **ACTIVE** |
| **Cloudflare AI** | `cloudflare` | Bearer Token | `https://api.cloudflare.com/client/v4/accounts/{acc}/ai` | Supported | Partial | **ACTIVE** |
| **NVIDIA NIM** | `nvidia` | Bearer Token | `https://integrate.api.nvidia.com/v1` | Supported | Supported | **ACTIVE** |
| **Ollama** | `ollama` | Optional / Local | `http://localhost:11434/v1` | Supported | Partial | **ACTIVE** |
| **Hugging Face** | `huggingface` | Bearer Token | `https://api-inference.huggingface.co/models` | Supported | Partial | **ACTIVE** |
| **DeepSeek (Custom)**| `deepseek` | Bearer Token | `https://api.deepseek.com/v1` | Supported | Supported | **ACTIVE** |
| **Mistral (Custom)** | `mistral` | Bearer Token | `https://api.mistral.ai/v1` | Supported | Supported | **ACTIVE** |
| **Together AI** | `together` | Bearer Token | `https://api.together.xyz/v1` | Supported | Supported | **ACTIVE** |
| **vLLM (Custom)** | `vllm` | Custom Base URL | Configurable | Supported | Supported | **ACTIVE** |
| **Airforce** | `airforce` | Bearer Token | `https://api.airforce/v1` | Supported | Partial | **ACTIVE** |
| **Bazaarlink** | `bazaarlink` | Bearer Token | `https://api.bazaarlink.com/v1` | Supported | Partial | **ACTIVE** |
| **Chutes** | `chutes` | Bearer Token | `https://api.chutes.ai/v1` | Supported | Partial | **ACTIVE** |
| **Coqui TTS** | `coqui` | Custom / Local | Configurable | Unsupported | N/A | **ACTIVE** |
| **Edge-TTS** | `edgetts` | WebSocket / Stream | Internal WebSocket | Supported | N/A | **ACTIVE** |
| **Devin** | `devin` | Bearer Token | `https://api.devin.ai/v1` | Supported | Partial | **ACTIVE** |
| **SearXNG** | `searxng` | Query Param / Custom| Configurable Base URL | N/A | Search Tool | **ACTIVE** |

---

## 2. Developer Platform Adapters Summary (27 Providers)

Configured in `internal/platform/catalog.go`:

| Provider ID | Category | Auth Header / Scheme | Operations Supported | Free Tier |
|---|---|---|---|---|
| `cloudflare` | Developer | `Authorization: Bearer <token>` | `list_zones`, `dns_records`, `workers` | Unlimited DNS, 100k req/day |
| `github` | Developer | `Authorization: Bearer <pat>` | `get_repo`, `list_issues`, `trigger_workflow` | 5,000 req/hr |
| `vercel` | Developer | `Authorization: Bearer <token>` | `list_deployments`, `create_deployment` | Hobby tier |
| `supabase` | Developer | `Authorization: Bearer <token>` | `list_projects`, `run_query` | 2 projects, 500MB DB |
| `neon` | Developer | `Authorization: Bearer <token>` | `list_projects`, `create_branch` | 0.5 GiB storage |
| `resend` | Communication | `Authorization: Bearer <key>` | `send_email`, `list_domains` | 3,000 emails/month |
| `sendgrid` | Communication | `Authorization: Bearer <key>` | `mail_send`, `get_stats` | 100 emails/day |
| `twilio` | Communication | `Authorization: Basic <base64>` | `send_sms`, `make_call` | Trial credit |
| `discord` | Communication | `Authorization: Bot <token>` | `create_message`, `execute_webhook` | Free bot rate limits |
| `slack` | Communication | `Authorization: Bearer <token>` | `chat.postMessage`, `conversations.list` | Free workspace apps |
| `sentry` | Monitoring | `Authorization: Bearer <token>` | `list_issues`, `get_issue` | 5,000 errors/month |
| `betterstack` | Monitoring | `Authorization: Bearer <token>` | `list_monitors`, `create_monitor` | 10 monitors (3-min checks) |
| `webhook` | Automation | `Custom Header + HMAC` | `dispatch_event` | Self-hosted |
| `firecrawl` | Scraping & Data | `Authorization: Bearer <key>` | `scrape`, `crawl`, `map` | 500 credits |
| `tavily` | Scraping & Data | `api-key: <key>` | `search`, `extract` | 1,000 searches/month |
| `serpapi` | Scraping & Data | `Authorization: <key>` | `search` | 100 searches/month |
| `jina` | Scraping & Data | `Authorization: Bearer <key>` | `read_url`, `search` | 1,000,000 tokens |
| `cloudflare-r2` | Storage | `Authorization: Bearer <token>` | `list_buckets`, `get_bucket` | 10 GB/month |
| `stripe` | Payments | `Authorization: Bearer <key>` | `create_customer`, `create_payment_intent` | Unlimited test mode |
| `midtrans` | Payments | `Authorization: Basic <base64>` | `charge`, `status`, `cancel` | Free sandbox |
| `posthog` | Analytics | `Authorization: Bearer <key>` | `capture_event`, `list_cohorts` | 1,000,000 events/month |
| `umami` | Analytics | `Authorization: Bearer <key>` | `get_websites`, `get_stats` | 10k events/month |
| `mapbox` | Maps & Geo | Query Param / Custom Header | `geocoding`, `directions` | 50,000 req/month |
| `virustotal` | Security | `x-apikey: <key>` | `get_ip_report`, `get_domain_report` | 500 req/day |
| `abuseipdb` | Security | `Key: <key>` | `check_ip`, `report_ip` | 1,000 checks/day |
| `ipinfo` | Security | `Authorization: Bearer <token>` | `lookup_ip` | 50,000 lookups/month |
| `generic-rest` | Custom | Custom Header + SSRF check | `get`, `post`, `put`, `patch`, `delete` | Custom |

---

## 3. Provider Error Handling & Classification

All provider HTTP responses are converted into structured `ProviderError` instances using `providers.ClassifyHTTPError()`:
- `HTTP 400`: `ErrorClassInvalidRequest` (Non-transient)
- `HTTP 401 / 403`: `ErrorClassAuth` (Triggers credential disable / cooldown)
- `HTTP 402 / Quota Exceeded`: `ErrorClassQuota` (Triggers credential cooldown & rotation)
- `HTTP 429`: `ErrorClassRateLimit` (Triggers exponential backoff cooldown)
- `HTTP 500 / 502 / 503 / 504 / 529`: `ErrorClassUpstream5xx` (Triggers immediate fallback to secondary candidate)
