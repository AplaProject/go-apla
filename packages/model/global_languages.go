package model

type Languages struct {
	tableName  string
	Name       string `gorm:not null;primary key`
	Res        []byte `gorm:PostgreSQL(jsonb)`
	Conditions string `gorm:not_null`
	RbID       int64  `gorm:not_null`
}

func (l *Languages) SetTableName(tableName string) {
	l.tableName = tableName
}

func (l *Languages) TableName() string {
	return l.tableName
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
