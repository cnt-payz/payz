package walletusecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	paymentpb "github.com/cnt-payz/payz/crypto-service/api/payment/v1"
	walletdto "github.com/cnt-payz/payz/crypto-service/internal/application/dto"
	domainrepo "github.com/cnt-payz/payz/crypto-service/internal/domain/repository"
	walletdomain "github.com/cnt-payz/payz/crypto-service/internal/domain/wallet"
	dbrepo "github.com/cnt-payz/payz/crypto-service/internal/infra/repository/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrInvalidTakerAddress  = errors.New("invalid taker_address")
	ErrInvalidTargetAddress = errors.New("invalid target_address")
	ErrInvalidShopID        = errors.New("invalid shop_id")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrNotEnoughAmount      = errors.New("not enough amount")
	ErrAmountMustBeBigger   = errors.New("amount must be bigger")
)

type ETHSDK interface {
	ConfirmDeposit(ctx context.Context, shopAddress, takerAddress string, amount uint64) error
	GetUSDTAmount(ctx context.Context, address string) (uint64, error)
	Withdraw(ctx context.Context, fromAddress, toAddress string, amount uint64) error
}

type WalletUsecase interface {
	Deposit(ctx context.Context, req *walletdto.DepositReq) (*walletdto.ExTransaction, error)
	GetWalletByShopID(ctx context.Context, shopID string) (*walletdto.Wallet, error)
	Withdraw(ctx context.Context, req *walletdto.WithdrawReq) error
	Wait()
}

type walletUsecase struct {
	dbRepo               domainrepo.DatabaseRepo
	ethSDK               ETHSDK
	paymentServiceClient paymentpb.PaymentClient
	paymentPrivateKey    string
	log                  *slog.Logger
	wg                   sync.WaitGroup
}

func NewWalletUsecase(dbRepo domainrepo.DatabaseRepo, ethSDK ETHSDK, paymentServiceClient paymentpb.PaymentClient, paymentPrivateKey string, log *slog.Logger) WalletUsecase {
	return &walletUsecase{
		dbRepo:               dbRepo,
		ethSDK:               ethSDK,
		paymentServiceClient: paymentServiceClient,
		paymentPrivateKey:    paymentPrivateKey,
		log:                  log,
		wg:                   sync.WaitGroup{},
	}
}

func (wu *walletUsecase) Deposit(ctx context.Context, req *walletdto.DepositReq) (*walletdto.ExTransaction, error) {
	if ok := common.IsHexAddress(req.TakerAddress); !ok {
		return nil, ErrInvalidTakerAddress
	}
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, ErrInvalidShopID
	}
	amount, err := strconv.ParseUint(req.Amount, 10, 64)
	if err != nil {
		return nil, ErrInvalidAmount
	}
	if amount < 1e6 {
		return nil, ErrAmountMustBeBigger
	}

	wallet, err := wu.dbRepo.GetWalletByShopID(ctx, shopID)
	if err != nil {
		if errors.Is(err, dbrepo.ErrWalletNotFound) {
			w, err := walletdomain.NewWallet(shopID)
			if err != nil {
				return nil, err
			}

			wallet, err = wu.dbRepo.CreateWallet(ctx, w)
			if err != nil {
				return nil, err
			}
		}
	}

	wu.wg.Go(func() {
		confirmContext, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		wu.ConfirmDeposit(confirmContext, req.TransactionID, req.ShopID, wallet.Address().String(), req.TakerAddress, amount)
	})

	return &walletdto.ExTransaction{
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		TargetAddress: wallet.Address().String(),
		TakerAddress:  req.TakerAddress,
	}, nil
}

func (wu *walletUsecase) GetWalletByShopID(ctx context.Context, shopID string) (*walletdto.Wallet, error) {
	sID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidShopID
	}

	wallet, err := wu.dbRepo.GetWalletByShopID(ctx, sID)
	if err != nil {
		return nil, err
	}

	amount, err := wu.ethSDK.GetUSDTAmount(ctx, wallet.Address().String())
	if err != nil {
		return nil, err
	}

	return &walletdto.Wallet{
		ID:           wallet.ID().String(),
		ShopID:       wallet.ShopID().String(),
		Address:      wallet.Address().String(),
		FrozenAmount: wallet.FrozenAmount().String(),
		Amount:       fmt.Sprintf("%d", amount),
	}, nil
}

func (wu *walletUsecase) Withdraw(ctx context.Context, req *walletdto.WithdrawReq) error {
	if ok := common.IsHexAddress(req.TargetAddress); !ok {
		return ErrInvalidTargetAddress
	}
	sID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return ErrInvalidShopID
	}
	amount, err := strconv.ParseUint(req.Amount, 10, 64)
	if err != nil {
		return ErrInvalidAmount
	}

	if amount < 1e6 {
		return ErrAmountMustBeBigger
	}

	wallet, err := wu.dbRepo.GetWalletByShopID(ctx, sID)
	if err != nil {
		return err
	}

	totalAmount, err := wu.ethSDK.GetUSDTAmount(ctx, wallet.Address().String())
	if err != nil {
		return err
	}

	if totalAmount-wallet.FrozenAmount().Uint64() < amount {
		return ErrNotEnoughAmount
	}

	return wu.ethSDK.Withdraw(ctx, wallet.Address().String(), req.TargetAddress, amount)
}

type ConfirmDepositReq struct {
	TXID         string
	ShopAddress  string
	TakerAddress string
	Amount       uint64
}

func (wu *walletUsecase) ConfirmDeposit(
	ctx context.Context,
	txID, shopID,
	shopAddress, takerAddress string,
	amount uint64,
) {
	md := metadata.New(map[string]string{
		"idempotency-key": txID,
	})

	if err := wu.ethSDK.ConfirmDeposit(ctx, shopAddress, takerAddress, amount); err != nil {
		wu.log.Debug("failed to confirm deposit", slog.String("err", err.Error()))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for range 3 {
			timestamp := time.Now().UTC()
			_, err := wu.paymentServiceClient.CancelExTransaction(metadata.NewOutgoingContext(ctx, md), &paymentpb.ActionExRequest{
				TransactionId: txID,
				Signature:     wu.generateSignature(timestamp, txID),
				Timestamp:     timestamppb.New(timestamp),
			})
			if err != nil {
				wu.log.Error("failed to call cancel transaction to payment-service", slog.String("err", err.Error()))
				continue
			}

			return
		}

		return
	}

	for range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		timestamp := time.Now().UTC()
		_, err := wu.paymentServiceClient.ConfirmExTransaction(metadata.NewOutgoingContext(ctx, md), &paymentpb.ActionExRequest{
			TransactionId: txID,
			Signature:     wu.generateSignature(timestamp, txID),
			Timestamp:     timestamppb.New(timestamp),
		})
		if err != nil {
			wu.log.Error("failed to call confirm transaction to payment-service", slog.String("err", err.Error()))
			continue
		}

		return
	}
}

func (wu *walletUsecase) generateSignature(timestamp time.Time, txID string) string {
	hash := sha256.Sum256(fmt.Appendf(nil, "%d:%s:%s", timestamp.UnixMilli(), txID, wu.paymentPrivateKey))
	return hex.EncodeToString(hash[:])
}

func (wu *walletUsecase) Wait() {
	wu.wg.Wait()
}
