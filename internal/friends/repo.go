package friends

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for the friends domain.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new friends Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// ListFriends returns accepted friendships for the given user.
func (r *Repo) ListFriends(ctx context.Context, userID string) ([]Friendship, error) {
	return nil, nil
}

// CreateRequest inserts a pending friend request.
func (r *Repo) CreateRequest(ctx context.Context, userID, friendID string) (*Friendship, error) {
	return nil, nil
}

// RespondToRequest updates the status of a friend request.
func (r *Repo) RespondToRequest(ctx context.Context, requestID, recipientID string, accept bool) (*Friendship, error) {
	return nil, nil
}

// GetFriendCollection returns collection entries for a user's friend.
func (r *Repo) GetFriendCollection(ctx context.Context, ownerUserID, friendID string) ([]interface{}, error) {
	return nil, nil
}
