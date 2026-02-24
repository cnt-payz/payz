package domainrepo

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	walletdomain "github.com/cnt-payz/payz/crypto-service/internal/domain/wallet"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DatabaseRepo interface {
	SaveKey(ctx context.Context, tx *gorm.DB, walletID uuid.UUID, privateKey *ecdsa.PrivateKey) error
	GetKeys(ctx context.Context, walletID uuid.UUID) (string, *ecdsa.PrivateKey, error)
	CreateWallet(ctx context.Context, wallet *walletdomain.Wallet) (*walletdomain.Wallet, error)
	GetWalletByShopID(ctx context.Context, shopID uuid.UUID) (*walletdomain.Wallet, error)
	UpdateWalletFrozenAmountByShopID(ctx context.Context, frozenAmount uint64, shopID uuid.UUID) (*walletdomain.Wallet, error)
	SignTransaction(ctx context.Context, fromAddress string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	Close() error
}
