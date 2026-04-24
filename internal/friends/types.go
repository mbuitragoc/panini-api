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

// FriendSyncRecord is a friendship enriched with the friend's profile and collection count.
// Returned by the sync endpoint so clients populate Friendship models in one round-trip.
type FriendSyncRecord struct {
	FriendID         uuid.UUID        `json:"friendId"`
	FriendUsername   string           `json:"friendUsername"`
	FriendHandle     string           `json:"friendHandle"`
	FriendOwnedCount int              `json:"friendOwnedCount"`
	Status           FriendshipStatus `json:"status"`
	SentByMe         bool             `json:"sentByMe"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// SendFriendRequestRequest is the payload for initiating a friend request.
type SendFriendRequestRequest struct {
	FriendID uuid.UUID `json:"friendId"`
}

// RespondFriendRequestRequest is the payload for accepting or declining a request.
type RespondFriendRequestRequest struct {
	Accept bool `json:"accept"`
}

// FriendCollectionItem is a single sticker entry from a friend's collection.
type FriendCollectionItem struct {
	StickerID     string `json:"stickerID"`
	QuantityOwned int    `json:"quantityOwned"`
	Wishlisted    bool   `json:"wishlisted"`
	Blacklisted   bool   `json:"blacklisted"`
}
