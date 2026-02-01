package extransactionpg

import (
	"time"

	"github.com/google/uuid"
)

type ExTransactionModel struct {
	ID           uuid.UUID `gorm:"primarykey"`
	ShopID       uuid.UUID `gorm:"not null"`
	UserID       uuid.UUID `gorm:"not null"`
	StatusID     int       `gorm:"not null"`
	TypID        int       `gorm:"not null"`
	Amount       string
	CallbackURL  string
	UserMetadata []byte `gorm:"type:jsonb"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
