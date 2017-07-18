package model

type StateParameters struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Value      string `gorm:"not null"`
	ByteCode   []byte `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp *StateParameters) TableName() string {
	return sp.tableName
}

func (sp *StateParameters) SetTableName(tableName string) {
	sp.tableName = tableName
}

func (sp *StateParameters) GetByName(name string) error {
	return DBConn.Where("name = ?", name).First(sp).Error
}

func (sp *StateParameters) GetByParameter(parameter string) error {
	return DBConn.Where("parameter = ?", parameter).First(sp).Error
}

func (sp *StateParameters) GetAllStateParameters(tablePrefix string) ([]StateParameters, error) {
	parameters := new([]StateParameters)
	err := DBConn.Find(parameters).Error
	if err != nil {
		return nil, err
	}
	return *parameters, nil
}

func CreateStateTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_state_parameters" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"bytecode" bytea  NOT NULL DEFAULT '',
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_state_parameters" ADD CONSTRAINT "` + stateID +
		`_state_parameters_pkey" PRIMARY KEY (name);`).Error
}

func CreateStateConditions(stateID string, sid string, psid string, currency string, country string, walletID int64) error {
	return DBConn.Exec(`INSERT INTO "`+stateID+`_state_parameters" (name, value, bytecode, conditions) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?)`,
		"restore_access_condition", sid, "", psid,
		"new_table", sid, "", psid,
		"new_column", sid, "", psid,
		"changing_tables", sid, "", psid,
		"changing_language", sid, "", psid,
		"changing_signature", sid, "", psid,
		"changing_smart_contracts", sid, "", psid,
		"changing_menu", sid, "", psid,
		"changing_page", sid, "", psid,
		"currency_name", currency, "", psid,
		"gender_list", "male,female", "", psid,
		"money_digit", "0", "", psid,
		"tx_fiat_limit", "10", "", psid,
		"state_name", country, "", psid,
		"gov_account", walletID, "", psid,
		"dlt_spending", walletID, "", psid,
		"state_flag", "", "", psid,
		"state_coords", ``, "", psid,
		"citizenship_price", "1000000", "", psid).Error
}
