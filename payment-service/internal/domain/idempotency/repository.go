package idempotency

import (
	"context"

	"github.com/google/uuid"
)

type IdempotencyRepository interface {
	Save(ctx context.Context, idempotencyKey string, id uuid.UUID) error
	Get(ctx context.Context, idempotencyKey string, id uuid.UUID) error
}
