package userpg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRepository struct {
	user.UserRepository
	log *slog.Logger
	db  *gorm.DB
}

func New(log *slog.Logger, db *gorm.DB) (*UserRepository, error) {
	if log == nil || db == nil {
		return nil, consts.ErrNilArgs
	}

	return &UserRepository{
		log: log,
		db:  db,
	}, nil
}

func (ur *UserRepository) Save(ctx context.Context, user *user.User) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ur.log.Debug("converting user from domain", slog.String("user_id", user.ID().String()))
	model := toModel(user)

	ur.log.Debug("saving user into table", slog.String("user_id", user.ID().String()))
	if err := ur.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if ur.isUniqueError(err) {
			return nil, consts.ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("failed to save user into table: %w", err)
	}

	ur.log.Debug("saved successfully", slog.String("user_id", user.ID().String()))
	return toDomain(model), nil
}

func (ur *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ur.log.Debug("deleting user", slog.String("user_id", id.String()))
	result := ur.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{})
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) || errors.Is(result.Error, context.Canceled) {
			return result.Error
		}

		return fmt.Errorf("failed to delete user from db: %w", result.Error)
	}

	ur.log.Debug("checking exist")
	if result.RowsAffected == 0 {
		return consts.ErrUserDoesntExist
	}

	return nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model UserModel

	ur.log.Debug("fetching user by email", slog.String("email", email.Value()))
	if err := ur.db.WithContext(ctx).First(&model, "email = ?", email.Value()).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return toDomain(&model), nil
}

func (ur *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model UserModel

	ur.log.Debug("fetching user by id", slog.String("user_id", id.String()))
	if err := ur.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return toDomain(&model), nil
}

func (ur *UserRepository) isUniqueError(err error) bool {
	return errors.Is(postgres.Dialector{}.Translate(err), gorm.ErrDuplicatedKey)
}
