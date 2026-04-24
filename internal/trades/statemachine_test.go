package trades

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestTransition(t *testing.T) {
	proposer := uuid.New()
	recipient := uuid.New()
	bystander := uuid.New()

	t.Run("valid transitions", func(t *testing.T) {
		tests := []struct {
			name    string
			current TradeStatus
			event   TradeEvent
			actor   uuid.UUID
			want    TradeStatus
		}{
			{
				name:    "recipient accepts proposed trade",
				current: StatusProposed,
				event:   EventAccept,
				actor:   recipient,
				want:    StatusAccepted,
			},
			{
				name:    "recipient declines proposed trade",
				current: StatusProposed,
				event:   EventDecline,
				actor:   recipient,
				want:    StatusDeclined,
			},
			{
				name:    "proposer confirms accepted trade",
				current: StatusAccepted,
				event:   EventConfirm,
				actor:   proposer,
				want:    StatusProposerConfirmed,
			},
			{
				name:    "recipient completes proposer-confirmed trade",
				current: StatusProposerConfirmed,
				event:   EventConfirm,
				actor:   recipient,
				want:    StatusCompleted,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := Transition(tc.current, tc.event, tc.actor, proposer, recipient)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("got status %q, want %q", got, tc.want)
				}
			})
		}
	})

	t.Run("invalid transitions", func(t *testing.T) {
		tests := []struct {
			name    string
			current TradeStatus
			event   TradeEvent
			actor   uuid.UUID
		}{
			{
				name:    "cannot confirm a proposed trade",
				current: StatusProposed,
				event:   EventConfirm,
				actor:   recipient,
			},
			{
				name:    "cannot accept an accepted trade",
				current: StatusAccepted,
				event:   EventAccept,
				actor:   recipient,
			},
			{
				name:    "cannot decline an accepted trade",
				current: StatusAccepted,
				event:   EventDecline,
				actor:   recipient,
			},
			{
				name:    "cannot accept a proposer_confirmed trade",
				current: StatusProposerConfirmed,
				event:   EventAccept,
				actor:   recipient,
			},
			{
				name:    "cannot decline a proposer_confirmed trade",
				current: StatusProposerConfirmed,
				event:   EventDecline,
				actor:   recipient,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := Transition(tc.current, tc.event, tc.actor, proposer, recipient)
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				if !errors.Is(err, ErrInvalidTransition) {
					t.Errorf("expected ErrInvalidTransition, got: %v", err)
				}
			})
		}
	})

	t.Run("wrong actor", func(t *testing.T) {
		tests := []struct {
			name    string
			current TradeStatus
			event   TradeEvent
			actor   uuid.UUID
		}{
			{
				name:    "proposer cannot accept their own trade",
				current: StatusProposed,
				event:   EventAccept,
				actor:   proposer,
			},
			{
				name:    "proposer cannot decline their own trade",
				current: StatusProposed,
				event:   EventDecline,
				actor:   proposer,
			},
			{
				name:    "bystander cannot accept a proposed trade",
				current: StatusProposed,
				event:   EventAccept,
				actor:   bystander,
			},
			{
				name:    "recipient cannot confirm an accepted trade",
				current: StatusAccepted,
				event:   EventConfirm,
				actor:   recipient,
			},
			{
				name:    "bystander cannot confirm an accepted trade",
				current: StatusAccepted,
				event:   EventConfirm,
				actor:   bystander,
			},
			{
				name:    "proposer cannot complete a proposer-confirmed trade",
				current: StatusProposerConfirmed,
				event:   EventConfirm,
				actor:   proposer,
			},
			{
				name:    "bystander cannot complete a proposer-confirmed trade",
				current: StatusProposerConfirmed,
				event:   EventConfirm,
				actor:   bystander,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := Transition(tc.current, tc.event, tc.actor, proposer, recipient)
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				if !errors.Is(err, ErrWrongActor) {
					t.Errorf("expected ErrWrongActor, got: %v", err)
				}
			})
		}
	})

	t.Run("terminal states reject all events", func(t *testing.T) {
		terminalStates := []TradeStatus{StatusCompleted, StatusDeclined}
		allEvents := []TradeEvent{EventAccept, EventDecline, EventConfirm}
		actors := []uuid.UUID{proposer, recipient, bystander}

		for _, state := range terminalStates {
			for _, event := range allEvents {
				for _, actor := range actors {
					name := string(state) + "/" + string(event)
					t.Run(name, func(t *testing.T) {
						_, err := Transition(state, event, actor, proposer, recipient)
						if err == nil {
							t.Fatal("expected an error but got nil")
						}
						if !errors.Is(err, ErrInvalidTransition) {
							t.Errorf("expected ErrInvalidTransition, got: %v", err)
						}
					})
				}
			}
		}
	})
}
