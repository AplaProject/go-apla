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

// Новое меню
type NewMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

// Редактирование меню
type EditMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}
