package model

type Apps struct {
	Name   string `gorm:"primary_key;not null;size:100"`
	Done   int32  `gorm:"not null"`
	Blocks string `gorm:"not null"`
}

func CreateStateAppsTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_apps" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"done" integer NOT NULL DEFAULT '0',
				"blocks" text  NOT NULL DEFAULT ''
				);
				ALTER TABLE ONLY "` + stateID + `_apps" ADD CONSTRAINT "` + stateID + `_apps_pkey" PRIMARY KEY (name);
			`).Error
}
