// Package trades manages sticker trade proposals between users.
package trades

import (
	"time"

	"github.com/google/uuid"
)

// TradeStatus represents the current state of a trade proposal.
type TradeStatus string

const (
	StatusProposed           TradeStatus = "proposed"
	StatusAccepted           TradeStatus = "accepted"
	StatusDeclined           TradeStatus = "declined"
	StatusProposerConfirmed  TradeStatus = "proposer_confirmed"
	StatusCompleted          TradeStatus = "completed"
)

// TradeEvent represents an action that triggers a trade state transition.
type TradeEvent string

const (
	EventAccept  TradeEvent = "accept"
	EventDecline TradeEvent = "decline"
	EventConfirm TradeEvent = "confirm"
)

// Trade represents a sticker trade proposal between two users.
type Trade struct {
	ID                uuid.UUID   `json:"id"`
	ProposerID        uuid.UUID   `json:"proposerId"`
	RecipientID       uuid.UUID   `json:"recipientId"`
	Status            TradeStatus `json:"status"`
	OfferedStickers   []string    `json:"offeredStickers"`
	RequestedStickers []string    `json:"requestedStickers"`
	ProposedAt        time.Time   `json:"proposedAt"`
	ResolvedAt        *time.Time  `json:"resolvedAt,omitempty"`
	CompletedAt       *time.Time  `json:"completedAt,omitempty"`
	UpdatedAt         time.Time   `json:"updatedAt"`
}

// CreateTradeRequest is the payload for proposing a new trade.
type CreateTradeRequest struct {
	RecipientID       uuid.UUID `json:"recipientId"`
	OfferedStickers   []string  `json:"offeredStickers"`
	RequestedStickers []string  `json:"requestedStickers"`
}
