# EkaRouter Webhooks API & Event Dispatcher

## 1. Overview

EkaRouter features an asynchronous, reliable webhook event notification engine. When critical infrastructure events occur—such as credential cooldowns, rate limit breaches, secret rotations, or provider degradations—EkaRouter dispatches signed HTTP POST notifications to registered subscriber endpoints.

---

## 2. Event Types Catalog

| Event Name | Trigger Condition | Severity |
|---|---|---|
| `credential.cooldown` | A provider credential encountered consecutive rate limits or HTTP 429/5xx errors and was temporarily removed from rotation. | Warning |
| `credential.recovered` | A credential cooldown expired and the credential successfully passed a health test. | Info |
| `credential.rotated` | An administrator or automated agent rotated an API key in the vault. | Info |
| `provider.health_degraded` | Upstream provider latency exceeded threshold or periodic probe failed. | Error |
| `quota.threshold_reached` | An environment or project reached 80% or 100% of its monthly token/spend quota. | Warning |
| `route.fallback_triggered` | Primary model route failed; request was transparently rerouted to secondary provider. | Info |

---

## 3. Webhook Payload Structure

All webhook events deliver a JSON body with consistent top-level fields:

```json
{
  "id": "evt_01j8k9m012345678",
  "event": "credential.cooldown",
  "created_at": "2026-09-15T04:55:00Z",
  "data": {
    "credential_id": "cred_01j8k9m0",
    "provider_id": "openai",
    "masked_key": "sk-proj-***4b",
    "reason": "HTTP 429 Too Many Requests: Rate limit exceeded for organization",
    "cooldown_duration_seconds": 60,
    "cooldown_expires_at": "2026-09-15T04:56:00Z",
    "consecutive_failures": 3
  }
}
```

---

## 4. Security & Signature Verification

To prevent spoofing and tampering, every outgoing webhook request includes a cryptographic signature in the `X-EkaRouter-Signature` HTTP header:

```http
X-EkaRouter-Signature: sha256=a5b3e2...c8d1
```

The signature is computed as an HMAC-SHA256 of the raw request payload bytes using the webhook secret configured during registration.

### Signature Verification Examples

#### Node.js / TypeScript
```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signatureHeader, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  const digest = 'sha256=' + hmac.update(payload).digest('hex');
  return crypto.timingSafeEqual(Buffer.from(digest), Buffer.from(signatureHeader));
}
```

#### Python 3
```python
import hmac
import hashlib

def verify_webhook_signature(payload_bytes: bytes, signature_header: str, secret: str) -> bool:
    expected = "sha256=" + hmac.new(secret.encode('utf-8'), payload_bytes, hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, signature_header)
```

#### Go
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func VerifySignature(secret string, payload []byte, sigHeader string) bool {
    parts := strings.SplitN(sigHeader, "=", 2)
    if len(parts) != 2 || parts[0] != "sha256" {
        return false
    }
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMac := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expectedMac), []byte(parts[1]))
}
```

---

## 5. Delivery Guarantees & Retry Policy

- **Initial Delivery**: Instant dispatch via background goroutine upon event occurrence.
- **Retries**: If the receiving endpoint returns any status code other than `2xx` or times out (10s default), EkaRouter retries with exponential backoff:
  - 1st retry: after 30 seconds
  - 2nd retry: after 2 minutes
  - 3rd retry: after 10 minutes
- **SSRF Restriction**: Subscriber target URLs are strictly validated against EkaRouter's socket blocklist. Webhook destinations pointing to `127.0.0.1`, loopbacks, private RFC1918 addresses, or cloud metadata endpoints are immediately rejected.
