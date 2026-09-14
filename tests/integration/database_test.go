package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/vault"
)

func TestDatabaseMigrationsAndIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration_test.db")
	migPath := filepath.Join("..", "..", "migrations")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open failed: %v", err)
	}
	defer database.Close()

	t.Run("Apply migrations initially", func(t *testing.T) {
		if err := database.Migrate(migPath); err != nil {
			t.Fatalf("Initial migration failed: %v", err)
		}
	})

	t.Run("Migrations are idempotent when run again", func(t *testing.T) {
		if err := database.Migrate(migPath); err != nil {
			t.Fatalf("Idempotent migration run failed: %v", err)
		}
	})

	t.Run("Foreign key cascade behavior", func(t *testing.T) {
		ctx := context.Background()
		_, err := database.ExecContext(ctx, "INSERT INTO providers (id, key, name, kind, base_url) VALUES ('prov_casc', 'casc', 'Casc', 'openai', 'https://api.openai.com')")
		if err != nil {
			t.Fatalf("Insert provider failed: %v", err)
		}
		_, err = database.ExecContext(ctx, "INSERT INTO accounts (id, provider_id, name, auth_type) VALUES ('acc_casc', 'prov_casc', 'Casc Acc', 'api_key')")
		if err != nil {
			t.Fatalf("Insert account failed: %v", err)
		}
		_, err = database.ExecContext(ctx, "INSERT INTO credentials (id, account_id, encrypted_access) VALUES ('cred_casc', 'acc_casc', 'enc_val')")
		if err != nil {
			t.Fatalf("Insert credential failed: %v", err)
		}

		_, err = database.ExecContext(ctx, "DELETE FROM providers WHERE id = 'prov_casc'")
		if err != nil {
			t.Fatalf("Delete provider failed: %v", err)
		}

		var count int
		_ = database.QueryRowContext(ctx, "SELECT COUNT(1) FROM credentials WHERE id = 'cred_casc'").Scan(&count)
		if count != 0 {
			t.Fatalf("Expected credential to be cascaded and deleted, got count=%d", count)
		}
	})

	t.Run("Master key rotation in Vault", func(t *testing.T) {
		ctx := context.Background()
		key1 := "master-key-initial-123456789012"
		v1, err := vault.NewVault(key1)
		if err != nil {
			t.Fatalf("NewVault v1: %v", err)
		}
		vStore := vault.NewStore(database.DB, v1)

		rawVal := "live_secret_token_to_reencrypt"
		cred := &vault.Credential{
			Name:        "Rotate Master Key Cred",
			ProviderID:  "github",
			Environment: "production",
		}
		if err := vStore.CreateCredential(ctx, cred, rawVal); err != nil {
			t.Fatalf("CreateCredential: %v", err)
		}

		key2 := "master-key-secondary-9876543210"
		v2, err := vault.NewVault(key2)
		if err != nil {
			t.Fatalf("NewVault v2: %v", err)
		}

		if err := vStore.RotateMasterKey(ctx, v2); err != nil {
			t.Fatalf("RotateMasterKey: %v", err)
		}

		vStore2 := vault.NewStore(database.DB, v2)
		decrypted, err := vStore2.GetDecryptedSecret(ctx, cred.ID)
		if err != nil || decrypted != rawVal {
			t.Fatalf("Decrypted secret mismatch after key rotation: got %q, want %q, err: %v", decrypted, rawVal, err)
		}
	})

	t.Run("Database Backup operation", func(t *testing.T) {
		backupPath := filepath.Join(tempDir, "backup_test.db")
		if err := database.Backup(backupPath); err != nil {
			t.Fatalf("Database backup failed: %v", err)
		}
	})
}
