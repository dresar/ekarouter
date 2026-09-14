# Security & SSRF Protection Architecture

## 1. Vault Cryptography
All third-party credentials and sensitive connection parameters are stored using authenticated AES-256-GCM symmetric encryption:
- Key Derivation: SHA-256 derivation from `EKAROUTER_SECRET_KEY`.
- Nonce: 12-byte CSPRNG random nonce per encryption operation.
- Ciphertext Prefix: Versioned payload format (`v1:<base64(nonce + ciphertext + tag)>`).
- Tamper Protection: Built-in GCM authentication tag verification rejects any altered payload with an authentication failure.

## 2. Credential Masking Standard
Plaintext secrets are masked upon intake and never surfaced in full in logs, telemetry, API responses, or error messages:
- Tokens with prefixes (`sk-proj-1234567890abcdef`): `sk-****cdef`
- GitHub tokens (`ghp_1234567890abcdef1234567890abcdef`): `ghp_****cdef`
- Bearer headers (`Bearer secret-token`): `Bearer secr****oken`
- Generic keys: First 2-4 characters + `****` + last 2-4 characters.

## 3. Server-Side Request Forgery (SSRF) Guard
Any user-configured outbound URL (tools, webhooks, custom providers, OpenAPI import) passes through strict SSRF validation (`platform.ValidateSSRF`):
- Protocol Filtering: Strictly allows `http://` and `https://`. Rejects `file://`, `ftp://`, `gopher://`, etc.
- Hostname Filtering: Prohibits `localhost`, `0.0.0.0`, `127.0.0.1`, `::1`, and domains ending in `.local`, `.internal`, `.localhost`.
- IP Range Blacklisting:
  - Loopback (`127.0.0.0/8`, `::1/128`)
  - RFC 1918 Private IPv4 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`)
  - Link-Local IPv4 & IPv6 (`169.254.0.0/16`, `fe80::/10`)
  - Cloud Metadata Services (`169.254.169.254`, `metadata.google.internal`)

## 4. Header Injection & CRLF Protection
Template interpolation across request headers explicitly sanitizes both header names and values:
- Rejects any header key containing `:`, `\r`, or `\n`.
- Rejects any header value containing `\r` or `\n` to eliminate HTTP response splitting.
