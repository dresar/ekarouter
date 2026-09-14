package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSystemEndpoints(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("GET /health returns 200 and healthy", func(t *testing.T) {
		res := env.Request(http.MethodGet, "/health", nil, "")
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /live returns 200", func(t *testing.T) {
		res := env.Request(http.MethodGet, "/live", nil, "")
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /ready returns 200", func(t *testing.T) {
		res := env.Request(http.MethodGet, "/ready", nil, "")
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /api/v1/health with auth returns 200", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/health", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /api/v1/health/providers returns 200", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/health/providers", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /api/v1/health/credentials returns 200", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/health/credentials", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("Unimplemented system endpoints return 404", func(t *testing.T) {
		for _, path := range []string{"/version", "/metrics"} {
			res := env.Request(http.MethodGet, path, nil, "")
			if res.Code != http.StatusNotFound {
				t.Fatalf("Expected 404 for %s, got %d", path, res.Code)
			}
		}
		res := env.AdminReq(http.MethodGet, "/api/v1/system/settings", nil)
		if res.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 for /api/v1/system/settings, got %d", res.Code)
		}
	})
}

func TestAuthAndSessionEndpoints(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("POST /api/auth/login with wrong password fails 401", func(t *testing.T) {
		res := env.Request(http.MethodPost, "/api/auth/login", map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		}, "")
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401, got %d", res.Code)
		}
	})

	t.Run("GET /api/auth/me returns current user", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/auth/me", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
		var payload struct {
			User string `json:"user"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &payload)
		if payload.User != "admin" {
			t.Fatalf("Expected admin user, got %s", payload.User)
		}
	})

	t.Run("POST /api/auth/logout revokes session", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/auth/logout", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
		resAfter := env.AdminReq(http.MethodGet, "/api/auth/me", nil)
		if resAfter.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 after logout, got %d", resAfter.Code)
		}
	})

	t.Run("Unimplemented auth endpoints return 404", func(t *testing.T) {
		for _, path := range []string{"/auth/register", "/auth/refresh", "/auth/change-password"} {
			res := env.Request(http.MethodPost, path, nil, "")
			if res.Code != http.StatusNotFound {
				t.Fatalf("Expected 404 for %s, got %d", path, res.Code)
			}
		}
	})
}

func TestPlatformEntitiesLifecycle(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("GET /api/v1/providers lists providers", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/providers", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
	})

	t.Run("GET /api/v1/providers/cloudflare metadata", func(t *testing.T) {
		res := env.AdminReq(http.MethodGet, "/api/v1/providers/cloudflare", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.Code)
		}
		resCaps := env.AdminReq(http.MethodGet, "/api/v1/providers/cloudflare/capabilities", nil)
		if resCaps.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resCaps.Code)
		}
		resDocs := env.AdminReq(http.MethodGet, "/api/v1/providers/cloudflare/docs", nil)
		if resDocs.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resDocs.Code)
		}
	})

	var projectID, envID string
	t.Run("Create & Read Project", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/v1/projects", map[string]string{
			"name":        "E2E QA Project",
			"description": "Integration testing project",
		})
		if res.Code != http.StatusOK && res.Code != http.StatusCreated {
			t.Fatalf("Create project failed: %d %s", res.Code, res.Body.String())
		}
		var r struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &r)
		projectID = r.Data.ID

		resGet := env.AdminReq(http.MethodGet, "/api/v1/projects/"+projectID, nil)
		if resGet.Code != http.StatusOK {
			t.Fatalf("Get project failed: %d", resGet.Code)
		}
	})

	t.Run("Create & Read Environment", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/v1/environments", map[string]string{
			"name":        "staging",
			"project_id":  projectID,
			"description": "Staging environment",
		})
		if res.Code != http.StatusOK && res.Code != http.StatusCreated {
			t.Fatalf("Create environment failed: %d %s", res.Code, res.Body.String())
		}
		var r struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &r)
		envID = r.Data.ID

		resList := env.AdminReq(http.MethodGet, "/api/v1/environments", nil)
		if resList.Code != http.StatusOK {
			t.Fatalf("List environments failed: %d", resList.Code)
		}
	})

	var credID string
	t.Run("Create, Mask, Rotate and Delete Credential", func(t *testing.T) {
		res := env.AdminReq(http.MethodPost, "/api/v1/credentials", map[string]any{
			"provider_id":  "cloudflare",
			"name":         "Production Cloudflare Token",
			"secret_value": "cf_sec_1234567890abcdef1234567890",
			"environment":  "production",
			"project_id":   projectID,
		})
		if res.Code != http.StatusOK && res.Code != http.StatusCreated {
			t.Fatalf("Create credential failed: %d %s", res.Code, res.Body.String())
		}
		var r struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &r)
		credID = r.Data.ID

		resGet := env.AdminReq(http.MethodGet, "/api/v1/credentials/"+credID, nil)
		if resGet.Code != http.StatusOK {
			t.Fatalf("Get credential failed: %d", resGet.Code)
		}
		if bytes.Contains(resGet.Body.Bytes(), []byte("cf_sec_1234567890abcdef1234567890")) {
			t.Fatalf("SECURITY VIOLATION: Raw credential leaked in GET /api/v1/credentials/:id")
		}

		resDis := env.AdminReq(http.MethodPost, "/api/v1/credentials/"+credID+"/disable", nil)
		if resDis.Code != http.StatusOK {
			t.Fatalf("Disable credential failed: %d", resDis.Code)
		}
		resEn := env.AdminReq(http.MethodPost, "/api/v1/credentials/"+credID+"/enable", nil)
		if resEn.Code != http.StatusOK {
			t.Fatalf("Enable credential failed: %d", resEn.Code)
		}

		resRot := env.AdminReq(http.MethodPost, "/api/v1/credentials/"+credID+"/rotate", map[string]string{
			"new_secret": "cf_sec_9999999999abcdef9999999999",
		})
		if resRot.Code != http.StatusOK {
			t.Fatalf("Rotate credential failed: %d %s", resRot.Code, resRot.Body.String())
		}

		resHealth := env.AdminReq(http.MethodGet, "/api/v1/credentials/"+credID+"/health", nil)
		if resHealth.Code != http.StatusOK {
			t.Fatalf("Credential health failed: %d", resHealth.Code)
		}
		resUsage := env.AdminReq(http.MethodGet, "/api/v1/credentials/"+credID+"/usage", nil)
		if resUsage.Code != http.StatusOK {
			t.Fatalf("Credential usage failed: %d", resUsage.Code)
		}

		resDel := env.AdminReq(http.MethodDelete, "/api/v1/credentials/"+credID, nil)
		if resDel.Code != http.StatusOK {
			t.Fatalf("Delete credential failed: %d", resDel.Code)
		}
	})

	_ = env.AdminReq(http.MethodDelete, "/api/v1/environments/"+envID, nil)
	_ = env.AdminReq(http.MethodDelete, "/api/v1/projects/"+projectID, nil)
}

func TestAdminManagementEndpoints(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	t.Run("API Keys Management", func(t *testing.T) {
		resList := env.AdminReq(http.MethodGet, "/api/keys", nil)
		if resList.Code != http.StatusOK {
			t.Fatalf("List keys failed: %d", resList.Code)
		}

		resCreate := env.AdminReq(http.MethodPost, "/api/keys", map[string]string{
			"name":   "Automated QA Key",
			"scopes": "*",
		})
		if resCreate.Code != http.StatusOK && resCreate.Code != http.StatusCreated {
			t.Fatalf("Create key failed: %d", resCreate.Code)
		}
		var created struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(resCreate.Body.Bytes(), &created)
		if created.ID != "" {
			_ = env.AdminReq(http.MethodDelete, "/api/keys/"+created.ID, nil)
		}
	})

	t.Run("Settings Management", func(t *testing.T) {
		resGet := env.AdminReq(http.MethodGet, "/api/settings", nil)
		if resGet.Code != http.StatusOK {
			t.Fatalf("Get settings failed: %d", resGet.Code)
		}

		resPut := env.AdminReq(http.MethodPut, "/api/settings", map[string]string{
			"key":   "qa_setting",
			"value": "qa_value_123",
		})
		if resPut.Code != http.StatusOK {
			t.Fatalf("Put settings failed: %d", resPut.Code)
		}
	})

	t.Run("Credential Pools CRUD", func(t *testing.T) {
		resCreate := env.AdminReq(http.MethodPost, "/api/credential-pools", map[string]string{
			"name":        "QA OpenAI Pool",
			"environment": "production",
		})
		if resCreate.Code != http.StatusOK && resCreate.Code != http.StatusCreated {
			t.Fatalf("Create pool failed: %d %s", resCreate.Code, resCreate.Body.String())
		}
		var pool struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(resCreate.Body.Bytes(), &pool)

		resGet := env.AdminReq(http.MethodGet, "/api/credential-pools/"+pool.ID, nil)
		if resGet.Code != http.StatusOK {
			t.Fatalf("Get pool failed: %d", resGet.Code)
		}

		resPause := env.AdminReq(http.MethodPost, "/api/credential-pools/"+pool.ID+"/pause", nil)
		if resPause.Code != http.StatusOK {
			t.Fatalf("Pause pool failed: %d", resPause.Code)
		}

		resResume := env.AdminReq(http.MethodPost, "/api/credential-pools/"+pool.ID+"/resume", nil)
		if resResume.Code != http.StatusOK {
			t.Fatalf("Resume pool failed: %d", resResume.Code)
		}

		resDel := env.AdminReq(http.MethodDelete, "/api/credential-pools/"+pool.ID, nil)
		if resDel.Code != http.StatusOK {
			t.Fatalf("Delete pool failed: %d", resDel.Code)
		}
	})

	t.Run("Free Tiers Catalog Endpoints", func(t *testing.T) {
		resList := env.AdminReq(http.MethodGet, "/api/free-tiers", nil)
		if resList.Code != http.StatusOK {
			t.Fatalf("List free tiers failed: %d", resList.Code)
		}

		resCats := env.AdminReq(http.MethodGet, "/api/free-tiers/categories", nil)
		if resCats.Code != http.StatusOK {
			t.Fatalf("List categories failed: %d", resCats.Code)
		}

		resVer := env.AdminReq(http.MethodGet, "/api/free-tiers/verified", nil)
		if resVer.Code != http.StatusOK {
			t.Fatalf("List verified free tiers failed: %d", resVer.Code)
		}
	})

	t.Run("Devtools Template and cURL Endpoints", func(t *testing.T) {
		resTpl := env.AdminReq(http.MethodGet, "/api/devtools/templates", nil)
		if resTpl.Code != http.StatusOK {
			t.Fatalf("List devtools templates failed: %d", resTpl.Code)
		}

		resCurl := env.AdminReq(http.MethodPost, "/api/devtools/generate-curl", map[string]any{
			"method":   "POST",
			"base_url": "https://api.openai.com/v1",
			"path":     "/chat/completions",
			"headers":  "{\"Authorization\": \"Bearer {{SECRET_TOKEN}}\"}",
			"body":     "{\"model\":\"gpt-4o\",\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}",
		})
		if resCurl.Code != http.StatusOK {
			t.Fatalf("Generate cURL failed: %d %s", resCurl.Code, resCurl.Body.String())
		}
	})
}

func TestGatewayRoutesAndCompletions(t *testing.T) {
	env := SetupTestEnv(t)
	defer env.Teardown()

	env.SeedDummyRoute(t, "mock-gpt-4o")

	t.Run("GET /v1/models returns model list", func(t *testing.T) {
		res := env.GatewayReq(http.MethodGet, "/v1/models", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("List models failed: %d", res.Code)
		}
		var payload struct {
			Data []map[string]any `json:"data"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &payload)
		found := false
		for _, m := range payload.Data {
			if m["id"] == "mock-gpt-4o" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected mock-gpt-4o in /v1/models response, got %s", res.Body.String())
		}
	})

	t.Run("POST /v1/chat/completions requires valid auth", func(t *testing.T) {
		res := env.Request(http.MethodPost, "/v1/chat/completions", map[string]any{
			"model": "mock-gpt-4o",
			"messages": []map[string]string{
				{"role": "user", "content": "hi"},
			},
		}, "")
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 without auth, got %d", res.Code)
		}
	})

	t.Run("POST /v1/chat/completions rejects empty model", func(t *testing.T) {
		res := env.GatewayReq(http.MethodPost, "/v1/chat/completions", map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "hi"},
			},
		})
		if res.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400 for empty model, got %d", res.Code)
		}
	})
}
