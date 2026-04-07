package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/cnt-payz/payz/payment-service/internal/domain/session"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/payment-service/pkg/consts"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	session.JWTManager
	cfg    *config.Config
	public *rsa.PublicKey
}

func New(cfg *config.Config) (*JWTManager, error) {
	if cfg == nil {
		return nil, consts.ErrNilArgs
	}

	public, err := loadPublic(cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		cfg:    cfg,
		public: public,
	}, nil
}

func (jm *JWTManager) Validate(tokenString string) (*session.PayzClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &session.PayzClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, consts.ErrInvalidToken
		}

		return jm.public, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*session.PayzClaims); ok {
		return claims, nil
	}

	return nil, consts.ErrInvalidToken
}

func loadPublic(path string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}
