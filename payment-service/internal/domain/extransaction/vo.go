package extransaction

import (
	"net/url"
	"strconv"

	"github.com/cnt-payz/payz/payment-service/pkg/consts"
)

type status int

const (
	STATUS_UNKNOWN status = iota
	STATUS_PENDING
	STATUS_SUCCESS
	STATUS_CANCELED
	STATUS_DEADLINE
)

type Status struct {
	status status
}

func (s Status) String() string {
	switch s.status {
	case STATUS_PENDING:
		return "pending"
	case STATUS_SUCCESS:
		return "success"
	case STATUS_CANCELED:
		return "canceled"
	case STATUS_DEADLINE:
		return "deadline"
	default:
		return "unknown"
	}
}

func (s Status) Value() status {
	return s.status
}

func NewStatusID(raw int) (Status, error) {
	status := status(raw)
	if status <= STATUS_UNKNOWN || status > STATUS_DEADLINE {
		return Status{}, consts.ErrInvalidStatus
	}

	return Status{
		status: status,
	}, nil
}

func NewStatus(status status) Status {
	return Status{
		status: status,
	}
}

type typ int

const (
	TYPE_UNKNOWN typ = iota
	TYPE_DEPOSIT
	TYPE_WITHDRAW
)

type Typ struct {
	typ typ
}

func (t Typ) String() string {
	switch t.typ {
	case TYPE_DEPOSIT:
		return "deposit"
	case TYPE_WITHDRAW:
		return "withdraw"
	default:
		return "unknown"
	}
}

func (t Typ) Value() typ {
	return t.typ
}

func NewTypID(raw int) (Typ, error) {
	typ := typ(raw)
	if typ <= TYPE_UNKNOWN || typ > TYPE_WITHDRAW {
		return Typ{}, consts.ErrInvalidTyp
	}

	return Typ{
		typ: typ,
	}, nil
}

func NewTyp(typ typ) Typ {
	return Typ{
		typ: typ,
	}
}

type Amount string

func (a Amount) Value() string {
	return string(a)
}

func NewAmount(raw string) (Amount, error) {
	if _, err := strconv.ParseFloat(raw, 64); err != nil {
		return "", consts.ErrInvalidAmount
	}

	return Amount(raw), nil
}

type CallbackURL string

func (curl CallbackURL) Value() string {
	return string(curl)
}

func NewCallbackURL(raw string) (CallbackURL, error) {
	if _, err := url.Parse(raw); err != nil {
		return "", consts.ErrInvalidCallbackURL
	}

	return CallbackURL(raw), nil
}
