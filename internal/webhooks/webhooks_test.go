package webhooks_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/webhooks"
)

func TestWebhookSignAndVerify(t *testing.T) {
	secret, err := webhooks.GenerateSecret()
	if err != nil || len(secret) < 32 {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	payload := []byte(`{"event":"credential.rotated","id":"cred_123"}`)
	sig := webhooks.SignPayload(secret, payload)

	if !webhooks.VerifySignature(secret, payload, sig) {
		t.Fatal("VerifySignature failed for valid signature")
	}

	if !webhooks.VerifySignature(secret, payload, "sha256="+sig) {
		t.Fatal("VerifySignature failed with sha256= prefix")
	}

	if webhooks.VerifySignature(secret, payload, "bad-signature") {
		t.Fatal("VerifySignature passed for bad signature")
	}

	if webhooks.VerifySignature("wrong-secret", payload, sig) {
		t.Fatal("VerifySignature passed for wrong secret")
	}
}

func TestWebhookDispatch(t *testing.T) {
	var receivedSig string
	var receivedEvent string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-EkaRouter-Signature")
		receivedEvent = r.Header.Get("X-EkaRouter-Event")
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_wh.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	migPath := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migPath); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	mgr := webhooks.NewManager(database.DB, true)
	ctx := context.Background()

	wh, err := mgr.CreateWebhook(ctx, "Test Hook", server.URL, "tool.executed", "enc_secret")
	if err != nil {
		t.Fatalf("CreateWebhook: %v", err)
	}

	rawSecret := "test-raw-secret-123"
	payload := []byte(`{"status":"success"}`)

	delivery, err := mgr.Dispatch(ctx, wh.ID, rawSecret, "tool.executed", payload)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if delivery.ResponseStatus != http.StatusOK {
		t.Fatalf("Delivery response status expected 200, got %d", delivery.ResponseStatus)
	}

	if receivedEvent != "tool.executed" || receivedBody != `{"status":"success"}` {
		t.Fatalf("Server received unexpected payload: event=%s body=%s", receivedEvent, receivedBody)
	}

	if !webhooks.VerifySignature(rawSecret, payload, receivedSig) {
		t.Fatal("Received signature is not valid HMAC-SHA256")
	}

	deliveries, err := mgr.ListDeliveries(ctx, wh.ID)
	if err != nil || len(deliveries) != 1 {
		t.Fatalf("ListDeliveries failed: len=%d, err=%v", len(deliveries), err)
	}
}
