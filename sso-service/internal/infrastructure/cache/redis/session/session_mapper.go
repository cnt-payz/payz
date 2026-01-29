package sessionredis

import (
	"encoding/json"
	"fmt"

	"github.com/cnt-payz/payz/sso-service/internal/domain/session"
)

func toDomain(bytes []byte) (*session.Session, error) {
	var model SessionModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}

	return session.From(
		model.UserID,
		model.Email,
		session.FingerPrint(model.FingerPrint),
		model.CreatedAt,
	), nil
}

func toModel(domain *session.Session) ([]byte, error) {
	bytes, err := json.Marshal(&SessionModel{
		UserID:      domain.UserID(),
		Email:       domain.Email(),
		FingerPrint: domain.FingerPrint().Value(),
		CreatedAt:   domain.CreatedAt(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model: %w", err)
	}

	return bytes, nil
}
