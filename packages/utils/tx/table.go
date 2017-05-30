package tx

// Новая таблица
type NewTable struct {
	Header
	Global  string
	Name    string
	Columns string
}

// Редактировать таблицу
type EditTable struct {
	Header
	Name          string
	GeneralUpdate string
	Insert        string
	NewColumn     string
}
