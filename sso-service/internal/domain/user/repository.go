package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(context.Context, *User) (*User, error)
	Delete(context.Context, uuid.UUID) error
	GetByEmail(context.Context, Email) (*User, error)
	GetByID(context.Context, uuid.UUID) (*User, error)
}

type UserCache interface {
	SetByEmail(context.Context, *User) error
	GetByEmail(context.Context, Email) (*User, error)
	SetByID(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
}
