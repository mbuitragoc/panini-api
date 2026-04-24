package collections

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for user collections.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new collections Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// ListByUser retrieves all collection entries for a given user.
func (r *Repo) ListByUser(ctx context.Context, userID string) ([]UserCollection, error) {
	return nil, nil
}

// Upsert creates or updates a single collection entry.
func (r *Repo) Upsert(ctx context.Context, userID, stickerID string, req UpsertCollectionRequest) (*UserCollection, error) {
	return nil, nil
}
