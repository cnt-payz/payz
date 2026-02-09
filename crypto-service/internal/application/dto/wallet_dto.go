package walletdto

type ExTransaction struct {
	TransactionID string
	Amount        string
	TargetAddress string
	TakerAddress  string
}

type Wallet struct {
	ID           string
	ShopID       string
	Address      string
	FrozenAmount string
	Amount       string
}

type DepositReq struct {
	TransactionID string
	ShopID        string
	Amount        string
	TakerAddress  string
}

type WithdrawReq struct {
	ShopID        string
	Amount        string
	TargetAddress string
}
