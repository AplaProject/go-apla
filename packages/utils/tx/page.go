package tx

import "fmt"

type AppendPage struct {
	Header
	Global string
	Name   string
	Value  string
}

func (s AppendPage) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value)
}

type EditPage struct {
	Header
	Global     string
	Name       string
	Value      string
	Menu       string
	Conditions string
}

func (s EditPage) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Menu, s.Conditions)
}

type NewPage struct {
	Header
	Global     string
	Name       string
	Value      string
	Menu       string
	Conditions string
}

func (s NewPage) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Value, s.Menu, s.Conditions)
}
