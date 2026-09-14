# Vault Package

## Purpose
The `vault` package provides secure authenticated encryption (AES-256-GCM), versioning, masking, and SQLite persistence for user-owned third-party credentials, tokens, and secrets.

## Responsibilities
- AES-256-GCM authenticated symmetric encryption with 12-byte random nonces and versioned ciphertext (`v1:<base64>`).
- Uniform secret masking (`sk-****abcd`, `ghp_****91xz`, `Bearer ****ef12`) preventing raw secret leakage.
- SQLite persistence for credential records, version history, and cooldown records.
- Concurrency-safe usage tracking and health state management.

## Public Interfaces
- `NewVault(secretKey string) (*Vault, error)`
- `(*Vault) Encrypt(plaintext string) (string, error)`
- `(*Vault) Decrypt(payload string) (string, error)`
- `MaskCredential(raw string) string`
- `NewStore(db *sql.DB, vault *Vault) *Store`
- `(*Store) CreateCredential(ctx context.Context, cred *Credential, rawSecret string) error`
- `(*Store) GetCredential(ctx context.Context, id string) (*Credential, error)`
- `(*Store) GetDecryptedSecret(ctx context.Context, id string) (string, error)`
- `(*Store) RotateSecret(ctx context.Context, id string, newSecret string) error`
- `(*Store) SetCooldown(ctx context.Context, id string, duration time.Duration, reason string) error`

## Security Considerations
- Plaintext secrets are never stored in the database.
- Ciphertext strings are excluded from JSON serialization via `json:"-"`.
- Errors never echo decrypted plaintext.
