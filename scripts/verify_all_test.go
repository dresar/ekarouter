package scripts

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/httpapi"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/proxy"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	_ "modernc.org/sqlite"
)

type dummyGatewayAdapter struct{}

func (d *dummyGatewayAdapter) Kind() string { return "openai" }
func (d *dummyGatewayAdapter) Models(ctx context.Context, creds *providers.Credentials) ([]providers.ModelInfo, error) {
	return nil, nil
}
func (d *dummyGatewayAdapter) Execute(ctx context.Context, req *providers.Request, creds *providers.Credentials) (*providers.Response, error) {
	return &providers.Response{
		ID:           "verify-resp-1",
		Model:        req.Model,
		Role:         "assistant",
		Content:      "EkaRouter all-endpoint verification passed",
		FinishReason: "stop",
		Usage:        providers.Usage{PromptTokens: 12, CompletionTokens: 8, TotalTokens: 20},
	}, nil
}
func (d *dummyGatewayAdapter) ExecuteStream(ctx context.Context, req *providers.Request, creds *providers.Credentials) (<-chan providers.StreamEvent, error) {
	ch := make(chan providers.StreamEvent, 3)
	ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "chunk1 "}
	ch <- providers.StreamEvent{Type: providers.StreamEventDelta, Delta: "chunk2"}
	ch <- providers.StreamEvent{Type: providers.StreamEventDone}
	close(ch)
	return ch, nil
}

func TestVerifyDatabaseImportState(t *testing.T) {
	dbPath := filepath.Join("..", "data", "ekarouter.db")
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	expectedCounts := map[string]int{
		"accounts":       182,
		"credentials":    182,
		"proxy_profiles": 16,
		"routes":         2,
	}

	for tbl, minCount := range expectedCounts {
		var cnt int
		err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl)).Scan(&cnt)
		if err != nil {
			t.Fatalf("table %s query error: %v", tbl, err)
		}
		if cnt < minCount {
			t.Fatalf("table %s has %d rows, expected at least %d", tbl, cnt, minCount)
		}
		t.Logf("Table %-16s: %4d rows (PASS)", tbl, cnt)
	}

	var provCount, modelCount, routeItemCount int
	_ = database.QueryRow("SELECT COUNT(*) FROM providers").Scan(&provCount)
	_ = database.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount)
	_ = database.QueryRow("SELECT COUNT(*) FROM route_items").Scan(&routeItemCount)
	t.Logf("Table providers       : %4d rows", provCount)
	t.Logf("Table models          : %4d rows", modelCount)
	t.Logf("Table route_items     : %4d rows", routeItemCount)
}

func TestVerifyAllProxyProfilesLive(t *testing.T) {
	dbPath := filepath.Join("..", "data", "ekarouter.db")
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, name, scheme, host, port, username, encrypted_password, enabled FROM proxy_profiles")
	if err != nil {
		t.Fatalf("failed to query proxy_profiles: %v", err)
	}
	defer rows.Close()

	crypto, _ := auth.NewCryptoService("test-secret-key-32-bytes-long-now!!")

	type ProxyResult struct {
		ID        string
		Name      string
		Host      string
		Status    int
		LatencyMs int64
		Healthy   bool
		Error     string
	}

	var results []ProxyResult
	healthyCount := 0

	for rows.Next() {
		var id, name, scheme, host, username, encPass sql.NullString
		var port, enabled int
		if err := rows.Scan(&id, &name, &scheme, &host, &port, &username, &encPass, &enabled); err != nil {
			continue
		}

		pass, _ := crypto.Decrypt(encPass.String)
		prof := &proxy.Profile{
			ID:       id.String,
			Name:     name.String,
			Scheme:   scheme.String,
			Host:     host.String,
			Port:     port,
			Username: username.String,
			Password: pass,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		ok, status, latency, testErr := proxy.TestProfile(ctx, prof, 7*time.Second)
		cancel()

		errStr := ""
		if testErr != nil {
			errStr = testErr.Error()
		}

		if ok {
			healthyCount++
		}

		results = append(results, ProxyResult{
			ID:        id.String,
			Name:      name.String,
			Host:      host.String,
			Status:    status,
			LatencyMs: latency,
			Healthy:   ok,
			Error:     errStr,
		})
	}

	t.Logf("Total Proxy Profiles Tested: %d | Healthy: %d | Down/Ephemeral: %d", len(results), healthyCount, len(results)-healthyCount)
	for _, r := range results {
		mark := "FAIL"
		if r.Healthy {
			mark = "OK  "
		}
		t.Logf("[%s] %-36s | Status: %3d | %4dms | %-32s | %s", mark, r.ID, r.Status, r.LatencyMs, r.Name, r.Host)
	}

	if healthyCount < 8 {
		t.Fatalf("expected at least 8 healthy proxies, got %d", healthyCount)
	}
}

func TestVerifyAllEndpointsE2E(t *testing.T) {
	tempDir := t.TempDir()
	testDbPath := filepath.Join(tempDir, "test_e2e.db")

	srcDb := filepath.Join("..", "data", "ekarouter.db")
	srcData, err := os.ReadFile(srcDb)
	if err == nil {
		_ = os.WriteFile(testDbPath, srcData, 0644)
	}

	database, err := db.Open(testDbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	migDir := filepath.Join("..", "migrations")
	if err := database.Migrate(migDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.DatabasePath = testDbPath
	crypto, _ := auth.NewCryptoService(cfg.SecretKey)
	usageRec := usage.NewRecorder(database.DB, 100)
	ts := tokensaver.New("safe")
	checker := health.NewChecker(database.DB)

	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)
	_ = router.LoadFromDB(context.Background(), database.DB)

	router.SetAccount(&routing.Account{ID: "acc-mock", ProviderID: "prov-mock", State: "active", Enabled: true})
	router.SetRoute(&routing.Route{
		Name:     "mock-test-model",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "prov-mock", ProviderKind: "openai", AccountID: "acc-mock", ModelName: "mock-test-model", Priority: 1, Enabled: true},
		},
	})

	reg := providers.NewRegistry()
	reg.Register("openai", &dummyGatewayAdapter{})
	reg.Register("custom", &dummyGatewayAdapter{})

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		return &providers.Credentials{APIKey: "sk-test-creds"}, nil
	}

	gw := gateway.NewGateway(router, reg, ts, cd, usageRec, credResolver)
	oauthMgr := oauth.NewManager()
	server := httpapi.NewServer(cfg, database.DB, gw, crypto, usageRec, ts, checker, router, oauthMgr, nil, nil, nil, nil)

	tsHttp := httptest.NewServer(server)
	defer tsHttp.Close()

	client := tsHttp.Client()

	healthResp, err := client.Get(tsHttp.URL + "/health")
	if err != nil || healthResp.StatusCode != http.StatusOK {
		t.Fatalf("/health failed: status=%v err=%v", healthResp.StatusCode, err)
	}
	t.Logf("GET /health -> %d OK", healthResp.StatusCode)

	readyResp, err := client.Get(tsHttp.URL + "/ready")
	if err != nil || readyResp.StatusCode != http.StatusOK {
		t.Fatalf("/ready failed: status=%v err=%v", readyResp.StatusCode, err)
	}
	t.Logf("GET /ready -> %d OK", readyResp.StatusCode)

	loginBody := `{"username":"admin","password":"admin12345"}`
	loginResp, err := client.Post(tsHttp.URL+"/api/auth/login", "application/json", strings.NewReader(loginBody))
	if err != nil || loginResp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/auth/login failed: %v", err)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&loginData)
	loginResp.Body.Close()
	token := loginData.Token
	if token == "" {
		t.Fatal("empty token from login")
	}
	t.Logf("POST /api/auth/login -> 200 OK (Token: %s...)", token[:10])

	authReq := func(method, path, body string) (*http.Response, error) {
		var rdr *strings.Reader
		if body != "" {
			rdr = strings.NewReader(body)
		} else {
			rdr = strings.NewReader("")
		}
		req, err := http.NewRequest(method, tsHttp.URL+path, rdr)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		return client.Do(req)
	}

	resp, err := authReq(http.MethodGet, "/api/auth/me", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/auth/me failed: status=%v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/auth/me -> 200 OK")

	keyBody := `{"name":"E2E Test Key","scopes":"*"}`
	resp, err = authReq(http.MethodPost, "/api/keys", keyBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/keys failed: %v", resp.StatusCode)
	}
	var keyData struct {
		ApiKey string `json:"api_key"`
		ID     string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&keyData)
	resp.Body.Close()
	apiKey := keyData.ApiKey
	t.Logf("POST /api/keys -> 201 Created (API Key: %s...)", keyData.ApiKey[:10])

	resp, err = authReq(http.MethodGet, "/api/keys", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/keys failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/keys -> 200 OK")

	resp, err = authReq(http.MethodGet, "/api/providers", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/providers failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/providers -> 200 OK")

	provBody := `{"id":"test-prov","key":"test-prov","name":"Test Prov","kind":"openai","base_url":"https://api.openai.com/v1"}`
	resp, err = authReq(http.MethodPost, "/api/providers", provBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/providers failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/providers -> 201 Created")

	resp, err = authReq(http.MethodGet, "/api/accounts", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/accounts failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/accounts -> 200 OK")

	accBody := `{"id":"test-acc","provider_id":"test-prov","name":"Test Acc","auth_type":"apiKey","priority":1,"api_key":"sk-test"}`
	resp, err = authReq(http.MethodPost, "/api/accounts", accBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/accounts failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/accounts -> 201 Created")

	resp, err = authReq(http.MethodGet, "/api/credentials", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/credentials failed: %v", resp.StatusCode)
	}
	var credsList []httpapi.CredentialResponse
	_ = json.NewDecoder(resp.Body).Decode(&credsList)
	resp.Body.Close()
	t.Logf("GET /api/credentials -> 200 OK (%d credentials returned)", len(credsList))

	if len(credsList) > 0 {
		resp, err = authReq(http.MethodGet, "/api/credentials/"+credsList[0].ID, "")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/credentials/%s failed: %v", credsList[0].ID, resp.StatusCode)
		}
		resp.Body.Close()
		t.Logf("GET /api/credentials/%s -> 200 OK", credsList[0].ID)
	}

	resp, err = authReq(http.MethodGet, "/api/models", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/models failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/models -> 200 OK")

	modelBody := `{"id":"test-prov/test-model","provider_id":"test-prov","external_name":"test-model","display_name":"Test Model"}`
	resp, err = authReq(http.MethodPost, "/api/models", modelBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/models failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/models -> 201 Created")

	resp, err = authReq(http.MethodGet, "/api/routes", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/routes failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/routes -> 200 OK")

	routeBody := `{"id":"test-route","name":"test-route","strategy":"priority","items":[{"provider_id":"test-prov","account_id":"test-acc","model_id":"test-prov/test-model","priority":1}]}`
	resp, err = authReq(http.MethodPost, "/api/routes", routeBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/routes failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/routes -> 201 Created")

	resp, err = authReq(http.MethodGet, "/api/proxies", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/proxies failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/proxies -> 200 OK")

	mockRelay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"relay ok"}`))
	}))
	defer mockRelay.Close()

	uRelay, _ := url.Parse(mockRelay.URL)
	portRelay, _ := strconv.Atoi(uRelay.Port())
	pxBody := fmt.Sprintf(`{"id":"test-relay","name":"Test Relay","scheme":"relay","host":"%s","port":%d}`, uRelay.Hostname(), portRelay)
	resp, err = authReq(http.MethodPost, "/api/proxies", pxBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/proxies failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/proxies -> 201 Created")

	resp, err = authReq(http.MethodPost, "/api/proxies/test-relay/test", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/proxies/test-relay/test failed: %v", resp.StatusCode)
	}
	var pxTestRes struct {
		OK        bool  `json:"ok"`
		Status    int   `json:"status"`
		LatencyMs int64 `json:"latency_ms"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&pxTestRes)
	resp.Body.Close()
	t.Logf("POST /api/proxies/test-relay/test -> 200 OK (ok=%v, status=%d, latency=%dms)", pxTestRes.OK, pxTestRes.Status, pxTestRes.LatencyMs)

	resp, err = authReq(http.MethodGet, "/api/usage", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/usage failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/usage -> 200 OK")

	tsBody := `{"input":"line 1\nline 1\nline 1\nline 2"}`
	resp, err = authReq(http.MethodPost, "/api/tokensaver/preview", tsBody)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/tokensaver/preview failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/tokensaver/preview -> 200 OK")

	backupDest := filepath.Join(tempDir, "manual_backup.db")
	backupBody := fmt.Sprintf(`{"dest_path":"%s"}`, strings.ReplaceAll(backupDest, `\`, `/`))
	resp, err = authReq(http.MethodPost, "/api/backup", backupBody)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/backup failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /api/backup -> 201 Created (Saved to %s)", backupDest)

	resp, err = authReq(http.MethodGet, "/api/backup", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/backup failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("GET /api/backup -> 200 OK")

	gwReq := func(method, path, body string) (*http.Response, error) {
		var rdr *strings.Reader
		if body != "" {
			rdr = strings.NewReader(body)
		} else {
			rdr = strings.NewReader("")
		}
		req, err := http.NewRequest(method, tsHttp.URL+path, rdr)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		return client.Do(req)
	}

	resp, err = gwReq(http.MethodGet, "/v1/models", "")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/models failed: %v", resp.StatusCode)
	}
	var modelsList struct {
		Object string `json:"object"`
		Data   []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&modelsList)
	resp.Body.Close()
	t.Logf("GET /v1/models -> 200 OK (%d models listed)", len(modelsList.Data))

	chatPayload := `{"model":"test-route","messages":[{"role":"user","content":"Hello EkaRouter"}]}`
	resp, err = gwReq(http.MethodPost, "/v1/chat/completions", chatPayload)
	if err != nil || resp.StatusCode != http.StatusOK {
		var errB string
		if resp != nil {
			b, _ := io.ReadAll(resp.Body)
			errB = string(b)
			resp.Body.Close()
		}
		t.Fatalf("POST /v1/chat/completions normal failed: status=%v err=%v body=%s", resp.StatusCode, err, errB)
	}
	var chatRes struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&chatRes)
	resp.Body.Close()
	if len(chatRes.Choices) == 0 || chatRes.Choices[0].Message.Content == "" {
		t.Fatal("empty completion response")
	}
	t.Logf("POST /v1/chat/completions -> 200 OK (Content: %s)", chatRes.Choices[0].Message.Content)

	streamPayload := `{"model":"test-route","stream":true,"messages":[{"role":"user","content":"Stream test"}]}`
	resp, err = gwReq(http.MethodPost, "/v1/chat/completions", streamPayload)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /v1/chat/completions stream failed: %v", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	gotDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[DONE]") {
			gotDone = true
			break
		}
	}
	resp.Body.Close()
	if !gotDone {
		t.Fatal("stream did not complete with [DONE]")
	}
	t.Logf("POST /v1/chat/completions (stream=true) -> 200 OK (Stream received [DONE])")

	respPayload := `{"model":"test-route","input":"Normalized response test"}`
	resp, err = gwReq(http.MethodPost, "/v1/responses", respPayload)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /v1/responses failed: %v", resp.StatusCode)
	}
	resp.Body.Close()
	t.Logf("POST /v1/responses -> 200 OK")
}
