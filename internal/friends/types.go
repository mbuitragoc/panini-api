// Package friends manages friend relationships and friend requests.
package friends

import (
	"time"

	"github.com/google/uuid"
)

// FriendshipStatus represents the state of a friendship record.
type FriendshipStatus string

const (
	StatusPending  FriendshipStatus = "pending"
	StatusAccepted FriendshipStatus = "accepted"
	StatusDeclined FriendshipStatus = "declined"
)

// Friendship represents a directional friendship relationship between two users.
type Friendship struct {
	UserID    uuid.UUID        `json:"userId"`
	FriendID  uuid.UUID        `json:"friendId"`
	Status    FriendshipStatus `json:"status"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// SendFriendRequestRequest is the payload for initiating a friend request.
type SendFriendRequestRequest struct {
	FriendID uuid.UUID `json:"friendId"`
}

// RespondFriendRequestRequest is the payload for accepting or declining a request.
type RespondFriendRequestRequest struct {
	Accept bool `json:"accept"`
}
