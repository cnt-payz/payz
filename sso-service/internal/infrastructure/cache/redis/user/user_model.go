package userredis

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
