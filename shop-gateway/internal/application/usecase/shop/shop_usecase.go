package shopusecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	shoppb "github.com/cnt-payz/payz/shop-gateway/api/shop/v1"
	"github.com/cnt-payz/payz/shop-gateway/internal/application/dto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrEmptyID   = errors.New("empty id")
	ErrEmptyName = errors.New("empty name")
)

type ShopUsecase interface {
	CreateShop(ctx context.Context, name string) (*dto.Shop, error)
	GetShopByOwner(ctx context.Context, id string) (*dto.Shop, error)
	GetShopsByUserID(ctx context.Context) ([]*dto.Shop, error)
	UpdateShopByID(ctx context.Context, id string, name string) (*dto.Shop, error)
	DeleteShopByID(ctx context.Context, id string) error
}

type shopUsecase struct {
	client shoppb.ShopServiceClient
	log    *slog.Logger
}

func NewShopUsecase(client shoppb.ShopServiceClient, log *slog.Logger) ShopUsecase {
	return &shopUsecase{
		client: client,
		log:    log,
	}
}

func (su *shopUsecase) CreateShop(ctx context.Context, name string) (*dto.Shop, error) {
	su.log.Debug("creating shop", slog.String("shop_name", name))

	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyName
	}

	resp, err := su.client.CreateShop(ctx, &shoppb.CreateShopReq{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return &dto.Shop{
		ID:     resp.Shop.Id,
		UserID: resp.Shop.UserId,
		Name:   resp.Shop.Name,
	}, nil
}

func (su *shopUsecase) GetShopByOwner(ctx context.Context, id string) (*dto.Shop, error) {
	su.log.Debug("getting shop", slog.String("id", id))

	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyID
	}

	resp, err := su.client.GetShopByOwner(ctx, &shoppb.GetShopByOwnerReq{Id: id})
	if err != nil {
		return nil, err
	}

	return &dto.Shop{
		ID:     resp.Shop.Id,
		UserID: resp.Shop.UserId,
		Name:   resp.Shop.Name,
	}, nil
}

func (su *shopUsecase) GetShopsByUserID(ctx context.Context) ([]*dto.Shop, error) {
	su.log.Debug("getting shops")

	resp, err := su.client.GetShopsByUserID(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	shops := make([]*dto.Shop, len(resp.Shops))
	for i, shop := range resp.Shops {
		shops[i] = &dto.Shop{
			ID:     shop.Id,
			UserID: shop.UserId,
			Name:   shop.Name,
		}
	}

	return shops, nil
}

func (su *shopUsecase) UpdateShopByID(ctx context.Context, id string, name string) (*dto.Shop, error) {
	su.log.Debug("updating shop",
		slog.String("id", id),
		slog.String("name", name),
	)

	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyID
	}

	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyName
	}

	resp, err := su.client.UpdateShopByID(ctx, &shoppb.UpdateShopByIDReq{
		Id:   id,
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return &dto.Shop{
		ID:     resp.Shop.Id,
		UserID: resp.Shop.UserId,
		Name:   resp.Shop.Name,
	}, nil
}

func (su *shopUsecase) DeleteShopByID(ctx context.Context, id string) error {
	su.log.Debug("deleting shop", slog.String("id", id))

	if strings.TrimSpace(id) == "" {
		return ErrEmptyID
	}

	if _, err := su.client.DeleteShopByID(ctx, &shoppb.DeleteShopByIDReq{
		Id: id,
	}); err != nil {
		return err
	}

	return nil
}
