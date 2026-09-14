# Webhooks Package

## Purpose
The `webhooks` package manages inbound and outbound webhooks, cryptographically signs payloads with HMAC-SHA256, verifies inbound signatures in constant time, and dispatches outbound event notifications with delivery auditing.

## Responsibilities
- Cryptographic secret generation with 32-byte CSPRNG random hex.
- HMAC-SHA256 signing and constant-time signature verification (`hmac.Equal`).
- Outbound HTTP POST delivery with SSRF validation, delivery IDs, and timing metrics.
- SQLite persistence for webhook endpoints and delivery history in `webhook_deliveries`.

## Public Interfaces
- `GenerateSecret() (string, error)`
- `SignPayload(secret string, payload []byte) string`
- `VerifySignature(secret string, payload []byte, signature string) bool`
- `NewManager(db *sql.DB, allowLocal bool) *Manager`
- `(*Manager) CreateWebhook(ctx context.Context, name, targetURL, events, encryptedSecret string) (*Webhook, error)`
- `(*Manager) ListWebhooks(ctx context.Context) ([]*Webhook, error)`
- `(*Manager) Dispatch(ctx context.Context, webhookID, rawSecret string, event string, payload []byte) (*Delivery, error)`
- `(*Manager) ListDeliveries(ctx context.Context, webhookID string) ([]*Delivery, error)`

## Security Considerations
- Outbound deliveries enforce SSRF checks against internal networks.
- Inbound signature checks use constant-time comparison to prevent timing attacks.
- Webhook secret keys are never logged in delivery history.
