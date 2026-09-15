package oauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildAuthURL(t *testing.T) {
	m := NewManager()
	state, challenge, err := m.GenerateState("antigravity", 10*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	authURL, err := m.BuildAuthURL("antigravity", "http://localhost:8080/callback", state, challenge)
	if err != nil {
		t.Fatalf("build auth url error: %v", err)
	}

	if !strings.HasPrefix(authURL, "https://accounts.google.com/o/oauth2/v2/auth") {
		t.Errorf("expected google oauth url, got %s", authURL)
	}
	if !strings.Contains(authURL, "code_challenge="+challenge) {
		t.Errorf("expected code_challenge in url")
	}
	if !strings.Contains(authURL, "state="+state) {
		t.Errorf("expected state in url")
	}
}

func TestExchangeTokenMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock_access","refresh_token":"mock_refresh","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer ts.Close()

	providerConfigs["mock_provider"] = ProviderConfig{
		ID:           "mock_provider",
		ClientID:     "mock_id",
		ClientSecret: "mock_sec",
		TokenURL:     ts.URL,
		FlowType:     "authorization_code",
	}

	m := NewManager()
	res, err := m.ExchangeToken(context.Background(), "mock_provider", "auth_code", "http://callback", "verifier")
	if err != nil {
		t.Fatalf("unexpected exchange error: %v", err)
	}
	if res.AccessToken != "mock_access" {
		t.Errorf("expected mock_access, got %s", res.AccessToken)
	}
	if res.RefreshToken != "mock_refresh" {
		t.Errorf("expected mock_refresh, got %s", res.RefreshToken)
	}
}
