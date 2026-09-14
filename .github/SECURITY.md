# Security Policy — EkaRouter

The EkaRouter project treats security as a core architectural foundation.

---

## Supported Versions

| Version | Supported |
| :--- | :--- |
| `v1.x` (current) | :white_check_mark: |
| `< v1.0` | :x: |

---

## Reporting a Vulnerability

If you discover a security vulnerability in EkaRouter, please do **NOT** open a public GitHub issue. 

Instead, report it directly and privately to the project architect:
- **Lead Architect**: Eka Syarif Maulana ([@dresar](https://github.com/dresar))
- **Email**: [eka.ckp16799@gmail.com](mailto:eka.ckp16799@gmail.com)

Please include:
1. Description of the vulnerability and attack vector.
2. Steps to reproduce or proof-of-concept (PoC).
3. Potential impact on encrypted credentials, SSRF bypass, or denial of service.

We strive to acknowledge receipt within 24 hours and provide an evaluation and mitigation plan promptly.

---

## Core Security Architectures

1. **Vault Encryption**:
   - All credentials and provider keys stored in SQLite are protected with **AES-256-GCM** authenticated encryption.
   - Master secret key must meet minimum entropy requirements.
   - Encrypted values are versioned (`v1:<base64>`) and decrypted only in-memory at execution time.
2. **SSRF Protection**:
   - Outbound proxy and generic tool calls strictly validate target addresses via `internal/platform/ssrf.go`.
   - Blocks `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, IPv6 loopbacks, and cloud instance metadata services.
3. **Cryptographic Timings**:
   - Webhook HMAC-SHA256 signatures and authentication token validations use `crypto/subtle.ConstantTimeCompare` to eliminate timing side-channel leaks.
4. **Zero Hardcoded Secrets**:
   - Under no circumstances are secrets, API keys, or private tokens committed to the codebase.
