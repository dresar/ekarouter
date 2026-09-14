package platform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
)

func TestPlatformRegistry(t *testing.T) {
	r := platform.NewRegistry()
	platform.RegisterDefaultProviders(r)

	list := r.List()
	if len(list) < 15 {
		t.Fatalf("Expected at least 15 providers, got %d", len(list))
	}

	cf, ok := r.Get("cloudflare")
	if !ok {
		t.Fatal("Cloudflare provider not found")
	}
	if cf.Metadata().Category != platform.CategoryDeveloper {
		t.Fatalf("Expected category developer, got %s", cf.Metadata().Category)
	}

	devs := r.ListByCategory(platform.CategoryDeveloper)
	if len(devs) == 0 {
		t.Fatal("Expected developer category providers")
	}

	searched := r.Search("stripe")
	if len(searched) == 0 || searched[0].ID != "stripe" {
		t.Fatalf("Expected stripe in search results, got %v", searched)
	}
}

func TestSSRFProtection(t *testing.T) {
	blockedURLs := []string{
		"http://localhost:8080/admin",
		"http://127.0.0.1:9090",
		"http://127.0.0.2:3000",
		"http://0.0.0.0:80",
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.5/api",
		"http://192.168.1.1/secret",
		"http://172.16.0.10:8000",
		"http://[::1]:8080",
		"http://[::ffff:127.0.0.1]:8080",
		"http://[fd12:3456:789a:1::1]:8080",
		"http://[fe80::1]:8080",
		"http://100.100.100.200:80",
		"http://metadata.google.internal/computeMetadata/v1/",
		"http://internal.service.local/status",
		"ftp://example.com/file",
		"file:///etc/passwd",
	}

	for _, u := range blockedURLs {
		err := platform.ValidateSSRF(u)
		if err == nil {
			t.Fatalf("Expected SSRF validation error for %s, but got nil", u)
		}
	}

	allowedURLs := []string{
		"https://api.github.com/user",
		"https://api.stripe.com/v1/charges",
		"https://api.resend.com/emails",
		"https://api.cloudflare.com/client/v4/zones",
	}

	for _, u := range allowedURLs {
		err := platform.ValidateSSRF(u)
		if err != nil {
			t.Fatalf("Unexpected SSRF validation error for %s: %v", u, err)
		}
	}
}

func TestBaseAdapterExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","delivered":true}`))
	}))
	defer server.Close()

	meta := platform.ProviderMetadata{
		ID:       "mock-dev",
		Name:     "Mock Developer API",
		Category: platform.CategoryDeveloper,
		BaseURL:  server.URL,
		AuthType: platform.AuthTypeBearerToken,
		Enabled:  true,
	}

	adapter := platform.NewBaseAdapterWithOptions(meta, 5*time.Second, true)

	req := &platform.ExecutionRequest{
		Path:             server.URL + "/test",
		Method:           http.MethodPost,
		Body:             []byte(`{"action":"ping"}`),
		CredentialSecret: "test-api-key",
	}

	resp, err := adapter.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("adapter.Execute error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", resp.StatusCode, string(resp.Body))
	}

	hStatus, err := adapter.HealthCheck(context.Background(), "test-api-key")
	if err != nil || !hStatus.Healthy {
		t.Fatalf("Health check failed: %+v, err: %v", hStatus, err)
	}
}

func TestSafeHTTPClientRedirectBlock(t *testing.T) {
	// Server redirects to loopback
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:8080/evil", http.StatusFound)
	}))
	defer server.Close()

	client := platform.NewSafeHTTPClient(5*time.Second, false)
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected error following redirect to 127.0.0.1, got nil")
	}
}

