package trades

import (
	"context"

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

// ListForUser returns all trades where the user is proposer or recipient.
func (r *Repo) ListForUser(ctx context.Context, userID string) ([]Trade, error) {
	return nil, nil
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
