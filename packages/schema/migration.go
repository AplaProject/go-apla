package schema

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func Migration() {
	oldDbVersion, err := utils.DB.Single(`SELECT version FROM migration_history ORDER BY id DESC LIMIT 1`).String()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	if len(*utils.OldVersion) == 0 && consts.VERSION != oldDbVersion {
		*utils.OldVersion = oldDbVersion
	}

	log.Debug("*utils.OldVersion %v", *utils.OldVersion)
	if len(*utils.OldVersion) > 0 {
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("1.0.2b5") {
			log.Debug("%v", "ALTER TABLE config ADD COLUMN analytics_disabled smallint")
			err = utils.DB.ExecSql(`ALTER TABLE config ADD COLUMN analytics_disabled smallint`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.0.1b2") {
			log.Debug("%v", "ALTER TABLE config ADD COLUMN sqlite_db_url varchar(255)")
			err = utils.DB.ExecSql(`ALTER TABLE config ADD COLUMN sqlite_db_url varchar(255)`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.1.0a13") {
			community, err := utils.DB.GetCommunityUsers()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			if len(community) > 0 {
				for i := 0; i < len(community); i++ {
					err = utils.DB.ExecSql(`ALTER TABLE ` + utils.Int64ToStr(community[i]) + `_my_table ADD COLUMN pool_user_id int NOT NULL DEFAULT '0'`)
					if err != nil {
						log.Error("%v", utils.ErrInfo(err))
					}
				}
			} else {
				log.Debug(`ALTER TABLE my_table ADD COLUMN pool_user_id int NOT NULL DEFAULT '0'`)
				err = utils.DB.ExecSql(`ALTER TABLE my_table ADD COLUMN pool_user_id int NOT NULL DEFAULT '0'`)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
			}

			log.Debug(`ALTER TABLE config ADD COLUMN stat_host varchar(255)`)
			err = utils.DB.ExecSql(`ALTER TABLE config ADD COLUMN stat_host varchar(255)`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = utils.DB.ExecSql(`ALTER TABLE miners_data ADD COLUMN i_am_pool int NOT NULL DEFAULT '0'`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = utils.DB.ExecSql(`ALTER TABLE miners_data ADD COLUMN pool_user_id int  NOT NULL DEFAULT '0'`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = utils.DB.ExecSql(`ALTER TABLE miners_data ADD COLUMN pool_count_users int  NOT NULL DEFAULT '0'`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = utils.DB.ExecSql(`ALTER TABLE log_miners_data ADD COLUMN pool_user_id int NOT NULL DEFAULT '0'`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			err = utils.DB.ExecSql(`ALTER TABLE log_miners_data ADD COLUMN backup_pool_users text NOT NULL DEFAULT ''`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}

			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB

			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('auto_payments_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "commission", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "last_payment_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Когда был последний платеж. При создании авто-платежа пишется текущее время"}
			s2[5] = map[string]string{"name": "period", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s2[6] = map[string]string{"name": "sender", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s2[7] = map[string]string{"name": "recipient", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s2[8] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
			s2[9] = map[string]string{"name": "block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "Для отката новой записи об авто-платеже"}
			s2[10] = map[string]string{"name": "del_block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "Чистим по крону старые данные раз в сутки. Удалять нельзя, т.к. нужно откатывать"}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["auto_payments"] = s1
			schema_.S = s
			schema_.PrintSchema()

			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_ca_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["log_time_auto_payments"] = s1
			schema_.S = s
			schema_.PrintSchema()

			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_ca_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["log_time_del_user_from_pool"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.1.0a16") {
			err = utils.DB.ExecSql(`ALTER TABLE miners_data ADD COLUMN backup_pool_users text NOT NULL DEFAULT ''`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.1.0a23") {

			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "day", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[1] = map[string]string{"name": "month", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "year", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "dc", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[5] = map[string]string{"name": "promised_amount", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["AI"] = "id"
			s1["PRIMARY"] = []string{"day", "month", "year", "currency_id"}
			s1["comment"] = ""
			s["stats"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.3a1") {

			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('stats_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "day", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "month", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "year", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[5] = map[string]string{"name": "dc", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[6] = map[string]string{"name": "promised_amount", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["UNIQ"] = []string{"day", "month", "year", "currency_id"}
			s1["PRIMARY"] = []string{"id"}
			s1["comment"] = ""
			s["stats"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.3a3") {

			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('migration_history_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "version", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "date_applied", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["migration_history"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.3a4") {

			err = utils.DB.ExecSql(`ALTER TABLE migration_history ADD COLUMN test_migration int NOT NULL DEFAULT '0'`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.3a8") {
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('migration_history_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["promised_amount_restricted"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.4a1") {
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_arbitrator_conditions_log_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "last_payment_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "prev_log_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"log_id"}
			s1["AI"] = "log_id"
			s1["comment"] = ""
			s["log_auto_payments"] = s1
			schema_.S = s
			schema_.PrintSchema()

			err = utils.DB.ExecSql(`ALTER TABLE auto_payments ADD COLUMN log_id bigint(20) NOT NULL DEFAULT '0';`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.4a2") {

			err = utils.DB.ExecSql(`ALTER TABLE config ADD COLUMN getpool_host varchar(255)`)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.2.5b3") {
			log.Debug("*utils.OldVersion", *utils.OldVersion)
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_user_upgrade_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["log_time_user_upgrade"] = s1
			schema_.S = s
			schema_.PrintSchema()


			schema_ = &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "sn_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
			s2[1] = map[string]string{"name": "sn_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
			s2[2] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[5] = map[string]string{"name": "status", "mysql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "comment": ""}
			s2[6] = map[string]string{"name": "sn_attempts", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s["users"] = s1
			schema_.S = s
			schema_.AddColumn = true
			schema_.PrintSchema()

			schema_ = &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "sn_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
			s2[1] = map[string]string{"name": "sn_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
			s2[2] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[5] = map[string]string{"name": "status", "mysql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "comment": ""}
			s2[6] = map[string]string{"name": "sn_attempts", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s["log_users"] = s1
			schema_.S = s
			schema_.AddColumn = true
			schema_.PrintSchema()

			schema_ = &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Кто голосует"}
			s2[1] = map[string]string{"name": "voting_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "За что голосует. тут может быть id geolocation и пр"}
			s2[2] = map[string]string{"name": "type", "mysql": "enum('null','votes_miners','promised_amount','sn_user') NOT NULL", "sqlite": "varchar(100)  NOT NULL", "postgresql": "enum('null','votes_miners','promised_amount','sn_user') NOT NULL", "comment": "Нужно для voting_id' DEFAULT 'null"}
			s2[3] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было удаление. Нужно для чистки по крону старых данных и для откатов."}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"user_id", "voting_id", "type"}
			s1["comment"] = "Чтобы 1 юзер не смог проголосовать 2 раза за одно и тоже"
			s["log_votes"] = s1
			schema_.S = s
			schema_.ChangeType = true
			schema_.PrintSchema()


			schema_ = &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s = make(Recmap)
			s1 = make(Recmap)
			s2 = make(Recmapi)
			s2[0] = map[string]string{"name": "dc_amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Списанная сумма намайненного"}
			s2[1] = map[string]string{"name": "last_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время последнего перевода намайненного на счет"}
			s2[2] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s["promised_amount_restricted"] = s1
			schema_.S = s
			schema_.AddColumn = true
			schema_.PrintSchema()
		}

		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.3.1b5") {
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_promised_amount_log_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "dc_amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Списанная сумма намайненного"}
			s2[2] = map[string]string{"name": "last_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время последнего перевода намайненного на счет"}
			s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
			s2[4] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s1["fields"] = s2
			s1["PRIMARY"] = []string{"log_id"}
			s1["AI"] = "log_id"
			s1["comment"] = ""
			s["log_promised_amount_restricted"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.3.3b2") {
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('notifications_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[3] = map[string]string{"name": "cmd_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[4] = map[string]string{"name": "params", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}

			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["notifications"] = s1
			schema_.S = s
			schema_.PrintSchema()
		}
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.3.4b3") {
			schema_ := &SchemaStruct{}
			schema_.DbType = utils.DB.ConfigIni["db_type"]
			schema_.DCDB = utils.DB
			s := make(Recmap)
			s1 := make(Recmap)
			s2 := make(Recmapi)
			s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('notifications_id_seq')", "comment": ""}
			s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[2] = map[string]string{"name": "subject", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
			s2[3] = map[string]string{"name": "topic", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
			s2[4] = map[string]string{"name": "idroot", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[5] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
			s2[6] = map[string]string{"name": "status", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
			s2[7] = map[string]string{"name": "uptime", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}

			s1["fields"] = s2
			s1["PRIMARY"] = []string{"id"}
			s1["AI"] = "id"
			s1["comment"] = ""
			s["e_tickets"] = s1
			schema_.S = s
			schema_.PrintSchema()
			
			schema_.DB.Exec(`CREATE INDEX e_ticket_idroot ON e_tickets (idroot)`)
			schema_.DB.Exec(`CREATE INDEX e_ticket_uptime ON e_tickets (uptime)`)
		}
		if utils.VersionOrdinal(*utils.OldVersion) < utils.VersionOrdinal("2.3.4b4") {
			if err = utils.DB.ExecSql(`ALTER TABLE notifications ADD COLUMN isread tinyint(3) NOT NULL DEFAULT '0'`); err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			if err = utils.DB.ExecSql(`CREATE INDEX notifications_ur ON notifications (user_id,isread)`); err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
		}
		err = utils.DB.ExecSql(`INSERT INTO migration_history (version, date_applied) VALUES (?, ?)`, consts.VERSION, utils.Time())
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
	}
}

