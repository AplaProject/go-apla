package tx

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
