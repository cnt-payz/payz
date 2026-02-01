package extransaction

import (
	"time"

	"github.com/google/uuid"
)

type ExTransaction struct {
	id           uuid.UUID
	shopID       uuid.UUID
	userID       uuid.UUID
	status       Status
	typ          Typ
	amount       Amount
	callbackURL  CallbackURL
	userMetadata []byte
	createdAt    time.Time
	updatedAt    time.Time
}

func New(
	shopID, userID uuid.UUID,
	typ Typ,
	amount Amount,
	callbackURL CallbackURL,
	userMetadata []byte,
) *ExTransaction {
	return &ExTransaction{
		id:           uuid.New(),
		shopID:       shopID,
		userID:       userID,
		status:       NewStatus(STATUS_PENDING),
		typ:          typ,
		amount:       amount,
		callbackURL:  callbackURL,
		userMetadata: userMetadata,
		createdAt:    time.Now().UTC(),
		updatedAt:    time.Now().UTC(),
	}
}

func From(
	id, shopID, userID uuid.UUID,
	status Status,
	typ Typ,
	amount Amount,
	callbackURL CallbackURL,
	userMetadata []byte,
	createdAt, updatedAt time.Time,
) *ExTransaction {
	return &ExTransaction{
		id:           id,
		shopID:       shopID,
		userID:       userID,
		status:       status,
		typ:          typ,
		amount:       amount,
		callbackURL:  callbackURL,
		userMetadata: userMetadata,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (et *ExTransaction) ID() uuid.UUID {
	return et.id
}

func (et *ExTransaction) ShopID() uuid.UUID {
	return et.shopID
}

func (et *ExTransaction) UserID() uuid.UUID {
	return et.userID
}

func (et *ExTransaction) Status() Status {
	return et.status
}

func (et *ExTransaction) Typ() Typ {
	return et.typ
}

func (et *ExTransaction) Amount() Amount {
	return et.amount
}

func (et *ExTransaction) CallbackURL() CallbackURL {
	return et.callbackURL
}

func (et *ExTransaction) UserMetadata() []byte {
	return et.userMetadata
}

func (et *ExTransaction) CreatedAt() time.Time {
	return et.createdAt
}

func (et *ExTransaction) UpdatedAt() time.Time {
	return et.updatedAt
}
