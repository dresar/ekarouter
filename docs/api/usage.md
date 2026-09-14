# EkaRouter Usage & Analytics API

## 1. Overview

EkaRouter provides multi-dimensional telemetry tracking for token consumption, request counts, response latency, and TokenSaver compression efficiency. Telemetry is tracked asynchronously in SQLite to preserve sub-millisecond gateway overhead.

---

## 2. Admin Usage Overview: `GET /api/usage`

Returns high-level request volume and aggregate token metrics for the management console.

- **URL**: `/api/usage`
- **Method**: `GET`
- **Authentication**: Admin Session Cookie or Bearer Token
- **Query Parameters**:
  - `since` (string, ISO 8601 timestamp, default: 24h ago)
  - `until` (string, ISO 8601 timestamp, default: now)
  - `limit` (integer, default: 50, max: 200)
- **Response Body (200 OK)**:
```json
{
  "total_requests": 1420,
  "successful_requests": 1412,
  "failed_requests": 8,
  "total_tokens": 1284500,
  "prompt_tokens": 945000,
  "completion_tokens": 339500,
  "tokens_saved": 189200,
  "avg_latency_ms": 345.8
}
```

---

## 3. Platform Usage APIs (`/api/v1/usage/*`)

### 3.1 Aggregated Summary: `GET /api/v1/usage/summary`
Returns metric rollups segmented by time windows.

- **URL**: `/api/v1/usage/summary`
- **Method**: `GET`
- **Authentication**: Platform Token / Admin Session
- **Query Parameters**:
  - `period` (string: `1h`, `24h`, `7d`, `30d`, default: `24h`)
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "period": "24h",
    "requests": {
      "total": 4820,
      "success": 4801,
      "error": 19,
      "error_rate_pct": 0.39
    },
    "tokens": {
      "prompt": 3410200,
      "completion": 1120400,
      "total": 4530600,
      "compressed_savings": 498000
    },
    "latency": {
      "p50_ms": 210,
      "p95_ms": 680,
      "p99_ms": 1420
    }
  }
}
```

### 3.2 Breakdown by Provider: `GET /api/v1/usage/providers`
Returns usage metrics grouped by upstream AI and platform providers.

- **URL**: `/api/v1/usage/providers`
- **Method**: `GET`
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": [
    {
      "provider_id": "openai",
      "provider_name": "OpenAI",
      "requests": 2840,
      "tokens": 2940000,
      "errors": 4,
      "avg_latency_ms": 280
    },
    {
      "provider_id": "anthropic",
      "provider_name": "Anthropic Claude",
      "requests": 1450,
      "tokens": 1240000,
      "errors": 2,
      "avg_latency_ms": 320
    },
    {
      "provider_id": "groq",
      "provider_name": "Groq Cloud",
      "requests": 530,
      "tokens": 350600,
      "errors": 13,
      "avg_latency_ms": 85
    }
  ]
}
```

### 3.3 Breakdown by Credential: `GET /api/v1/usage/credentials`
Shows token consumption and request velocity across individual credential vault entries. Useful for balancing costs across multiple team members or accounts.

- **URL**: `/api/v1/usage/credentials`
- **Method**: `GET`
- **Query Parameters**:
  - `provider_id` (optional string, filter by provider)
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": [
    {
      "credential_id": "cred_01j8k9m0",
      "name": "Prod OpenAI Key Primary",
      "provider_id": "openai",
      "masked_key": "sk-proj-***4b",
      "requests_count": 2100,
      "tokens_consumed": 2200400,
      "cooldown_triggers": 1,
      "last_used_at": "2026-09-15T04:52:10Z"
    }
  ]
}
```

### 3.4 Breakdown by Project: `GET /api/v1/usage/projects`
Aggregates usage grouped by developer projects.

- **URL**: `/api/v1/usage/projects`
- **Method**: `GET`
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": [
    {
      "project_id": "proj_prod_app",
      "project_name": "Production Web App",
      "requests": 3940,
      "tokens": 3890000
    },
    {
      "project_id": "proj_internal_tools",
      "project_name": "Internal Agentic Bots",
      "requests": 880,
      "tokens": 640600
    }
  ]
}
```

### 3.5 Specific Credential Counters: `GET /api/v1/credentials/{id}/usage`
Returns granular counters for a single credential.

- **URL**: `/api/v1/credentials/{id}/usage`
- **Method**: `GET`
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "credential_id": "cred_01j8k9m0",
    "total_requests": 2100,
    "total_errors": 1,
    "consecutive_errors": 0,
    "tokens_consumed": 2200400,
    "last_used_at": "2026-09-15T04:52:10Z"
  }
}
```

---

## 4. TokenSaver Telemetry & Compression Preview

### Preview Compaction: `POST /api/tokensaver/preview`
Enables developers to test how much whitespace, formatting redundancy, and prompt slop can be stripped before sending to LLM.

- **URL**: `/api/tokensaver/preview`
- **Method**: `POST`
- **Request Body**:
```json
{
  "text": "Please summarize this text: \n\n\n\n\n\n   The brown fox jumps over the lazy dog.    \n\n  "
}
```
- **Response Body (200 OK)**:
```json
{
  "original_length": 68,
  "compressed_length": 51,
  "reduction_percentage": 25.0,
  "estimated_original_tokens": 17,
  "estimated_compressed_tokens": 13,
  "compressed_text": "Please summarize this text:\nThe brown fox jumps over the lazy dog."
}
```
