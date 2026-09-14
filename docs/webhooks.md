# Webhooks Management System

EkaRouter provides HMAC-SHA256 signature verification for inbound webhooks and reliable outbound webhook dispatching.

## Inbound Webhook Signature Verification
Inbound webhooks use constant-time comparison (`hmac.Equal`) to prevent timing side-channel attacks:

```go
isValid := webhooks.VerifySignature(secret, payloadBytes, req.Header.Get("X-EkaRouter-Signature"))
```

## Outbound Webhook Delivery
- Signatures are generated using `X-EkaRouter-Signature: sha256=<hex_hmac>`.
- Deliveries are logged in `webhook_deliveries` with HTTP status code, response time, and errors.
- Outbound endpoints undergo strict SSRF validation to prevent internal network scanning.
