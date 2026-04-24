package trades

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for the trades domain.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new trades Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// ListForUser returns trades where the user is proposer or recipient, optionally filtered by since.
func (r *Repo) ListForUser(ctx context.Context, userID string, since *time.Time) ([]Trade, error) {
	query := `
		SELECT id, proposer_id, recipient_id, status, offered_stickers, requested_stickers,
		       proposed_at, resolved_at, completed_at, updated_at
		FROM trades
		WHERE proposer_id = $1 OR recipient_id = $1`

	args := []any{userID}
	if since != nil {
		query += ` AND updated_at > $2`
		args = append(args, *since)
	}
	query += ` ORDER BY proposed_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("trades: list for user: %w", err)
	}
	defer rows.Close()

	var results []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(
			&t.ID,
			&t.ProposerID,
			&t.RecipientID,
			&t.Status,
			&t.OfferedStickers,
			&t.RequestedStickers,
			&t.ProposedAt,
			&t.ResolvedAt,
			&t.CompletedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("trades: scan: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// GetByID retrieves a single trade by primary key.
// Returns nil, nil when not found.
func (r *Repo) GetByID(ctx context.Context, id string) (*Trade, error) {
	return nil, nil
}

// Create inserts a new trade proposal and returns the created record.
func (r *Repo) Create(ctx context.Context, proposerID string, req CreateTradeRequest) (*Trade, error) {
	return nil, nil
}

// UpdateStatus persists a new trade status.
func (r *Repo) UpdateStatus(ctx context.Context, tradeID string, status TradeStatus) (*Trade, error) {
	return nil, nil
}
