package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/cnt-payz/payz/sso-service/internal/domain/session"
	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	session.JWTManager
	cfg     *config.Config
	public  *rsa.PublicKey
	private *rsa.PrivateKey
}

func New(cfg *config.Config) (*JWTManager, error) {
	if cfg == nil {
		return nil, consts.ErrNilArgs
	}

	private, err := loadPrivate(cfg.Secrets.JWT.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	public, err := loadPublic(cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		cfg:     cfg,
		public:  public,
		private: private,
	}, nil
}

func (jm *JWTManager) GetAccess(userID uuid.UUID, email user.Email) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &session.PayzClaims{
		UserID: userID,
		Email:  email.Value(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "payz",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.cfg.Secrets.JWT.AccessTTL)),
		},
	})

	tokenString, err := token.SignedString(jm.private)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (jm *JWTManager) GetRefresh() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read buffer: %w", err)
	}

	return hex.EncodeToString(buf), nil
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

func loadPrivate(path string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
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
