package trades

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	const query = `
		SELECT id, proposer_id, recipient_id, status, offered_stickers, requested_stickers,
		       proposed_at, resolved_at, completed_at, updated_at
		FROM trades
		WHERE id = $1`

	var t Trade
	err := r.db.QueryRow(ctx, query, id).Scan(
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
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("trades: get by id: %w", err)
	}
	return &t, nil
}

// Create inserts a new trade proposal and returns the created record.
func (r *Repo) Create(ctx context.Context, proposerID string, req CreateTradeRequest) (*Trade, error) {
	const query = `
		INSERT INTO trades (
			id, proposer_id, recipient_id, status,
			offered_stickers, requested_stickers,
			proposed_at, resolved_at, completed_at, updated_at
		) VALUES (
			$1, $2, $3, 'proposed',
			$4, $5,
			NOW(), NULL, NULL, NOW()
		)
		RETURNING id, proposer_id, recipient_id, status, offered_stickers, requested_stickers,
		          proposed_at, resolved_at, completed_at, updated_at`

	proposerUUID, err := uuid.Parse(proposerID)
	if err != nil {
		return nil, fmt.Errorf("trades: create: invalid proposer id: %w", err)
	}

	id := uuid.New()

	var t Trade
	err = r.db.QueryRow(ctx, query,
		id,
		proposerUUID,
		req.RecipientID,
		req.OfferedStickers,
		req.RequestedStickers,
	).Scan(
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
	)
	if err != nil {
		return nil, fmt.Errorf("trades: create: %w", err)
	}
	return &t, nil
}

// UpdateStatus persists a new trade status.
func (r *Repo) UpdateStatus(ctx context.Context, tradeID string, status TradeStatus) (*Trade, error) {
	// resolved_at is set when the recipient makes a decision (accept or decline).
	// completed_at is set when the trade is fully done.
	const query = `
		UPDATE trades
		SET
			status      = $2,
			updated_at  = NOW(),
			resolved_at = CASE
				WHEN $2 IN ('accepted', 'declined') THEN NOW()
				ELSE resolved_at
			END,
			completed_at = CASE
				WHEN $2 = 'completed' THEN NOW()
				ELSE completed_at
			END
		WHERE id = $1
		RETURNING id, proposer_id, recipient_id, status, offered_stickers, requested_stickers,
		          proposed_at, resolved_at, completed_at, updated_at`

	var t Trade
	err := r.db.QueryRow(ctx, query, tradeID, status).Scan(
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
	)
	if err != nil {
		return nil, fmt.Errorf("trades: update status: %w", err)
	}
	return &t, nil
}
