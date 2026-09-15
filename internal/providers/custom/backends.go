package custom

import (
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

const (
	BackendGeneric     = "custom"
	BackendOllama      = "ollama"
	BackendVLLM        = "vllm"
	BackendDeepSeek    = "deepseek"
	BackendLocalAI     = "localai"
	BackendGroq        = "groq"
	BackendOpenRouter  = "openrouter"
	BackendMistral     = "mistral"
	BackendTogether    = "together"
	BackendXAI         = "xai"
	BackendSambaNova   = "sambanova"
	BackendFireworks   = "fireworks"
	BackendSiliconFlow = "siliconflow"
	BackendNebius      = "nebius"
	BackendPerplexity  = "perplexity"
	BackendVenice      = "venice"
	BackendTencent     = "tencent"
	BackendMiniMax     = "minimax"
	BackendGLM         = "glm"
	BackendCohere      = "cohere"
	BackendFeatherless = "featherless"
	BackendPoolside    = "poolside"
	BackendByteplus    = "byteplus"
	BackendCursor      = "cursor"
	BackendWindsurf    = "windsurf"
	BackendCline       = "cline"
	BackendTrae        = "trae"
	BackendZed         = "zed"
	BackendQoder       = "qoder"
	BackendIFlow       = "iflow"
	BackendKimi        = "kimi"
	BackendKiloCode    = "kilocode"
	BackendCodeBuddy   = "codebuddy"
	BackendAzure       = "azure"
	BackendVertex      = "vertex"
)

func DefaultBaseURL(backend string) string {
	switch strings.ToLower(backend) {
	case BackendOllama:
		return "http://localhost:11434/v1"
	case BackendVLLM:
		return "http://localhost:8000/v1"
	case BackendDeepSeek:
		return "https://api.deepseek.com/v1"
	case BackendLocalAI:
		return "http://localhost:8080/v1"
	case BackendGroq:
		return "https://api.groq.com/openai/v1"
	case BackendOpenRouter:
		return "https://openrouter.ai/api/v1"
	case BackendMistral:
		return "https://api.mistral.ai/v1"
	case BackendTogether:
		return "https://api.together.xyz/v1"
	case BackendXAI:
		return "https://api.x.ai/v1"
	case BackendSambaNova:
		return "https://api.sambanova.ai/v1"
	case BackendFireworks:
		return "https://api.fireworks.ai/inference/v1"
	case BackendSiliconFlow:
		return "https://api.siliconflow.com/v1"
	case BackendNebius:
		return "https://api.studio.nebius.ai/v1"
	case BackendPerplexity:
		return "https://api.perplexity.ai"
	case BackendVenice:
		return "https://api.venice.ai/api/v1"
	case BackendTencent:
		return "https://api.hunyuan.cloud.tencent.com/v1"
	case BackendMiniMax:
		return "https://api.minimax.chat/v1"
	case BackendGLM:
		return "https://open.bigmodel.cn/api/paas/v4"
	case BackendCohere:
		return "https://api.cohere.ai/v1"
	case BackendFeatherless:
		return "https://api.featherless.ai/v1"
	case BackendPoolside:
		return "https://api.poolside.ai/v1"
	case BackendByteplus:
		return "https://ark.cn-beijing.volces.com/api/v3"
	case BackendCursor:
		return "https://api2.cursor.sh/v1"
	case BackendWindsurf:
		return "https://server.codeium.com/v1"
	case BackendCline:
		return "https://api.cline.bot/api/v1"
	case BackendTrae:
		return "https://api.trae.ai/v1"
	case BackendZed:
		return "https://cloud.zed.dev/v1"
	case BackendQoder:
		return "https://api.qoder.co/v1"
	case BackendIFlow:
		return "https://api.iflow.cn/v1"
	case BackendKimi:
		return "https://api.moonshot.cn/v1"
	case BackendKiloCode:
		return "https://api.kilo.ai/v1"
	case BackendCodeBuddy:
		return "https://copilot.tencent.com/v1"
	case BackendAzure:
		return "https://models.inference.ai.azure.com"
	case BackendVertex:
		return "https://aiplatform.googleapis.com/v1"
	default:
		return ""
	}
}

func ResolveBaseURL(backend string, creds *providers.Credentials) string {
	if creds != nil && creds.BaseURL != "" {
		return strings.TrimRight(creds.BaseURL, "/")
	}
	def := DefaultBaseURL(backend)
	if def != "" {
		return def
	}
	return "http://localhost:8080/v1"
}

func ApplyBackendHeaders(backend string, creds *providers.Credentials, h http.Header) {
	h.Set("Content-Type", "application/json")
	if strings.ToLower(backend) == BackendOpenRouter {
		h.Set("HTTP-Referer", "https://github.com/dresar/ekarouter")
		h.Set("X-Title", "EkaRouter")
	}

	if creds != nil {
		if creds.APIKey != "" {
			h.Set("Authorization", "Bearer "+creds.APIKey)
		} else if creds.AccessToken != "" {
			h.Set("Authorization", "Bearer "+creds.AccessToken)
		}
		for k, v := range creds.Headers {
			h.Set(k, v)
		}
	}
}

func ApplyBackendCustomHeaders(backend string, h map[string]string) {
	if strings.ToLower(backend) == BackendOpenRouter {
		h["HTTP-Referer"] = "https://github.com/dresar/ekarouter"
		h["X-Title"] = "EkaRouter"
	}
}
