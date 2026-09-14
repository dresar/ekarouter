package rbac_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/rbac"
)

func TestPasswordHashing(t *testing.T) {
	pwd := "SuperSecretPassword123!"
	encoded, err := rbac.HashPassword(pwd)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !rbac.CheckPassword(pwd, encoded) {
		t.Fatal("CheckPassword failed for correct password")
	}

	if rbac.CheckPassword("WrongPassword", encoded) {
		t.Fatal("CheckPassword succeeded for wrong password")
	}
}

func TestRBACAndTokens(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_rbac.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	svc := rbac.NewService(database.DB)
	ctx := context.Background()

	u, err := svc.CreateUser(ctx, "dev@example.com", "devPass123", "Developer One", "developer")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	authU, err := svc.AuthenticateUser(ctx, "dev@example.com", "devPass123")
	if err != nil || authU.ID != u.ID {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}

	if !svc.CheckPermission("developer", "tools.execute") {
		t.Fatal("Expected developer to have tools.execute")
	}

	if svc.CheckPermission("developer", "system.manage") {
		t.Fatal("Expected developer not to have system.manage")
	}

	if !svc.CheckPermission("admin", "system.manage") {
		t.Fatal("Expected admin to have all permissions")
	}

	proj, err := svc.CreateProject(ctx, "Fintech App", u.ID, "production", "Financial services API tools")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	rawToken, ct, err := svc.CreateClientToken(ctx, "CI Token", u.ID, proj.ID, "*", 30)
	if err != nil {
		t.Fatalf("CreateClientToken: %v", err)
	}

	verified, err := svc.VerifyClientToken(ctx, rawToken)
	if err != nil || verified.ID != ct.ID {
		t.Fatalf("VerifyClientToken failed: %v", err)
	}
}
