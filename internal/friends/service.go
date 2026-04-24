package friends

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mbuitragoc/panini-api/internal/notify"
)

// authTokenGetter is the subset of auth.Repo used by this service.
type authTokenGetter interface {
	GetDeviceToken(ctx context.Context, userID string) (string, error)
}

// Service implements business logic for the friends domain.
type Service struct {
	repo      *Repo
	notifySvc *notify.Service
	authRepo  authTokenGetter
}

// NewService creates a new friends Service.
func NewService(repo *Repo, notifySvc *notify.Service, authRepo authTokenGetter) *Service {
	return &Service{repo: repo, notifySvc: notifySvc, authRepo: authRepo}
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

	go s.notify(ctx, friendID, "Friend request", "Someone sent you a friend request")

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

// notify is a helper that retrieves the device token for userID and sends a push notification.
// Errors are logged and swallowed so callers are not affected.
func (s *Service) notify(ctx context.Context, userID, title, body string) {
	token, err := s.authRepo.GetDeviceToken(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "friends: notify: get device token", "userID", userID, "err", err)
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
		slog.WarnContext(ctx, "friends: notify: send", "userID", userID, "err", err)
	}
}
