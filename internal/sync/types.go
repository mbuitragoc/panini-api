// Package sync provides a bulk-pull endpoint so clients can hydrate all their data in one call.
package sync

import (
	"time"

	"github.com/mbuitragoc/panini-api/internal/collections"
	"github.com/mbuitragoc/panini-api/internal/friends"
	"github.com/mbuitragoc/panini-api/internal/trades"
)

// SyncResponse bundles a user's collections, trades, and friendships for a single round-trip.
type SyncResponse struct {
	Collections []collections.UserCollection `json:"collections"`
	Trades      []trades.Trade               `json:"trades"`
	Friendships []friends.FriendSyncRecord   `json:"friendships"`
	SyncedAt    time.Time                    `json:"syncedAt"`
}
