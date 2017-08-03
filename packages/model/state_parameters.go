package model

type StateParameter struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Value      string `gorm:"not null"`
	ByteCode   []byte `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp *StateParameter) TableName() string {
	return sp.tableName
}

func (sp *StateParameter) SetTablePrefix(tablePrefix string) {
	sp.tableName = tablePrefix + "_state_parameters"
}

func (sp *StateParameter) GetByName(name string) error {
	return DBConn.Where("name = ?", name).First(sp).Error
}

func (sp *StateParameter) GetByParameter(parameter string) error {
	return DBConn.Where("parameter = ?", parameter).First(sp).Error
}

func (sp *StateParameter) GetAllStateParameters(tablePrefix string) ([]StateParameter, error) {
	parameters := new([]StateParameter)
	err := DBConn.Table(tablePrefix + "_state_parameters").Find(parameters).Error
	if err != nil {
		return nil, err
	}
	return *parameters, nil
}

func (sp *StateParameter) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = sp.Name
	result["value"] = sp.Value
	result["byte_code"] = string(sp.ByteCode)
	result["conditions"] = sp.Conditions
	result["rb_id"] = string(sp.RbID)
	return result
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

func CreateStateAnonymsTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_anonyms" (
				"id_citizen" bigint NOT NULL DEFAULT '0',
				"id_anonym" bigint NOT NULL DEFAULT '0',
				"encrypted" bytea  NOT NULL DEFAULT ''
			    );
			    CREATE INDEX "` + stateID + `_anonyms_index_id" ON "` + stateID + `_anonyms" (id_citizen);`).Error
}
