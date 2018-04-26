package model

// Page is model
type Page struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Value      string `gorm:"not null" json:"value"`
	Menu       string `gorm:"not null;size:255" json:"menu"`
	Conditions string `gorm:"not null" json:"conditions"`
}

// SetTablePrefix is setting table prefix
func (p *Page) SetTablePrefix(prefix string) {
	p.tableName = prefix + "_pages"
}

// TableName returns name of table
func (p *Page) TableName() string {
	return p.tableName
}

// Get is retrieving model from database
func (p *Page) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(p))
}

// Count returns count of records in table
func (p *Page) Count() (count int64, err error) {
	err = DBConn.Table(p.TableName()).Count(&count).Error
	return
}
