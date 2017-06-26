package tx

import "fmt"

type NewTable struct {
	Header
	Global  string
	Name    string
	Columns string
}

func (s NewTable) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Global, s.Name, s.Columns)
}

type EditTable struct {
	Header
	Name          string
	GeneralUpdate string
	Insert        string
	NewColumn     string
}

func (s EditTable) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Name, s.GeneralUpdate, s.Insert, s.NewColumn)
}
