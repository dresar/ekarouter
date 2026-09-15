package routing

import (
	"testing"
)

func TestParseModel_Prefixes(t *testing.T) {
	cases := []struct {
		input            string
		expectedProvider string
		expectedModel    string
		expectedAlias    string
	}{
		{"ag/gemini-2.5-flash", "antigravity", "gemini-2.5-flash", "ag"},
		{"antigravity/gemini-pro", "antigravity", "gemini-pro", "antigravity"},
		{"oa/gpt-4o", "openai", "gpt-4o", "oa"},
		{"openai/o3-mini", "openai", "o3-mini", "openai"},
		{"an/claude-3-7-sonnet", "anthropic", "claude-3-7-sonnet", "an"},
		{"claude/claude-3-5-sonnet", "anthropic", "claude-3-5-sonnet", "claude"},
		{"gm/gemini-2.5-flash", "gemini", "gemini-2.5-flash", "gm"},
		{"or/meta-llama/llama-3.3-70b-instruct", "openrouter", "meta-llama/llama-3.3-70b-instruct", "or"},
		{"gr/llama-3.3-70b-versatile", "groq", "llama-3.3-70b-versatile", "gr"},
		{"kr/claude-sonnet-4-5", "kiro", "claude-sonnet-4-5", "kr"},
		{"oc/gpt-4o", "opencode", "gpt-4o", "oc"},
		{"ocg/muse-spark", "opencode-go", "muse-spark", "ocg"},
		{"gcli/grok-3", "grok-cli", "grok-3", "gcli"},
		{"cx/gpt-4o-review", "codex", "gpt-4o-review", "cx"},
		{"qd/deepseek-v3", "qoder", "deepseek-v3", "qd"},
		{"kc/code-model", "kilocode", "code-model", "kc"},
		{"cl/cline-custom", "cline", "cline-custom", "cl"},
		{"cb/buddy-code", "codebuddy-intl", "buddy-code", "cb"},
		{"nv/deepseek-r1", "nvidia", "deepseek-r1", "nv"},
		{"ch/deepseek-ai", "chutes", "deepseek-ai", "ch"},
		{"ol/llama3.2", "ollama", "llama3.2", "ol"},
		{"cf/meta-llama-3-8b", "cloudflare-ai", "meta-llama-3-8b", "cf"},
		{"mm/mimo-preview", "xiaomi-mimo", "mimo-preview", "mm"},
		{"mmf/mimo-free", "mimo-free", "mimo-free", "mmf"},
		{"dv/devin-agent", "devin", "devin-agent", "dv"},
	}

	for _, c := range cases {
		info := ParseModel(c.input)
		if info.Provider != c.expectedProvider {
			t.Errorf("input %s: expected provider %s, got %s", c.input, c.expectedProvider, info.Provider)
		}
		if info.Model != c.expectedModel {
			t.Errorf("input %s: expected model %s, got %s", c.input, c.expectedModel, info.Model)
		}
		if info.ProviderAlias != c.expectedAlias {
			t.Errorf("input %s: expected alias %s, got %s", c.input, c.expectedAlias, info.ProviderAlias)
		}
		if info.IsAlias {
			t.Errorf("input %s: expected IsAlias=false", c.input)
		}
	}
}

func TestParseModel_ThinkingSuffixes(t *testing.T) {
	cases := []struct {
		input          string
		expectedModel  string
		expectedSuffix string
	}{
		{"gemini-3.8-flash-high(high)", "gemini-3.8-flash-high", "(high)"},
		{"claude-3-7-sonnet(medium)", "claude-3-7-sonnet", "(medium)"},
		{"ag/gemini-3.7-flash-tiered(low)", "gemini-3.7-flash-tiered", "(low)"},
		{"gpt-oss-120b(thought)", "gpt-oss-120b", "(thought)"},
	}

	for _, c := range cases {
		info := ParseModel(c.input)
		if info.Model != c.expectedModel {
			t.Errorf("input %s: expected model %s, got %s", c.input, c.expectedModel, info.Model)
		}
		if info.ThinkingSuffix != c.expectedSuffix {
			t.Errorf("input %s: expected suffix %s, got %s", c.input, c.expectedSuffix, info.ThinkingSuffix)
		}
	}
}

func TestParseModel_BareFamilies(t *testing.T) {
	cases := []struct {
		input            string
		expectedProvider string
	}{
		{"gpt-4o", "openai"},
		{"gpt-4o-mini", "openai"},
		{"o1-preview", "openai"},
		{"o3-mini", "openai"},
		{"text-embedding-3-small", "openai"},
		{"claude-3-5-sonnet", "anthropic"},
		{"claude-3-7-sonnet", "anthropic"},
		{"gemini-2.5-flash", "gemini"},
		{"gemini-2.5-pro", "gemini"},
		{"gemma-2-9b-it", "gemini"},
		{"llama-3.3-70b-versatile", "groq"},
		{"mixtral-8x7b-32768", "groq"},
		{"deepseek-chat", "deepseek"},
		{"deepseek-reasoner", "deepseek"},
		{"qwen-2.5-72b-instruct", "openrouter"},
		{"grok-2", "xai"},
		{"mistral-large-latest", "mistral"},
		{"codestral-2501", "mistral"},
		{"command-r-plus", "cohere"},
	}

	for _, c := range cases {
		info := ParseModel(c.input)
		if info.Provider != c.expectedProvider {
			t.Errorf("input %s: expected provider %s, got %s", c.input, c.expectedProvider, info.Provider)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"claude-sonnet-4-5", "claude-sonnet-4.5"},
		{"gemini-2-5-flash", "gemini-2.5-flash"},
		{"claude-3-7-sonnet", "claude-3.7-sonnet"},
	}

	for _, c := range cases {
		got := NormalizeVersion(c.input)
		if got != c.expected {
			t.Errorf("input %s: expected %s, got %s", c.input, c.expected, got)
		}
	}
}

func TestRouter_DynamicModelResolutionWithAliases(t *testing.T) {
	router := NewRouter(nil)
	router.SetProvider("p-ag", "antigravity")
	router.SetProvider("p-oa", "openai")
	router.SetProvider("p-an", "anthropic")
	router.SetProvider("p-gr", "groq")

	router.SetAccount(&Account{ID: "acc-ag", ProviderID: "p-ag", State: "active", Enabled: true})
	router.SetAccount(&Account{ID: "acc-oa", ProviderID: "p-oa", State: "active", Enabled: true})
	router.SetAccount(&Account{ID: "acc-an", ProviderID: "p-an", State: "active", Enabled: true})
	router.SetAccount(&Account{ID: "acc-gr", ProviderID: "p-gr", State: "active", Enabled: true})

	targetsAg, err := router.SelectTargets("ag/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("unexpected error for ag prefix: %v", err)
	}
	if len(targetsAg) != 1 || targetsAg[0].AccountID != "acc-ag" || targetsAg[0].ModelName != "gemini-2.5-flash" {
		t.Errorf("unexpected target for ag: %+v", targetsAg)
	}

	targetsOa, err := router.SelectTargets("oa/gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error for oa prefix: %v", err)
	}
	if len(targetsOa) != 1 || targetsOa[0].AccountID != "acc-oa" || targetsOa[0].ModelName != "gpt-4o" {
		t.Errorf("unexpected target for oa: %+v", targetsOa)
	}

	targetsAn, err := router.SelectTargets("an/claude-3-7-sonnet(high)")
	if err != nil {
		t.Fatalf("unexpected error for an prefix: %v", err)
	}
	if len(targetsAn) != 1 || targetsAn[0].AccountID != "acc-an" || targetsAn[0].ModelName != "claude-3-7-sonnet(high)" {
		t.Errorf("unexpected target for an: %+v", targetsAn)
	}

	targetsGr, err := router.SelectTargets("gr/llama-3.3-70b-versatile")
	if err != nil {
		t.Fatalf("unexpected error for gr prefix: %v", err)
	}
	if len(targetsGr) != 1 || targetsGr[0].AccountID != "acc-gr" || targetsGr[0].ModelName != "llama-3.3-70b-versatile" {
		t.Errorf("unexpected target for gr: %+v", targetsGr)
	}
}
