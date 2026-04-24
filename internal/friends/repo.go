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
// It returns both outgoing records (sentByMe=true) and incoming pending requests (sentByMe=false).
func (r *Repo) ListFriends(ctx context.Context, userID string, since *time.Time) ([]FriendSyncRecord, error) {
	sinceFilter := ""
	args := []any{userID}
	if since != nil {
		sinceFilter = ` AND f.updated_at > $2`
		args = append(args, *since)
	}

	query := `
		SELECT
		    f.friend_id,
		    COALESCE(u.username, '') AS friend_username,
		    COALESCE(u.handle, '')   AS friend_handle,
		    (SELECT COUNT(*) FROM user_collections uc WHERE uc.user_id = f.friend_id AND uc.quantity_owned > 0) AS friend_owned_count,
		    f.status,
		    TRUE AS sent_by_me,
		    f.updated_at
		FROM friendships f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1` + sinceFilter + `

		UNION ALL

		SELECT
		    f.user_id,
		    COALESCE(u.username, '') AS friend_username,
		    COALESCE(u.handle, '')   AS friend_handle,
		    (SELECT COUNT(*) FROM user_collections uc WHERE uc.user_id = f.user_id AND uc.quantity_owned > 0) AS friend_owned_count,
		    f.status,
		    FALSE AS sent_by_me,
		    f.updated_at
		FROM friendships f
		JOIN users u ON u.id = f.user_id
		WHERE f.friend_id = $1 AND f.status = 'pending'` + sinceFilter + `

		ORDER BY updated_at DESC`

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
			&f.SentByMe,
			&f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("friends: scan: %w", err)
		}
		results = append(results, f)
	}
	return results, rows.Err()
}

// CreateRequest inserts a pending friend request (sender's direction only).
func (r *Repo) CreateRequest(ctx context.Context, userID, friendID string) (*Friendship, error) {
	const q = `
		INSERT INTO friendships (user_id, friend_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'pending', updated_at = now()
		RETURNING user_id, friend_id, status, created_at, updated_at`
	var f Friendship
	err := r.db.QueryRow(ctx, q, userID, friendID).Scan(&f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("friends: create request: %w", err)
	}
	return &f, nil
}

// RespondToRequest updates the sender's outgoing record and, when accepted,
// inserts the recipient's mirrored row so both sides see accepted status.
func (r *Repo) RespondToRequest(ctx context.Context, senderID, recipientID string, accept bool) (*Friendship, error) {
	newStatus := FriendshipStatus("declined")
	if accept {
		newStatus = StatusAccepted
	}

	const q1 = `UPDATE friendships SET status = $1, updated_at = now() WHERE user_id = $2 AND friend_id = $3`
	if _, err := r.db.Exec(ctx, q1, newStatus, senderID, recipientID); err != nil {
		return nil, fmt.Errorf("friends: respond: %w", err)
	}

	if accept {
		const q2 = `
			INSERT INTO friendships (user_id, friend_id, status)
			VALUES ($1, $2, 'accepted')
			ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'accepted', updated_at = now()`
		if _, err := r.db.Exec(ctx, q2, recipientID, senderID); err != nil {
			return nil, fmt.Errorf("friends: respond accept: %w", err)
		}
	}

	return &Friendship{Status: newStatus}, nil
}

// GetFriendCollection returns collection entries for a user's friend.
// It first verifies an accepted friendship exists between requesterID and friendID.
func (r *Repo) GetFriendCollection(ctx context.Context, requesterID, friendID string) ([]FriendCollectionItem, error) {
	const checkQ = `
        SELECT COUNT(*) FROM friendships
        WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted'`
	var count int
	if err := r.db.QueryRow(ctx, checkQ, requesterID, friendID).Scan(&count); err != nil {
		return nil, fmt.Errorf("friends: collection: check friendship: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("friends: collection: not friends")
	}

	const q = `
        SELECT sticker_id, quantity_owned, wishlisted, blacklisted
        FROM user_collections
        WHERE user_id = $1 AND quantity_owned > 0
        ORDER BY sticker_id`
	rows, err := r.db.Query(ctx, q, friendID)
	if err != nil {
		return nil, fmt.Errorf("friends: collection: query: %w", err)
	}
	defer rows.Close()

	var items []FriendCollectionItem
	for rows.Next() {
		var item FriendCollectionItem
		if err := rows.Scan(&item.StickerID, &item.QuantityOwned, &item.Wishlisted, &item.Blacklisted); err != nil {
			return nil, fmt.Errorf("friends: collection: scan: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []FriendCollectionItem{}
	}
	return items, rows.Err()
}
