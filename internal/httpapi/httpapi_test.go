package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
)

type dummyAdapter struct{}

func (d *dummyAdapter) Kind() string { return "openai" }
func (d *dummyAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return nil, nil
}
func (d *dummyAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	return &providers.Response{
		ID:           "test-resp",
		Model:        req.Model,
		Role:         "assistant",
		Content:      "Gateway response success",
		FinishReason: "stop",
		Usage:        providers.Usage{PromptTokens: 5, CompletionTokens: 5, TotalTokens: 10},
	}, nil
}
func (d *dummyAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	ch := make(chan providers.StreamEvent, 3)
	ch <- providers.StreamEvent{Type: providers.StreamEventReasoning, Reasoning: "deep reasoning"}
	ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "chunk"}
	ch <- providers.StreamEvent{Type: providers.StreamEventDone}
	close(ch)
	return ch, nil
}

func setupTestServer(t *testing.T) (*Server, *db.DB, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "httpapi_test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(migrationsDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.DatabasePath = dbPath
	crypto, _ := auth.NewCryptoService(cfg.SecretKey)
	usageRec := usage.NewRecorder(database.DB, 100)
	ts := tokensaver.New("safe")
	checker := health.NewChecker(database.DB)

	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	router.SetAccount(&routing.Account{ID: "acc-1", ProviderID: "prov-1", State: "active", Enabled: true})
	router.SetRoute(&routing.Route{
		Name:     "gpt-4o",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "prov-1", ProviderKind: "openai", AccountID: "acc-1", ModelName: "gpt-4o", Priority: 10, Enabled: true},
		},
	})

	reg := providers.NewRegistry()
	reg.Register("openai", &dummyAdapter{})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "sk-test"}, nil
	}

	gw := gateway.NewGateway(router, reg, ts, cd, usageRec, credResolver)
	oauthMgr := oauth.NewManager()
	server := NewServer(cfg, database.DB, gw, crypto, usageRec, ts, checker, router, oauthMgr, nil, nil, nil, nil)

	rawKey, prefix, hash, _ := auth.GenerateApiKey()
	_, _ = database.Exec("INSERT INTO api_keys (id, name, prefix, hash, scopes, enabled) VALUES ('key1', 'Test Key', ?, ?, '*', 1)", prefix, hash)

	return server, database, rawKey
}

func TestHealthEndpoints(t *testing.T) {
	server, database, _ := setupTestServer(t)
	defer database.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got %d", rec.Code)
	}

	reqReady := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recReady := httptest.NewRecorder()
	server.ServeHTTP(recReady, reqReady)

	if recReady.Code != http.StatusOK {
		t.Errorf("expected 200 for /ready, got %d", recReady.Code)
	}
}

func TestGatewayAuthAndChatCompletions(t *testing.T) {
	server, database, validKey := setupTestServer(t)
	defer database.Close()

	reqBody := `{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}`

	reqUnauth := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	reqUnauth.Header.Set("Content-Type", "application/json")
	recUnauth := httptest.NewRecorder()
	server.ServeHTTP(recUnauth, reqUnauth)

	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 unauthorized without key, got %d", recUnauth.Code)
	}

	reqAuth := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	reqAuth.Header.Set("Content-Type", "application/json")
	reqAuth.Header.Set("Authorization", "Bearer "+validKey)
	recAuth := httptest.NewRecorder()
	server.ServeHTTP(recAuth, reqAuth)

	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid key, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(recAuth.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "Gateway response success" {
		t.Errorf("unexpected completion content: %v", resp)
	}

	streamReqBody := `{"model":"gpt-4o","stream":true,"messages":[{"role":"user","content":"Hello stream"}]}`
	reqStream := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(streamReqBody))
	reqStream.Header.Set("Content-Type", "application/json")
	reqStream.Header.Set("Authorization", "Bearer "+validKey)
	recStream := httptest.NewRecorder()
	server.ServeHTTP(recStream, reqStream)

	if recStream.Code != http.StatusOK {
		t.Fatalf("expected 200 for stream, got %d: %s", recStream.Code, recStream.Body.String())
	}
	streamOutput := recStream.Body.String()
	if !strings.Contains(streamOutput, "reasoning_content") || !strings.Contains(streamOutput, "deep reasoning") {
		t.Errorf("expected stream to contain reasoning_content, got: %s", streamOutput)
	}
	if !strings.Contains(streamOutput, "data: [DONE]") {
		t.Errorf("expected stream to finish with [DONE], got: %s", streamOutput)
	}
}

func TestAdminLoginAndSessionFlow(t *testing.T) {
	server, database, _ := setupTestServer(t)
	defer database.Close()

	loginBody := `{"username":"admin","password":"admin12345"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	server.ServeHTTP(recLogin, reqLogin)

	if recLogin.Code != http.StatusOK {
		t.Fatalf("login failed, expected 200 got %d", recLogin.Code)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(recLogin.Body).Decode(&loginResp)
	if loginResp.Token == "" {
		t.Fatal("expected session token in login response")
	}

	reqMe := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResp.Token)
	recMe := httptest.NewRecorder()
	server.ServeHTTP(recMe, reqMe)

	if recMe.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/auth/me with session token, got %d", recMe.Code)
	}

	createKeyBody := `{"name":"Generated Key","scopes":"*"}`
	reqKey := httptest.NewRequest(http.MethodPost, "/api/keys", strings.NewReader(createKeyBody))
	reqKey.Header.Set("Authorization", "Bearer "+loginResp.Token)
	recKey := httptest.NewRecorder()
	server.ServeHTTP(recKey, reqKey)

	if recKey.Code != http.StatusCreated {
		t.Fatalf("expected 201 for /api/keys, got %d", recKey.Code)
	}

	var keyResp struct {
		ApiKey string `json:"api_key"`
	}
	_ = json.NewDecoder(recKey.Body).Decode(&keyResp)
	if !strings.HasPrefix(keyResp.ApiKey, "eka_live_") {
		t.Errorf("expected eka_live_ key, got %s", keyResp.ApiKey)
	}
}

func TestBodyLimitEnforcement(t *testing.T) {
	server, database, validKey := setupTestServer(t)
	defer database.Close()

	giantBody := bytes.Repeat([]byte("a"), 11*1024*1024)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(giantBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+validKey)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected body size limit error, got %d", rec.Code)
	}
}

func TestResponsesEndpointWithInput(t *testing.T) {
	server, database, validKey := setupTestServer(t)
	defer database.Close()

	body := `{"model":"gpt-4o","input":"Hello from responses API"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+validKey)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /v1/responses, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTokenSaverOptOutHeader(t *testing.T) {
	server, database, validKey := setupTestServer(t)
	defer database.Close()

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hello world"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+validKey)
	req.Header.Set("X-Token-Saver", "off")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with X-Token-Saver off, got %d", rec.Code)
	}
}

func TestAdminEntityManagement(t *testing.T) {
	server, database, _ := setupTestServer(t)
	defer database.Close()

	loginBody := `{"username":"admin","password":"admin12345"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	server.ServeHTTP(recLogin, reqLogin)

	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(recLogin.Body).Decode(&loginResp)
	token := loginResp.Token

	provBody := `{"id":"p_test","key":"key_test","name":"Test Provider","kind":"openai","base_url":"https://api.test.com","enabled":true}`
	reqProv := httptest.NewRequest(http.MethodPost, "/api/providers", strings.NewReader(provBody))
	reqProv.Header.Set("Authorization", "Bearer "+token)
	recProv := httptest.NewRecorder()
	server.ServeHTTP(recProv, reqProv)
	if recProv.Code != http.StatusCreated {
		t.Fatalf("create provider failed, got %d", recProv.Code)
	}

	modelBody := `{"id":"m_test","provider_id":"p_test","external_name":"test-model","display_name":"Test Model","context_limit":8192,"streaming":true}`
	reqModel := httptest.NewRequest(http.MethodPost, "/api/models", strings.NewReader(modelBody))
	reqModel.Header.Set("Authorization", "Bearer "+token)
	recModel := httptest.NewRecorder()
	server.ServeHTTP(recModel, reqModel)
	if recModel.Code != http.StatusCreated {
		t.Fatalf("create model failed, got %d", recModel.Code)
	}

	accBody := `{"id":"a_test","provider_id":"p_test","name":"Test Acc","auth_type":"api_key","priority":1,"api_key":"sk-secret"}`
	reqAcc := httptest.NewRequest(http.MethodPost, "/api/accounts", strings.NewReader(accBody))
	reqAcc.Header.Set("Authorization", "Bearer "+token)
	recAcc := httptest.NewRecorder()
	server.ServeHTTP(recAcc, reqAcc)
	if recAcc.Code != http.StatusCreated {
		t.Fatalf("create account failed, got %d", recAcc.Code)
	}

	proxyBody := `{"id":"px_test","name":"Test Proxy","scheme":"http","host":"103.253.213.185","port":8080,"username":"user","password":"pass"}`
	reqPx := httptest.NewRequest(http.MethodPost, "/api/proxy-profiles", strings.NewReader(proxyBody))
	reqPx.Header.Set("Authorization", "Bearer "+token)
	recPx := httptest.NewRecorder()
	server.ServeHTTP(recPx, reqPx)
	if recPx.Code != http.StatusCreated {
		t.Fatalf("create proxy profile failed, got %d", recPx.Code)
	}

	settingBody := `{"key":"theme","value":"dark"}`
	reqSet := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(settingBody))
	reqSet.Header.Set("Authorization", "Bearer "+token)
	recSet := httptest.NewRecorder()
	server.ServeHTTP(recSet, reqSet)
	if recSet.Code != http.StatusOK {
		t.Fatalf("update setting failed, got %d", recSet.Code)
	}
}

func TestOAuthStartAndCallback(t *testing.T) {
	server, database, _ := setupTestServer(t)
	defer database.Close()

	loginBody := `{"username":"admin","password":"admin12345"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	server.ServeHTTP(recLogin, reqLogin)

	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(recLogin.Body).Decode(&loginResp)
	token := loginResp.Token

	provBody := `{"id":"p_oauth","key":"key_oauth","name":"OAuth Provider","kind":"openai","base_url":"https://api.test.com","enabled":true}`
	reqProv := httptest.NewRequest(http.MethodPost, "/api/providers", strings.NewReader(provBody))
	reqProv.Header.Set("Authorization", "Bearer "+token)
	recProv := httptest.NewRecorder()
	server.ServeHTTP(recProv, reqProv)

	startBody := `{"provider_id":"p_oauth"}`
	reqStart := httptest.NewRequest(http.MethodPost, "/api/accounts/oauth/start", strings.NewReader(startBody))
	reqStart.Header.Set("Authorization", "Bearer "+token)
	recStart := httptest.NewRecorder()
	server.ServeHTTP(recStart, reqStart)

	if recStart.Code != http.StatusOK {
		t.Fatalf("expected 200 for oauth start, got %d", recStart.Code)
	}

	var startResp struct {
		State string `json:"state"`
	}
	_ = json.NewDecoder(recStart.Body).Decode(&startResp)
	if startResp.State == "" {
		t.Fatal("expected state in oauth start response")
	}

	cbBody := `{"state":"` + startResp.State + `","code":"auth_code_123","account_name":"My OAuth Acc"}`
	reqCb := httptest.NewRequest(http.MethodPost, "/api/accounts/oauth/callback", strings.NewReader(cbBody))
	reqCb.Header.Set("Authorization", "Bearer "+token)
	recCb := httptest.NewRecorder()
	server.ServeHTTP(recCb, reqCb)

	if recCb.Code != http.StatusCreated {
		t.Fatalf("expected 201 for oauth callback, got %d: %s", recCb.Code, recCb.Body.String())
	}
}

func TestAllEndpointsComprehensive(t *testing.T) {
	server, database, validKey := setupTestServer(t)
	defer database.Close()

	recHealth := httptest.NewRecorder()
	server.ServeHTTP(recHealth, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recHealth.Code != http.StatusOK {
		t.Fatalf("GET /health failed: %d", recHealth.Code)
	}

	recReady := httptest.NewRecorder()
	server.ServeHTTP(recReady, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recReady.Code != http.StatusOK {
		t.Fatalf("GET /ready failed: %d", recReady.Code)
	}

	reqModels := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	reqModels.Header.Set("Authorization", "Bearer "+validKey)
	recModels := httptest.NewRecorder()
	server.ServeHTTP(recModels, reqModels)
	if recModels.Code != http.StatusOK {
		t.Fatalf("GET /v1/models failed: %d", recModels.Code)
	}

	chatBody := `{"model":"gpt-4o","messages":[{"role":"user","content":"ping"}]}`
	reqChat := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(chatBody))
	reqChat.Header.Set("Content-Type", "application/json")
	reqChat.Header.Set("Authorization", "Bearer "+validKey)
	recChat := httptest.NewRecorder()
	server.ServeHTTP(recChat, reqChat)
	if recChat.Code != http.StatusOK {
		t.Fatalf("POST /v1/chat/completions failed: %d", recChat.Code)
	}

	respBody := `{"model":"gpt-4o","input":"ping"}`
	reqResp := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(respBody))
	reqResp.Header.Set("Content-Type", "application/json")
	reqResp.Header.Set("Authorization", "Bearer "+validKey)
	recResp := httptest.NewRecorder()
	server.ServeHTTP(recResp, reqResp)
	if recResp.Code != http.StatusOK {
		t.Fatalf("POST /v1/responses failed: %d", recResp.Code)
	}

	loginBody := `{"username":"admin","password":"admin12345"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	server.ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK {
		t.Fatalf("POST /api/auth/login failed: %d", recLogin.Code)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(recLogin.Body).Decode(&loginData)
	tok := loginData.Token

	adminGet := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)
		return rec
	}

	adminPost := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)
		return rec
	}

	if r := adminGet("/api/providers"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/providers failed: %d", r.Code)
	}
	if r := adminPost("/api/providers", `{"id":"p2","key":"k2","name":"P2","kind":"openai","base_url":"https://api.test","enabled":true}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/providers failed: %d", r.Code)
	}

	if r := adminGet("/api/accounts"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/accounts failed: %d", r.Code)
	}
	if r := adminPost("/api/accounts", `{"id":"a2","provider_id":"p2","name":"A2","auth_type":"apiKey","priority":1,"api_key":"sk-test"}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/accounts failed: %d", r.Code)
	}

	if r := adminGet("/api/credentials"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/credentials failed: %d", r.Code)
	}
	if r := adminGet("/api/credentials/cred_a2"); r.Code != http.StatusOK && r.Code != http.StatusNotFound {
		t.Fatalf("GET /api/credentials/id failed: %d", r.Code)
	}

	if r := adminGet("/api/models"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/models failed: %d", r.Code)
	}
	if r := adminPost("/api/models", `{"id":"m2","provider_id":"p2","external_name":"m2-ext","display_name":"M2"}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/models failed: %d", r.Code)
	}

	if r := adminGet("/api/routes"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/routes failed: %d", r.Code)
	}
	if r := adminPost("/api/routes", `{"id":"r2","name":"combo2","strategy":"priority","items":[{"provider_id":"p2","priority":1}]}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/routes failed: %d", r.Code)
	}

	mockRelay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer mockRelay.Close()

	uRelay, _ := url.Parse(mockRelay.URL)
	portRelay, _ := strconv.Atoi(uRelay.Port())

	if r := adminGet("/api/proxies"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/proxies failed: %d", r.Code)
	}
	proxyPayload := fmt.Sprintf(`{"id":"px2","name":"PX2","scheme":"relay","host":"%s","port":%d}`, uRelay.Hostname(), portRelay)
	if r := adminPost("/api/proxies", proxyPayload); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/proxies failed: %d", r.Code)
	}
	if r := adminPost("/api/proxies/px2/test", ""); r.Code != http.StatusOK {
		t.Fatalf("POST /api/proxies/px2/test failed: %d", r.Code)
	}

	if r := adminGet("/api/keys"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/keys failed: %d", r.Code)
	}
	if r := adminPost("/api/keys", `{"name":"test key"}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/keys failed: %d", r.Code)
	}

	if r := adminGet("/api/usage"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/usage failed: %d", r.Code)
	}

	backupFile := filepath.Join(t.TempDir(), "backup_test.db")
	if r := adminPost("/api/backup", `{"dest_path":"`+strings.ReplaceAll(backupFile, `\`, `/`)+`"}`); r.Code != http.StatusCreated {
		t.Fatalf("POST /api/backup failed: %d", r.Code)
	}
	if r := adminGet("/api/backup"); r.Code != http.StatusOK {
		t.Fatalf("GET /api/backup failed: %d", r.Code)
	}
}
