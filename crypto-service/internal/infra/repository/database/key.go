package dbrepo

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// in prod private key must be encrypted
type key struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	WalletID   uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	Address    string    `gorm:"not null"`
	PrivateKey string    `gorm:"not null"`
}

func (dr *DatabaseRepo) SaveKey(ctx context.Context, tx *gorm.DB, walletID uuid.UUID, privateKey *ecdsa.PrivateKey) error {
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("failed to create wallet")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	key := key{
		ID:         uuid.New(),
		WalletID:   walletID,
		Address:    address,
		PrivateKey: privateKeyHex,
	}

	return tx.WithContext(ctx).Create(&key).Error
}

func (dr *DatabaseRepo) GetKeys(ctx context.Context, walletID uuid.UUID) (string, *ecdsa.PrivateKey, error) {
	var key key
	if err := dr.db.WithContext(ctx).Where("wallet_id = ?", walletID).First(&key).Error; err != nil {
		return "", nil, err
	}

	privateKeyBytes, err := hex.DecodeString(key.PrivateKey)
	if err != nil {
		return "", nil, err
	}

	pk, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return "", nil, err
	}

	return key.Address, pk, nil
}
