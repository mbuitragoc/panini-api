package friends

import (
	"context"
	"fmt"
	"time"

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

// ListFriends returns friendship records enriched with friend profile data,
// optionally filtered to rows updated after since.
func (r *Repo) ListFriends(ctx context.Context, userID string, since *time.Time) ([]FriendSyncRecord, error) {
	query := `
		SELECT f.friend_id,
		       COALESCE(u.username, '') AS friend_username,
		       COALESCE(u.handle, '')   AS friend_handle,
		       (SELECT COUNT(*) FROM user_collections uc
		        WHERE uc.user_id = f.friend_id AND uc.quantity_owned > 0) AS friend_owned_count,
		       f.status,
		       f.updated_at
		FROM friendships f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1`

	args := []any{userID}
	if since != nil {
		query += ` AND f.updated_at > $2`
		args = append(args, *since)
	}
	query += ` ORDER BY f.updated_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("friends: list: %w", err)
	}
	defer rows.Close()

	var results []FriendSyncRecord
	for rows.Next() {
		var f FriendSyncRecord
		if err := rows.Scan(
			&f.FriendID,
			&f.FriendUsername,
			&f.FriendHandle,
			&f.FriendOwnedCount,
			&f.Status,
			&f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("friends: scan: %w", err)
		}
		results = append(results, f)
	}
	return results, rows.Err()
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
