// Package collections manages each user's personal sticker collection.
package collections

import (
	"time"

	"github.com/google/uuid"
)

// UserCollection represents the state of a single sticker in a user's collection.
type UserCollection struct {
	UserID          uuid.UUID  `json:"userId"`
	StickerID       string     `json:"stickerId"`
	QuantityOwned   int        `json:"quantityOwned"`
	Wishlisted      bool       `json:"wishlisted"`
	Blacklisted     bool       `json:"blacklisted"`
	FirstAcquiredAt *time.Time `json:"firstAcquiredAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// UpsertCollectionRequest carries the fields that may be updated for a collection entry.
// All fields are pointers so the caller can partial-update (only send what changed).
type UpsertCollectionRequest struct {
	QuantityOwned *int  `json:"quantityOwned"`
	Wishlisted    *bool `json:"wishlisted"`
	Blacklisted   *bool `json:"blacklisted"`
}
