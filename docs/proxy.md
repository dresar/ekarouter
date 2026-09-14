# Controlled API Proxy Mode

EkaRouter provides a controlled reverse API proxy mode that injects vault credentials, strips sensitive headers, enforces rate limits and SSRF defenses, and logs usage.

## Proxy Endpoint

```http
ANY /api/v1/proxy/{provider}/*
Authorization: Bearer <client_token>
```

Example request to Cloudflare DNS through EkaRouter proxy:
```bash
curl -X GET http://localhost:8080/api/v1/proxy/cloudflare/zones \
  -H "Authorization: Bearer eka_pat_1234567890abcdef..."
```

EkaRouter will:
1. Verify the client token and evaluate RBAC permissions.
2. Select an eligible, healthy credential for `cloudflare` from the vault using priority rotation.
3. Decrypt the secret in memory.
4. Inject the authorization header (`Authorization: Bearer <token>`).
5. Execute the upstream request with SSRF validation and timeout enforcement.
6. Stream/relay the upstream response back to the client.
7. Record request metrics and update credential health.
