package session

import (
	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager interface {
	GetAccess(uuid.UUID, user.Email) (string, error)
	GetRefresh() (string, error)
	Validate(string) (*PayzClaims, error)
}

type PayzClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}
