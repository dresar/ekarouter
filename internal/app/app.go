package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dresar/ekarouter/internal/audit"
	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/credpool"
	"github.com/dresar/ekarouter/internal/db"
	"github.com/dresar/ekarouter/internal/executor"
	"github.com/dresar/ekarouter/internal/freetier"
	"github.com/dresar/ekarouter/internal/gateway"
	"github.com/dresar/ekarouter/internal/health"
	"github.com/dresar/ekarouter/internal/httpapi"
	"github.com/dresar/ekarouter/internal/limits"
	"github.com/dresar/ekarouter/internal/oauth"
	"github.com/dresar/ekarouter/internal/platform"
	"github.com/dresar/ekarouter/internal/providers"
	"github.com/dresar/ekarouter/internal/providers/airforce"
	"github.com/dresar/ekarouter/internal/providers/anthropic"
	"github.com/dresar/ekarouter/internal/providers/bazaarlink"
	"github.com/dresar/ekarouter/internal/providers/cerebras"
	"github.com/dresar/ekarouter/internal/providers/chutes"
	"github.com/dresar/ekarouter/internal/providers/cloudflare"
	"github.com/dresar/ekarouter/internal/providers/coqui"
	"github.com/dresar/ekarouter/internal/providers/custom"
	"github.com/dresar/ekarouter/internal/providers/devin"
	"github.com/dresar/ekarouter/internal/providers/edgetts"
	"github.com/dresar/ekarouter/internal/providers/gemini"
	"github.com/dresar/ekarouter/internal/providers/groq"
	"github.com/dresar/ekarouter/internal/providers/huggingface"
	"github.com/dresar/ekarouter/internal/providers/kilogateway"
	"github.com/dresar/ekarouter/internal/providers/kimchi"
	"github.com/dresar/ekarouter/internal/providers/kiro"
	"github.com/dresar/ekarouter/internal/providers/mimofree"
	"github.com/dresar/ekarouter/internal/providers/nvidia"
	"github.com/dresar/ekarouter/internal/providers/ollama"
	"github.com/dresar/ekarouter/internal/providers/openai"
	"github.com/dresar/ekarouter/internal/providers/opencode"
	"github.com/dresar/ekarouter/internal/providers/openrouter"
	"github.com/dresar/ekarouter/internal/providers/searxng"
	"github.com/dresar/ekarouter/internal/proxy"
	"github.com/dresar/ekarouter/internal/rbac"
	"github.com/dresar/ekarouter/internal/rotator"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/scheduler"
	"github.com/dresar/ekarouter/internal/tokensaver"
	"github.com/dresar/ekarouter/internal/usage"
	"github.com/dresar/ekarouter/internal/vault"
)

type Application struct {
	Config    *config.Config
	DB        *db.DB
	Server    *http.Server
	UsageRec  *usage.Recorder
	Scheduler *scheduler.Scheduler
	AuditLog  *audit.Logger
	Router    *routing.Router
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
	registry.Register("openai", openai.NewAdapter(sharedClient))
	registry.Register("codex", openai.NewCodexAdapter(sharedClient))
	registry.Register("anthropic", anthropic.NewAdapter(sharedClient))
	geminiAdapter := gemini.NewAdapter(sharedClient)
	registry.Register("gemini", geminiAdapter)
	registry.Register("gemini-cli", gemini.NewCLIAdapter(sharedClient))
	antigravityAdapter := gemini.NewAntigravityAdapter(sharedClient)
	registry.Register("antigravity", antigravityAdapter)
	registry.Register("gemini-agy", antigravityAdapter)
	customAdapter := custom.NewAdapter(sharedClient)
	registry.Register("custom", customAdapter)
	registry.Register("opencode", opencode.NewAdapter(sharedClient))
	registry.Register("oc", opencode.NewAdapter(sharedClient))
	registry.Register("mimo-free", mimofree.NewAdapter(sharedClient))
	registry.Register("mmf", mimofree.NewAdapter(sharedClient))
	registry.Register("devin", devin.NewAdapter(sharedClient))
	registry.Register("devin-cli", devin.NewAdapter(sharedClient))
	registry.Register("groq", groq.NewAdapter(sharedClient))
	registry.Register("cerebras", cerebras.NewAdapter(sharedClient))
	registry.Register("openrouter", openrouter.NewAdapter(sharedClient))
	registry.Register("cloudflare-ai", cloudflare.NewAdapter(sharedClient))
	registry.Register("cloudflare", cloudflare.NewAdapter(sharedClient))
	registry.Register("cf", cloudflare.NewAdapter(sharedClient))
	registry.Register("nvidia", nvidia.NewAdapter(sharedClient))
	registry.Register("api-airforce", airforce.NewAdapter(sharedClient))
	registry.Register("airforce", airforce.NewAdapter(sharedClient))
	registry.Register("af", airforce.NewAdapter(sharedClient))
	registry.Register("bazaarlink", bazaarlink.NewAdapter(sharedClient))
	registry.Register("bzl", bazaarlink.NewAdapter(sharedClient))
	registry.Register("kilo-gateway", kilogateway.NewAdapter(sharedClient))
	registry.Register("kgw", kilogateway.NewAdapter(sharedClient))
	registry.Register("kimchi", kimchi.NewAdapter(sharedClient))
	registry.Register("ollama", ollama.NewAdapter(sharedClient))
	registry.Register("chutes", chutes.NewAdapter(sharedClient))
	registry.Register("huggingface", huggingface.NewAdapter(sharedClient))
	registry.Register("hf", huggingface.NewAdapter(sharedClient))
	registry.Register("kiro", kiro.NewAdapter(sharedClient))
	registry.Register("searxng", searxng.NewAdapter(sharedClient))
	registry.Register("edge-tts", edgetts.NewAdapter(sharedClient))
	registry.Register("coqui", coqui.NewAdapter(sharedClient))
	registry.Register("deepseek", custom.NewBackendAdapter("deepseek", sharedClient))
	registry.Register("mistral", custom.NewBackendAdapter("mistral", sharedClient))
	registry.Register("together", custom.NewBackendAdapter("together", sharedClient))
	registry.Register("vllm", custom.NewBackendAdapter("vllm", sharedClient))
	registry.Register("localai", custom.NewBackendAdapter("localai", sharedClient))

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

		var client *http.Client
		if secretKey != "" {
			var meta map[string]any
			if err := json.Unmarshal([]byte(secretKey), &meta); err == nil {
				poolID, _ := meta["proxyPoolId"].(string)
				if poolID == "" {
					if psd, ok := meta["providerSpecificData"].(map[string]any); ok {
						poolID, _ = psd["proxyPoolId"].(string)
					}
				}
				if poolID != "" {
					var pName, pScheme, pHost, pUser, pPass sql.NullString
					var pPort, pEnabled int
					_ = database.QueryRowContext(ctx, "SELECT name, scheme, host, port, username, encrypted_password, enabled FROM proxy_profiles WHERE id = ?", poolID).Scan(
						&pName, &pScheme, &pHost, &pPort, &pUser, &pPass, &pEnabled)
					if pEnabled == 1 && pHost.Valid && pHost.String != "" {
						pass, _ := crypto.Decrypt(pPass.String)
						prof := &proxy.Profile{
							ID:       poolID,
							Name:     pName.String,
							Scheme:   pScheme.String,
							Host:     pHost.String,
							Port:     pPort,
							Username: pUser.String,
							Password: pass,
						}
						client, _ = proxyMgr.GetClient(prof, 60*time.Second)
					} else if pEnabled == 0 && pName.Valid && pName.String != "" {
						var fID, fName, fScheme, fHost, fUser, fPass sql.NullString
						var fPort int
						err := database.QueryRowContext(ctx, "SELECT id, name, scheme, host, port, username, encrypted_password FROM proxy_profiles WHERE enabled = 1 AND (name = ? OR name LIKE '%' || ? || '%') LIMIT 1", pName.String, pName.String).Scan(
							&fID, &fName, &fScheme, &fHost, &fPort, &fUser, &fPass)
						if err == nil && fHost.Valid && fHost.String != "" {
							pass, _ := crypto.Decrypt(fPass.String)
							prof := &proxy.Profile{
								ID:       fID.String,
								Name:     fName.String,
								Scheme:   fScheme.String,
								Host:     fHost.String,
								Port:     fPort,
								Username: fUser.String,
								Password: pass,
							}
							client, _ = proxyMgr.GetClient(prof, 60*time.Second)
						}
					}
				}
			}
		}

		return &providers.Credentials{
			APIKey:     apiKey,
			SecretKey:  secretKey,
			BaseURL:    baseURL,
			HTTPClient: client,
		}, nil
	}

	oauthMgr := oauth.NewManager()

	poolStore := credpool.NewStore(database.DB)
	poolEngine := credpool.NewEngine(poolStore)
	healthCheckerPool := credpool.NewHealthChecker(poolStore, registry, credResolver, 3)
	catalogStore := freetier.NewCatalogStore(database.DB)

	_ = freetier.SeedCatalog(context.Background(), catalogStore)

	platRegistry := platform.NewRegistry()
	platform.RegisterDefaultProviders(platRegistry)

	v, err := vault.NewVault(cfg.SecretKey)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("vault: %w", err)
	}
	vStore := vault.NewStore(database.DB, v)
	limEng := limits.NewEngine(database.DB)
	rot := rotator.NewRotator()
	exec := executor.NewExecutor(database.DB, platRegistry, vStore, cfg.AllowLocalProviders, rot)
	rbacSvc := rbac.NewService(database.DB)
	auditLog := audit.NewLogger(database.DB, 1000)
	platHandler := httpapi.NewPlatformHandler(database.DB, platRegistry, vStore, limEng, rot, exec, rbacSvc, auditLog, cfg.AllowLocalProviders)
	sched := scheduler.NewScheduler(database.DB, platRegistry, cfg.LogRetentionDays, 60*time.Second)

	gw := gateway.NewGateway(router, registry, ts, cd, usageRec, credResolver)
	httpServerHandler := httpapi.NewServer(cfg, database.DB, gw, crypto, usageRec, ts, checker, router, oauthMgr,
		poolStore, poolEngine, healthCheckerPool, catalogStore, platHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      httpServerHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Application{
		Config:    cfg,
		DB:        database,
		Server:    srv,
		UsageRec:  usageRec,
		Scheduler: sched,
		AuditLog:  auditLog,
		Router:    router,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	a.Scheduler.Start(ctx)
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
		a.Scheduler.Stop()
		a.AuditLog.Close()
		a.UsageRec.Close()
		_ = a.DB.Close()
		return nil

	case err := <-errChan:
		a.Scheduler.Stop()
		a.AuditLog.Close()
		a.UsageRec.Close()
		_ = a.DB.Close()
		return err
	}
}
