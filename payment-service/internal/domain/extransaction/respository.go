package extransaction

import (
	"context"

	"github.com/google/uuid"
)

type ExTransactionRepository interface {
	Save(context.Context, *ExTransaction) (*ExTransaction, error)
	Confirm(context.Context, uuid.UUID) (*ExTransaction, error)
	Cancel(context.Context, uuid.UUID) (*ExTransaction, error)
	GetExHistory(context.Context, uuid.UUID, uint32, uint32) ([]*ExTransaction, error)
	GetExTransaction(context.Context, uuid.UUID) (*ExTransaction, error)
}
