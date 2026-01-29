package user

import (
	"fmt"
	"strings"

	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHash string

func NewPasswordHash(raw string) (PasswordHash, error) {
	raw = strings.TrimSpace(raw)
	if len([]rune(raw)) < 6 {
		return "", consts.ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash from password: %w", err)
	}

	return PasswordHash(hash), nil
}

func (ph PasswordHash) Compare(raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(ph), []byte(raw)) == nil
}

func (ph PasswordHash) Value() string {
	return string(ph)
}
