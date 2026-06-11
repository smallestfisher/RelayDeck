package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/config"
	"github.com/smallestfisher/relaydeck/backend/internal/http/admin"
	"github.com/smallestfisher/relaydeck/backend/internal/http/gateway"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
	"github.com/smallestfisher/relaydeck/backend/internal/store/postgres"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	memoryStore := store.NewMemoryStoreWithBootstrap(cfg.BootstrapOwnerEmail, cfg.BootstrapOwnerPassword)
	adminStore := store.NewAdminStore(memoryStore, memoryStore)
	if cfg.DatabaseURL != "" {
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
		adminStore = store.NewAdminStore(memoryStore, users)
	}
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
	gatewayHandler := gateway.New(memoryStore, upstream.NewClient(cfg.GatewayRequestTimeout), time.Now)
	adminHandler := admin.New(adminStore, sessions, time.Now)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.Handle("/v1/", gatewayHandler)
	mux.Handle("/api/admin/", adminHandler)
	return mux
}
