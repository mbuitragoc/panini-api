// Package app wires together the application's shared dependencies.
package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbuitragoc/panini-api/internal/config"
)

// App holds the application's shared infrastructure dependencies.
type App struct {
	DB     *pgxpool.Pool
	Config *config.Config
}

// New creates a new App, opens a pgxpool connection, and verifies connectivity.
// The caller must call Close() when the App is no longer needed.
func New(cfg *config.Config) (*App, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("app: open pgxpool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("app: ping database: %w", err)
	}

	return &App{
		DB:     pool,
		Config: cfg,
	}, nil
}

// Close releases all resources held by the App.
func (a *App) Close() {
	a.DB.Close()
}
