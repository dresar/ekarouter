package freetier

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupCatalogTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "freetier_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.Close()

	db, err := sql.Open("sqlite", "file:"+tmpFile.Name()+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
CREATE TABLE IF NOT EXISTS free_tier_catalog (
    id TEXT PRIMARY KEY, provider_name TEXT NOT NULL, category TEXT NOT NULL DEFAULT 'general',
    official_website TEXT NOT NULL DEFAULT '', docs_url TEXT NOT NULL DEFAULT '',
    pricing_url TEXT NOT NULL DEFAULT '', free_tier_status TEXT NOT NULL DEFAULT 'unverified',
    free_quota TEXT NOT NULL DEFAULT '', reset_interval TEXT NOT NULL DEFAULT '',
    supported_regions TEXT NOT NULL DEFAULT '', signup_steps TEXT NOT NULL DEFAULT '',
    auth_method TEXT NOT NULL DEFAULT 'api_key', required_scopes TEXT NOT NULL DEFAULT '',
    expiration_behavior TEXT NOT NULL DEFAULT '', restrictions TEXT NOT NULL DEFAULT '',
    payment_required INTEGER NOT NULL DEFAULT 0, personal_key_required INTEGER NOT NULL DEFAULT 1,
    sandbox_support INTEGER NOT NULL DEFAULT 0, verified_at DATETIME,
    confidence TEXT NOT NULL DEFAULT 'low', status TEXT NOT NULL DEFAULT 'unverified',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS free_tier_sources (
    id TEXT PRIMARY KEY, catalog_id TEXT NOT NULL REFERENCES free_tier_catalog(id) ON DELETE CASCADE,
    source_url TEXT NOT NULL, source_type TEXT NOT NULL DEFAULT 'documentation',
    verified_at DATETIME, notes TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCatalogCRUD(t *testing.T) {
	db := setupCatalogTestDB(t)
	store := NewCatalogStore(db)
	ctx := context.Background()

	entry := &CatalogEntry{
		ProviderName:    "TestProvider",
		Category:        "ai_inference",
		OfficialWebsite: "https://test.example.com",
		FreeTierStatus:  "free_tier_available",
		AuthMethod:      "api_key",
		Confidence:      ConfidenceHigh,
		Status:          StatusVerified,
	}

	if err := store.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if entry.ID == "" {
		t.Fatal("entry ID empty")
	}

	got, err := store.GetEntry(ctx, entry.ID)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if got.ProviderName != "TestProvider" || got.Category != "ai_inference" {
		t.Errorf("entry mismatch: %+v", got)
	}

	entries, err := store.ListEntries(ctx, nil)
	if err != nil || len(entries) != 1 {
		t.Fatalf("list entries: err=%v, count=%d", err, len(entries))
	}

	filtered, err := store.ListEntries(ctx, &CatalogFilter{Category: "ai_inference"})
	if err != nil || len(filtered) != 1 {
		t.Fatalf("filter by category: err=%v, count=%d", err, len(filtered))
	}

	noResults, err := store.ListEntries(ctx, &CatalogFilter{Category: "nonexistent"})
	if err != nil || len(noResults) != 0 {
		t.Fatalf("expected 0 results for nonexistent category, got %d", len(noResults))
	}

	if err := store.DeleteEntry(ctx, entry.ID); err != nil {
		t.Fatalf("delete entry: %v", err)
	}
}

func TestCatalogSources(t *testing.T) {
	db := setupCatalogTestDB(t)
	store := NewCatalogStore(db)
	ctx := context.Background()

	entry := &CatalogEntry{ProviderName: "SourceTest", Category: "test"}
	_ = store.CreateEntry(ctx, entry)

	src := &Source{
		CatalogID:  entry.ID,
		SourceURL:  "https://docs.example.com/pricing",
		SourceType: "documentation",
		Notes:      "official pricing page",
	}
	if err := store.AddSource(ctx, src); err != nil {
		t.Fatalf("add source: %v", err)
	}

	sources, err := store.GetSources(ctx, entry.ID)
	if err != nil || len(sources) != 1 {
		t.Fatalf("get sources: err=%v, count=%d", err, len(sources))
	}
	if sources[0].SourceURL != "https://docs.example.com/pricing" {
		t.Errorf("source URL mismatch: %s", sources[0].SourceURL)
	}
}

func TestCatalogCategories(t *testing.T) {
	db := setupCatalogTestDB(t)
	store := NewCatalogStore(db)
	ctx := context.Background()

	_ = store.CreateEntry(ctx, &CatalogEntry{ProviderName: "A", Category: "ai_inference"})
	_ = store.CreateEntry(ctx, &CatalogEntry{ProviderName: "B", Category: "search"})
	_ = store.CreateEntry(ctx, &CatalogEntry{ProviderName: "C", Category: "tts"})

	cats, err := store.GetCategories(ctx)
	if err != nil {
		t.Fatalf("get categories: %v", err)
	}
	if len(cats) != 3 {
		t.Errorf("expected 3 categories, got %d: %v", len(cats), cats)
	}
}

func TestCatalogStatusUpdate(t *testing.T) {
	db := setupCatalogTestDB(t)
	store := NewCatalogStore(db)
	ctx := context.Background()

	entry := &CatalogEntry{ProviderName: "StatusTest", Category: "test", Status: StatusUnverified}
	_ = store.CreateEntry(ctx, entry)

	if err := store.UpdateStatus(ctx, entry.ID, StatusVerified); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, _ := store.GetEntry(ctx, entry.ID)
	if got.Status != StatusVerified {
		t.Errorf("expected verified, got %s", got.Status)
	}
}

func TestSeedCatalog(t *testing.T) {
	db := setupCatalogTestDB(t)
	store := NewCatalogStore(db)
	ctx := context.Background()

	if err := SeedCatalog(ctx, store); err != nil {
		t.Fatalf("seed catalog: %v", err)
	}

	entries, _ := store.ListEntries(ctx, nil)
	if len(entries) < 10 {
		t.Errorf("expected at least 10 seeded entries, got %d", len(entries))
	}

	if err := SeedCatalog(ctx, store); err != nil {
		t.Fatalf("re-seed should be idempotent: %v", err)
	}
	entries2, _ := store.ListEntries(ctx, nil)
	if len(entries2) != len(entries) {
		t.Errorf("re-seed created duplicates: %d vs %d", len(entries), len(entries2))
	}

	verified, _ := store.GetVerified(ctx)
	if len(verified) == 0 {
		t.Error("expected some verified entries after seed")
	}
}
