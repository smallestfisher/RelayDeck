package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/config"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/http/admin"
	"github.com/smallestfisher/relaydeck/backend/internal/http/gateway"
	"github.com/smallestfisher/relaydeck/backend/internal/secretbox"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
	"github.com/smallestfisher/relaydeck/backend/internal/store/postgres"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	gatewayStore := store.NewStaticGatewayConfigStore(nil, nil, nil)
	if cfg.UpstreamSecretKey == "" {
		panic(errors.New("APP_UPSTREAM_SECRET_KEY is required"))
	}
	upstreamSecrets, err := secretbox.New([]byte(cfg.UpstreamSecretKey))
	if err != nil {
		panic(fmt.Errorf("create upstream secretbox: %w", err))
	}
	accountAdapters := upstream.DefaultAccountAdapterRegistry(nil, cfg.GatewayRequestTimeout)

	if cfg.DatabaseURL == "" {
		panic(errors.New("DATABASE_URL is required"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("open postgres: %w", err))
	}
	users := postgres.NewUserStore(db)
	if err := users.EnsureSchema(ctx); err != nil {
		panic(fmt.Errorf("ensure users schema: %w", err))
	}
	if err := users.BootstrapOwner(ctx, cfg.BootstrapOwnerEmail, cfg.BootstrapOwnerPassword); err != nil {
		panic(fmt.Errorf("bootstrap owner: %w", err))
	}
	upstreams := postgres.NewUpstreamStore(db)
	if err := upstreams.EnsureSchema(ctx); err != nil {
		panic(fmt.Errorf("ensure upstream schema: %w", err))
	}
	adminStore := store.NewAdminStore(users, gatewayStore, upstreams)

	var sessions auth.SessionStore = auth.NewMemorySessionStore(time.Now)
	if cfg.RedisURL != "" {
		redisSessions, err := auth.NewRedisSessionStore(cfg.RedisURL, time.Now)
		if err != nil {
			panic(fmt.Errorf("create redis session store: %w", err))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisSessions.Ping(ctx); err != nil {
			panic(fmt.Errorf("connect redis session store: %w", err))
		}
		sessions = redisSessions
	}
	gatewayHandler := gateway.New(noGatewayConfig{}, upstream.NewClient(cfg.GatewayRequestTimeout), time.Now)
	adminHandler := admin.NewWithDependencies(adminStore, sessions, time.Now, upstreamSecrets, accountAdapters)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.Handle("/v1/", gatewayHandler)
	mux.Handle("/api/admin/", adminHandler)
	return mux
}

type noGatewayConfig struct{}

func (noGatewayConfig) APIKeys() []domain.APIKey     { return nil }
func (noGatewayConfig) Models() []domain.Model       { return nil }
func (noGatewayConfig) Sites() []domain.UpstreamSite { return nil }
func (noGatewayConfig) Mappings() []domain.SiteModel { return nil }
