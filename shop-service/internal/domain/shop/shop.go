package shopdomain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmptyShopName = errors.New("empty shop name")
)

type Shop struct {
	id     uuid.UUID
	userID uuid.UUID
	name   Name
}

func NewShop(userID uuid.UUID, name string) (*Shop, error) {
	n, err := NewName(name)
	if err != nil {
		return nil, err
	}

	return &Shop{
		id:     uuid.New(),
		userID: userID,
		name:   n,
	}, nil
}

func ShopFromRepo(id, userID uuid.UUID, name string) (*Shop, error) {
	n, err := NewName(name)
	if err != nil {
		return nil, err
	}

	return &Shop{
		id:     id,
		userID: userID,
		name:   n,
	}, nil
}

func (s Shop) ID() uuid.UUID {
	return s.id
}

func (s Shop) UserID() uuid.UUID {
	return s.userID
}

func (s Shop) Name() Name {
	return s.name
}
