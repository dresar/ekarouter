package scripts

import (
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func TestVerifyCompleteDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", "../data/ekarouter.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	tables := []string{
		"accounts",
		"credentials",
		"providers",
		"models",
		"routes",
		"route_items",
		"proxy_profiles",
		"api_keys",
		"settings",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			t.Errorf("error querying table %s: %v", table, err)
		} else {
			t.Logf("Table %-16s: %4d rows", table, count)
		}
	}

	rows, err := db.Query("SELECT id, name, strategy FROM routes")
	if err != nil {
		t.Fatalf("failed to query routes: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, strategy string
		if err := rows.Scan(&id, &name, &strategy); err != nil {
			t.Fatal(err)
		}
		var itemCount int
		_ = db.QueryRow("SELECT COUNT(*) FROM route_items WHERE route_id = ?", id).Scan(&itemCount)
		t.Logf("Route %s (%s) -> %d items", name, strategy, itemCount)
	}

	var proxyCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM proxy_profiles").Scan(&proxyCount)
	t.Logf("Total Proxy Profiles: %d", proxyCount)
}

func TestEnsureFreeProvidersInDB(t *testing.T) {
	db, err := sql.Open("sqlite", "../data/ekarouter.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	type Prov struct {
		ID      string
		Key     string
		Name    string
		Kind    string
		BaseURL string
	}

	freeProviders := []Prov{
		{"opencode", "opencode", "OpenCode Free", "opencode", "https://opencode.ai"},
		{"mimo-free", "mimo-free", "MiMo Code Free", "mimo-free", "https://api.xiaomimimo.com/api/free-ai/openai/chat"},
		{"devin", "devin", "Devin Free", "devin", "https://api.devin.ai/v1"},
		{"groq", "groq", "Groq Cloud", "groq", "https://api.groq.com/openai/v1"},
		{"cerebras", "cerebras", "Cerebras AI", "cerebras", "https://api.cerebras.ai/v1"},
		{"openrouter", "openrouter", "OpenRouter", "openrouter", "https://openrouter.ai/api/v1"},
		{"cloudflare-ai", "cloudflare-ai", "Cloudflare Workers AI", "cloudflare-ai", "https://api.cloudflare.com/client/v4/accounts/ai/v1"},
		{"nvidia", "nvidia", "NVIDIA NIM", "nvidia", "https://integrate.api.nvidia.com/v1"},
		{"api-airforce", "api-airforce", "API.airforce", "api-airforce", "https://api.airforce/v1"},
		{"bazaarlink", "bazaarlink", "Bazaarlink", "bazaarlink", "https://api.bazaarlink.com/v1"},
		{"kilo-gateway", "kilo-gateway", "Kilo Gateway", "kilo-gateway", "https://gateway.kilo.ai/v1"},
		{"kimchi", "kimchi", "Kimchi Dev", "kimchi", "https://llm.kimchi.dev/openai/v1"},
		{"ollama", "ollama", "Ollama", "ollama", "http://localhost:11434/v1"},
		{"chutes", "chutes", "Chutes AI", "chutes", "https://api.chutes.ai/v1"},
		{"huggingface", "huggingface", "HuggingFace", "huggingface", "https://router.huggingface.co/hf-inference/v1"},
		{"kiro", "kiro", "Kiro AI", "kiro", "https://runtime.us-east-1.kiro.dev"},
		{"searxng", "searxng", "SearXNG", "searxng", "http://localhost:8080/search"},
		{"edge-tts", "edge-tts", "Edge TTS", "edge-tts", "http://localhost:5050/v1/audio/speech"},
		{"coqui", "coqui", "Coqui TTS", "coqui", "http://localhost:5002/api/tts"},
	}

	for _, p := range freeProviders {
		_, err := db.Exec(`
INSERT INTO providers (id, key, name, kind, base_url, enabled)
VALUES (?, ?, ?, ?, ?, 1)
ON CONFLICT(id) DO UPDATE SET
	key = excluded.key,
	name = excluded.name,
	kind = excluded.kind,
	base_url = excluded.base_url,
	enabled = 1
`, p.ID, p.Key, p.Name, p.Kind, p.BaseURL)
		if err != nil {
			t.Errorf("failed to register provider %s: %v", p.ID, err)
		} else {
			t.Logf("Provider registered/updated: %s (%s)", p.ID, p.Kind)
		}
	}
}
