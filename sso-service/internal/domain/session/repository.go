package session

import (
	"context"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Set(context.Context, string, *Session) error
	Get(context.Context, string) (*Session, error)
	Del(context.Context, string) error
	LogoutAll(context.Context, uuid.UUID) error
}
