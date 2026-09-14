package freetier

import (
	"context"
	"time"
)

func SeedCatalog(ctx context.Context, store *CatalogStore) error {
	now := time.Now()

	entries := []struct {
		entry   CatalogEntry
		sources []Source
	}{
		{
			entry: CatalogEntry{
				ID: "ft_groq", ProviderName: "Groq", Category: "ai_inference",
				OfficialWebsite: "https://groq.com", DocsURL: "https://console.groq.com/docs",
				PricingURL: "https://groq.com/pricing", FreeTierStatus: "free_tier_available",
				FreeQuota: "14400 requests/day, 6000 tokens/min", ResetInterval: "daily",
				SupportedRegions: "global", SignupSteps: "email_signup",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://console.groq.com/docs/rate-limits", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_cerebras", ProviderName: "Cerebras", Category: "ai_inference",
				OfficialWebsite: "https://cerebras.ai", DocsURL: "https://inference-docs.cerebras.ai",
				PricingURL: "https://cerebras.ai/pricing", FreeTierStatus: "free_tier_available",
				FreeQuota: "rate limited free tier", ResetInterval: "per_minute",
				SupportedRegions: "global", SignupSteps: "email_signup",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://inference-docs.cerebras.ai/api-reference", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_cloudflare", ProviderName: "Cloudflare Workers AI", Category: "ai_inference",
				OfficialWebsite: "https://ai.cloudflare.com", DocsURL: "https://developers.cloudflare.com/workers-ai",
				PricingURL:     "https://developers.cloudflare.com/workers-ai/platform/pricing",
				FreeTierStatus: "free_tier_available",
				FreeQuota:      "10000 neurons/day", ResetInterval: "daily",
				SupportedRegions: "global", SignupSteps: "cloudflare_account",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://developers.cloudflare.com/workers-ai/platform/pricing/", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_nvidia", ProviderName: "NVIDIA NIM", Category: "ai_inference",
				OfficialWebsite: "https://build.nvidia.com", DocsURL: "https://docs.api.nvidia.com",
				PricingURL: "https://build.nvidia.com/explore", FreeTierStatus: "free_tier_available",
				FreeQuota: "1000 API credits free", ResetInterval: "one_time",
				SupportedRegions: "global", SignupSteps: "nvidia_account",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceMedium, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://build.nvidia.com", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_huggingface", ProviderName: "HuggingFace Inference", Category: "ai_inference",
				OfficialWebsite: "https://huggingface.co", DocsURL: "https://huggingface.co/docs/api-inference",
				PricingURL: "https://huggingface.co/pricing", FreeTierStatus: "free_tier_available",
				FreeQuota: "rate limited free inference", ResetInterval: "per_minute",
				SupportedRegions: "global", SignupSteps: "huggingface_account",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://huggingface.co/docs/api-inference/rate-limits", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_ollama", ProviderName: "Ollama", Category: "ai_inference",
				OfficialWebsite: "https://ollama.com", DocsURL: "https://github.com/ollama/ollama/blob/main/docs/api.md",
				FreeTierStatus: "self_hosted_free", FreeQuota: "unlimited (self-hosted)",
				SupportedRegions: "local", SignupSteps: "install_binary",
				AuthMethod: "none", PersonalKeyRequired: false, SandboxSupport: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://github.com/ollama/ollama", SourceType: "github", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_openrouter", ProviderName: "OpenRouter", Category: "ai_inference",
				OfficialWebsite: "https://openrouter.ai", DocsURL: "https://openrouter.ai/docs",
				PricingURL: "https://openrouter.ai/models", FreeTierStatus: "free_models_available",
				FreeQuota: "free models at $0/request, rate limited", ResetInterval: "per_minute",
				SupportedRegions: "global", SignupSteps: "email_signup",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://openrouter.ai/docs/limits", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_searxng", ProviderName: "SearXNG", Category: "search",
				OfficialWebsite: "https://searxng.org", DocsURL: "https://docs.searxng.org",
				FreeTierStatus: "self_hosted_free", FreeQuota: "unlimited (self-hosted)",
				SupportedRegions: "local", SignupSteps: "install_docker",
				AuthMethod: "none", PersonalKeyRequired: false, SandboxSupport: true,
				Confidence: ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://github.com/searxng/searxng", SourceType: "github", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_edgetts", ProviderName: "Edge TTS", Category: "tts",
				OfficialWebsite: "https://azure.microsoft.com/en-us/products/ai-services/text-to-speech",
				DocsURL:         "https://github.com/rany2/edge-tts", FreeTierStatus: "free_public_api",
				FreeQuota: "rate limited", SupportedRegions: "global",
				AuthMethod: "none", PersonalKeyRequired: false,
				Confidence: ConfidenceMedium, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://github.com/rany2/edge-tts", SourceType: "github", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_chutes", ProviderName: "Chutes AI", Category: "ai_inference",
				OfficialWebsite: "https://chutes.ai", DocsURL: "https://docs.chutes.ai",
				FreeTierStatus: "free_tier_available", FreeQuota: "limited free credits",
				SupportedRegions: "global", SignupSteps: "email_signup",
				AuthMethod: "api_key", PersonalKeyRequired: true,
				Confidence: ConfidenceMedium, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://chutes.ai", SourceType: "documentation", VerifiedAt: &now}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_airforce", ProviderName: "API Airforce", Category: "ai_inference",
				OfficialWebsite: "https://api.airforce", DocsURL: "https://api.airforce",
				FreeTierStatus: "free_public_api", FreeQuota: "rate limited",
				SupportedRegions: "global", AuthMethod: "none", PersonalKeyRequired: false,
				Confidence: ConfidenceLow, Status: StatusUnverified,
			},
			sources: []Source{{SourceURL: "https://api.airforce", SourceType: "website"}},
		},
		{
			entry: CatalogEntry{
				ID: "ft_coqui", ProviderName: "Coqui TTS", Category: "tts",
				OfficialWebsite: "https://github.com/coqui-ai/TTS",
				DocsURL:         "https://tts.readthedocs.io", FreeTierStatus: "self_hosted_free",
				FreeQuota: "unlimited (self-hosted)", SupportedRegions: "local",
				SignupSteps: "install_package", AuthMethod: "none", PersonalKeyRequired: false,
				SandboxSupport: true,
				Confidence:     ConfidenceHigh, Status: StatusVerified, VerifiedAt: &now,
			},
			sources: []Source{{SourceURL: "https://github.com/coqui-ai/TTS", SourceType: "github", VerifiedAt: &now}},
		},
	}

	for _, item := range entries {
		var count int
		err := store.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM free_tier_catalog WHERE id = ?", item.entry.ID).Scan(&count)
		if err != nil || count > 0 {
			continue
		}

		if err := store.CreateEntry(ctx, &item.entry); err != nil {
			continue
		}

		for _, src := range item.sources {
			src.CatalogID = item.entry.ID
			_ = store.AddSource(ctx, &src)
		}
	}

	return nil
}
