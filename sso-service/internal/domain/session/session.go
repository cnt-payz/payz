package session

import (
	"time"

	"github.com/google/uuid"
)

type CtxKey string

type Session struct {
	userID      uuid.UUID
	email       string
	fingerPrint FingerPrint
	createdAt   time.Time
}

func New(userID uuid.UUID, email string, fingerPrint FingerPrint) *Session {
	return &Session{
		userID:      userID,
		email:       email,
		fingerPrint: fingerPrint,
		createdAt:   time.Now().UTC(),
	}
}

func From(
	userID uuid.UUID,
	email string,
	fingerPrint FingerPrint,
	createdAt time.Time,
) *Session {
	return &Session{
		userID:      userID,
		email:       email,
		fingerPrint: fingerPrint,
		createdAt:   createdAt,
	}
}

func (s *Session) UserID() uuid.UUID {
	return s.userID
}

func (s *Session) Email() string {
	return s.email
}

func (s *Session) FingerPrint() FingerPrint {
	return s.fingerPrint
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}
