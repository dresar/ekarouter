package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProxyProfilesCRUD(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	var profileID string
	rawPass := "super_secret_proxy_pass_12345"

	t.Run("Create Proxy Profile with credentials", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/proxies", map[string]any{
			"name":     "QA Residential Proxy",
			"scheme":   "http",
			"host":     "proxy.qa-partner.com",
			"port":     8080,
			"username": "qa_user",
			"password": rawPass,
			"enabled":  true,
		})
		if res.Code != http.StatusOK && res.Code != http.StatusCreated {
			t.Fatalf("Create proxy failed: %d %s", res.Code, res.Body.String())
		}
		var created struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &created)
		if created.ID == "" {
			t.Fatalf("Expected proxy ID in response, got empty: %s", res.Body.String())
		}
		profileID = created.ID

		if bytes.Contains(res.Body.Bytes(), []byte(rawPass)) {
			t.Fatalf("SECURITY VIOLATION: Plaintext proxy password returned in create response")
		}
	})

	t.Run("List Proxies and verify password masking", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/proxies", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("List proxies failed: %d", res.Code)
		}

		if bytes.Contains(res.Body.Bytes(), []byte(rawPass)) {
			t.Fatalf("SECURITY VIOLATION: Plaintext proxy password returned in list response")
		}
	})

	t.Run("List Proxy Profiles alias endpoint", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/proxy-profiles", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("List proxy-profiles failed: %d", res.Code)
		}
	})

	t.Run("Test Proxy Connection Endpoint", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/proxies/"+profileID+"/test", nil)
		if res.Code == http.StatusInternalServerError {
			t.Fatalf("Test proxy connection resulted in 500 internal server error: %s", res.Body.String())
		}
	})

	t.Run("Delete Proxy Profile", func(t *testing.T) {
		res := env.AdminReq(http.MethodDelete, "/api/proxies/"+profileID, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Delete proxy failed: %d", res.Code)
		}
	})
}

func TestUniversalPlatformProxyEndpoint(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("Universal proxy endpoint requires valid authentication", func(t *testing.T) {
		res := env.Request(http.MethodGet, "/api/v1/proxy/cloudflare/zones", nil, "")
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 unauthorized without auth, got %d", res.Code)
		}
	})

	t.Run("Universal proxy endpoint with auth and missing credential returns helpful error", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/proxy/cloudflare/zones", nil)
		if res.Code == http.StatusInternalServerError {
			t.Fatalf("Expected structured 4xx error for missing provider credential, got 500: %s", res.Body.String())
		}
	})
}
