package collections

import (
	"context"
	"fmt"
)

// Service implements business logic for the collections domain.
type Service struct {
	repo *Repo
}

// NewService creates a new collections Service.
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ListForUser returns all collection entries belonging to a user.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]UserCollection, error) {
	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("collections: list for user: %w", err)
	}

	return items, nil
}

// Upsert creates or updates a collection entry for a sticker.
func (s *Service) Upsert(ctx context.Context, userID, stickerID string, req UpsertCollectionRequest) (*UserCollection, error) {
	item, err := s.repo.Upsert(ctx, userID, stickerID, req)
	if err != nil {
		return nil, fmt.Errorf("collections: upsert: %w", err)
	}

	return item, nil
}
