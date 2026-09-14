package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
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
) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(RequestIDMiddleware)
	r.Use(BodyLimitMiddleware(cfg.MaxRequestBodyBytes))
	r.Use(CORSMiddleware(cfg.CORSOrigins))

	r.Get("/health", checker.HealthHandler)
	r.Get("/ready", checker.ReadyHandler)

	gwHandler := NewGatewayHandler(gw, db)
	adminHandler := NewAdminHandler(db, cfg, crypto, usageRec, ts, router, oauthMgr)

	r.Route("/v1", func(v1 chi.Router) {
		v1.Use(GatewayAuthMiddleware(db))
		v1.Get("/models", gwHandler.ListModels)
		v1.Post("/chat/completions", gwHandler.ChatCompletions)
		v1.Post("/responses", gwHandler.Responses)
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

			authApi.Get("/routes", adminHandler.ListRoutes)
			authApi.Post("/routes", adminHandler.CreateRoute)
			authApi.Delete("/routes/{id}", adminHandler.DeleteRoute)

			authApi.Get("/models", adminHandler.ListModelsAdmin)
			authApi.Post("/models", adminHandler.CreateModel)
			authApi.Delete("/models/{id}", adminHandler.DeleteModel)

			authApi.Get("/proxy-profiles", adminHandler.ListProxyProfiles)
			authApi.Post("/proxy-profiles", adminHandler.CreateProxyProfile)
			authApi.Delete("/proxy-profiles/{id}", adminHandler.DeleteProxyProfile)

			authApi.Get("/keys", adminHandler.ListApiKeys)
			authApi.Post("/keys", adminHandler.CreateApiKey)
			authApi.Delete("/keys/{id}", adminHandler.DeleteApiKey)

			authApi.Get("/settings", adminHandler.GetSettings)
			authApi.Put("/settings", adminHandler.UpdateSetting)

			authApi.Get("/usage", adminHandler.GetUsageSummary)
			authApi.Post("/tokensaver/preview", adminHandler.PreviewTokenSaver)
		})
	})

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
