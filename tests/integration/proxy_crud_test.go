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

	t.Run("Read Proxy Profile by ID and verify password masking", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/proxies/"+profileID, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Get proxy failed: %d %s", res.Code, res.Body.String())
		}

		if bytes.Contains(res.Body.Bytes(), []byte(rawPass)) {
			t.Fatalf("SECURITY VIOLATION: Plaintext proxy password returned in get response")
		}

		var prof struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			Host           string `json:"host"`
			Port           int    `json:"port"`
			MaskedPassword string `json:"masked_password"`
			Enabled        bool   `json:"enabled"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &prof)
		if prof.ID != profileID || prof.Name != "QA Residential Proxy" {
			t.Fatalf("Profile data mismatch: got %+v", prof)
		}
		if prof.MaskedPassword == "" || prof.MaskedPassword == rawPass {
			t.Fatalf("Expected masked password, got: %s", prof.MaskedPassword)
		}

		resAlias := env.AdminReq(http.MethodGet, "/api/proxy-profiles/"+profileID, nil)
		if resAlias.Code != http.StatusOK {
			t.Fatalf("Get proxy via alias failed: %d", resAlias.Code)
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

	t.Run("Update Proxy Profile", func(t *testing.T) {
		updatedName := "QA Updated Residential Proxy"
		updatedPort := 9090
		resUpdate := env.AdminReq(http.MethodPut, "/api/proxies/"+profileID, map[string]any{
			"name": updatedName,
			"port": updatedPort,
		})
		if resUpdate.Code != http.StatusOK {
			t.Fatalf("Update proxy failed: %d %s", resUpdate.Code, resUpdate.Body.String())
		}

		resGet := env.AdminReq(http.MethodGet, "/api/proxies/"+profileID, nil)
		var prof struct {
			Name string `json:"name"`
			Port int    `json:"port"`
		}
		_ = json.Unmarshal(resGet.Body.Bytes(), &prof)
		if prof.Name != updatedName || prof.Port != updatedPort {
			t.Fatalf("Update verification failed: got %+v", prof)
		}
	})

	t.Run("Disable and Enable Proxy Profile", func(t *testing.T) {
		resDis := env.AdminReq(http.MethodPost, "/api/proxies/"+profileID+"/disable", nil)
		if resDis.Code != http.StatusOK {
			t.Fatalf("Disable proxy failed: %d %s", resDis.Code, resDis.Body.String())
		}
		resGet1 := env.AdminReq(http.MethodGet, "/api/proxies/"+profileID, nil)
		var p1 struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.Unmarshal(resGet1.Body.Bytes(), &p1)
		if p1.Enabled {
			t.Fatalf("Expected proxy to be disabled, got enabled=true")
		}

		resEn := env.AdminReq(http.MethodPost, "/api/proxies/"+profileID+"/enable", nil)
		if resEn.Code != http.StatusOK {
			t.Fatalf("Enable proxy failed: %d %s", resEn.Code, resEn.Body.String())
		}
		resGet2 := env.AdminReq(http.MethodGet, "/api/proxies/"+profileID, nil)
		var p2 struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.Unmarshal(resGet2.Body.Bytes(), &p2)
		if !p2.Enabled {
			t.Fatalf("Expected proxy to be enabled, got enabled=false")
		}
	})

	t.Run("Test Proxy Connection Endpoint", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/proxies/"+profileID+"/test", nil)
		if res.Code == http.StatusInternalServerError {
			t.Fatalf("Test proxy connection resulted in 500 internal server error: %s", res.Body.String())
		}
	})

	t.Run("Delete Proxy Profile and verify deletion", func(t *testing.T) {
		res := env.AdminReq(http.MethodDelete, "/api/proxies/"+profileID, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Delete proxy failed: %d", res.Code)
		}

		resAfter := env.AdminReq(http.MethodGet, "/api/proxies/"+profileID, nil)
		if resAfter.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 after deletion, got %d", resAfter.Code)
		}

		resDelAgain := env.AdminReq(http.MethodDelete, "/api/proxies/"+profileID, nil)
		if resDelAgain.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 on duplicate delete, got %d", resDelAgain.Code)
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
