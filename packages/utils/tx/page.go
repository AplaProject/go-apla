package tx

// Добавление новой страницы
type AppendPage struct {
	Header
	Global string
	Name   string
	Value  string
}

// Редактирование существующей страницы
type EditPage struct {
	Header
	Global     string
	Name       string
	Value      string
	Menu       string
	Conditions string
}

// Новая страница
type NewPage struct {
	Header
	Global     string
	Name       string
	Value      string
	Menu       string
	Conditions string
}
