package userpg

import "github.com/cnt-payz/payz/sso-service/internal/domain/user"

func toDomain(model *UserModel) *user.User {
	return user.From(
		model.ID,
		user.Email(model.Email),
		user.PasswordHash(model.PasswordHash),
		model.CreatedAt,
	)
}

func toModel(domain *user.User) *UserModel {
	return &UserModel{
		ID:           domain.ID(),
		Email:        domain.Email().Value(),
		PasswordHash: domain.PasswordHash().Value(),
		CreatedAt:    domain.CreatedAt(),
	}
}
