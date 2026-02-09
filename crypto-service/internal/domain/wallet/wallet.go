package walletdomain

import (
	"errors"

	"github.com/google/uuid"
)

type ETHSDK interface {
}

var (
	ErrInvalidAddress                = errors.New("invalid wallet address")
	ErrInvalidPrivateKey             = errors.New("invalid private key")
	ErrFrozenAnountMustBeNonNegative = errors.New("frozen amount must be non negative")
)

type Wallet struct {
	id           uuid.UUID
	shopID       uuid.UUID
	address      Address
	frozenAmount FrozenAmount
}

func NewWallet(shopID uuid.UUID) (*Wallet, error) {
	fa := NewFrozenAmount(0)

	return &Wallet{
		id:           uuid.New(),
		shopID:       shopID,
		frozenAmount: fa,
	}, nil
}

func NewWalletFromDB(ID uuid.UUID, shopID uuid.UUID, address string, frozenAmount uint64) (*Wallet, error) {
	a, err := NewAddress(address)
	if err != nil {
		return nil, err
	}

	fa := NewFrozenAmount(frozenAmount)

	return &Wallet{
		id:           ID,
		shopID:       shopID,
		address:      a,
		frozenAmount: fa,
	}, nil
}

func (w *Wallet) ID() uuid.UUID {
	return w.id
}

func (w *Wallet) ShopID() uuid.UUID {
	return w.shopID
}

func (w *Wallet) SetAddress(address string) error {
	a, err := NewAddress(address)
	if err != nil {
		return err
	}

	w.address = a

	return nil
}

func (w *Wallet) Address() Address {
	return w.address
}

func (w *Wallet) FrozenAmount() FrozenAmount {
	return w.frozenAmount
}
