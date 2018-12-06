package model

import "github.com/AplaProject/go-apla/packages/converter"

// Page is model
type Page struct {
	ecosystem     int64
	ID            int64  `gorm:"primary_key;not null" json:"id"`
	Name          string `gorm:"not null" json:"name"`
	Value         string `gorm:"not null" json:"value"`
	Menu          string `gorm:"not null;size:255" json:"menu"`
	ValidateCount int64  `gorm:"not null" json:"nodesCount"`
	AppID         int64  `gorm:"column:app_id;not null" json:"app_id"`
	Conditions    string `gorm:"not null" json:"conditions"`
}

// SetTablePrefix is setting table prefix
func (p *Page) SetTablePrefix(prefix string) {
	p.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (p *Page) TableName() string {
	if p.ecosystem == 0 {
		p.ecosystem = 1
	}
	return `1_pages`
}

// Get is retrieving model from database
func (p *Page) Get(name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and name = ?", p.ecosystem, name).First(p))
}

// Count returns count of records in table
func (p *Page) Count() (count int64, err error) {
	err = DBConn.Table(p.TableName()).Count(&count).Error
	return
}
