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

	gwHandler := NewGatewayHandler(gw, db)
	adminHandler := NewAdminHandler(db, cfg, crypto, usageRec, ts, router, oauthMgr, extra...)

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
		exec := executor.NewExecutor(db, reg, vStore, cfg.AllowLocalProviders)
		rbacSvc := rbac.NewService(db)
		auditLog := audit.NewLogger(db, 1000)
		platformHandler = NewPlatformHandler(db, reg, vStore, limEng, rot, exec, rbacSvc, auditLog)
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
		apiV1.Get("/environments", platformHandler.ListEnvironments)
		apiV1.Post("/environments", platformHandler.CreateEnvironment)

		apiV1.Get("/tools", platformHandler.ListTools)
		apiV1.Post("/tools", platformHandler.CreateTool)
		apiV1.Get("/tools/{id}", platformHandler.GetTool)
		apiV1.Post("/tools/{id}/execute", platformHandler.ExecuteTool)
		apiV1.Post("/tools/{id}/test", platformHandler.ExecuteTool)

		apiV1.Get("/usage", platformHandler.GetUsageSummary)
		apiV1.Get("/usage/summary", platformHandler.GetUsageSummary)
		apiV1.Get("/health", platformHandler.HealthSummary)
		apiV1.Get("/health/providers", platformHandler.HealthProviders)
		apiV1.Get("/health/credentials", platformHandler.HealthCredentials)

		apiV1.Get("/audit-logs", platformHandler.GetAuditLogs)
		apiV1.Get("/events", platformHandler.GetAuditLogs)

		apiV1.HandleFunc("/proxy/{provider}/*", platformHandler.Proxy)
	})

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", adminHandler.Login)

		api.Group(func(authApi chi.Router) {
			authApi.Use(SessionAuthMiddleware(db))
			authApi.Post("/auth/logout", adminHandler.Logout)
			authApi.Get("/auth/me", adminHandler.Me)

			authApi.Get("/providers", adminHandler.ListProviders)
			authApi.Post("/providers", adminHandler.CreateProvider)
			authApi.Delete("/providers/{id}", adminHandler.DeleteProvider)

			authApi.Get("/accounts", adminHandler.ListAccounts)
			authApi.Post("/accounts", adminHandler.CreateAccount)
			authApi.Delete("/accounts/{id}", adminHandler.DeleteAccount)
			authApi.Post("/accounts/oauth/start", adminHandler.OAuthStart)
			authApi.Post("/accounts/oauth/callback", adminHandler.OAuthCallback)

			authApi.Get("/credentials", adminHandler.ListCredentials)
			authApi.Get("/credentials/{id}", adminHandler.GetCredential)

			authApi.Get("/routes", adminHandler.ListRoutes)
			authApi.Post("/routes", adminHandler.CreateRoute)
			authApi.Delete("/routes/{id}", adminHandler.DeleteRoute)

			authApi.Get("/models", adminHandler.ListModelsAdmin)
			authApi.Post("/models", adminHandler.CreateModel)
			authApi.Delete("/models/{id}", adminHandler.DeleteModel)

			authApi.Get("/proxy-profiles", adminHandler.ListProxyProfiles)
			authApi.Post("/proxy-profiles", adminHandler.CreateProxyProfile)
			authApi.Delete("/proxy-profiles/{id}", adminHandler.DeleteProxyProfile)
			authApi.Post("/proxy-profiles/{id}/test", adminHandler.TestProxyProfile)

			authApi.Get("/proxies", adminHandler.ListProxyProfiles)
			authApi.Post("/proxies", adminHandler.CreateProxyProfile)
			authApi.Delete("/proxies/{id}", adminHandler.DeleteProxyProfile)
			authApi.Post("/proxies/{id}/test", adminHandler.TestProxyProfile)

			authApi.Get("/keys", adminHandler.ListApiKeys)
			authApi.Post("/keys", adminHandler.CreateApiKey)
			authApi.Delete("/keys/{id}", adminHandler.DeleteApiKey)

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
