package model

import "strconv"

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

func (p *Page) Get(name string) error {
	return DBConn.Where("name = ?", name).First(p).Error
}

func (p *Page) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(p).Error
}

func (p *Page) GetWithMenu(prefix string) ([]Page, error) {
	pages := new([]Page)
	err := DBConn.Table(prefix + "_pages").Where("menu != '0'").Order("name").Find(&pages).Error
	return *pages, err
}

func (p *Page) GetWithoutMenu(prefix string) ([]Page, error) {
	pages := new([]Page)
	err := DBConn.Table(prefix + "_pages").Where("menu = '0'").Order("name").Find(&pages).Error
	return *pages, err
}

func (p *Page) ToMap() map[string]string {
	result := make(map[string]string)
	result["name"] = p.Name
	result["value"] = p.Value
	result["menu"] = p.Menu
	result["conditions"] = p.Conditions
	result["rb_id"] = strconv.FormatInt(p.RbID, 10)
	return result
}

func CreateStatePagesTable(transaction *DbTransaction, stateID string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + stateID + `_pages" (
			    "name" varchar(255)  NOT NULL DEFAULT '',
			    "value" text  NOT NULL DEFAULT '',
			    "menu" varchar(255)  NOT NULL DEFAULT '',
			    "conditions" bytea  NOT NULL DEFAULT '',
			    "rb_id" bigint NOT NULL DEFAULT '0'
			    );
			    ALTER TABLE ONLY "` + stateID + `_pages" ADD CONSTRAINT "` + stateID + `_pages_pkey" PRIMARY KEY (name);`).Error
}
