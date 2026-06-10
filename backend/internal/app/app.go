package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/config"
	"github.com/smallestfisher/relaydeck/backend/internal/http/admin"
	"github.com/smallestfisher/relaydeck/backend/internal/http/gateway"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	memoryStore := store.NewMemoryStore()
	gatewayHandler := gateway.New(memoryStore, upstream.NewClient(cfg.GatewayRequestTimeout), time.Now)
	adminHandler := admin.New(memoryStore)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.Handle("/v1/", gatewayHandler)
	mux.Handle("/api/admin/", adminHandler)
	return mux
}
