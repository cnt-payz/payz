package walletdomain

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Address string
type PrivateKey string
type FrozenAmount uint64

func NewAddress(address string) (Address, error) {
	a := Address(address)
	if err := a.validate(); err != nil {
		return Address(""), err
	}

	return a, nil
}

func (a Address) validate() error {
	if !strings.HasPrefix(a.String(), "0x") {
		return ErrInvalidAddress
	}
	if !common.IsHexAddress(a.String()) {
		return ErrInvalidAddress
	}

	addr := common.HexToAddress(a.String())

	if addr == (common.Address{}) {
		return ErrInvalidAddress
	}

	return nil
}

func (a Address) String() string {
	return string(a)
}

func NewFrozenAmount(frozenAmount uint64) FrozenAmount {
	return FrozenAmount(frozenAmount)
}

func (fa FrozenAmount) String() string {
	return fmt.Sprintf("%d", fa)
}

func (fa FrozenAmount) Uint64() uint64 {
	return uint64(fa)
}
