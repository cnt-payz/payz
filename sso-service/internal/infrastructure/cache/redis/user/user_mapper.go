package userredis

import (
	"encoding/json"

	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
)

func toDomain(bytes []byte) (*user.User, error) {
	var model UserModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, err
	}

	return user.From(
		model.ID,
		user.Email(model.Email),
		"",
		model.CreatedAt,
	), nil
}

func toModel(domain *user.User) ([]byte, error) {
	bytes, err := json.Marshal(&UserModel{
		ID:        domain.ID(),
		Email:     domain.Email().Value(),
		CreatedAt: domain.CreatedAt(),
	})
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
