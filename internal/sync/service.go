package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/mbuitragoc/panini-api/internal/collections"
	"github.com/mbuitragoc/panini-api/internal/friends"
	"github.com/mbuitragoc/panini-api/internal/trades"
)

// Service assembles a SyncResponse from multiple downstream services.
type Service struct {
	collectionsRepo *collections.Repo
	tradesRepo      *trades.Repo
	friendsRepo     *friends.Repo
}

// NewService creates a new sync Service.
func NewService(
	collectionsRepo *collections.Repo,
	tradesRepo *trades.Repo,
	friendsRepo *friends.Repo,
) *Service {
	return &Service{
		collectionsRepo: collectionsRepo,
		tradesRepo:      tradesRepo,
		friendsRepo:     friendsRepo,
	}
}

// Sync fetches all relevant data for a user and returns it as a single SyncResponse.
// When since is non-nil only records updated after that timestamp are included.
func (s *Service) Sync(ctx context.Context, userID string, since *time.Time) (*SyncResponse, error) {
	userCollections, err := s.collectionsRepo.ListByUser(ctx, userID, since)
	if err != nil {
		return nil, fmt.Errorf("sync: collections: %w", err)
	}

	userTrades, err := s.tradesRepo.ListForUser(ctx, userID, since)
	if err != nil {
		return nil, fmt.Errorf("sync: trades: %w", err)
	}

	userFriendships, err := s.friendsRepo.ListFriends(ctx, userID, since)
	if err != nil {
		return nil, fmt.Errorf("sync: friendships: %w", err)
	}

	if userCollections == nil {
		userCollections = []collections.UserCollection{}
	}

	if userTrades == nil {
		userTrades = []trades.Trade{}
	}

	if userFriendships == nil {
		userFriendships = []friends.FriendSyncRecord{}
	}

	return &SyncResponse{
		Collections: userCollections,
		Trades:      userTrades,
		Friendships: userFriendships,
		SyncedAt:    time.Now().UTC(),
	}, nil
}
