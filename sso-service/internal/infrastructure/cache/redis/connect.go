package redis

import (
	"context"
	"fmt"
	"net"

	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(
			cfg.Secrets.Redis.Host,
			fmt.Sprint(cfg.Secrets.Redis.Port),
		),
		DB:       cfg.Secrets.Redis.DB,
		Password: cfg.Secrets.Redis.Password,
		Username: cfg.Secrets.Redis.User,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func Close(client *redis.Client) error {
	if err := client.Conn().Close(); err != nil {
		return err
	}

	return nil
}
