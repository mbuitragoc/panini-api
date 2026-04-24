package trades

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service implements business logic for the trades domain.
type Service struct {
	repo *Repo
}

// NewService creates a new trades Service.
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ListForUser returns all trades involving the given user.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]Trade, error) {
	trades, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("trades: list for user: %w", err)
	}

	return trades, nil
}

// Create proposes a new trade.
func (s *Service) Create(ctx context.Context, proposerID string, req CreateTradeRequest) (*Trade, error) {
	trade, err := s.repo.Create(ctx, proposerID, req)
	if err != nil {
		return nil, fmt.Errorf("trades: create: %w", err)
	}

	return trade, nil
}

// Apply applies a trade event from a given actor and persists the new status.
func (s *Service) Apply(ctx context.Context, tradeID string, actorID uuid.UUID, event TradeEvent) (*Trade, error) {
	trade, err := s.repo.GetByID(ctx, tradeID)
	if err != nil {
		return nil, fmt.Errorf("trades: get trade: %w", err)
	}

	if trade == nil {
		return nil, fmt.Errorf("trades: trade %s not found", tradeID)
	}

	next, err := Transition(trade.Status, event, actorID, trade.ProposerID, trade.RecipientID)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateStatus(ctx, tradeID, next)
	if err != nil {
		return nil, fmt.Errorf("trades: update status: %w", err)
	}

	return updated, nil
}
