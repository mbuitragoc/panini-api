package friends

import (
	"context"
	"fmt"
)

// Service implements business logic for the friends domain.
type Service struct {
	repo *Repo
}

// NewService creates a new friends Service.
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ListFriends returns all accepted friends for a user.
func (s *Service) ListFriends(ctx context.Context, userID string) ([]FriendSyncRecord, error) {
	friends, err := s.repo.ListFriends(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("friends: list: %w", err)
	}

	return friends, nil
}

// SendRequest creates a pending friend request.
func (s *Service) SendRequest(ctx context.Context, userID, friendID string) (*Friendship, error) {
	f, err := s.repo.CreateRequest(ctx, userID, friendID)
	if err != nil {
		return nil, fmt.Errorf("friends: send request: %w", err)
	}

	return f, nil
}

// GetFriendCollection returns the sticker collection for a friend of the requester.
func (s *Service) GetFriendCollection(ctx context.Context, requesterID, friendID string) ([]FriendCollectionItem, error) {
	return s.repo.GetFriendCollection(ctx, requesterID, friendID)
}

// RespondToRequest accepts or declines an incoming friend request.
func (s *Service) RespondToRequest(ctx context.Context, requestID, recipientID string, accept bool) (*Friendship, error) {
	f, err := s.repo.RespondToRequest(ctx, requestID, recipientID, accept)
	if err != nil {
		return nil, fmt.Errorf("friends: respond to request: %w", err)
	}

	return f, nil
}
