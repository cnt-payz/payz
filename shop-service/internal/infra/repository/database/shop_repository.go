package dbrepo

import (
	"context"
	"errors"
	"log/slog"

	shopdomain "github.com/cnt-payz/payz/shop-service/internal/domain/shop"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrShopNotFound = errors.New("shop not found")
)

type shop struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID      `gorm:"not null;type:uuid"`
	Name      string         `gorm:"not null;size:32"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (dr *DatabaseRepo) CreateShop(ctx context.Context, shop *shopdomain.Shop) (*shopdomain.Shop, error) {
	dr.log.Debug("creating shop",
		slog.String("shop_id", shop.ID().String()),
		slog.String("user_id", shop.UserID().String()),
		slog.String("shop_name", shop.Name().String()),
	)

	s := dr.ShopDomainToDBModel(shop)
	if err := dr.db.WithContext(ctx).Create(s).Error; err != nil {
		return nil, err
	}

	return dr.ShopDBModelToDomain(s)
}

func (dr *DatabaseRepo) GetShopByOwner(ctx context.Context, id, userID uuid.UUID) (*shopdomain.Shop, error) {
	dr.log.Debug("getting shop", slog.String("id", id.String()))

	var shop shop
	if err := dr.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&shop).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}

		return nil, err
	}

	return dr.ShopDBModelToDomain(&shop)
}

func (dr *DatabaseRepo) GetShopsByUserID(ctx context.Context, userID uuid.UUID) ([]*shopdomain.Shop, error) {
	dr.log.Debug("getting shops", slog.String("user_id", userID.String()))

	var shops []*shop
	if err := dr.db.WithContext(ctx).Where("user_id = ?", userID).Find(&shops).Error; err != nil {
		return nil, err
	}

	return dr.ShopDBModelsToDomain(shops)
}

func (dr *DatabaseRepo) CountShopsByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	dr.log.Debug("counting shops", slog.String("user_id", userID.String()))

	var count int64
	if err := dr.db.WithContext(ctx).Model(&shop{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (dr *DatabaseRepo) UpdateShopByID(ctx context.Context, id, userID uuid.UUID, name shopdomain.Name) (*shopdomain.Shop, error) {
	dr.log.Debug("updating shop",
		slog.String("id", id.String()),
		slog.String("new_name", name.String()),
	)

	var newShop shop
	result := dr.db.WithContext(ctx).Model(&shop{}).Clauses(clause.Returning{}).Where("id = ? AND user_id = ?", id, userID).Update("name", name.String()).Scan(&newShop)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrShopNotFound
	}

	return dr.ShopDBModelToDomain(&newShop)
}

func (dr *DatabaseRepo) DeleteShopByID(ctx context.Context, id, userID uuid.UUID) error {
	dr.log.Debug("deleting shop", slog.String("id", id.String()))

	result := dr.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&shop{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrShopNotFound
	}

	return nil
}

func (dr *DatabaseRepo) ShopDomainToDBModel(s *shopdomain.Shop) *shop {
	return &shop{
		ID:     s.ID(),
		UserID: s.UserID(),
		Name:   s.Name().String(),
	}
}

func (dr *DatabaseRepo) ShopDBModelToDomain(s *shop) (*shopdomain.Shop, error) {
	return shopdomain.ShopFromRepo(s.ID, s.UserID, s.Name)
}

func (dr *DatabaseRepo) ShopDBModelsToDomain(shops []*shop) ([]*shopdomain.Shop, error) {
	domainShops := make([]*shopdomain.Shop, len(shops))
	for i, s := range shops {
		domainShop, err := shopdomain.ShopFromRepo(s.ID, s.UserID, s.Name)
		if err != nil {
			return nil, err
		}

		domainShops[i] = domainShop
	}

	return domainShops, nil
}
