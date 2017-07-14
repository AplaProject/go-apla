package tx

import "fmt"

type NewState struct {
	Header
	StateName    string
	CurrencyName string
}

func (s NewState) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.StateName, s.CurrencyName)
}

type EditStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}

func (s EditStateParameters) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Name, s.Value, s.Conditions)
}

type NewStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}

func (s NewStateParameters) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Name, s.Value, s.Conditions)
}
