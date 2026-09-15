package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
)

func TestStorageService(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "storage_test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.Migrate(migrationsDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	uploadDir := filepath.Join(tempDir, "uploads")
	svc := NewService(database.DB, uploadDir)
	ctx := context.Background()

	cfg, err := svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.Provider != "local" {
		t.Errorf("expected local provider, got %s", cfg.Provider)
	}

	err = svc.SaveConfig(ctx, &Config{
		Provider:            "cloudinary",
		CloudinaryCloudName: "my-cloud",
		CloudinaryAPIKey:    "my-key",
		CloudinaryAPISecret: "my-secret",
		CloudinaryFolder:    "ai-icons",
		ImageKitPublicKey:   "ik-pub",
		ImageKitPrivateKey:  "ik-priv",
		ImageKitURLEndpoint: "https://ik.imagekit.io/test",
		ImageKitFolder:      "/icons",
	})
	if err != nil {
		t.Fatalf("save config: %v", err)
	}

	cfg2, err := svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("get config2: %v", err)
	}
	if cfg2.CloudinaryCloudName != "my-cloud" {
		t.Errorf("expected my-cloud, got %s", cfg2.CloudinaryCloudName)
	}
	if cfg2.ImageKitPublicKey != "ik-pub" {
		t.Errorf("expected ik-pub, got %s", cfg2.ImageKitPublicKey)
	}

	localRes, err := svc.uploadLocal("test.png", []byte("sample-data"), "")
	if err != nil {
		t.Fatalf("upload local: %v", err)
	}
	if localRes.Provider != "local" {
		t.Errorf("expected provider local, got %s", localRes.Provider)
	}

	err = svc.SaveCustomIcon(ctx, "gemini", "https://res.cloudinary.com/demo/image/upload/gemini.png", "Google Gemini Pro", "cloudinary")
	if err != nil {
		t.Fatalf("save icon: %v", err)
	}

	icon, err := svc.GetCustomIcon(ctx, "gemini")
	if err != nil {
		t.Fatalf("get icon: %v", err)
	}
	if icon.IconURL != "https://res.cloudinary.com/demo/image/upload/gemini.png" {
		t.Errorf("unexpected icon url: %s", icon.IconURL)
	}

	icons, err := svc.ListCustomIcons(ctx)
	if err != nil {
		t.Fatalf("list icons: %v", err)
	}
	if len(icons) != 1 {
		t.Errorf("expected 1 icon, got %d", len(icons))
	}

	err = svc.DeleteCustomIcon(ctx, "gemini")
	if err != nil {
		t.Fatalf("delete icon: %v", err)
	}

	iconsAfter, _ := svc.ListCustomIcons(ctx)
	if len(iconsAfter) != 0 {
		t.Errorf("expected 0 icons, got %d", len(iconsAfter))
	}
}
