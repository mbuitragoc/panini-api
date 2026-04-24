// Command api starts the panini-api HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mbuitragoc/panini-api/internal/app"
	"github.com/mbuitragoc/panini-api/internal/config"
	apphttp "github.com/mbuitragoc/panini-api/internal/http"
)

func main() {
	// Load .env in development — silently ignored when the file is absent.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialise app", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	router := apphttp.NewRouter(application)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine so we can listen for shutdown signals.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr, "env", cfg.Env)
		serverErr <- srv.ListenAndServe()
	}()

	// Block until we receive a termination signal or the server errors out.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("server error", "error", err)
		os.Exit(1)

	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	// Give in-flight requests up to 15 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
