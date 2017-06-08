package tx

import "fmt"

type NewTable struct {
	Header
	Global  string
	Name    string
	Columns string
}

func (n NewTable) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", n.Header.Type, n.Header.Time, n.Header.UserID, n.Header.StateID, n.Global, n.Name, n.Columns)
}

type EditTable struct {
	Header
	Name          string
	GeneralUpdate string
	Insert        string
	NewColumn     string
}

func (e EditTable) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.Name, e.GeneralUpdate, e.Insert, e.NewColumn)
}
