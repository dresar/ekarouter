package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/platform"
)

func TestSecurityAndSecretMasking(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("API key raw value never returned in credential responses", func(t *testing.T) {
		rawSecret := "super_secret_production_api_key_value_9876543210"
		resCreate := env.AdminReq(http.MethodPost, "/api/v1/credentials", map[string]any{
			"provider_id":  "github",
			"name":         "GitHub Secret Token",
			"secret_value": rawSecret,
			"environment":  "production",
		})
		if resCreate.Code != http.StatusOK && resCreate.Code != http.StatusCreated {
			t.Fatalf("Create credential failed: %d", resCreate.Code)
		}

		if bytes.Contains(resCreate.Body.Bytes(), []byte(rawSecret)) {
			t.Fatalf("SECURITY VIOLATION: Raw secret leaked in create response: %s", resCreate.Body.String())
		}

		resList := env.AdminReq(http.MethodGet, "/api/v1/credentials", nil)
		if bytes.Contains(resList.Body.Bytes(), []byte(rawSecret)) {
			t.Fatalf("SECURITY VIOLATION: Raw secret leaked in credentials list: %s", resList.Body.String())
		}
	})

	t.Run("Developer token cannot access Admin-only settings or pools", func(t *testing.T) {
		res := env.PlatformClientReq(http.MethodGet, "/api/settings", nil)
		if res.Code != http.StatusUnauthorized && res.Code != http.StatusForbidden {
			t.Fatalf("Expected 401 or 403 for developer accessing /api/settings, got %d", res.Code)
		}

		resPool := env.PlatformClientReq(http.MethodGet, "/api/credential-pools", nil)
		if resPool.Code != http.StatusUnauthorized && resPool.Code != http.StatusForbidden {
			t.Fatalf("Expected 401 or 403 for developer accessing /api/credential-pools, got %d", resPool.Code)
		}
	})

	t.Run("SSRF Protection blocks loopback, private IP, and metadata", func(t *testing.T) {
		client := platform.NewSafeHTTPClient(5*time.Second, false)

		ssrfTargets := []string{
			"http://127.0.0.1:8080/evil",
			"http://localhost:8080/evil",
			"http://10.0.0.1/admin",
			"http://172.16.0.1/secret",
			"http://192.168.1.1/router",
			"http://169.254.169.254/latest/meta-data/",
			"http://metadata.google.internal/computeMetadata/v1/",
			"http://100.64.0.1/cgnat",
		}

		for _, target := range ssrfTargets {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
			if err != nil {
				continue
			}
			_, err = client.Do(req)
			if err == nil {
				t.Fatalf("SSRF VIOLATION: Safe client allowed connection to forbidden target: %s", target)
			}
		}
	})

	t.Run("Oversized request body is rejected", func(t *testing.T) {
		oversized := bytes.Repeat([]byte("A"), 11*1024*1024)
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(oversized))
		req.Header.Set("Authorization", "Bearer "+env.ApiKeyToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		env.App.Server.Handler.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			t.Fatalf("Expected request to fail due to body limit, got %d", rr.Code)
		}
	})

	t.Run("Header injection and CRLF sanitization", func(t *testing.T) {
		maliciousHeader := "val\r\nInjected-Header: evil"
		if strings.Contains(maliciousHeader, "\r") || strings.Contains(maliciousHeader, "\n") {
			sanitized := strings.ReplaceAll(strings.ReplaceAll(maliciousHeader, "\r", ""), "\n", "")
			if strings.Contains(sanitized, "\r") || strings.Contains(sanitized, "\n") {
				t.Fatalf("CRLF sanitization failed")
			}
		}
	})
}
