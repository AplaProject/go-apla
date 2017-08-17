package model

import "strconv"

type Language struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Res        string `gorm:"type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gotm:"not null"`
}

func (l *Language) SetTablePrefix(tablePrefix string) {
	l.tableName = tablePrefix + "_languages"
}

func (l *Language) TableName() string {
	return l.tableName
}

func (l *Language) Get(name string) error {
	return DBConn.Where("name = ?", name).First(l).Error
}

func (l *Language) GetAll(prefix string) ([]Language, error) {
	var result []Language
	err := DBConn.Table(prefix + "_languages").Order("name").Find(result).Error
	return result, err
}

func (l *Language) IsExistsByName(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(l)
	return !query.RecordNotFound(), query.Error
}

func CreateLanguagesStateTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_languages" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"res" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_languages" ADD CONSTRAINT "` + stateID + `_languages_pkey" PRIMARY KEY (name);
		`).Error
}

func CreateStateDefaultLanguages(stateID, conditions string) error {
	return DBConn.Exec(`INSERT INTO "`+stateID+`_languages" (name, res, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?)`,
		`dateformat`, `{"en": "YYYY-MM-DD", "ru": "DD.MM.YYYY"}`, conditions,
		`timeformat`, `{"en": "YYYY-MM-DD HH:MI:SS", "ru": "DD.MM.YYYY HH:MI:SS"}`, conditions,
		`Gender`, `{"en": "Gender", "ru": "Пол"}`, conditions,
		`male`, `{"en": "Male", "ru": "Мужской"}`, conditions,
		`female`, `{"en": "Female", "ru": "Женский"}`, conditions).Error
}

func (l *Language) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = l.Name
	result["res"] = l.Res
	result["conditions"] = l.Conditions
	result["rb_id"] = strconv.FormatInt(l.RbID, 10)
	return result
}
