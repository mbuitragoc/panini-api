package trades

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mbuitragoc/panini-api/internal/notify"
)

// authTokenGetter is the subset of auth.Repo used by this service.
type authTokenGetter interface {
	GetDeviceToken(ctx context.Context, userID string) (string, error)
}

// Service implements business logic for the trades domain.
type Service struct {
	repo      *Repo
	notifySvc *notify.Service
	authRepo  authTokenGetter
}

// NewService creates a new trades Service.
func NewService(repo *Repo, notifySvc *notify.Service, authRepo authTokenGetter) *Service {
	return &Service{repo: repo, notifySvc: notifySvc, authRepo: authRepo}
}

// ListForUser returns all trades involving the given user.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]Trade, error) {
	trades, err := s.repo.ListForUser(ctx, userID, nil)
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

	proposerDisplayName := proposerID
	go s.notify(ctx, trade.RecipientID.String(), "Trade proposal received", proposerDisplayName+" wants to trade with you")

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

	proposerID := trade.ProposerID.String()
	recipientID := trade.RecipientID.String()

	switch next {
	case StatusAccepted:
		go s.notify(ctx, proposerID, "Trade accepted", "Your trade proposal was accepted")
	case StatusDeclined:
		go s.notify(ctx, proposerID, "Trade declined", "Your trade proposal was declined")
	case StatusProposerConfirmed:
		go s.notify(ctx, recipientID, "Confirm the exchange", "Confirm the exchange to complete the trade")
	case StatusCompleted:
		go s.notify(ctx, proposerID, "Trade completed!", "Trade completed!")
		go s.notify(ctx, recipientID, "Trade completed!", "Trade completed!")
	}

	return updated, nil
}

// notify is a helper that retrieves the device token for userID and sends a push notification.
// Errors are logged and swallowed so callers are not affected.
func (s *Service) notify(ctx context.Context, userID, title, body string) {
	token, err := s.authRepo.GetDeviceToken(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "trades: notify: get device token", "userID", userID, "err", err)
		return
	}
	if token == "" {
		return
	}
	if err := s.notifySvc.Send(ctx, notify.Notification{
		DeviceToken: token,
		Title:       title,
		Body:        body,
	}); err != nil {
		slog.WarnContext(ctx, "trades: notify: send", "userID", userID, "err", err)
	}
}
