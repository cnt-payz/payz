package consts

import "errors"

var (
	// Domain's
	ErrInvalidEmail    = errors.New("invalid mail")
	ErrInvalidPassword = errors.New("invalid password")

	// General's
	ErrNilArgs      = errors.New("some args are nil")
	ErrInvalidToken = errors.New("invalid token")

	// Service's
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidIP          = errors.New("invalid ip")
	ErrInvalidUserAgent   = errors.New("invalid user-agent")
	ErrInvalidFingerprint = errors.New("invalid finger print")
	ErrInvalidUserID      = errors.New("invalid user's id")
	ErrInvalidRequest     = errors.New("invalid request error")

	// Repository's
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserDoesntExist    = errors.New("user doesn't exist")
	ErrSessionDoesntExist = errors.New("there isn't any sessions")
	ErrInvalidPath        = errors.New("path to file is invalid")
)
