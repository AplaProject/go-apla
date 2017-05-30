package tx

// Новый контракт
type NewContract struct {
	Header
	Global     string
	Name       string
	Value      string
	Conditions string
}

// Редактировать контракт
type EditContract struct {
	Header
	Global     string
	Id         string
	Value      string
	Conditions string
}

// Активация контракта
type ActivateContract struct {
	Header
	Global string
	Id     string
}
