package consts

import "errors"

var (
	// Domain's
	ErrInvalidStatus       = errors.New("invalid status of transaction")
	ErrInvalidTyp          = errors.New("invalid type of transaction")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidCallbackURL  = errors.New("invalid callback url")
	ErrInvalidShopID       = errors.New("invalid shop id")
	ErrInvalidUserMetadata = errors.New("invalid user's metadata")

	// General's
	ErrInvalidArgs  = errors.New("some args are invalid")
	ErrInvalidToken = errors.New("invalid token")
	ErrNilArgs      = errors.New("some args are nil")
	ErrNilRequest   = errors.New("request cannot be empty")

	// Repository's
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrTransactionDoesntExist   = errors.New("transaction doesn't exist")
	ErrTransactionIsntPending   = errors.New("transaction isn't pending")
	ErrIdempotencyNotFound      = errors.New("idempotency key not found")
	ErrIdempotencyCheck         = errors.New("idempotency key already used")
	ErrInternalServer           = errors.New("internal server error")
	ErrTooOldRequest            = errors.New("too old request")
	ErrInvalidSignature         = errors.New("invalid signature")
	ErrInvalidTransactionID     = errors.New("invalid transaction's id")
)
