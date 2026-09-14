# EkaRouter Fallback Engine Report

**Audit Date:** 2026-09-15  
**Component:** `internal/routing` and `internal/gateway`  
**Test Suite:** `TestProviderFallbackExecution`, `TestGatewayRoutesAndCompletions`

---

## 1. Fallback Chain Architecture

EkaRouter supports multi-level fallback:
1. **Intra-Provider Fallback**: Rotates to next eligible credential within the same account or credential pool.
2. **Inter-Provider Route Fallback**: Cascades across ordered route items configured in `routes` and `route_items`.

---

## 2. Fallback Scenarios & Test Verification

| Scenario | Primary Candidate | Secondary Candidate | Trigger Condition | Actual Behavior | Result |
|---|---|---|---|---|---|
| **Upstream 5xx Outage** | OpenAI (GPT-4o) | Anthropic (Claude 3.5 Sonnet) | Upstream HTTP 500/502/503 | Cooldown assigned to primary; secondary immediately queried | **PASS** |
| **Rate Limit (429)** | Groq (Llama 3.3) | Cerebras (Llama 3.1) | Upstream 429 Too Many Requests | Exponential backoff cooldown assigned; traffic routed to secondary | **PASS** |
| **Quota Exhaustion** | OpenAI Key A | OpenAI Key B | Insufficient Quota error | Key A disabled/cooldown; rotated to Key B | **PASS** |
| **Network Timeout** | Provider A | Provider B | Upstream read timeout reached | Context timeout caught; next route target executed | **PASS** |
| **All Providers Down** | Provider A | Provider B (all failed) | All upstream targets exhausted | Clean 502 Bad Gateway returned without hanging | **PASS** |

---

## 3. Cooldown Multiplier Algorithm

When an upstream account fails, `routing.CooldownManager.MarkFailure(id, baseCooldown)` applies exponential backoff:
- 1st failure: `1x baseCooldown`
- 2nd failure: `2x baseCooldown`
- 3rd failure: `4x baseCooldown`
- 4th failure: `8x baseCooldown`
- 5th+ failure: `16x baseCooldown` (maximum cap)

Upon any successful request through that account, `CooldownManager.MarkSuccess(id)` immediately clears the failure counter and resets cooldown state.
