package idempotency

import (
	"context"
	"errors"
	"fmt"

	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/payment-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IdempotencyRepository struct {
	cfg    *config.Config
	client *redis.Client
}

func New(cfg *config.Config, client *redis.Client) *IdempotencyRepository {
	return &IdempotencyRepository{
		cfg:    cfg,
		client: client,
	}
}

func (ir *IdempotencyRepository) Save(ctx context.Context, idempotencyKey string, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := ir.key(idempotencyKey, id)
	if err := ir.client.Set(ctx, key, idempotencyKey, ir.cfg.Service.Idempotency.TTL).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to set idempotency key: %w", err)
	}

	return nil
}

func (ir *IdempotencyRepository) Get(ctx context.Context, idempotencyKey string, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := ir.key(idempotencyKey, id)
	if err := ir.client.Get(ctx, key).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, redis.Nil) {
			return consts.ErrIdempotencyNotFound
		}

		return fmt.Errorf("failed to get idempotency key: %w", err)
	}

	return nil
}

func (ir *IdempotencyRepository) key(idempotency string, id uuid.UUID) string {
	return fmt.Sprintf("ik:%s:%s", idempotency, id.String())
}
