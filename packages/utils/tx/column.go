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

// Редактирование колонки в реестре
type EditColumn struct {
	Header
	TableName   string
	ColumnName  string
	Permissions string
}

func (e EditColumn) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", e.Header.Type, e.Header.Time, e.Header.UserID, e.Header.StateID, e.TableName, e.ColumnName, e.Permissions)
}
