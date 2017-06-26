package tx

import (
	"encoding/hex"
	"fmt"
)

type EditWallet struct {
	Header
	WalletID         string
	SpendingContract string
	Conditions       string
}

func (s EditWallet) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.WalletID, s.Header.StateID, s.WalletID, s.SpendingContract, s.Conditions)
}

type EditNewLang struct {
	Header
	Name  string
	Trans string
}

func (s EditNewLang) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Name, s.Trans)
}

type EditNewSign struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (s EditNewSign) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Conditions)
}

type NewAccount struct {
	Header
	PublicKey []byte
}

func (s NewAccount) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, hex.EncodeToString(s.PublicKey))
}

type ChangeNodeKey struct {
	Header
	NewNodePublicKey []byte
}

func (s ChangeNodeKey) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.NewNodePublicKey)
}

type UpdFullNodes struct {
	Header
}

func (s UpdFullNodes) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d", s.Header.Type, s.Header.Time, s.Header.UserID, 0)
}
