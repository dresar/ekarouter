package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/tests/mocks"
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

	t.Run("GET /version returns 200 and version 1.0.0", func(t *testing.T) {
		res := env.Request(http.MethodGet, "/version", nil, "")
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200 for /version, got %d", res.Code)
		}
		var payload struct {
			Version string `json:"version"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &payload)
		if payload.Version != "1.0.0" {
			t.Fatalf("Expected version 1.0.0, got %s", payload.Version)
		}
	})

	t.Run("GET & PATCH /api/v1/system/settings", func(t *testing.T) {
		resGet := env.AdminReq(http.MethodGet, "/api/v1/system/settings", nil)
		if resGet.Code != http.StatusOK {
			t.Fatalf("Expected 200 for GET /api/v1/system/settings, got %d", resGet.Code)
		}

		resPatch := env.AdminReq(http.MethodPatch, "/api/v1/system/settings", map[string]string{
			"key":   "system_mode",
			"value": "production_ready",
		})
		if resPatch.Code != http.StatusOK {
			t.Fatalf("Expected 200 for PATCH /api/v1/system/settings, got %d", resPatch.Code)
		}
	})

	t.Run("Unimplemented system endpoints return 404", func(t *testing.T) {
		for _, path := range []string{"/metrics"} {
			res := env.Request(http.MethodGet, path, nil, "")
			if res.Code != http.StatusNotFound {
				t.Fatalf("Expected 404 for %s, got %d", path, res.Code)
			}
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

	t.Run("POST /auth/login and GET /auth/me root alias", func(t *testing.T) {
		resLogin := env.Request(http.MethodPost, "/auth/login", map[string]string{
			"username": "admin",
			"password": "admin12345",
		}, "")
		if resLogin.Code != http.StatusOK {
			t.Fatalf("Expected 200 for /auth/login alias, got %d", resLogin.Code)
		}
		var sessionTok string
		for _, c := range resLogin.Result().Cookies() {
			if c.Name == "session_token" {
				sessionTok = c.Value
				break
			}
		}
		if sessionTok != "" {
			resMe := env.Request(http.MethodGet, "/auth/me", nil, "Bearer "+sessionTok)
			if resMe.Code != http.StatusOK {
				t.Fatalf("Expected 200 for /auth/me alias, got %d", resMe.Code)
			}
		}
	})

	t.Run("GET /auth/sessions and DELETE /auth/sessions/:id", func(t *testing.T) {
		dummyToken := "test_revokable_session_token_123"
		_, err := env.DB.Exec("INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES ('sess_revoke_target', 'admin', ?, datetime('now', '+1 hour'))", auth.HashToken(dummyToken))
		if err != nil {
			t.Fatalf("Insert dummy session: %v", err)
		}

		resSessions := env.AdminReq(http.MethodGet, "/auth/sessions", nil)
		if resSessions.Code != http.StatusOK {
			t.Fatalf("Expected 200 for /auth/sessions, got %d", resSessions.Code)
		}
		var sessions []struct {
			ID     string `json:"id"`
			UserID string `json:"user_id"`
		}
		_ = json.Unmarshal(resSessions.Body.Bytes(), &sessions)
		if len(sessions) == 0 {
			t.Fatalf("Expected at least 1 session in /auth/sessions")
		}

		resRevoke := env.AdminReq(http.MethodDelete, "/auth/sessions/sess_revoke_target", nil)
		if resRevoke.Code != http.StatusOK {
			t.Fatalf("Expected 200 for DELETE /auth/sessions/:id, got %d", resRevoke.Code)
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

	t.Run("POST /v1/chat/completions non-streaming execution with mock upstream", func(t *testing.T) {
		mockUpstream := mocks.NewMockUpstreamServer(mocks.UpstreamBehavior{
			StatusCode:   200,
			ResponseBody: `{"id":"chatcmpl-live-qa","object":"chat.completion","created":1700000000,"model":"mock-live-gpt4","choices":[{"index":0,"message":{"role":"assistant","content":"Hello from verified mock!"},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":6,"total_tokens":18}}`,
		})
		defer mockUpstream.Close()

		env.SeedLiveMockRoute(t, "mock-live-gpt4", mockUpstream.URL+"/v1")

		res := env.GatewayReq(http.MethodPost, "/v1/chat/completions", map[string]any{
			"model": "mock-live-gpt4",
			"messages": []map[string]string{
				{"role": "user", "content": "hello verified test"},
			},
		})
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200 for chat completions, got %d: %s", res.Code, res.Body.String())
		}
		var resp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &resp)
		if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "Hello from verified mock!" {
			t.Fatalf("Unexpected completion response: %s", res.Body.String())
		}
	})

	t.Run("POST /v1/chat/completions SSE streaming execution with mock upstream", func(t *testing.T) {
		mockStream := mocks.NewMockUpstreamServer(mocks.UpstreamBehavior{
			StreamChunks: []string{
				`{"id":"chatcmpl-chunk-1","object":"chat.completion.chunk","choices":[{"delta":{"content":"Live stream chunk 1"}}]}`,
				`{"id":"chatcmpl-chunk-2","object":"chat.completion.chunk","choices":[{"delta":{"content":" and chunk 2"}}]}`,
			},
		})
		defer mockStream.Close()

		env.SeedLiveMockRoute(t, "mock-live-stream", mockStream.URL+"/v1")

		res := env.GatewayReq(http.MethodPost, "/v1/chat/completions", map[string]any{
			"model":  "mock-live-stream",
			"stream": true,
			"messages": []map[string]string{
				{"role": "user", "content": "stream test"},
			},
		})
		if res.Code != http.StatusOK {
			t.Fatalf("Expected 200 for stream, got %d: %s", res.Code, res.Body.String())
		}
		bodyStr := res.Body.String()
		if !strings.Contains(bodyStr, "Live stream chunk 1") || !strings.Contains(bodyStr, "[DONE]") {
			t.Fatalf("Streaming response missing expected chunks: %s", bodyStr)
		}
	})
}
