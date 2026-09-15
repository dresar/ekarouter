package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type ProviderConfig struct {
	ID            string
	Name          string
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	UserInfoURL   string
	DeviceCodeURL string
	Scopes        []string
	FlowType      string
}

type TokenResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
	Email        string    `json:"email,omitempty"`
	ProjectID    string    `json:"project_id,omitempty"`
}

type DeviceCodeResult struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func decodeCred(envKey string, rawBytes []byte) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	res := make([]byte, len(rawBytes))
	for i, b := range rawBytes {
		res[i] = b ^ 0x5A
	}
	return string(res)
}

var (
	agyIDBytes  = []byte{107, 106, 109, 107, 106, 106, 108, 106, 108, 106, 111, 99, 107, 119, 46, 55, 50, 41, 41, 51, 52, 104, 50, 104, 107, 54, 57, 40, 63, 104, 105, 111, 44, 46, 53, 54, 53, 48, 50, 110, 61, 110, 106, 105, 63, 42, 116, 59, 42, 42, 41, 116, 61, 53, 53, 61, 54, 63, 47, 41, 63, 40, 57, 53, 52, 46, 63, 52, 46, 116, 57, 53, 55}
	agySecBytes = []byte{29, 21, 25, 9, 10, 2, 119, 17, 111, 98, 28, 13, 8, 110, 98, 108, 22, 62, 22, 16, 107, 55, 22, 24, 98, 41, 2, 25, 110, 32, 108, 43, 30, 27, 60}
	gemIDBytes  = []byte{108, 98, 107, 104, 111, 111, 98, 106, 99, 105, 99, 111, 119, 53, 53, 98, 60, 46, 104, 53, 42, 40, 62, 40, 52, 42, 99, 63, 105, 59, 43, 60, 108, 59, 44, 105, 50, 55, 62, 51, 56, 107, 105, 111, 48, 116, 59, 42, 42, 41, 116, 61, 53, 53, 61, 54, 63, 47, 41, 63, 40, 57, 53, 52, 46, 63, 52, 46, 116, 57, 53, 55}
	gemSecBytes = []byte{29, 21, 25, 9, 10, 2, 119, 110, 47, 18, 61, 23, 10, 55, 119, 107, 53, 109, 9, 49, 119, 61, 63, 12, 108, 25, 47, 111, 57, 54, 2, 28, 41, 34, 54}
)

var providerConfigs = map[string]ProviderConfig{
	"antigravity": {
		ID:           "antigravity",
		Name:         "Antigravity",
		ClientID:     decodeCred("ANTIGRAVITY_OAUTH_CLIENT_ID", agyIDBytes),
		ClientSecret: decodeCred("ANTIGRAVITY_OAUTH_CLIENT_SECRET", agySecBytes),
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v1/userinfo",
		Scopes: []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/cclog",
			"https://www.googleapis.com/auth/experimentsandconfigs",
		},
		FlowType: "authorization_code",
	},
	"gemini-agy": {
		ID:           "gemini-agy",
		Name:         "Antigravity",
		ClientID:     decodeCred("ANTIGRAVITY_OAUTH_CLIENT_ID", agyIDBytes),
		ClientSecret: decodeCred("ANTIGRAVITY_OAUTH_CLIENT_SECRET", agySecBytes),
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v1/userinfo",
		Scopes: []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/cclog",
			"https://www.googleapis.com/auth/experimentsandconfigs",
		},
		FlowType: "authorization_code",
	},
	"gemini-cli": {
		ID:           "gemini-cli",
		Name:         "Gemini CLI",
		ClientID:     decodeCred("GEMINI_CLI_OAUTH_CLIENT_ID", gemIDBytes),
		ClientSecret: decodeCred("GEMINI_CLI_OAUTH_CLIENT_SECRET", gemSecBytes),
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v1/userinfo",
		Scopes: []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		FlowType: "authorization_code",
	},
	"claude": {
		ID:       "claude",
		Name:     "Claude Code",
		ClientID: "9d1c250a-e391-49fa-9624-e59325996db7",
		AuthURL:  "https://cli.auth.anthropic.com/authorize",
		TokenURL: "https://cli.auth.anthropic.com/token",
		Scopes:   []string{"openid", "profile", "email"},
		FlowType: "authorization_code_pkce",
	},
	"codex": {
		ID:       "codex",
		Name:     "OpenAI Codex",
		ClientID: "app-p64m2v01",
		AuthURL:  "https://auth.openai.com/oauth/authorize",
		TokenURL: "https://auth.openai.com/oauth/token",
		Scopes:   []string{"openid", "profile", "email", "model.request"},
		FlowType: "authorization_code_pkce",
	},
	"github": {
		ID:            "github",
		Name:          "GitHub Copilot",
		ClientID:      "Iv1.b507a08c87ecfe81",
		DeviceCodeURL: "https://github.com/login/device/code",
		TokenURL:      "https://github.com/login/oauth/access_token",
		Scopes:        []string{"read:user"},
		FlowType:      "device_code",
	},
	"qoder": {
		ID:            "qoder",
		Name:          "Qoder",
		ClientID:      "qoder-ide",
		DeviceCodeURL: "https://api.qoder.co/v1/auth/device/code",
		TokenURL:      "https://api.qoder.co/v1/auth/token",
		Scopes:        []string{"openid", "profile"},
		FlowType:      "device_code",
	},
	"grok-cli": {
		ID:            "grok-cli",
		Name:          "Grok CLI",
		ClientID:      "b1a00492-073a-47ea-816f-4c329264a828",
		DeviceCodeURL: "https://auth.x.ai/oauth2/device/code",
		TokenURL:      "https://auth.x.ai/oauth2/token",
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
			"grok-cli:access",
			"api:access",
			"conversations:read",
			"conversations:write",
		},
		FlowType: "device_code",
	},
	"xai": {
		ID:       "xai",
		Name:     "xAI",
		ClientID: "b1a00492-073a-47ea-816f-4c329264a828",
		AuthURL:  "https://auth.x.ai/oauth2/authorize",
		TokenURL: "https://auth.x.ai/oauth2/token",
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
			"api:access",
		},
		FlowType: "authorization_code_pkce",
	},
	"kimi": {
		ID:            "kimi",
		Name:          "Kimi Code",
		ClientID:      "17e5f671-d194-4dfb-9706-5516cb48c098",
		DeviceCodeURL: "https://auth.kimi.com/api/oauth/device_authorization",
		TokenURL:      "https://auth.kimi.com/api/oauth/token",
		FlowType:      "device_code",
	},
	"kimi-coding": {
		ID:            "kimi-coding",
		Name:          "Kimi Coding",
		ClientID:      "17e5f671-d194-4dfb-9706-5516cb48c098",
		DeviceCodeURL: "https://auth.kimi.com/api/oauth/device_authorization",
		TokenURL:      "https://auth.kimi.com/api/oauth/token",
		FlowType:      "device_code",
	},
	"iflow": {
		ID:           "iflow",
		Name:         "iFlow",
		ClientID:     "10009311001",
		ClientSecret: "4Z3YjXycVsQvyGF1etiNlIBB4RsqSDtW",
		AuthURL:      "https://iflow.cn/oauth",
		TokenURL:     "https://iflow.cn/oauth/token",
		UserInfoURL:  "https://iflow.cn/api/oauth/getUserInfo",
		FlowType:     "authorization_code",
	},
	"gitlab": {
		ID:       "gitlab",
		Name:     "GitLab Duo",
		ClientID: "gitlab-duo",
		AuthURL:  "https://gitlab.com/oauth/authorize",
		TokenURL: "https://gitlab.com/oauth/token",
		Scopes:   []string{"read_user", "api"},
		FlowType: "authorization_code_pkce",
	},
	"cline": {
		ID:       "cline",
		Name:     "Cline",
		ClientID: "cline-cli",
		AuthURL:  "https://app.cline.bot/oauth/authorize",
		TokenURL: "https://api.cline.bot/api/v1/auth/token",
		FlowType: "authorization_code",
	},
	"clinepass": {
		ID:       "clinepass",
		Name:     "ClinePass",
		ClientID: "clinepass",
		AuthURL:  "https://pass.cline.bot/oauth/authorize",
		TokenURL: "https://pass.cline.bot/oauth/token",
		FlowType: "authorization_code",
	},
	"codebuddy-cn": {
		ID:       "codebuddy-cn",
		Name:     "CodeBuddy CN",
		ClientID: "codebuddy-cn",
		AuthURL:  "https://copilot.tencent.com/oauth/authorize",
		TokenURL: "https://copilot.tencent.com/v2/plugin/auth/token",
		FlowType: "authorization_code",
	},
	"codebuddy-intl": {
		ID:       "codebuddy-intl",
		Name:     "CodeBuddy Intl",
		ClientID: "codebuddy-intl",
		AuthURL:  "https://www.codebuddy.ai/oauth/authorize",
		TokenURL: "https://www.codebuddy.ai/v2/plugin/auth/token",
		FlowType: "authorization_code",
	},
	"kilocode": {
		ID:            "kilocode",
		Name:          "KiloCode",
		ClientID:      "kilocode",
		DeviceCodeURL: "https://api.kilo.ai/api/device-auth/codes",
		TokenURL:      "https://api.kilo.ai/api/device-auth/codes",
		FlowType:      "device_code",
	},
	"kimchi": {
		ID:       "kimchi",
		Name:     "Kimchi",
		ClientID: "kimchi",
		AuthURL:  "https://app.kimchi.dev",
		TokenURL: "https://app.kimchi.dev/api/v1/me",
		FlowType: "import_token",
	},
	"trae": {
		ID:           "trae",
		Name:         "Trae",
		ClientID:     "ono9krqynydwx5",
		ClientSecret: "-",
		AuthURL:      "https://www.trae.ai/authorization",
		TokenURL:     "https://api.trae.ai/cloudide/api/v3/trae/oauth/ExchangeToken",
		FlowType:     "authorization_code",
	},
	"windsurf": {
		ID:       "windsurf",
		Name:     "Windsurf",
		ClientID: "3GUryQ7ldAeKEuD2obYnppsnmj58eP5u",
		AuthURL:  "https://www.windsurf.com/windsurf/signin",
		TokenURL: "https://register.windsurf.com/exa.seat_management_pb.SeatManagementService/RegisterUser",
		FlowType: "import_token",
	},
	"cursor": {
		ID:       "cursor",
		Name:     "Cursor",
		ClientID: "cursor",
		FlowType: "import_token",
	},
	"zed": {
		ID:       "zed",
		Name:     "Zed",
		ClientID: "zed",
		AuthURL:  "https://zed.dev/native_app_signin",
		TokenURL: "https://cloud.zed.dev",
		FlowType: "native_app_signin",
	},
	"xiaomi-mimo": {
		ID:       "xiaomi-mimo",
		Name:     "Xiaomi MiMo",
		ClientID: "mimocode",
		AuthURL:  "https://platform.xiaomimimo.com/authorize",
		FlowType: "custom_ecdh",
	},
}

func GetProviderConfig(providerID string) (ProviderConfig, bool) {
	norm := strings.ToLower(strings.TrimSpace(providerID))
	cfg, ok := providerConfigs[norm]
	return cfg, ok
}

func (m *Manager) BuildAuthURL(providerID, redirectURI, state, codeChallenge string) (string, error) {
	cfg, ok := GetProviderConfig(providerID)
	if !ok {
		return "/oauth/authorize?provider=" + url.QueryEscape(providerID) + "&state=" + url.QueryEscape(state) + "&code_challenge=" + url.QueryEscape(codeChallenge), nil
	}

	if cfg.AuthURL == "" {
		return "", fmt.Errorf("provider %s does not support authorization url flow", providerID)
	}

	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)

	if len(cfg.Scopes) > 0 {
		params.Set("scope", strings.Join(cfg.Scopes, " "))
	}

	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	if strings.Contains(cfg.AuthURL, "accounts.google.com") {
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	}

	return fmt.Sprintf("%s?%s", cfg.AuthURL, params.Encode()), nil
}

func (m *Manager) RequestDeviceCode(ctx context.Context, providerID string) (*DeviceCodeResult, error) {
	cfg, ok := GetProviderConfig(providerID)
	if !ok || cfg.DeviceCodeURL == "" {
		return nil, fmt.Errorf("provider %s does not support device code flow", providerID)
	}

	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	if len(cfg.Scopes) > 0 {
		data.Set("scope", strings.Join(cfg.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.DeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if providerID == "grok-cli" {
		req.Header.Set("User-Agent", "grok-pager/0.2.93 grok-shell/0.2.93 (linux; x86_64)")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read device code response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("device code request failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var dcr DeviceCodeResult
	if err := json.Unmarshal(bodyBytes, &dcr); err != nil {
		return nil, fmt.Errorf("parse device code response: %w", err)
	}
	return &dcr, nil
}

func (m *Manager) PollDeviceToken(ctx context.Context, providerID, deviceCode string) (*TokenResult, error) {
	cfg, ok := GetProviderConfig(providerID)
	if !ok || cfg.TokenURL == "" {
		return nil, fmt.Errorf("provider %s does not have a token endpoint configured", providerID)
	}

	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if providerID == "grok-cli" {
		req.Header.Set("User-Agent", "grok-pager/0.2.93 grok-shell/0.2.93 (linux; x86_64)")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poll token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read poll token response: %w", err)
	}

	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(bodyBytes, &errResp)
	if errResp.Error != "" {
		return nil, errors.New(errResp.Error)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("poll token failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var tr TokenResult
	if err := json.Unmarshal(bodyBytes, &tr); err != nil {
		vals, errParse := url.ParseQuery(string(bodyBytes))
		if errParse == nil && vals.Get("access_token") != "" {
			tr.AccessToken = vals.Get("access_token")
			tr.RefreshToken = vals.Get("refresh_token")
			tr.TokenType = vals.Get("token_type")
			tr.Scope = vals.Get("scope")
		} else {
			return nil, fmt.Errorf("parse token response: %w", err)
		}
	}

	if tr.ExpiresIn <= 0 {
		tr.ExpiresIn = 3600
	}
	tr.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)

	return &tr, nil
}

func (m *Manager) ExchangeToken(ctx context.Context, providerID, code, redirectURI, codeVerifier string) (*TokenResult, error) {
	cfg, ok := GetProviderConfig(providerID)
	if !ok {
		return &TokenResult{
			AccessToken: code,
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}, nil
	}

	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("provider %s does not have a token endpoint configured", providerID)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", cfg.ClientID)
	if cfg.ClientSecret != "" {
		data.Set("client_secret", cfg.ClientSecret)
	}
	data.Set("code", code)
	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	}
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token exchange response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var tr TokenResult
	if err := json.Unmarshal(bodyBytes, &tr); err != nil {
		vals, errParse := url.ParseQuery(string(bodyBytes))
		if errParse == nil && vals.Get("access_token") != "" {
			tr.AccessToken = vals.Get("access_token")
			tr.RefreshToken = vals.Get("refresh_token")
			tr.TokenType = vals.Get("token_type")
			tr.Scope = vals.Get("scope")
		} else {
			return nil, fmt.Errorf("parse token response: %w", err)
		}
	}

	if tr.ExpiresIn <= 0 {
		tr.ExpiresIn = 3600
	}
	tr.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)

	if cfg.UserInfoURL != "" && tr.AccessToken != "" {
		userInfoReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL+"?alt=json", nil)
		if err == nil {
			userInfoReq.Header.Set("Authorization", "Bearer "+tr.AccessToken)
			userInfoReq.Header.Set("x-request-source", "local")
			if userResp, err := client.Do(userInfoReq); err == nil {
				defer userResp.Body.Close()
				var uInfo struct {
					Email string `json:"email"`
				}
				if json.NewDecoder(userResp.Body).Decode(&uInfo) == nil && uInfo.Email != "" {
					tr.Email = uInfo.Email
				}
			}
		}
	}

	if (providerID == "antigravity" || providerID == "gemini-agy" || providerID == "gemini-cli") && tr.AccessToken != "" {
		metaReq := map[string]any{
			"metadata": map[string]any{
				"ideType":    9,
				"platform":   5,
				"pluginType": 2,
			},
		}
		jsonBody, _ := json.Marshal(metaReq)
		loadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", bytes.NewReader(jsonBody))
		if err == nil {
			loadReq.Header.Set("Authorization", "Bearer "+tr.AccessToken)
			loadReq.Header.Set("Content-Type", "application/json")
			loadReq.Header.Set("User-Agent", "antigravity/ide/2.11.0 darwin/arm64")
			loadReq.Header.Set("x-request-source", "local")
			if loadResp, err := client.Do(loadReq); err == nil {
				defer loadResp.Body.Close()
				var loadData struct {
					CloudAICompanionProject any `json:"cloudaicompanionProject"`
				}
				if json.NewDecoder(loadResp.Body).Decode(&loadData) == nil {
					switch v := loadData.CloudAICompanionProject.(type) {
					case string:
						tr.ProjectID = v
					case map[string]any:
						if id, ok := v["id"].(string); ok {
							tr.ProjectID = id
						}
					}
				}
			}
		}
	}

	return &tr, nil
}

func (m *Manager) RefreshToken(ctx context.Context, providerID, refreshToken string) (*TokenResult, error) {
	if refreshToken == "" {
		return nil, errors.New("empty refresh token")
	}

	cfg, ok := GetProviderConfig(providerID)
	if !ok || cfg.TokenURL == "" {
		return nil, fmt.Errorf("provider %s does not support token refresh", providerID)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", cfg.ClientID)
	if cfg.ClientSecret != "" {
		data.Set("client_secret", cfg.ClientSecret)
	}
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read refresh token response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("refresh token failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var tr TokenResult
	if err := json.Unmarshal(bodyBytes, &tr); err != nil {
		return nil, fmt.Errorf("unmarshal refresh token response: %w", err)
	}

	if tr.RefreshToken == "" {
		tr.RefreshToken = refreshToken
	}

	if tr.ExpiresIn <= 0 {
		tr.ExpiresIn = 3600
	}
	tr.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)

	return &tr, nil
}
