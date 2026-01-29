package sessionredis

import (
	"time"

	"github.com/google/uuid"
)

type SessionModel struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	FingerPrint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}
