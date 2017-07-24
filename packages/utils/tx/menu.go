package tx

import "fmt"

type AppendMenu struct {
	Header
	Global string
	Name   string
	Value  string
}

func (s AppendMenu) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value)
}

type NewMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (s NewMenu) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Conditions)
}

// Редактирование меню
type EditMenu struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

func (s EditMenu) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Conditions)
}
