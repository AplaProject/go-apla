package tx

import "fmt"

type NewState struct {
	Header
	StateName    string
	CurrencyName string
}

func (n NewState) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s,%s", n.Header.Type, n.Header.Time, n.Header.UserID, n.StateName, n.CurrencyName)
}

type EditStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}

func (n EditStateParameters) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", n.Header.Type, n.Header.Time, n.Header.UserID, n.Header.StateID, n.Name, n.Value, n.Conditions)
}

type NewStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}

func (n NewStateParameters) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", n.Header.Type, n.Header.Time, n.Header.UserID, n.Header.StateID, n.Name, n.Value, n.Conditions)
}
