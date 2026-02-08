package cacherepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	shopdomain "github.com/cnt-payz/payz/shop-service/internal/domain/shop"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrShopNotFound = errors.New("shop not found")
)

type shop struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

func (cr *CacheRepo) SetShop(ctx context.Context, shop *shopdomain.Shop) error {
	cr.log.Debug("saving shop to cache",
		slog.String("id", shop.ID().String()),
		slog.String("user_id", shop.UserID().String()),
		slog.String("shop_name", shop.Name().String()),
	)

	bytes, err := json.Marshal(cr.ShopDomainToCacheModel(shop))
	if err != nil {
		return fmt.Errorf("failed to marshal shop: %v", err)
	}

	if err := cr.client.Set(ctx, fmt.Sprintf("shop:%s", shop.ID().String()), bytes, 5*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (cr *CacheRepo) GetShopByOwner(ctx context.Context, id, userID uuid.UUID) (*shopdomain.Shop, error) {
	cr.log.Debug("getting shop from cache", slog.String("id", id.String()))

	bytes, err := cr.client.Get(ctx, fmt.Sprintf("shop:%s", id.String())).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrShopNotFound
		}

		return nil, err
	}

	var shop shop
	if err := json.Unmarshal([]byte(bytes), &shop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shop: %v", err)
	}

	if shop.UserID != userID {
		return nil, ErrShopNotFound
	}

	return cr.ShopCacheModelToDomain(&shop)
}

func (cr *CacheRepo) UpdateShop(ctx context.Context, shop *shopdomain.Shop) error {
	cr.log.Debug("updating shop in cache",
		slog.String("id", shop.ID().String()),
		slog.String("user_id", shop.UserID().String()),
		slog.String("shop_name", shop.Name().String()),
	)

	bytes, err := json.Marshal(cr.ShopDomainToCacheModel(shop))
	if err != nil {
		return fmt.Errorf("failed to marshal shop: %v", err)
	}

	if err := cr.client.Set(ctx, fmt.Sprintf("shop:%s", shop.ID().String()), bytes, redis.KeepTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (cr *CacheRepo) DeleteShopByID(ctx context.Context, id uuid.UUID) error {
	cr.log.Debug("deleting shop from cache", slog.String("id", id.String()))

	return cr.client.Del(ctx, fmt.Sprintf("shop:%s", id.String())).Err()
}

func (cr *CacheRepo) ShopDomainToCacheModel(s *shopdomain.Shop) *shop {
	return &shop{
		ID:     s.ID(),
		UserID: s.UserID(),
		Name:   s.Name().String(),
	}
}

func (cr *CacheRepo) ShopCacheModelToDomain(s *shop) (*shopdomain.Shop, error) {
	return shopdomain.ShopFromRepo(s.ID, s.UserID, s.Name)
}
