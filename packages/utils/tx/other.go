package tx

// Редактирование кошелька
type EditWallet struct {
	Header
	WalletID         string
	SpendingContract string
	Conditions       string
}

// Редактировать/создать новый язык
type EditNewLang struct {
	Header
	Name  string
	Trans string
}

type EditNewSign struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

type NewAccount struct {
	Header
}

type ChangeNodeKey struct {
	Header
}
