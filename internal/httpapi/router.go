package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/limits"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	"github.com/dresar/ekarouter/internal/vault"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router chi.Router
}

func NewServer(
	cfg *config.Config,
	db *sql.DB,
	gw *gateway.Gateway,
	crypto *auth.CryptoService,
	usageRec *usage.Recorder,
	ts *tokensaver.TokenSaver,
	checker *health.Checker,
	router *routing.Router,
	oauthMgr *oauth.Manager,
	extra ...any,
) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(RequestIDMiddleware)
	r.Use(BodyLimitMiddleware(cfg.MaxRequestBodyBytes))
	r.Use(CORSMiddleware(cfg.CORSOrigins))

	r.Get("/health", checker.HealthHandler)
	r.Get("/ready", checker.ReadyHandler)
	r.Get("/live", checker.HealthHandler)
	r.Get("/version", checker.VersionHandler)

	gwHandler := NewGatewayHandler(gw, db)
	adminHandler := NewAdminHandler(db, cfg, crypto, usageRec, ts, router, oauthMgr, extra...)
	aiDocsHandler := NewAIDocsHandler(db, cfg, router)
	mcpHandler := NewMCPHandler(db, gw, ts, router, cfg)

	r.Get("/docs", aiDocsHandler.ServeHTMLDocs)
	r.Get("/docs/raw", aiDocsHandler.GetDocs)
	r.Get("/api/ai/docs", aiDocsHandler.GetDocs)
	r.Get("/api/ai/prompt", aiDocsHandler.GetPrompt)
	r.Get("/api/ai/skills", aiDocsHandler.GetSkills)
	r.Get("/api/ai/status", aiDocsHandler.GetStatus)
	r.Get("/api/ai/keys", aiDocsHandler.GetKeys)
	r.Post("/api/ai/keys", aiDocsHandler.CreateKey)
	r.Get("/api/ai/models", aiDocsHandler.GetModels)

	r.Get("/mcp", mcpHandler.HandleJSONRPC)
	r.Post("/mcp", mcpHandler.HandleJSONRPC)
	r.Get("/mcp/sse", mcpHandler.HandleSSE)
	r.Post("/mcp/messages", mcpHandler.HandleMessages)
	r.Get("/api/mcp", mcpHandler.HandleJSONRPC)
	r.Post("/api/mcp", mcpHandler.HandleJSONRPC)
	r.Get("/api/mcp/sse", mcpHandler.HandleSSE)
	r.Post("/api/mcp/messages", mcpHandler.HandleMessages)

	r.Route("/auth", func(authRouter chi.Router) {
		authRouter.Post("/login", adminHandler.Login)
		authRouter.Group(func(authGroup chi.Router) {
			authGroup.Use(SessionAuthMiddleware(db))
			authGroup.Post("/logout", adminHandler.Logout)
			authGroup.Get("/me", adminHandler.Me)
			authGroup.Get("/sessions", adminHandler.ListSessions)
			authGroup.Delete("/sessions/{id}", adminHandler.RevokeSession)
		})
	})

	var platformHandler *PlatformHandler
	for _, opt := range extra {
		if v, ok := opt.(*PlatformHandler); ok {
			platformHandler = v
			break
		}
	}
	if platformHandler == nil {
		reg := platform.NewRegistry()
		platform.RegisterDefaultProviders(reg)
		v, _ := vault.NewVault(cfg.SecretKey)
		vStore := vault.NewStore(db, v)
		limEng := limits.NewEngine(db)
		rot := rotator.NewRotator()
		exec := executor.NewExecutor(db, reg, vStore, cfg.AllowLocalProviders, rot)
		rbacSvc := rbac.NewService(db)
		auditLog := audit.NewLogger(db, 1000)
		platformHandler = NewPlatformHandler(db, reg, vStore, limEng, rot, exec, rbacSvc, auditLog, cfg.AllowLocalProviders)
	}

	r.Route("/v1", func(v1 chi.Router) {
		v1.Use(GatewayAuthMiddleware(db))
		v1.Get("/models", gwHandler.ListModels)
		v1.Post("/chat/completions", gwHandler.ChatCompletions)
		v1.Post("/responses", gwHandler.Responses)
	})

	r.Route("/api/v1", func(apiV1 chi.Router) {
		apiV1.Use(PlatformAuthMiddleware(db))

		apiV1.Get("/providers", platformHandler.ListProviders)
		apiV1.Post("/providers", platformHandler.CreateProvider)
		apiV1.Get("/providers/{id}", platformHandler.GetProvider)
		apiV1.Post("/providers/{id}/validate", platformHandler.ValidateProvider)
		apiV1.Post("/providers/{id}/health", platformHandler.HealthProvider)
		apiV1.Get("/providers/{id}/capabilities", platformHandler.GetProviderCapabilities)
		apiV1.Get("/providers/{id}/docs", platformHandler.GetProviderDocs)

		apiV1.Get("/credentials", platformHandler.ListCredentials)
		apiV1.Post("/credentials", platformHandler.CreateCredential)
		apiV1.Get("/credentials/{id}", platformHandler.GetCredential)
		apiV1.Patch("/credentials/{id}", platformHandler.UpdateCredential)
		apiV1.Delete("/credentials/{id}", platformHandler.DeleteCredential)
		apiV1.Post("/credentials/{id}/test", platformHandler.TestCredential)
		apiV1.Post("/credentials/{id}/validate", platformHandler.TestCredential)
		apiV1.Post("/credentials/{id}/enable", platformHandler.EnableCredential)
		apiV1.Post("/credentials/{id}/disable", platformHandler.DisableCredential)
		apiV1.Post("/credentials/{id}/rotate", platformHandler.RotateCredential)
		apiV1.Get("/credentials/{id}/usage", platformHandler.GetCredentialUsage)
		apiV1.Get("/credentials/{id}/health", platformHandler.GetCredentialHealth)
		apiV1.Get("/credentials/{id}/events", platformHandler.GetCredentialHealth)

		apiV1.Get("/projects", platformHandler.ListProjects)
		apiV1.Post("/projects", platformHandler.CreateProject)
		apiV1.Get("/projects/{id}", platformHandler.GetProject)
		apiV1.Patch("/projects/{id}", platformHandler.UpdateProject)
		apiV1.Delete("/projects/{id}", platformHandler.DeleteProject)

		apiV1.Get("/environments", platformHandler.ListEnvironments)
		apiV1.Post("/environments", platformHandler.CreateEnvironment)
		apiV1.Patch("/environments/{id}", platformHandler.UpdateEnvironment)
		apiV1.Delete("/environments/{id}", platformHandler.DeleteEnvironment)

		apiV1.Get("/tools", platformHandler.ListTools)
		apiV1.Post("/tools", platformHandler.CreateTool)
		apiV1.Get("/tools/{id}", platformHandler.GetTool)
		apiV1.Get("/tools/{id}/schema", platformHandler.GetToolSchema)
		apiV1.Post("/tools/{id}/execute", platformHandler.ExecuteTool)
		apiV1.Post("/tools/{id}/test", platformHandler.ExecuteTool)

		apiV1.Get("/request-templates", platformHandler.ListRequestTemplates)
		apiV1.Post("/request-templates", platformHandler.CreateRequestTemplate)
		apiV1.Get("/request-templates/{id}", platformHandler.GetRequestTemplate)
		apiV1.Patch("/request-templates/{id}", platformHandler.UpdateRequestTemplate)
		apiV1.Delete("/request-templates/{id}", platformHandler.DeleteRequestTemplate)
		apiV1.Post("/request-templates/{id}/execute", platformHandler.ExecuteRequestTemplate)

		apiV1.Get("/usage", platformHandler.GetUsageSummary)
		apiV1.Get("/usage/summary", platformHandler.GetUsageSummary)
		apiV1.Get("/usage/providers", platformHandler.GetUsageProviders)
		apiV1.Get("/usage/credentials", platformHandler.GetUsageCredentials)
		apiV1.Get("/usage/projects", platformHandler.GetUsageProjects)

		apiV1.Get("/health", platformHandler.HealthSummary)
		apiV1.Get("/health/providers", platformHandler.HealthProviders)
		apiV1.Get("/health/credentials", platformHandler.HealthCredentials)

		apiV1.Get("/audit-logs", platformHandler.GetAuditLogs)
		apiV1.Get("/events", platformHandler.GetAuditLogs)

		apiV1.Get("/webhooks", platformHandler.ListWebhooks)
		apiV1.Post("/webhooks", platformHandler.CreateWebhook)
		apiV1.Get("/webhooks/{id}", platformHandler.GetWebhook)
		apiV1.Delete("/webhooks/{id}", platformHandler.DeleteWebhook)
		apiV1.Post("/webhooks/{id}/test", platformHandler.TestWebhook)
		apiV1.Get("/webhooks/{id}/deliveries", platformHandler.GetWebhookDeliveries)

		apiV1.Get("/system/settings", adminHandler.GetSettings)
		apiV1.Patch("/system/settings", adminHandler.UpdateSetting)
		apiV1.Put("/system/settings", adminHandler.UpdateSetting)

		apiV1.Get("/proxy/routes", adminHandler.ListRoutes)
		apiV1.Post("/proxy/routes", adminHandler.CreateRoute)
		apiV1.Get("/proxy/routes/{id}", adminHandler.GetRoute)
		apiV1.Delete("/proxy/routes/{id}", adminHandler.DeleteRoute)

		apiV1.HandleFunc("/proxy/{provider}/*", platformHandler.Proxy)
	})

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", adminHandler.Login)
		api.Get("/accounts/oauth/callback", adminHandler.OAuthCallback)

		api.Group(func(authApi chi.Router) {
			authApi.Use(SessionAuthMiddleware(db))
			authApi.Post("/auth/logout", adminHandler.Logout)
			authApi.Get("/auth/me", adminHandler.Me)
			authApi.Get("/auth/sessions", adminHandler.ListSessions)
			authApi.Delete("/auth/sessions/{id}", adminHandler.RevokeSession)

			authApi.Get("/providers", adminHandler.ListProviders)
			authApi.Post("/providers", adminHandler.CreateProvider)
			authApi.Delete("/providers/{id}", adminHandler.DeleteProvider)

			authApi.Get("/accounts", adminHandler.ListAccounts)
			authApi.Post("/accounts", adminHandler.CreateAccount)
			authApi.Post("/accounts/batch-proxy", adminHandler.ApplyBatchProxy)
			authApi.Post("/accounts/test-all", adminHandler.TestAllAccounts)
			authApi.Post("/accounts/{id}/test", adminHandler.TestAccount)
			authApi.Put("/accounts/{id}", adminHandler.UpdateAccount)
			authApi.Patch("/accounts/{id}", adminHandler.UpdateAccount)
			authApi.Delete("/accounts/{id}", adminHandler.DeleteAccount)
			authApi.Post("/accounts/oauth/start", adminHandler.OAuthStart)
			authApi.Post("/accounts/oauth/callback", adminHandler.OAuthCallback)
			authApi.Get("/accounts/oauth/callback", adminHandler.OAuthCallback)

			authApi.Get("/quota", adminHandler.GetQuotas)
			authApi.Post("/quota/refresh", adminHandler.GetQuotas)
			authApi.Post("/quota/{id}/refresh", adminHandler.RefreshQuota)

			authApi.Get("/credentials", adminHandler.ListCredentials)
			authApi.Get("/credentials/{id}", adminHandler.GetCredential)

			authApi.Get("/routes", adminHandler.ListRoutes)
			authApi.Post("/routes", adminHandler.CreateRoute)
			authApi.Get("/routes/{id}", adminHandler.GetRoute)
			authApi.Delete("/routes/{id}", adminHandler.DeleteRoute)

			authApi.Get("/models", adminHandler.ListModelsAdmin)
			authApi.Post("/models", adminHandler.CreateModel)
			authApi.Post("/models/batch-toggle", adminHandler.BatchToggleModels)
			authApi.Put("/models/{id}", adminHandler.UpdateModel)
			authApi.Patch("/models/{id}", adminHandler.UpdateModel)
			authApi.Delete("/models/{id}", adminHandler.DeleteModel)

			authApi.Get("/proxy-profiles", adminHandler.ListProxyProfiles)
			authApi.Post("/proxy-profiles", adminHandler.CreateProxyProfile)
			authApi.Get("/proxy-profiles/{id}", adminHandler.GetProxyProfile)
			authApi.Put("/proxy-profiles/{id}", adminHandler.UpdateProxyProfile)
			authApi.Patch("/proxy-profiles/{id}", adminHandler.UpdateProxyProfile)
			authApi.Delete("/proxy-profiles/{id}", adminHandler.DeleteProxyProfile)
			authApi.Post("/proxy-profiles/{id}/test", adminHandler.TestProxyProfile)
			authApi.Post("/proxy-profiles/{id}/enable", adminHandler.EnableProxyProfile)
			authApi.Post("/proxy-profiles/{id}/disable", adminHandler.DisableProxyProfile)

			authApi.Get("/proxies", adminHandler.ListProxyProfiles)
			authApi.Post("/proxies", adminHandler.CreateProxyProfile)
			authApi.Get("/proxies/{id}", adminHandler.GetProxyProfile)
			authApi.Put("/proxies/{id}", adminHandler.UpdateProxyProfile)
			authApi.Patch("/proxies/{id}", adminHandler.UpdateProxyProfile)
			authApi.Delete("/proxies/{id}", adminHandler.DeleteProxyProfile)
			authApi.Post("/proxies/{id}/test", adminHandler.TestProxyProfile)
			authApi.Post("/proxies/{id}/enable", adminHandler.EnableProxyProfile)
			authApi.Post("/proxies/{id}/disable", adminHandler.DisableProxyProfile)

			authApi.Get("/keys", adminHandler.ListApiKeys)
			authApi.Post("/keys", adminHandler.CreateApiKey)
			authApi.Delete("/keys/{id}", adminHandler.DeleteApiKey)
			authApi.Post("/keys/health", adminHandler.CheckKeyHealth)

			authApi.Get("/settings", adminHandler.GetSettings)
			authApi.Put("/settings", adminHandler.UpdateSetting)

			authApi.Get("/usage", adminHandler.GetUsageSummary)
			authApi.Post("/tokensaver/preview", adminHandler.PreviewTokenSaver)

			authApi.Get("/backup", adminHandler.ListBackups)
			authApi.Post("/backup", adminHandler.CreateBackup)

			authApi.Route("/credential-pools", func(pools chi.Router) {
				pools.Get("/", adminHandler.ListPools)
				pools.Post("/", adminHandler.CreatePool)
				pools.Get("/{id}", adminHandler.GetPool)
				pools.Delete("/{id}", adminHandler.DeletePool)
				pools.Get("/{id}/credentials", adminHandler.ListPoolMembers)
				pools.Post("/{id}/credentials", adminHandler.AddPoolMember)
				pools.Delete("/{id}/credentials/{mid}", adminHandler.RemovePoolMember)
				pools.Get("/{id}/policy", adminHandler.GetPoolPolicy)
				pools.Put("/{id}/policy", adminHandler.UpdatePoolPolicy)
				pools.Post("/{id}/health", adminHandler.PoolHealthCheck)
				pools.Post("/{id}/rotate", adminHandler.PoolRotate)
				pools.Post("/{id}/pause", adminHandler.PausePool)
				pools.Post("/{id}/resume", adminHandler.ResumePool)
				pools.Get("/{id}/usage", adminHandler.PoolUsage)
			})

			authApi.Route("/free-tiers", func(ft chi.Router) {
				ft.Get("/", adminHandler.ListFreeTiers)
				ft.Get("/categories", adminHandler.ListFreeTierCategories)
				ft.Get("/verified", adminHandler.ListVerifiedFreeTiers)
				ft.Post("/refresh", adminHandler.RefreshFreeTiers)
				ft.Get("/{id}", adminHandler.GetFreeTier)
				ft.Get("/{id}/sources", adminHandler.GetFreeTierSources)
			})

			authApi.Route("/devtools", func(dt chi.Router) {
				dt.Get("/templates", adminHandler.ListTemplates)
				dt.Post("/templates", adminHandler.CreateTemplate)
				dt.Post("/import-openapi", adminHandler.ImportOpenAPI)
				dt.Post("/confirm-openapi", adminHandler.ConfirmOpenAPIImport)
				dt.Post("/generate-curl", adminHandler.GenerateCurl)
			})
		})
	})

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
