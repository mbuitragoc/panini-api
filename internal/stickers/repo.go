package stickers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for the stickers catalog.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new stickers Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// GetByID retrieves a single sticker by its primary key.
// Returns nil, nil when no sticker is found.
func (r *Repo) GetByID(ctx context.Context, id string) (*Sticker, error) {
	return nil, nil
}

// ListByTeam retrieves all stickers for a given country code.
func (r *Repo) ListByTeam(ctx context.Context, countryCode string) ([]Sticker, error) {
	return nil, nil
}

// List retrieves the full sticker catalog.
func (r *Repo) List(ctx context.Context) ([]Sticker, error) {
	return nil, nil
}
