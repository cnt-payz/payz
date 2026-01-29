package sessionredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cnt-payz/payz/sso-service/internal/domain/session"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	session.SessionRepository
	cfg    *config.Config
	log    *slog.Logger
	client *redis.Client
}

func New(cfg *config.Config, log *slog.Logger, client *redis.Client) (*SessionRepository, error) {
	if cfg == nil || log == nil {
		return nil, consts.ErrNilArgs
	}

	return &SessionRepository{
		cfg:    cfg,
		log:    log,
		client: client,
	}, nil
}

func (sr *SessionRepository) Set(ctx context.Context, refresh string, domain *session.Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	bytes, err := toModel(domain)
	if err != nil {
		return fmt.Errorf("failed to convert model: %w", err)
	}

	key := sr.getKey(refresh)
	pipe := sr.client.TxPipeline()

	sr.log.Debug("set main session")
	pipe.Set(ctx, key, bytes, sr.cfg.Secrets.Redis.SessionTTL)

	sr.log.Debug("add session into group")
	pipe.SAdd(ctx, sr.getSessionKey(domain.UserID()), refresh)

	sr.log.Debug("update ttl of all session")
	pipe.Expire(ctx, sr.getSessionKey(domain.UserID()), sr.cfg.Secrets.Redis.SessionTTL*2)

	sr.log.Debug("exec pipe")
	if _, err := pipe.Exec(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to exec redis pipeline")
	}

	return nil
}

func (sr *SessionRepository) Get(ctx context.Context, refresh string) (*session.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := sr.getKey(refresh)
	sr.log.Debug("fetching session by refresh key")
	bytes, err := sr.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, redis.Nil) {
			return nil, consts.ErrSessionDoesntExist
		}

		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return toDomain(bytes)
}

func (sr *SessionRepository) Del(ctx context.Context, refresh string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := sr.getKey(refresh)
	sr.log.Debug("start to delete session")
	if err := sr.client.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		if errors.Is(err, redis.Nil) {
			return consts.ErrSessionDoesntExist
		}

		return fmt.Errorf("failed to delete session from cache: %w", err)
	}

	return nil
}

func (sr *SessionRepository) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := sr.getSessionKey(userID)

	sr.log.Debug("start to logout all sessions")
	refreshTokens, err := sr.client.SMembers(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to get all sessions: %w", err)
	}

	oldKeys := make([]string, len(refreshTokens))
	for idx, key := range refreshTokens {
		oldKeys[idx] = sr.getKey(key)
	}

	sr.log.Debug("starting pipe")
	pipe := sr.client.TxPipeline()

	sr.log.Debug("delete all keys")
	if len(oldKeys) > 0 {
		pipe.Del(ctx, oldKeys...)
	}

	sr.log.Debug("delete session key")
	pipe.Del(ctx, key)

	sr.log.Debug("start pipe")
	if _, err := pipe.Exec(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to exec pipe: %w", err)
	}

	return nil
}

func (sr *SessionRepository) getKey(refresh string) string {
	return fmt.Sprintf("rtk:%s", refresh)
}

func (sr *SessionRepository) getSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("sessions:%s", userID.String())
}
