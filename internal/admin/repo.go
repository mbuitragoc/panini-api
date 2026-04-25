package admin

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for admin operations.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new admin Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// ReportMissingRatings inserts missing-rating reports for the given sticker IDs.
// The UNIQUE constraint on (sticker_id, reported_by) makes this idempotent.
func (r *Repo) ReportMissingRatings(ctx context.Context, userID string, stickerIDs []string) error {
	if len(stickerIDs) == 0 {
		return nil
	}
	for _, id := range stickerIDs {
		_, err := r.db.Exec(ctx, `
			INSERT INTO missing_rating_reports (sticker_id, reported_by)
			VALUES ($1, $2)
			ON CONFLICT (sticker_id, reported_by) DO NOTHING`,
			id, userID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
