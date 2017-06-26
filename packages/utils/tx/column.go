package tx

import "fmt"

// Новая колонка в реестре
type NewColumn struct {
	Header
	TableName   string
	ColumnName  string
	ColumnType  string
	Permissions string
	Index       string
}

func (s NewColumn) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.TableName, s.ColumnName, s.Permissions, s.Index, s.ColumnType)
}

// Редактирование колонки в реестре
type EditColumn struct {
	Header
	TableName   string
	ColumnName  string
	Permissions string
}

func (s EditColumn) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.TableName, s.ColumnName, s.Permissions)
}
