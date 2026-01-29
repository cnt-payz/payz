package consts

import "errors"

var (
	// Domain's
	ErrInvalidEmail = errors.New("invalid email")

	// General's
	ErrNilArgs             = errors.New("some args are nil")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidToken        = errors.New("invalid token")

	// Service's
	ErrBadGateway     = errors.New("bad gateway")
	ErrInternalServer = errors.New("internal server error")
)
