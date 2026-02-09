package dbrepo

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"math/big"

	walletdomain "github.com/cnt-payz/payz/crypto-service/internal/domain/wallet"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type wallet struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	ShopID       uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	FrozenAmount uint64    `gorm:"not null"`
}

var (
	ErrWalletNotFound = errors.New("wallet not found")
)

func (dr *DatabaseRepo) CreateWallet(ctx context.Context, wallet *walletdomain.Wallet) (*walletdomain.Wallet, error) {
	dr.log.Debug("creating wallet",
		slog.String("ID", wallet.ID().String()),
		slog.String("ShopID", wallet.ShopID().String()),
		slog.String("Address", wallet.Address().String()),
	)

	tx := dr.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		tx.Rollback()
		return nil, errors.New("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	if err := wallet.SetAddress(address.Hex()); err != nil {
		tx.Rollback()
		return nil, err
	}

	w, _, err := dr.WalletDomainToDBModel(wallet)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to convert wallet domain to db model: %v", err)
	}

	if err := tx.WithContext(ctx).Create(&w).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	if err := dr.SaveKey(ctx, tx, w.ID, privateKey); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save key: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	dr.log.Debug("wallet created successfully",
		slog.String("ID", wallet.ID().String()),
		slog.String("Address", address.Hex()),
	)

	return wallet, nil
}

func (dr *DatabaseRepo) GetWalletByShopID(ctx context.Context, shopID uuid.UUID) (*walletdomain.Wallet, error) {
	dr.log.Debug("getting wallet by shop id", slog.String("ShopID", shopID.String()))

	var wallet wallet
	if err := dr.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}

		return nil, err
	}

	address, _, err := dr.GetKeys(ctx, wallet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys")
	}

	return dr.WalletDBModelToDomain(&wallet, address)
}

func (dr *DatabaseRepo) UpdateWalletFrozenAmountByShopID(ctx context.Context, frozenAmount uint64, shopID uuid.UUID) (*walletdomain.Wallet, error) {
	dr.log.Debug("updating wallet by shop id", slog.String("ShopID", shopID.String()))

	var w wallet
	if err := dr.db.WithContext(ctx).Model(&wallet{}).Clauses(clause.Returning{}).Where("shop_id = ?", shopID).Update("frozen_amount", frozenAmount).Scan(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}

		return nil, err
	}

	address, _, err := dr.GetKeys(ctx, w.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys")
	}

	return dr.WalletDBModelToDomain(&w, address)
}

func (dr *DatabaseRepo) SignTransaction(ctx context.Context, fromAddress string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	var key key
	if err := dr.db.WithContext(ctx).Where("address = ?", fromAddress).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to find keys for address %s: %v", fromAddress, err)
	}

	privateKey, err := crypto.HexToECDSA(key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex to private key: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return signedTx, nil
}

func (dr *DatabaseRepo) WalletDBModelToDomain(w *wallet, address string) (*walletdomain.Wallet, error) {
	wallet, err := walletdomain.NewWalletFromDB(w.ID, w.ShopID, address, w.FrozenAmount)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (dr *DatabaseRepo) WalletDomainToDBModel(w *walletdomain.Wallet) (*wallet, string, error) {
	return &wallet{
		ID:           w.ID(),
		ShopID:       w.ShopID(),
		FrozenAmount: w.FrozenAmount().Uint64(),
	}, w.Address().String(), nil
}
