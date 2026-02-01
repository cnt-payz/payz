package session

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager interface {
	Validate(string) (*PayzClaims, error)
}

type PayzClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type CtxKey string
