package model

import (
	"github.com/GenesisKernel/go-genesis/packages/converter"
)

// Contract represents record of 1_contracts table
type Contract struct {
	tableName   string
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	WalletID    int64  `json:"wallet_id,omitempty"`
	TokenID     int64  `json:"token_id,omitempty"`
	Active      bool   `json:"active,omitempty"`
	Conditions  string `json:"conditions,omitempty"`
	AppID       int64  `json:"app_id,omitempty"`
	EcosystemID int64  `json:"ecosystem_id,omitempty"`
}

// TableName returns name of table
func (c *Contract) TableName() string {
	return "1_contracts"
}

func (c *Contract) GetList(offset, limit int64) ([]Contract, error) {
	var list []Contract
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).
		Order("id").Where("ecosystem = ?", c.EcosystemID).
		Find(&list).Error
	return list, err
}

func (c *Contract) Count() (n int64, err error) {
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

// GetByApp returns all contracts belonging to selected app
func (c *Contract) GetByApp(appID int64) ([]Contract, error) {
	var result []Contract
	err := DBConn.Select("id, name").Where("app_id = ?", appID).Find(&result).Error
	return result, err
}
