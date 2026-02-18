package storage

import (
	"context"

	"github.com/google/uuid"
)

type APIKeyRepositoryInterface interface {
	Create(ctx context.Context, key *APIKey) error
	FindByLookupHash(ctx context.Context, lookupHash string) (*APIKey, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)
	Update(ctx context.Context, key *APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
}

