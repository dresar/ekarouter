package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/httpapi"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/proxy"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
)

type Application struct {
	Config   *config.Config
	DB       *db.DB
	Server   *http.Server
	UsageRec *usage.Recorder
}

func Setup(cfg *config.Config, migrationsDir string) (*Application, error) {
	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := database.Migrate(migrationsDir); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	crypto, err := auth.NewCryptoService(cfg.SecretKey)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("crypto service: %w", err)
	}

	usageRec := usage.NewRecorder(database.DB, 1000)
	ts := tokensaver.New(cfg.TokenSaverMode)
	checker := health.NewChecker(database.DB)
	proxyMgr := proxy.NewManager(cfg.AllowLocalProviders)

	sharedClient, err := proxyMgr.GetClient(nil, 60*time.Second)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("proxy manager: %w", err)
	}

	registry := providers.NewRegistry()
	registry.Register("openai", providers.NewOpenAIAdapter(sharedClient))
	registry.Register("anthropic", providers.NewAnthropicAdapter(sharedClient))
	registry.Register("gemini", providers.NewGeminiAdapter(sharedClient))
	customAdapter := providers.NewCustomAdapter(sharedClient)
	registry.Register("custom", customAdapter)
	registry.Register("deepseek", customAdapter)
	registry.Register("groq", customAdapter)
	registry.Register("openrouter", customAdapter)
	registry.Register("ollama", customAdapter)
	registry.Register("mistral", customAdapter)
	registry.Register("together", customAdapter)

	cd := routing.NewCooldownManager()
	router := routing.NewRouter(cd)

	if err := router.LoadFromDB(context.Background(), database.DB); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("load routes: %w", err)
	}

	credResolver := func(ctx context.Context, accountID string) (*providers.Credentials, error) {
		if accountID == "" {
			return nil, errors.New("empty account id")
		}

		var encAccess, encSecret sql.NullString
		var baseURL string
		err := database.QueryRowContext(ctx, `
SELECT c.encrypted_access, c.encrypted_secret, p.base_url
FROM credentials c
JOIN accounts a ON a.id = c.account_id
JOIN providers p ON p.id = a.provider_id
WHERE c.account_id = ?`, accountID).Scan(&encAccess, &encSecret, &baseURL)

		if err != nil {
			return nil, fmt.Errorf("resolve credentials for %s: %w", accountID, err)
		}

		apiKey, _ := crypto.Decrypt(encAccess.String)
		secretKey, _ := crypto.Decrypt(encSecret.String)

		return &providers.Credentials{
			APIKey:    apiKey,
			SecretKey: secretKey,
			BaseURL:   baseURL,
		}, nil
	}

	oauthMgr := oauth.NewManager()

	gw := gateway.NewGateway(router, registry, ts, cd, usageRec, credResolver)
	httpServerHandler := httpapi.NewServer(cfg, database.DB, gw, crypto, usageRec, ts, checker, router, oauthMgr)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      httpServerHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Application{
		Config:   cfg,
		DB:       database,
		Server:   srv,
		UsageRec: usageRec,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = a.Server.Shutdown(shutdownCtx)
		a.UsageRec.Close()
		_ = a.DB.Close()
		return nil

	case err := <-errChan:
		a.UsageRec.Close()
		_ = a.DB.Close()
		return err
	}
}
