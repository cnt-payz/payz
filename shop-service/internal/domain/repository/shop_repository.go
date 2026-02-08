package domainrepo

import (
	"context"

	shopdomain "github.com/cnt-payz/payz/shop-service/internal/domain/shop"
	"github.com/google/uuid"
)

type DatabaseRepo interface {
	CreateShop(ctx context.Context, shop *shopdomain.Shop) (*shopdomain.Shop, error)
	GetShopByOwner(ctx context.Context, id, userID uuid.UUID) (*shopdomain.Shop, error)
	GetShopsByUserID(ctx context.Context, userID uuid.UUID) ([]*shopdomain.Shop, error)
	CountShopsByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateShopByID(ctx context.Context, id, userID uuid.UUID, name shopdomain.Name) (*shopdomain.Shop, error)
	DeleteShopByID(ctx context.Context, id, userID uuid.UUID) error
	Close() error
}

type CacheRepo interface {
	SetShop(ctx context.Context, shop *shopdomain.Shop) error
	GetShopByOwner(ctx context.Context, id, userID uuid.UUID) (*shopdomain.Shop, error)
	UpdateShop(ctx context.Context, shop *shopdomain.Shop) error
	DeleteShopByID(ctx context.Context, id uuid.UUID) error
	Close() error
}
