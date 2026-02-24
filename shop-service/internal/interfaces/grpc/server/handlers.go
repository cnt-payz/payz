package servergrpc

import (
	"context"
	"errors"
	"log/slog"

	shoppb "github.com/cnt-payz/payz/shop-service/api/shop/v1"
	shopusecase "github.com/cnt-payz/payz/shop-service/internal/application/usecase/shop"
	shopdomain "github.com/cnt-payz/payz/shop-service/internal/domain/shop"
	dbrepo "github.com/cnt-payz/payz/shop-service/internal/infra/repository/database"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *shopServiceServer) CreateShop(ctx context.Context, req *shoppb.CreateShopReq) (*shoppb.CreateShopResp, error) {
	s.log.Debug("received request to create shop", slog.String("name", req.Name))

	shop, err := s.shopUsecase.CreateShop(ctx, req.Name)
	if err != nil {
		return nil, s.grpcError(err)
	}

	return &shoppb.CreateShopResp{Shop: s.ShopDomainToPBModel(shop)}, nil
}

func (s *shopServiceServer) GetShopByOwner(ctx context.Context, req *shoppb.GetShopByOwnerReq) (*shoppb.GetShopByOwnerResp, error) {
	s.log.Debug("received request to get shop by owner", slog.String("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shop id")
	}

	shop, err := s.shopUsecase.GetShopByOwner(ctx, id)
	if err != nil {
		return nil, s.grpcError(err)
	}

	return &shoppb.GetShopByOwnerResp{Shop: s.ShopDomainToPBModel(shop)}, nil
}

func (s *shopServiceServer) GetShopsByUserID(ctx context.Context, _ *emptypb.Empty) (*shoppb.GetShopsByUserIDResp, error) {
	s.log.Debug("received request to get shops by user_id")

	shops, err := s.shopUsecase.GetShopsByUserID(ctx)
	if err != nil {
		return nil, s.grpcError(err)
	}

	shopspb := make([]*shoppb.Shop, len(shops))
	for i, shop := range shops {
		shopspb[i] = &shoppb.Shop{
			Id:     shop.ID().String(),
			UserId: shop.UserID().String(),
			Name:   shop.Name().String(),
		}
	}

	return &shoppb.GetShopsByUserIDResp{
		Shops: shopspb,
	}, nil
}

func (s *shopServiceServer) UpdateShopByID(ctx context.Context, req *shoppb.UpdateShopByIDReq) (*shoppb.UpdateShopByIDResp, error) {
	s.log.Debug("received request to update shop by id", slog.String("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shop id")
	}

	shop, err := s.shopUsecase.UpdateShopByID(ctx, id, req.Name)
	if err != nil {
		return nil, s.grpcError(err)
	}

	return &shoppb.UpdateShopByIDResp{Shop: s.ShopDomainToPBModel(shop)}, nil
}

func (s *shopServiceServer) DeleteShopByID(ctx context.Context, req *shoppb.DeleteShopByIDReq) (*emptypb.Empty, error) {
	s.log.Debug("received request to delete shop by id", slog.String("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shop id")
	}

	return &emptypb.Empty{}, s.grpcError(s.shopUsecase.DeleteShopByID(ctx, id))
}

func (s *shopServiceServer) ShopDomainToPBModel(shop *shopdomain.Shop) *shoppb.Shop {
	return &shoppb.Shop{
		Id:     shop.ID().String(),
		UserId: shop.UserID().String(),
		Name:   shop.Name().String(),
	}
}

func (s *shopServiceServer) grpcError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, shopdomain.ErrEmptyShopName):
		return status.Error(codes.InvalidArgument, "shop name cannot be empty")

	case errors.Is(err, shopusecase.ErrInvalidUserID),
		errors.Is(err, shopusecase.ErrNoUserIDInContext):
		return status.Error(codes.Unauthenticated, "user is not authenticated")

	case errors.Is(err, shopusecase.ErrReachedMaxCountOfShops):
		return status.Error(codes.FailedPrecondition, "maximum number of shops reached")

	case errors.Is(err, dbrepo.ErrShopNotFound):
		return status.Error(codes.NotFound, "shop not found")

	default:
		s.log.Error("unexpected error", slog.String("err", err.Error()))
		return status.Error(codes.Internal, "internal server error")
	}
}
