# EkaRouter Credential Validation Report

**Audit Date:** 2026-09-15  
**Encryption Standard:** AES-256-GCM with SHA-256 PBKDF2 Key Derivation  
**Masking Standard:** 4-character prefix + `***` + 2-character suffix

---

## 1. Vault Security Architecture

Every credential stored via the platform APIs (`/api/v1/credentials`) or admin APIs (`/api/accounts`) is protected using multi-layered defense:

1. **Envelope Encryption**: Secrets are encrypted at rest using AES-256-GCM.
2. **Deterministic Fingerprinting**: A SHA-256 hash fingerprint is maintained for duplicate detection without decrypting stored ciphertext.
3. **Strict In-Memory Decryption**: Ciphertext is decrypted only immediately before initiating an outbound HTTP connection to the verified upstream provider.
4. **Zero Cleartext Serialization**: The `vault.Credential` struct declares `EncryptedValue string json:"-"`, ensuring standard JSON encoders omit ciphertext completely.

---

## 2. API Key Display and Masking Verification

Automated tests verified that cleartext secrets never leak in:
- `GET /api/v1/credentials` (List)
- `GET /api/v1/credentials/{id}` (Detail)
- `POST /api/v1/credentials` (Creation Response)
- `PATCH /api/v1/credentials/{id}` (Update Response)
- `POST /api/v1/credentials/{id}/rotate` (Rotation Response)
- `GET /api/v1/credentials/{id}/health` (Health Logs)
- `GET /api/v1/credentials/{id}/usage` (Usage Snapshots)
- `GET /api/proxies` and `GET /api/proxy-profiles` (Proxy List)
- Error payloads & panic recovery traces

| Verification Target | Raw Secret Sent | Returned Value in Response | Status |
|---|---|---|---|
| Cloudflare API Token | `cf_sec_1234567890abcdef1234567890` | `cf_s***90` | **PASS (Masked)** |
| GitHub Personal Token | `ghp_abcdefghijklmnopqrstuvwxyz123456` | `ghp_***56` | **PASS (Masked)** |
| OpenAI API Key | `sk-proj-1234567890abcdef1234567890` | `sk-p***90` | **PASS (Masked)** |
| Outbound Proxy Password | `super_secret_proxy_pass_12345` | `(omitted / masked)` | **PASS (Masked)** |

---

## 3. Credential Lifecycle & State Machine

Every credential exists in one of the following states:

- **`active`**: Eligible for selection by rotator and router.
- **`disabled`**: Manually suspended by operator; router skips.
- **`cooldown`**: Temporarily suspended due to upstream rate limits (429) or transient 5xx errors; auto-recovers after backoff duration expires.
- **`expired`**: Validated expiration timestamp has passed; excluded from candidate pool.
- **`exhausted`**: Upstream returned insufficient quota (402 or quota error); excluded until next billing cycle or manual reset.

---

## 4. Master Key Rotation

The vault supports zero-downtime master key rotation via `Store.RotateMasterKey(ctx, newVault)`. In automated test `TestDatabaseMigrationsAndIntegrity/Master_key_rotation_in_Vault`:
1. Credential encrypted with Key 1.
2. Master key rotated to Key 2.
3. Database rows re-encrypted in an atomic transaction.
4. Verification proved Key 2 successfully decrypts secret, while Key 1 can no longer decrypt the updated ciphertext.
