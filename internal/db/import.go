package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
)

type ImportStats struct {
	Providers   int
	Accounts    int
	Credentials int
	Models      int
	Routes      int
	RouteItems  int
	ProxyPools  int
	APIKeys     int
	Settings    int
}

type backupRoot struct {
	Settings            map[string]any   `json:"settings"`
	ProviderConnections []map[string]any `json:"providerConnections"`
	ProviderNodes       []map[string]any `json:"providerNodes"`
	ProxyPools          []map[string]any `json:"proxyPools"`
	APIKeys             []map[string]any `json:"apiKeys"`
	Combos              []map[string]any `json:"combos"`
	CustomModels        []map[string]any `json:"customModels"`
}

var providerAliases = map[string]string{
	"ag":                  "antigravity",
	"gh":                  "github",
	"gc":                  "gemini-cli",
	"cl":                  "cline",
	"af":                  "api-airforce",
	"kgw":                 "kilo-gateway",
	"bzl":                 "bazaarlink",
	"bai":                 "node_b_ai_api",
	"9inf":                "node_9inference_cloud",
	"node_openagentic_id": "node_openagentic_id",
}

var defaultProviderURLs = map[string]struct{ BaseURL, Kind string }{
	"gemini":            {"https://generativelanguage.googleapis.com", "gemini"},
	"gemini-cli":        {"https://cloudcode-pa.googleapis.com", "gemini"},
	"antigravity":       {"https://daily-cloudcode-pa.googleapis.com", "gemini"},
	"anthropic":         {"https://api.anthropic.com", "anthropic"},
	"claude":            {"https://api.anthropic.com", "anthropic"},
	"openai":            {"https://api.openai.com/v1", "openai"},
	"groq":              {"https://api.groq.com/openai/v1", "openai"},
	"openrouter":        {"https://openrouter.ai/api/v1", "openai"},
	"mistral":           {"https://api.mistral.ai/v1", "openai"},
	"siliconflow":       {"https://api.siliconflow.cn/v1", "openai"},
	"venice":            {"https://api.venice.ai/api/v1", "openai"},
	"together":          {"https://api.together.xyz/v1", "openai"},
	"xai":               {"https://api.x.ai/v1", "openai"},
	"grok-cli":          {"https://api.x.ai/v1", "openai"},
	"nvidia":            {"https://integrate.api.nvidia.com/v1", "openai"},
	"cohere":            {"https://api.cohere.com/v2", "openai"},
	"hyperbolic":        {"https://api.hyperbolic.xyz/v1", "openai"},
	"chutes":            {"https://api.chutes.ai/v1", "openai"},
	"github":            {"https://models.inference.ai.azure.com", "openai"},
	"ollama":            {"http://localhost:11434/v1", "openai"},
	"qoder":             {"https://api.qoder.co/v1", "openai"},
	"kilocode":          {"https://api.kilocode.ai/v1", "openai"},
	"cline":             {"https://api.cline.bot/v1", "openai"},
	"clinepass":         {"https://api.cline.bot/v1", "openai"},
	"codebuddy-intl":    {"https://api.codebuddy.ai/v1", "openai"},
	"cloudflare-ai":     {"https://api.cloudflare.com/client/v4/accounts/ai/v1", "openai"},
	"tokenrouter":       {"https://api.tokenrouter.ai/v1", "openai"},
	"exa":               {"https://api.exa.ai", "custom"},
	"xiaomi-tokenplan":  {"https://api.mimo.mi.com/v1", "openai"},
	"vercel-ai-gateway": {"https://gateway.ai.cloudflare.com/v1", "openai"},
	"llm7":              {"https://api.llm7.com/v1", "openai"},
	"morph":             {"https://api.morph.so/v1", "openai"},
	"kimi":              {"https://api.moonshot.cn/v1", "openai"},
	"kiro":              {"https://codewhisperer.us-east-1.amazonaws.com", "custom"},
	"byteplus":          {"https://ark.cn-beijing.volces.com/api/v3", "openai"},
	"api-airforce":      {"https://api.airforce/v1", "openai"},
	"bazaarlink":        {"https://api.bazaarlink.com/v1", "openai"},
	"kilo-gateway":      {"https://gateway.kilo.ai/v1", "openai"},
	"poolside":          {"https://api.poolside.ai/v1", "openai"},
}

func Import9RouterBackup(database *DB, crypto *auth.CryptoService, filePath string) (*ImportStats, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read backup file: %w", err)
	}

	var root backupRoot
	if err := json.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse backup json: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stats := &ImportStats{}

	for k, v := range root.Settings {
		valBytes, _ := json.Marshal(v)
		_, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)", k, string(valBytes))
		if err == nil {
			stats.Settings++
		}
	}

	knownProviders := make(map[string]bool)

	for _, node := range root.ProviderNodes {
		id, _ := node["id"].(string)
		name, _ := node["name"].(string)
		baseUrl, _ := node["baseUrl"].(string)
		if baseUrl == "" {
			if dataMap, ok := node["data"].(map[string]any); ok {
				baseUrl, _ = dataMap["baseUrl"].(string)
			}
		}

		if id == "" {
			continue
		}
		if name == "" {
			name = id
		}
		if baseUrl == "" {
			baseUrl = "https://api.openai.com/v1"
		}

		_, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO providers (id, key, name, kind, base_url, enabled) VALUES (?, ?, ?, ?, ?, 1)",
			id, id, name, "custom", baseUrl)
		if err == nil {
			knownProviders[id] = true
			stats.Providers++
		}
	}

	ensureProvider := func(provID string) {
		if knownProviders[provID] {
			return
		}
		name := strings.Title(strings.ReplaceAll(provID, "-", " "))
		kind := "custom"
		baseUrl := "https://api.openai.com/v1"

		if def, ok := defaultProviderURLs[provID]; ok {
			kind = def.Kind
			baseUrl = def.BaseURL
		}

		_, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO providers (id, key, name, kind, base_url, enabled) VALUES (?, ?, ?, ?, ?, 1)",
			provID, provID, name, kind, baseUrl)
		if err == nil {
			knownProviders[provID] = true
			stats.Providers++
		}
	}

	for _, conn := range root.ProviderConnections {
		id, _ := conn["id"].(string)
		provider, _ := conn["provider"].(string)
		if id == "" || provider == "" {
			continue
		}

		ensureProvider(provider)

		name, _ := conn["displayName"].(string)
		if name == "" {
			name, _ = conn["name"].(string)
		}
		if name == "" {
			name, _ = conn["email"].(string)
		}
		if name == "" {
			name = id[:8]
		}

		authType, _ := conn["authType"].(string)
		if authType == "" {
			authType = "apiKey"
		}

		priority := 10
		if pVal, ok := conn["priority"].(float64); ok {
			priority = int(pVal)
		}

		enabled := 1
		if active, ok := conn["isActive"].(bool); ok && !active {
			enabled = 0
		}

		state := "active"
		if st, ok := conn["testStatus"].(string); ok && st != "" {
			state = st
		}

		var expiresAt *time.Time
		if expStr, ok := conn["expiresAt"].(string); ok && expStr != "" {
			if t, err := time.Parse(time.RFC3339, expStr); err == nil {
				expiresAt = &t
			}
		}

		_, err := tx.ExecContext(ctx, `
INSERT OR REPLACE INTO accounts (id, provider_id, name, auth_type, state, priority, enabled, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, id, provider, name, authType, state, priority, enabled, expiresAt)
		if err != nil {
			continue
		}
		stats.Accounts++

		accessToken, _ := conn["accessToken"].(string)
		if accessToken == "" {
			accessToken, _ = conn["apiKey"].(string)
		}
		refreshToken, _ := conn["refreshToken"].(string)

		secretData := make(map[string]any)
		for k, v := range conn {
			if strings.HasPrefix(k, "modelLock_") || k == "providerSpecificData" || k == "lastError" || k == "errorCode" || k == "backoffLevel" {
				secretData[k] = v
			}
		}
		secretBytes, _ := json.Marshal(secretData)

		encAccess, _ := crypto.Encrypt(accessToken)
		encRefresh, _ := crypto.Encrypt(refreshToken)
		encSecret, _ := crypto.Encrypt(string(secretBytes))

		h := sha256.Sum256([]byte(accessToken))
		fingerprint := hex.EncodeToString(h[:8])

		_, err = tx.ExecContext(ctx, `
INSERT OR REPLACE INTO credentials (id, account_id, encrypted_access, encrypted_refresh, encrypted_secret, key_fingerprint)
VALUES (?, ?, ?, ?, ?, ?)`, "cred-"+id, id, encAccess, encRefresh, encSecret, fingerprint)
		if err == nil {
			stats.Credentials++
		}
	}

	for _, pool := range root.ProxyPools {
		id, _ := pool["id"].(string)
		name, _ := pool["name"].(string)
		rawUrl, _ := pool["proxyUrl"].(string)
		if id == "" || rawUrl == "" {
			continue
		}
		if name == "" {
			name = id[:8]
		}

		parsed, err := url.Parse(rawUrl)
		scheme := "https"
		host := rawUrl
		port := 443
		username := ""
		password := ""

		if err == nil {
			scheme = parsed.Scheme
			host = parsed.Hostname()
			if parsed.Port() != "" {
				port, _ = strconv.Atoi(parsed.Port())
			} else if scheme == "http" {
				port = 80
			}
			if parsed.User != nil {
				username = parsed.User.Username()
				password, _ = parsed.User.Password()
			}
		}

		encPass, _ := crypto.Encrypt(password)

		enabled := 1
		if active, ok := pool["isActive"].(bool); ok && !active {
			enabled = 0
		}

		_, err = tx.ExecContext(ctx, `
INSERT OR REPLACE INTO proxy_profiles (id, name, scheme, host, port, username, encrypted_password, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, id, name, scheme, host, port, username, encPass, enabled)
		if err == nil {
			stats.ProxyPools++
		}
	}

	for _, m := range root.CustomModels {
		alias, _ := m["providerAlias"].(string)
		id, _ := m["id"].(string)
		name, _ := m["name"].(string)
		if id == "" || alias == "" {
			continue
		}
		if name == "" {
			name = id
		}

		ensureProvider(alias)

		modelKey := fmt.Sprintf("%s/%s", alias, id)
		_, err := tx.ExecContext(ctx, `
INSERT OR REPLACE INTO models (id, provider_id, external_name, display_name, context_limit, input_capability, output_capability, streaming, enabled)
VALUES (?, ?, ?, ?, 128000, 'text', 'text', 1, 1)`, modelKey, alias, id, name)
		if err == nil {
			stats.Models++
		}
	}

	for _, combo := range root.Combos {
		id, _ := combo["id"].(string)
		name, _ := combo["name"].(string)
		modelsRaw, _ := combo["models"].([]any)
		if id == "" || name == "" {
			continue
		}

		_, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO routes (id, name, strategy, enabled) VALUES (?, ?, 'priority', 1)", id, name)
		if err != nil {
			continue
		}
		stats.Routes++

		_ = tx.QueryRowContext(ctx, "DELETE FROM route_items WHERE route_id = ?", id)

		for idx, mRaw := range modelsRaw {
			modelStr, _ := mRaw.(string)
			if modelStr == "" {
				continue
			}

			parts := strings.SplitN(modelStr, "/", 2)
			if len(parts) < 2 {
				continue
			}

			alias := parts[0]
			modelName := parts[1]

			provID := alias
			if canonical, ok := providerAliases[alias]; ok {
				provID = canonical
			}

			ensureProvider(provID)

			modelKey := fmt.Sprintf("%s/%s", provID, modelName)
			_, _ = tx.ExecContext(ctx, `
INSERT OR IGNORE INTO models (id, provider_id, external_name, display_name, context_limit, input_capability, output_capability, streaming, enabled)
VALUES (?, ?, ?, ?, 128000, 'text', 'text', 1, 1)`, modelKey, provID, modelName, modelName)

			itemID := fmt.Sprintf("%s-%d", id, idx)
			_, err = tx.ExecContext(ctx, `
INSERT OR REPLACE INTO route_items (id, route_id, provider_id, model_id, priority, weight, enabled, timeout_ms, max_retries)
VALUES (?, ?, ?, ?, ?, 1, 1, 60000, 2)`, itemID, id, provID, modelKey, idx+1)
			if err == nil {
				stats.RouteItems++
			}
		}
	}

	for _, keyEntry := range root.APIKeys {
		id, _ := keyEntry["id"].(string)
		rawKey, _ := keyEntry["key"].(string)
		name, _ := keyEntry["name"].(string)
		if id == "" || rawKey == "" {
			continue
		}
		if name == "" {
			name = "API Key"
		}

		hash := auth.HashToken(rawKey)
		prefix := rawKey
		if len(prefix) > 10 {
			prefix = prefix[:10]
		}

		enabled := 1
		if active, ok := keyEntry["isActive"].(bool); ok && !active {
			enabled = 0
		}

		_, err := tx.ExecContext(ctx, `
INSERT OR REPLACE INTO api_keys (id, name, prefix, hash, scopes, enabled)
VALUES (?, ?, ?, ?, '*', ?)`, id, name, prefix, hash, enabled)
		if err == nil {
			stats.APIKeys++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return stats, nil
}
