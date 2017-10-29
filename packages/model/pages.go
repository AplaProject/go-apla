package model

type Page struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Menu       string `gorm:"not null;size:255"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (p *Page) SetTablePrefix(prefix string) {
	p.tableName = prefix + "_pages"
}

func (p *Page) TableName() string {
	return p.tableName
}

func (p *Page) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(p))
}
