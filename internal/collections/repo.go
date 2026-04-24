package collections

import (
	"context"
	"fmt"
	"time"

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

// ListByUser retrieves collection entries for a user, optionally filtered to rows updated after since.
func (r *Repo) ListByUser(ctx context.Context, userID string, since *time.Time) ([]UserCollection, error) {
	query := `
		SELECT user_id, sticker_id, quantity_owned, wishlisted, blacklisted, first_acquired_at, updated_at
		FROM user_collections
		WHERE user_id = $1`

	args := []any{userID}
	if since != nil {
		query += ` AND updated_at > $2`
		args = append(args, *since)
	}
	query += ` ORDER BY sticker_id`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("collections: list by user: %w", err)
	}
	defer rows.Close()

	var results []UserCollection
	for rows.Next() {
		var uc UserCollection
		if err := rows.Scan(
			&uc.UserID,
			&uc.StickerID,
			&uc.QuantityOwned,
			&uc.Wishlisted,
			&uc.Blacklisted,
			&uc.FirstAcquiredAt,
			&uc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("collections: scan: %w", err)
		}
		results = append(results, uc)
	}
	return results, rows.Err()
}

// Upsert creates or updates a single collection entry.
func (r *Repo) Upsert(ctx context.Context, userID, stickerID string, req UpsertCollectionRequest) (*UserCollection, error) {
	query := `
		INSERT INTO user_collections (user_id, sticker_id, quantity_owned, wishlisted, blacklisted, first_acquired_at, updated_at)
		VALUES ($1, $2,
			COALESCE($3, 0),
			COALESCE($4, false),
			COALESCE($5, false),
			CASE WHEN COALESCE($3, 0) > 0 THEN NOW() ELSE NULL END,
			NOW()
		)
		ON CONFLICT (user_id, sticker_id) DO UPDATE SET
			quantity_owned   = COALESCE($3, user_collections.quantity_owned),
			wishlisted       = COALESCE($4, user_collections.wishlisted),
			blacklisted      = COALESCE($5, user_collections.blacklisted),
			first_acquired_at = CASE
				WHEN COALESCE($3, 0) > 0 AND user_collections.first_acquired_at IS NULL
				THEN NOW()
				ELSE user_collections.first_acquired_at
			END,
			updated_at       = NOW()
		RETURNING user_id, sticker_id, quantity_owned, wishlisted, blacklisted, first_acquired_at, updated_at`

	var uc UserCollection
	err := r.db.QueryRow(ctx, query,
		userID, stickerID,
		req.QuantityOwned,
		req.Wishlisted,
		req.Blacklisted,
	).Scan(
		&uc.UserID,
		&uc.StickerID,
		&uc.QuantityOwned,
		&uc.Wishlisted,
		&uc.Blacklisted,
		&uc.FirstAcquiredAt,
		&uc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("collections: upsert: %w", err)
	}
	return &uc, nil
}
