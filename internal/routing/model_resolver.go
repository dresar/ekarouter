package routing

import (
	"regexp"
	"strings"
)

type ModelInfo struct {
	RawModel        string
	Provider        string
	Model           string
	IsAlias         bool
	ProviderAlias   string
	ThinkingSuffix  string
	NormalizedModel string
}

var ProviderAliasMap = map[string]string{
	"ag":                  "antigravity",
	"antigravity":         "antigravity",
	"oa":                  "openai",
	"openai":              "openai",
	"an":                  "anthropic",
	"anthropic":           "anthropic",
	"claude":              "anthropic",
	"cc":                  "anthropic",
	"gm":                  "gemini",
	"gemini":              "gemini",
	"gc":                  "gemini-cli",
	"gemini-cli":          "gemini-cli",
	"or":                  "openrouter",
	"openrouter":          "openrouter",
	"gr":                  "groq",
	"groq":                "groq",
	"kr":                  "kiro",
	"kiro":                "kiro",
	"oc":                  "opencode",
	"opencode":            "opencode",
	"ocg":                 "opencode-go",
	"opencode-go":         "opencode-go",
	"cx":                  "codex",
	"codex":               "codex",
	"qd":                  "qoder",
	"qoder":               "qoder",
	"kc":                  "kilocode",
	"kilocode":            "kilocode",
	"cl":                  "cline",
	"cline":               "cline",
	"clinepass":           "cline",
	"cb":                  "codebuddy-intl",
	"codebuddy":           "codebuddy-intl",
	"codebuddy-intl":      "codebuddy-intl",
	"codebuddy-cn":        "codebuddy-cn",
	"nv":                  "nvidia",
	"nvidia":              "nvidia",
	"ch":                  "chutes",
	"chutes":              "chutes",
	"ol":                  "ollama",
	"ollama":              "ollama",
	"cf":                  "cloudflare-ai",
	"cloudflare":          "cloudflare-ai",
	"cloudflare-ai":       "cloudflare-ai",
	"mm":                  "xiaomi-mimo",
	"xiaomi-mimo":         "xiaomi-mimo",
	"mmf":                 "mimo-free",
	"mimo-free":           "mimo-free",
	"dv":                  "devin",
	"devin":               "devin",
	"devin-cli":           "devin",
	"af":                  "api-airforce",
	"api-airforce":        "api-airforce",
	"bzl":                 "bazaarlink",
	"bazaarlink":          "bazaarlink",
	"kgw":                 "kilo-gateway",
	"kilo-gateway":        "kilo-gateway",
	"apx":                 "node_apinex",
	"node_apinex":         "node_apinex",
	"dahl":                "node_dahl_global",
	"node_dahl_global":    "node_dahl_global",
	"zans":                "node_zanslab_id",
	"node_zanslab_id":     "node_zanslab_id",
	"node_openagentic_id": "node_openagentic_id",
	"cur":                 "cursor",
	"cursor":              "cursor",
	"ws":                  "windsurf",
	"windsurf":            "windsurf",
	"gcli":                "grok-cli",
	"grok-cli":            "grok-cli",
	"xai":                 "xai",
	"deepseek":            "deepseek",
	"mistral":             "mistral",
	"together":            "together",
	"cerebras":            "cerebras",
	"cohere":              "cohere",
	"hyperbolic":          "hyperbolic",
	"siliconflow":         "siliconflow",
	"venice":              "venice",
	"kimi":                "kimi",
}

var thinkingRegex = regexp.MustCompile(`\([^()]+\)\s*$`)
var dotVersionRegex = regexp.MustCompile(`(\d+)-(\d+)`)

func ResolveProviderAlias(aliasOrID string) string {
	lower := strings.ToLower(strings.TrimSpace(aliasOrID))
	if mapped, ok := ProviderAliasMap[lower]; ok {
		return mapped
	}
	return lower
}

func ExtractThinkingSuffix(model string) (string, string) {
	loc := thinkingRegex.FindStringIndex(model)
	if loc == nil {
		return model, ""
	}
	base := strings.TrimSpace(model[:loc[0]])
	suffix := strings.TrimSpace(model[loc[0]:])
	return base, suffix
}

func NormalizeVersion(model string) string {
	return dotVersionRegex.ReplaceAllString(model, "$1.$2")
}

func DetectProviderFamily(model string) string {
	lower := strings.ToLower(model)
	switch {
	case strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "o4") || strings.HasPrefix(lower, "chatgpt") || strings.HasPrefix(lower, "text-embedding"):
		return "openai"
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	case strings.HasPrefix(lower, "gemini") || strings.HasPrefix(lower, "gemma"):
		return "gemini"
	case strings.HasPrefix(lower, "llama") || strings.HasPrefix(lower, "mixtral"):
		return "groq"
	case strings.HasPrefix(lower, "deepseek"):
		return "deepseek"
	case strings.HasPrefix(lower, "qwen"):
		return "openrouter"
	case strings.HasPrefix(lower, "grok"):
		return "xai"
	case strings.HasPrefix(lower, "mistral") || strings.HasPrefix(lower, "codestral"):
		return "mistral"
	case strings.HasPrefix(lower, "command"):
		return "cohere"
	default:
		return ""
	}
}

func ParseModel(modelStr string) ModelInfo {
	trimmed := strings.TrimSpace(modelStr)
	if trimmed == "" {
		return ModelInfo{}
	}

	baseModel, thinkingSuffix := ExtractThinkingSuffix(trimmed)

	if strings.Contains(baseModel, "/") {
		parts := strings.SplitN(baseModel, "/", 2)
		prefix := strings.ToLower(strings.TrimSpace(parts[0]))
		actualModel := strings.TrimSpace(parts[1])

		prov := ResolveProviderAlias(prefix)

		normalized := actualModel
		if prov == "kiro" || prov == "anthropic" || prov == "gemini" {
			normalized = NormalizeVersion(actualModel)
		}

		return ModelInfo{
			RawModel:        trimmed,
			Provider:        prov,
			Model:           actualModel,
			IsAlias:         false,
			ProviderAlias:   prefix,
			ThinkingSuffix:  thinkingSuffix,
			NormalizedModel: normalized,
		}
	}

	detectedFamily := DetectProviderFamily(baseModel)

	return ModelInfo{
		RawModel:        trimmed,
		Provider:        detectedFamily,
		Model:           baseModel,
		IsAlias:         detectedFamily == "",
		ProviderAlias:   "",
		ThinkingSuffix:  thinkingSuffix,
		NormalizedModel: baseModel,
	}
}
