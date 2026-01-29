package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	id           uuid.UUID
	email        Email
	passwordHash PasswordHash
	createdAt    time.Time
}

func New(
	email Email,
	passwordHash PasswordHash,
) *User {
	return &User{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		createdAt:    time.Now().UTC(),
	}
}

func From(
	id uuid.UUID,
	email Email,
	passwordHash PasswordHash,
	createdAt time.Time,
) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}
}

func ForSession(
	id uuid.UUID,
	email Email,
) *User {
	return &User{
		id:    id,
		email: email,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
