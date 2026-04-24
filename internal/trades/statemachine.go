package trades

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrInvalidTransition is returned when the requested event is not valid
// from the trade's current state.
var ErrInvalidTransition = errors.New("trades: invalid state transition")

// ErrWrongActor is returned when the actor attempting the transition does not
// have permission to trigger the event for this trade.
var ErrWrongActor = errors.New("trades: actor is not permitted to perform this action")

// Transition computes the next TradeStatus given the current status, the triggering
// event, and the identity of the acting user.
//
// Rules:
//   - Only the recipient may accept or decline a proposed trade.
//   - Only the proposer may confirm an accepted trade.
//   - Only the recipient may complete a proposer_confirmed trade (via confirm event).
//   - completed and declined are terminal: no further transitions are allowed.
//
// It is a pure function with no database I/O.
func Transition(
	current TradeStatus,
	event TradeEvent,
	actorID uuid.UUID,
	proposerID uuid.UUID,
	recipientID uuid.UUID,
) (TradeStatus, error) {
	// Terminal states reject all events before any actor check.
	if current == StatusCompleted || current == StatusDeclined {
		return "", fmt.Errorf("%w: trade is in terminal state %q", ErrInvalidTransition, current)
	}

	switch current {
	case StatusProposed:
		switch event {
		case EventAccept:
			if actorID != recipientID {
				return "", fmt.Errorf("%w: only the recipient may accept a proposed trade", ErrWrongActor)
			}
			return StatusAccepted, nil

		case EventDecline:
			if actorID != recipientID {
				return "", fmt.Errorf("%w: only the recipient may decline a proposed trade", ErrWrongActor)
			}
			return StatusDeclined, nil

		default:
			return "", fmt.Errorf("%w: event %q is not valid in state %q", ErrInvalidTransition, event, current)
		}

	case StatusAccepted:
		switch event {
		case EventConfirm:
			if actorID != proposerID {
				return "", fmt.Errorf("%w: only the proposer may confirm an accepted trade", ErrWrongActor)
			}
			return StatusProposerConfirmed, nil

		default:
			return "", fmt.Errorf("%w: event %q is not valid in state %q", ErrInvalidTransition, event, current)
		}

	case StatusProposerConfirmed:
		switch event {
		case EventConfirm:
			if actorID != recipientID {
				return "", fmt.Errorf("%w: only the recipient may complete a proposer-confirmed trade", ErrWrongActor)
			}
			return StatusCompleted, nil

		default:
			return "", fmt.Errorf("%w: event %q is not valid in state %q", ErrInvalidTransition, event, current)
		}

	default:
		return "", fmt.Errorf("%w: unknown state %q", ErrInvalidTransition, current)
	}
}
