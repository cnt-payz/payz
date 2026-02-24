package servergrpc

import (
	"context"
	"errors"
	"log/slog"

	cryptopb "github.com/cnt-payz/payz/crypto-service/api/crypto/v1"
	walletdto "github.com/cnt-payz/payz/crypto-service/internal/application/dto"
	walletusecase "github.com/cnt-payz/payz/crypto-service/internal/application/usecase"
	dbrepo "github.com/cnt-payz/payz/crypto-service/internal/infra/repository/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cryptoServiceServer) Deposit(ctx context.Context, req *cryptopb.DepositReq) (*cryptopb.ExTransaction, error) {
	tx, err := s.walletUsecase.Deposit(ctx, &walletdto.DepositReq{
		TransactionID: req.TransactionId,
		ShopID:        req.ShopId,
		Amount:        req.Amount,
		TakerAddress:  req.TakerAddress,
	})
	if err != nil {
		return nil, s.grpcError(err)
	}

	return &cryptopb.ExTransaction{
		TransactionId: tx.TransactionID,
		Amount:        tx.Amount,
		TargetAddress: tx.TargetAddress,
		TakerAddress:  tx.TakerAddress,
	}, nil
}

func (s *cryptoServiceServer) GetWalletByShopID(ctx context.Context, req *cryptopb.GetWalletByShopIDReq) (*cryptopb.Wallet, error) {
	wallet, err := s.walletUsecase.GetWalletByShopID(ctx, req.ShopId)
	if err != nil {
		return nil, s.grpcError(err)
	}

	return &cryptopb.Wallet{
		Id:           wallet.ID,
		ShopId:       wallet.ShopID,
		Address:      wallet.Address,
		FrozenAmount: wallet.FrozenAmount,
		Amount:       wallet.Amount,
	}, nil
}

func (s *cryptoServiceServer) Withdraw(ctx context.Context, req *cryptopb.WithdrawReq) (*emptypb.Empty, error) {
	if err := s.walletUsecase.Withdraw(ctx, &walletdto.WithdrawReq{
		ShopID:        req.ShopId,
		Amount:        req.Amount,
		TargetAddress: req.TargetAddress,
	}); err != nil {
		return nil, s.grpcError(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *cryptoServiceServer) grpcError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, walletusecase.ErrInvalidTakerAddress):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, walletusecase.ErrInvalidTargetAddress):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, walletusecase.ErrInvalidShopID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, walletusecase.ErrInvalidAmount):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, dbrepo.ErrWalletNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, walletusecase.ErrNotEnoughAmount):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, walletusecase.ErrAmountMustBeBigger):
		return status.Error(codes.Aborted, err.Error())
	default:
		s.log.Error("unexpected error", slog.String("err", err.Error()))
		return status.Error(codes.Internal, "internal server error")
	}
}
