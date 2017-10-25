package model

type SmartContract struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	Name       string `gorm:"not null;size:100"`
	Value      []byte `gorm:"not null"`
	WalletID   int64  `gorm:"not null"`
	Active     string `gorm:"not null;size:1"`
	Conditions string `gorm:"not null"`
	Variables  []byte `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sc *SmartContract) SetTablePrefix(tablePrefix string) {
	sc.tableName = tablePrefix + "_smart_contracts"
}

func (sc *SmartContract) TableName() string {
	return sc.tableName
}

func (sc *SmartContract) GetByName(contractName string) (bool, error) {
	return isFound(DBConn.Where("name = ?", contractName).Find(sc))
}
