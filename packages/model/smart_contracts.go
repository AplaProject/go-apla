package model

type SmartContracts struct {
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

func (sc *SmartContracts) SetTableName(newName string) {
	sc.tableName = newName
}

func (sc *SmartContracts) TableName() string {
	return sc.tableName
}

func GetAllSmartContracts(tablePrefix string) ([]SmartContracts, error) {
	contracts := new([]SmartContracts)
	err := DBConn.Order("id").Table(tablePrefix + "_smart_contracts").Find(contracts).Error
	if err != nil {
		return nil, err
	}
	return *contracts, nil
}

func (sc *SmartContracts) Create() error {
	return DBConn.Create(sc).Error
}

func (sc *SmartContracts) GetByID(contractID int64) error {
	return DBConn.Where("id = ?", contractID).Find(sc).Error
}

func (sc *SmartContracts) ExistsByID(contractID int64) (bool, error) {
	query := DBConn.Where("id = ?", contractID).First(sc)
	return !query.RecordNotFound(), query.Error
}

func (sc *SmartContracts) ExistsByName(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(sc)
	return !query.RecordNotFound(), query.Error
}

func (sc *SmartContracts) GetByName(contractName string) error {
	return DBConn.Where("name = ?", contractName).Find(sc).Error
}

func (sc *SmartContracts) UpdateConditions(conditions string) error {
	return DBConn.Model(sc).Update("conditions", conditions).Error
}

func CreateSmartContractTable(id string) error {
	return DBConn.Exec(`CREATE SEQUENCE "` + id + `_smart_contracts_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_smart_contracts" (
				"id" bigint NOT NULL  default nextval('` + id + `_smart_contracts_id_seq'),
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"wallet_id" bigint  NOT NULL DEFAULT '0',
				"active" character(1) NOT NULL DEFAULT '0',
				"conditions" text  NOT NULL DEFAULT '',
				"variables" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_smart_contracts_id_seq" owned by "` + id + `_smart_contracts".id;
				ALTER TABLE ONLY "` + id + `_smart_contracts" ADD CONSTRAINT "` + id + `_smart_contracts_pkey" PRIMARY KEY (id);
				CREATE INDEX "` + id + `_smart_contracts_index_name" ON "` + id + `_smart_contracts" (name);
				`).Error
}

func CreateSmartContractMainCondition(id string, walletID int64) error {
	return DBConn.Exec(`INSERT INTO "`+id+`_smart_contracts" (name, value, wallet_id, active) VALUES
		(?, ?, ?, ?)`,
		`MainCondition`, `contract MainCondition {
            data {}
            conditions {
                    if(StateVal("gov_account")!=$citizen)
                    {
                        warning "Sorry, you don't have access to this action."
                    }
            }
            action {}
    }`, walletID, 1,
	).Error
}

func (sc *SmartContracts) ToMap() map[string]string {
	result := make(map[string]string)
	result["id"] = string(sc.ID)
	result["name"] = sc.Name
	result["value"] = string(sc.Value)
	result["wallet_id"] = string(sc.WalletID)
	result["active"] = sc.Active
	result["conditions"] = sc.Conditions
	result["variables"] = string(sc.Variables)
	result["rb_id"] = string(sc.RbID)
	return result
}
