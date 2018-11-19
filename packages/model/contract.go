// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package model

import "github.com/GenesisKernel/go-genesis/packages/converter"

// Contract represents record of 1_contracts table
type Contract struct {
	ID          int64
	Name        string
	Value       string
	WalletID    int64
	TokenID     int64
	Conditions  string
	AppID       int64
	EcosystemID int64 `gorm:"column:ecosystem"`
}

// TableName returns name of table
func (c *Contract) TableName() string {
	return `1_contracts`
}

// GetList is retrieving records from database
func (c *Contract) GetList(offset, limit int64) ([]Contract, error) {
	result := new([]Contract)
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).Order("id asc").Find(&result).Error
	return *result, err
}

// GetFromEcosystem retrieving ecosystem contracts from database
func (c *Contract) GetFromEcosystem(db *DbTransaction, ecosystem int64) ([]Contract, error) {
	result := new([]Contract)
	err := GetDB(db).Table(c.TableName()).Where("ecosystem = ?", ecosystem).Order("id asc").Find(&result).Error
	return *result, err
}

// Count returns count of records in table
func (c *Contract) Count() (count int64, err error) {
	err = DBConn.Table(c.TableName()).Count(&count).Error
	return
}

func (c *Contract) GetListByEcosystem(offset, limit int64) ([]Contract, error) {
	var list []Contract
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).
		Order("id").Where("ecosystem = ?", c.EcosystemID).
		Find(&list).Error
	return list, err
}

func (c *Contract) CountByEcosystem() (n int64, err error) {
	err = DBConn.Table(c.TableName()).Where("ecosystem = ?", c.EcosystemID).Count(&n).Error
	return
}

func (c *Contract) ToMap() (v map[string]string) {
	v = make(map[string]string)
	v["id"] = converter.Int64ToStr(c.ID)
	v["name"] = c.Name
	v["value"] = c.Value
	v["wallet_id"] = converter.Int64ToStr(c.WalletID)
	v["token_id"] = converter.Int64ToStr(c.TokenID)
	v["conditions"] = c.Conditions
	v["app_id"] = converter.Int64ToStr(c.AppID)
	v["ecosystem_id"] = converter.Int64ToStr(c.EcosystemID)
	return
}
