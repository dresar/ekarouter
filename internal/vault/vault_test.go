package vault_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/vault"
)

func TestVaultEncryptionAndMasking(t *testing.T) {
	key := "super-secret-vault-master-key-32b"
	v, err := vault.NewVault(key)
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	secret := "sk-ant-api03-abcdef1234567890xyz"
	enc, err := v.Encrypt(secret)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == secret {
		t.Fatal("Ciphertext equals plaintext")
	}

	dec, err := v.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != secret {
		t.Fatalf("Decrypted mismatch: got %q, want %q", dec, secret)
	}

	masked := vault.MaskCredential(secret)
	if masked == secret {
		t.Fatalf("Masked leaked full secret: %s", masked)
	}
	if masked != "sk-****0xyz" {
		t.Logf("Masked format: %s", masked)
	}

	ghToken := "ghp_1234567890abcdef1234567890abcdef"
	ghMasked := vault.MaskCredential(ghToken)
	if ghMasked != "ghp_****cdef" {
		t.Logf("GitHub token masked: %s", ghMasked)
	}

	bearer := "Bearer my-secret-token-value"
	bearerMasked := vault.MaskCredential(bearer)
	if bearerMasked != "Bearer my-s****alue" {
		t.Logf("Bearer masked: %s", bearerMasked)
	}
}

func TestStoreCRUD(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_vault.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	v, _ := vault.NewVault("testing-secret-encryption-key-32")
	store := vault.NewStore(database.DB, v)

	ctx := context.Background()
	cred := &vault.Credential{
		Name:           "Production Stripe Key",
		CredentialType: vault.TypeAPIKey,
		ProviderID:     "stripe",
		Environment:    "production",
		Priority:       1,
		Tags:           "payments,live",
		Notes:          "Stripe live API key for checkout",
	}

	rawVal := "rk_live_51AbcDef1234567890"
	if err := store.CreateCredential(ctx, cred, rawVal); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	fetched, err := store.GetCredential(ctx, cred.ID)
	if err != nil {
		t.Fatalf("GetCredential: %v", err)
	}
	if fetched.Name != cred.Name || fetched.MaskedValue == "" {
		t.Fatalf("Invalid fetched credential: %+v", fetched)
	}

	decVal, err := store.GetDecryptedSecret(ctx, cred.ID)
	if err != nil {
		t.Fatalf("GetDecryptedSecret: %v", err)
	}
	if decVal != rawVal {
		t.Fatalf("Secret mismatch: got %q, want %q", decVal, rawVal)
	}

	newVal := "rk_live_99ZyxWvu9876543210"
	if err := store.RotateSecret(ctx, cred.ID, newVal); err != nil {
		t.Fatalf("RotateSecret: %v", err)
	}

	decNew, err := store.GetDecryptedSecret(ctx, cred.ID)
	if err != nil {
		t.Fatalf("GetDecryptedSecret after rotate: %v", err)
	}
	if decNew != newVal {
		t.Fatalf("Rotated mismatch: got %q, want %q", decNew, newVal)
	}

	if err := store.SetCooldown(ctx, cred.ID, 30*time.Second, "rate_limited_429"); err != nil {
		t.Fatalf("SetCooldown: %v", err)
	}

	if err := store.RecordUsage(ctx, cred.ID, false); err != nil {
		t.Fatalf("RecordUsage: %v", err)
	}

	list, err := store.ListCredentials(ctx, "stripe", "", "production")
	if err != nil || len(list) != 1 {
		t.Fatalf("ListCredentials: len=%d, err=%v", len(list), err)
	}

	if err := store.DeleteCredential(ctx, cred.ID); err != nil {
		t.Fatalf("DeleteCredential: %v", err)
	}
}
