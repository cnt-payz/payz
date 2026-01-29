package user

import (
	"net/mail"

	"github.com/cnt-payz/payz/sso-service/pkg/consts"
)

type Email string

func NewEmail(raw string) (Email, error) {
	if _, err := mail.ParseAddress(raw); err != nil {
		return "", consts.ErrInvalidEmail
	}

	return Email(raw), nil
}

func (e Email) Value() string {
	return string(e)
}
