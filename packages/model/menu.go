package model

type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (m *Menu) SetTableName(prefix string) {
	m.tableName = prefix + "_menus"
}

func (m *Menu) TableName() string {
	return m.tableName
}

func (m *Menu) Get(name string) error {
	return DBConn.Where("name = ?", name).First(m).Error
}

func (m *Menu) GetAll(prefix string) ([]Menu, error) {
	var result []Menu
	err := DBConn.Table(prefix + "_menus").Order("name").Find(result).Error
	return result, err
}

func (m *Menu) ToMap() map[string]string {
	result := make(map[string]string)
	result["name"] = m.Name
	result["value"] = m.Value
	result["conditions"] = m.Conditions
	result["rb_id"] = string(m.RbID)
	return result
}
