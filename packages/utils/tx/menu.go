package tx

import "fmt"

type AppendMenu struct {
	Header
	Global string
	Name   string
	Value  string
}

func (a AppendMenu) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", a.Header.Type, a.Header.Time, a.Header.UserID, a.Header.StateID, a.Global, a.Name, a.Value)
}

type NewMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (a NewMenu) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", a.Header.Type, a.Header.Time, a.Header.UserID, a.Header.StateID, a.Global, a.Name, a.Value, a.Conditions)
}

// Редактирование меню
type EditMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (a EditMenu) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", a.Header.Type, a.Header.Time, a.Header.UserID, a.Header.StateID, a.Global, a.Name, a.Value, a.Conditions)
}
