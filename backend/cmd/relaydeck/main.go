package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/app"
	"github.com/smallestfisher/relaydeck/backend/internal/config"
)

func main() {
	cfg := config.Load()
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           app.New(cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("relaydeck backend listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("backend server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("backend shutdown failed", "error", err)
		os.Exit(1)
	}
}
