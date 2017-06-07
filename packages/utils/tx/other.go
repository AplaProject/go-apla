package tx

import "fmt"

type EditWallet struct {
	Header
	WalletID         string
	SpendingContract string
	Conditions       string
}

func (e EditWallet) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", e.Header.Type, e.Header.Time, e.WalletID, e.Header.StateID, e.WalletID, e.SpendingContract, e.Conditions)
}

type EditNewLang struct {
	Header
	Name  string
	Trans string
}

func (e EditNewLang) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.Name, e.Trans)
}

type EditNewSign struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (e *EditNewSign) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.Global, e.Name, e.Value, e.Conditions)
}

type NewAccount struct {
	Header
}

type ChangeNodeKey struct {
	Header
}
