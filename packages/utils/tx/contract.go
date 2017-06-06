package tx

import "fmt"

type NewContract struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (e NewContract) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.Global, e.Name, e.Value, e.Conditions)
}

type EditContract struct {
	Header
	Global     string
	Id         string
	Value      string
	Conditions string
}

func (e EditContract) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.Global, e.Id, e.Value, e.Conditions)
}

type ActivateContract struct {
	Header
	Global string
	Id     string
}

func (a ActivateContract) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s", a.Header.Type, a.Header.Time, a.Header.UserID,
		a.Header.StateID, a.Global, a.Id)
}
