package cacherepo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cnt-payz/payz/shop-service/config"
	domainrepo "github.com/cnt-payz/payz/shop-service/internal/domain/repository"
	"github.com/redis/go-redis/v9"
)

type CacheRepo struct {
	client *redis.Client
	log    *slog.Logger
}

func NewCacheRepo(cfg *config.Config, log *slog.Logger) (domainrepo.CacheRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &CacheRepo{
		client: client,
		log:    log,
	}, nil
}

func (cr *CacheRepo) Close() error {
	if err := cr.client.Close(); err != nil {
		return nil
	}

	cr.log.Info("cache connection closed")
	return nil
}
