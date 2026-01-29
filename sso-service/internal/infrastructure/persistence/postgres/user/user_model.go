package userpg

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID           uuid.UUID `gorm:"primarykey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"created_at"`
}
