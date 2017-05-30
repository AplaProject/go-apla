package tx

// Новое государство
type NewState struct {
	Header
	StateName    string
	CurrencyName string
}

// Редактировать параметры государства
type EditStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}

// Новые параметры государства
type NewStateParameters struct {
	Header
	Name       string
	Value      string
	Conditions string
}
