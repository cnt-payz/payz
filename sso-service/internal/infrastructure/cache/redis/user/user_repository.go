package userredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	user.UserCache
	client *redis.Client
	cfg    *config.Config
	log    *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, client *redis.Client) *UserCache {
	return &UserCache{
		client: client,
		cfg:    cfg,
		log:    log,
	}
}

func (uc *UserCache) SetByEmail(ctx context.Context, user *user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	uc.log.Debug("conver model into bytes", slog.String("email", user.Email().Value()))
	bytes, err := toModel(user)
	if err != nil {
		return fmt.Errorf("failed to convert model: %w", err)
	}

	uc.log.Debug("saving user into cache")
	if err := uc.client.Set(ctx, uc.getEmailKey(user.Email().Value()), bytes, uc.cfg.Secrets.Redis.UserTTL).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to save user into cache: %w", err)
	}

	return nil
}

func (uc *UserCache) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := uc.getEmailKey(email.Value())

	uc.log.Debug("fetching user", slog.String("email", email.Value()))
	bytes, err := uc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, redis.Nil) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to fetch user from cache: %w", err)
	}

	return toDomain(bytes)
}

func (uc *UserCache) SetByID(ctx context.Context, user *user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	uc.log.Debug("conver model into bytes", slog.String("user_id", user.ID().String()))
	bytes, err := toModel(user)
	if err != nil {
		return fmt.Errorf("failed to convert model: %w", err)
	}

	uc.log.Debug("saving user into cache")
	if err := uc.client.Set(ctx, uc.getIDKey(user.ID()), bytes, uc.cfg.Secrets.Redis.UserTTL).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to save user into cache: %w", err)
	}

	return nil
}

func (uc *UserCache) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := uc.getIDKey(id)

	uc.log.Debug("fetching user", slog.String("user_id", id.String()))
	bytes, err := uc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, redis.Nil) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to fetch user from cache: %w", err)
	}

	return toDomain(bytes)
}

func (uc *UserCache) getEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func (uc *UserCache) getIDKey(id uuid.UUID) string {
	return fmt.Sprintf("user:id:%s", id.String())
}
