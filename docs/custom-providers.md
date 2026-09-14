# Custom API Providers

EkaRouter allows users and administrators to register arbitrary HTTP REST endpoints as custom providers while maintaining full SSRF security and credential vault injection.

## Configuring a Custom Provider

Custom providers are registered with:
- `BaseURL`: Upstream HTTP/HTTPS base URL.
- `AuthType`: `bearer_auth`, `api_key_auth`, `basic_auth`, or `custom_header`.
- `AuthHeaderName`: Custom header name (e.g. `X-Custom-Token`, `App-Key`).
- `AllowedDomains`: Optional domain allowlist.

## Security Constraints
All custom provider URLs are strictly checked against SSRF rules. Requests directed to private networks, loopback IP addresses, or internal cloud metadata addresses are rejected before socket establishment.
