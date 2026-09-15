package quota

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/db"
)

func TestQuotaTrackerGeneric(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	defer database.Close()

	if err := database.Migrate("../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	crypto, err := auth.NewCryptoService("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("crypto: %v", err)
	}

	_, _ = database.Exec(`
		INSERT INTO providers (id, key, name, kind, base_url, enabled)
		VALUES ('openai', 'key_openai', 'OpenAI', 'openai', 'https://api.openai.com', 1)
	`)
	_, _ = database.Exec(`
		INSERT INTO accounts (id, provider_id, name, auth_type, state, enabled)
		VALUES ('acc_test', 'openai', 'Test Account', 'apikey', 'active', 1)
	`)

	tracker := NewTracker(database, crypto, http.DefaultClient)
	quotas, err := tracker.GetAllQuotas(context.Background(), true)
	if err != nil {
		t.Fatalf("get all quotas error: %v", err)
	}
	if len(quotas) != 1 {
		t.Fatalf("expected 1 quota, got %d", len(quotas))
	}
	if quotas[0].AccountID != "acc_test" {
		t.Errorf("expected acc_test, got %s", quotas[0].AccountID)
	}
	if quotas[0].OverallRemaining != 100 {
		t.Errorf("expected 100%% remaining, got %f", quotas[0].OverallRemaining)
	}
}

func TestAntigravityQuotaParsing(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"models": {
				"gemini-3.8-flash-high": {
					"displayName": "Gemini 3.8 Flash (High)",
					"quotaInfo": {
						"remainingFraction": 0.85,
						"resetTime": "2026-09-15T18:00:00Z"
					}
				}
			}
		}`))
	}))
	defer mockServer.Close()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	defer database.Close()

	if err := database.Migrate("../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	crypto, _ := auth.NewCryptoService("0123456789abcdef0123456789abcdef")
	encToken, _ := crypto.Encrypt("mock_token")

	_, _ = database.Exec(`
		INSERT INTO providers (id, key, name, kind, base_url, enabled)
		VALUES ('antigravity', 'key_ag', 'Antigravity', 'gemini-agy', 'https://daily-cloudcode-pa.googleapis.com', 1)
	`)
	_, _ = database.Exec(`
		INSERT INTO accounts (id, provider_id, name, auth_type, state, enabled)
		VALUES ('acc_ag', 'antigravity', 'AG Account', 'oauth', 'active', 1)
	`)
	_, _ = database.Exec(`
		INSERT INTO credentials (id, account_id, encrypted_access)
		VALUES ('cred_ag', 'acc_ag', ?)
	`, encToken)

	tracker := NewTracker(database, crypto, mockServer.Client())
	base := AccountQuota{
		AccountID:   "acc_ag",
		ProviderID:  "antigravity",
		AccountName: "AG Account",
	}

	parsed := tracker.fetchAntigravityQuota(context.Background(), base, "acc_ag", "mock_token", "")
	if parsed.Plan != "Pro Tier" {
		t.Errorf("expected Pro Tier, got %s", parsed.Plan)
	}
}
