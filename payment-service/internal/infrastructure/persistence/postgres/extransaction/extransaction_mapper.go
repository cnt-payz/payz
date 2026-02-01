package extransactionpg

import "github.com/cnt-payz/payz/payment-service/internal/domain/extransaction"

func toDomain(model *ExTransactionModel) *extransaction.ExTransaction {
	status, _ := extransaction.NewStatusID(model.StatusID)
	typ, _ := extransaction.NewTypID(model.TypID)

	return extransaction.From(
		model.ID,
		model.ShopID,
		model.UserID,
		status,
		typ,
		extransaction.Amount(model.Amount),
		extransaction.CallbackURL(model.CallbackURL),
		model.UserMetadata,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func toModel(domain *extransaction.ExTransaction) *ExTransactionModel {
	return &ExTransactionModel{
		ID:           domain.ID(),
		ShopID:       domain.ShopID(),
		UserID:       domain.UserID(),
		StatusID:     int(domain.Status().Value()),
		TypID:        int(domain.Typ().Value()),
		Amount:       domain.Amount().Value(),
		CallbackURL:  domain.CallbackURL().Value(),
		UserMetadata: domain.UserMetadata(),
		CreatedAt:    domain.CreatedAt(),
		UpdatedAt:    domain.UpdatedAt(),
	}
}
