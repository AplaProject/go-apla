package tx

// Добавление нового меню
type AppendMenu struct {
	Header
	Global string
	Name   string
	Value  string
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
