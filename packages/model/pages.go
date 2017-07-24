package model

type Page struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Menu       string `gorm:"not null;size:255"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (p *Page) SetTableName(prefix string) {
	p.tableName = prefix + "_pages"
}

func (p *Page) TableName() string {
	return p.tableName
}

func (p *Page) Get(name string) error {
	return DBConn.Where("name = ?", name).First(p).Error
}

func (p *Page) ToMap() map[string]string {
	result := make(map[string]string)
	result["name"] = p.Name
	result["value"] = p.Value
	result["menu"] = p.Menu
	result["conditions"] = p.Conditions
	result["rb_id"] = string(p.RbID)
	return result
}
