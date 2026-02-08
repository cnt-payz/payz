package shopusecase

import (
	"context"
	"errors"
	"log/slog"

	domainrepo "github.com/cnt-payz/payz/shop-service/internal/domain/repository"
	shopdomain "github.com/cnt-payz/payz/shop-service/internal/domain/shop"
	userdomain "github.com/cnt-payz/payz/shop-service/internal/domain/user"
	cacherepo "github.com/cnt-payz/payz/shop-service/internal/infra/repository/cache"
	"github.com/google/uuid"
)

var (
	ErrNoUserIDInContext      = errors.New("no user_id in context")
	ErrInvalidUserID          = errors.New("invalid user_id")
	ErrReachedMaxCountOfShops = errors.New("reached max count of shops")
)

type ShopUsecase interface {
	CreateShop(ctx context.Context, name string) (*shopdomain.Shop, error)
	GetShopByOwner(ctx context.Context, id uuid.UUID) (*shopdomain.Shop, error)
	GetShopsByUserID(ctx context.Context) ([]*shopdomain.Shop, error)
	UpdateShopByID(ctx context.Context, shopID uuid.UUID, name string) (*shopdomain.Shop, error)
	DeleteShopByID(ctx context.Context, shopID uuid.UUID) error
}

type shopUsecase struct {
	dbRepo    domainrepo.DatabaseRepo
	cacheRepo domainrepo.CacheRepo
	log       *slog.Logger
}

func NewShopUsecase(dbRepo domainrepo.DatabaseRepo, cacheRepo domainrepo.CacheRepo, log *slog.Logger) ShopUsecase {
	return &shopUsecase{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		log:       log,
	}
}

func (su *shopUsecase) CreateShop(ctx context.Context, name string) (*shopdomain.Shop, error) {
	su.log.Debug("start creating shop", slog.String("name", name))

	userID, err := su.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	shopsCount, err := su.dbRepo.CountShopsByUserID(ctx, userID.UUID())
	if err != nil {
		return nil, err
	}

	if shopsCount >= 100 {
		return nil, ErrReachedMaxCountOfShops
	}

	shop, err := shopdomain.NewShop(userID.UUID(), name)
	if err != nil {
		return nil, err
	}

	return su.dbRepo.CreateShop(ctx, shop)
}

func (su *shopUsecase) GetShopByOwner(ctx context.Context, shopID uuid.UUID) (*shopdomain.Shop, error) {
	su.log.Debug("start getting shop by owner", slog.String("shop_id", shopID.String()))

	userID, err := su.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	shop, err := su.cacheRepo.GetShopByOwner(ctx, shopID, userID.UUID())
	if err != nil {
		if !errors.Is(err, cacherepo.ErrShopNotFound) {
			su.log.Error("failed to get shop from cache", slog.String("err", err.Error()))
		}

		shop, err = su.dbRepo.GetShopByOwner(ctx, shopID, userID.UUID())
		if err != nil {
			return nil, err
		}

		if err := su.cacheRepo.SetShop(ctx, shop); err != nil {
			su.log.Error("failed to set shop to cache", slog.String("err", err.Error()))
		}
	}

	return shop, nil
}

func (su *shopUsecase) GetShopsByUserID(ctx context.Context) ([]*shopdomain.Shop, error) {
	su.log.Debug("start getting shops by user_id")

	userID, err := su.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	return su.dbRepo.GetShopsByUserID(ctx, userID.UUID())
}

func (su *shopUsecase) UpdateShopByID(ctx context.Context, shopID uuid.UUID, name string) (*shopdomain.Shop, error) {
	su.log.Debug("start updating shop by id",
		slog.String("shop_id", shopID.String()),
		slog.String("name", name),
	)

	userID, err := su.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	n, err := shopdomain.NewName(name)
	if err != nil {
		return nil, err
	}

	shop, err := su.dbRepo.UpdateShopByID(ctx, shopID, userID.UUID(), n)
	if err != nil {
		return nil, err
	}

	if err := su.cacheRepo.UpdateShop(ctx, shop); err != nil {
		if !errors.Is(err, cacherepo.ErrShopNotFound) {
			su.log.Error("failed to update shop in cache", slog.String("err", err.Error()))
		}
	}

	return shop, nil
}

func (su *shopUsecase) DeleteShopByID(ctx context.Context, shopID uuid.UUID) error {
	su.log.Debug("starting deleting shop by id", slog.String("shop_id", shopID.String()))

	userID, err := su.getUserID(ctx)
	if err != nil {
		return err
	}

	if err := su.cacheRepo.DeleteShopByID(ctx, shopID); err != nil {
		if !errors.Is(err, cacherepo.ErrShopNotFound) {
			su.log.Error("failed to delete shop from cache", slog.String("err", err.Error()))
		}
	}

	return su.dbRepo.DeleteShopByID(ctx, shopID, userID.UUID())
}

func (su *shopUsecase) getUserID(ctx context.Context) (userdomain.UserID, error) {
	userIDRaw := ctx.Value("user_id")
	if userIDRaw == nil {
		return userdomain.UserID{}, ErrNoUserIDInContext
	}

	userIDString, ok := userIDRaw.(string)
	if !ok {
		return userdomain.UserID{}, ErrInvalidUserID
	}

	userID, err := userdomain.NewUserID(userIDString)
	if err != nil {
		return userdomain.UserID{}, err
	}

	return userID, nil
}
