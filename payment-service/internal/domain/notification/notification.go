package notification

import "github.com/google/uuid"

type Message struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	CallbackURL   string    `json:"callback_url"`
	UserMetadata  []byte    `json:"user_metadata"`
	Status        string    `json:"status"`
}
