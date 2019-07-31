// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package model

import "github.com/AplaProject/go-apla/packages/converter"

// Page is model
type Page struct {
	ecosystem     int64
	ID            int64  `gorm:"primary_key;not null" json:"id,omitempty"`
	Name          string `gorm:"not null" json:"name,omitempty"`
	Value         string `gorm:"not null" json:"value,omitempty"`
	Menu          string `gorm:"not null;size:255" json:"menu,omitempty"`
	ValidateCount int64  `gorm:"not null" json:"nodesCount,omitempty"`
	AppID         int64  `gorm:"column:app_id;not null" json:"app_id,omitempty"`
	Conditions    string `gorm:"not null" json:"conditions,omitempty"`
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

// GetByApp returns all pages belonging to selected app
func (p *Page) GetByApp(appID int64, ecosystemID int64) ([]Page, error) {
	var result []Page
	err := DBConn.Select("id, name").Where("app_id = ? and ecosystem = ?", appID, ecosystemID).Find(&result).Error
	return result, err
}
