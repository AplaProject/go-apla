package tx

import "fmt"

type NewContract struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (s NewContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Conditions)
}

type EditContract struct {
	Header
	Global     string
	Id         string
	Value      string
	Conditions string
}

func (s EditContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Id, s.Value, s.Conditions)
}

type ActivateContract struct {
	Header
	Global string
	Id     string
}

func (s ActivateContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID,
		s.Header.StateID, s.Global, s.Id)
}
