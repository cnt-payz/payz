package extransactionpg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cnt-payz/payz/payment-service/internal/domain/extransaction"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/payment-service/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ExTransactionRepository struct {
	extransaction.ExTransactionRepository
	cfg *config.Config
	log *slog.Logger
	db  *gorm.DB
}

func New(cfg *config.Config, log *slog.Logger, db *gorm.DB) *ExTransactionRepository {
	return &ExTransactionRepository{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (etr *ExTransactionRepository) Save(ctx context.Context, transaction *extransaction.ExTransaction) (*extransaction.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	etr.log.Debug("convert domain to model", slog.String("transaction_id", transaction.ID().String()))
	model := toModel(transaction)

	etr.log.Debug("saving transaction into db", slog.String("transaction_id", transaction.ID().String()))
	if err := etr.db.WithContext(ctx).Create(model).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if etr.isUniqueError(err) {
			return nil, consts.ErrTransactionAlreadyExists
		}

		return nil, fmt.Errorf("failed to save ex-transaction: %w", err)
	}

	return toDomain(model), nil
}

func (etr *ExTransactionRepository) Confirm(ctx context.Context, transactionID uuid.UUID) (*extransaction.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	var model ExTransactionModel
	if err := etr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, transactionID).Error; err != nil {
			return err
		}

		if model.StatusID != int(extransaction.STATUS_PENDING) {
			return consts.ErrTransactionIsntPending
		}

		result := tx.Model(&model).Where("id = ? AND status_id = ?", transactionID, int(extransaction.STATUS_PENDING)).Updates(
			map[string]any{
				"status_id":  int(extransaction.STATUS_SUCCESS),
				"updated_at": now,
			},
		)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return consts.ErrTransactionIsntPending
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, consts.ErrTransactionIsntPending) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrTransactionDoesntExist
		}

		return nil, fmt.Errorf("failed to confirm transaction: %w", err)
	}

	return toDomain(&model), nil
}

func (etr *ExTransactionRepository) Cancel(ctx context.Context, transactionID uuid.UUID) (*extransaction.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	var model ExTransactionModel
	if err := etr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, transactionID).Error; err != nil {
			return err
		}

		if model.StatusID != int(extransaction.STATUS_PENDING) {
			return consts.ErrTransactionIsntPending
		}

		result := tx.Model(&model).Where("id = ? AND status_id = ?", transactionID, int(extransaction.STATUS_PENDING)).Updates(
			map[string]any{
				"status_id":  int(extransaction.STATUS_CANCELED),
				"updated_at": now,
			},
		)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return consts.ErrTransactionIsntPending
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, consts.ErrTransactionIsntPending) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrTransactionDoesntExist
		}

		return nil, fmt.Errorf("failed to cancel ex-transaction: %w", err)
	}

	return toDomain(&model), nil
}

func (etr *ExTransactionRepository) GetExHistory(ctx context.Context, shopID uuid.UUID, page, size uint32) ([]*extransaction.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if page > 10_000 || size > 1000 {
		return nil, consts.ErrInvalidArgs
	}

	offset := size * (page - 1)
	query := etr.db.WithContext(ctx).Limit(int(size)).Offset(int(offset)).Where("shop_id = ?", shopID)

	var models []ExTransactionModel
	if err := query.Find(&models).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to get ex-history: %w", err)
	}

	transactions := make([]*extransaction.ExTransaction, len(models))
	for idx, transaction := range models {
		transactions[idx] = toDomain(&transaction)
	}

	return transactions, nil
}

func (etr *ExTransactionRepository) GetExTransaction(ctx context.Context, transactionID uuid.UUID) (*extransaction.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model ExTransactionModel
	if err := etr.db.WithContext(ctx).Take(&model, transactionID).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrTransactionDoesntExist
		}

		return nil, fmt.Errorf("failed to get ex-transaction: %w", err)
	}

	return toDomain(&model), nil
}

func (etr *ExTransactionRepository) HandleDeadlines(ctx context.Context) {
	ticker := time.NewTicker(etr.cfg.Secrets.Postgres.WorkerTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoffTime := time.Now().UTC().Add(-etr.cfg.Service.TransactionTimeout)

			result := etr.db.WithContext(ctx).Model(&ExTransactionModel{}).
				Where("status_id = ? AND created_at <= ?",
					int(extransaction.STATUS_PENDING),
					cutoffTime).
				Update("status_id", int(extransaction.STATUS_DEADLINE))
			if result.Error != nil {
				etr.log.Warn("failed to update status of transactions",
					slog.String("error", result.Error.Error()),
				)
			}
		}
	}
}

func (etr *ExTransactionRepository) isUniqueError(err error) bool {
	return errors.Is(postgres.Dialector{}.Translate(err), gorm.ErrDuplicatedKey)
}
