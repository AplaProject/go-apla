package controllers

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/schema"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"os"
)

type installStep1Struct struct {
	Lang map[string]string
}

// Шаг 1 - выбор либо стандартных настроек (sqlite и блокчейн с сервера) либо расширенных - pg/mysql и загрузка с нодов
func (c *Controller) InstallStep1() (string, error) {

	c.r.ParseForm()
	installType := c.r.FormValue("type")
	url := c.r.FormValue("url")
	setupPassword := c.r.FormValue("setup_password")
	firstLoad := c.r.FormValue("first_load")
	dbType := c.r.FormValue("db_type")
	dbHost := c.r.FormValue("host")
	dbPort := c.r.FormValue("port")
	dbName := c.r.FormValue("db_name")
	dbUsername := c.r.FormValue("username")
	dbPassword := c.r.FormValue("password")
	sqliteDbUrl := c.r.FormValue("sqlite_db_url")

	if installType == "standard" {
		dbType = "sqlite"
	} else {
		if len(url) == 0 {
			url = consts.BLOCKCHAIN_URL
		}
	}

	if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
		ioutil.WriteFile(*utils.Dir+"/config.ini", []byte(``), 0644)
	}
	confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	//confIni.Set("sql_log", "1")
	confIni.Set("error_log", "1")
	confIni.Set("log_level", "ERROR")
	confIni.Set("log", "0")
	confIni.Set("log_block_id_begin", "0")
	confIni.Set("log_block_id_end", "0")
	confIni.Set("bad_tx_log", "1")
	confIni.Set("nodes_ban_exit", "0")
	confIni.Set("log_tables", "")
	confIni.Set("log_fns", "")
	confIni.Set("sign_hash", "ip")
	confIni.Set("install_type", installType)	
	if len(sqliteDbUrl) > 0 && dbType == "sqlite" {
		utils.SqliteDbUrl = sqliteDbUrl
	}

	if dbType == "sqlite" {
		confIni.Set("db_type", "sqlite")
		confIni.Set("db_user", "")
		confIni.Set("db_host", "")
		confIni.Set("db_port", "")
		confIni.Set("db_password", "")
		confIni.Set("db_name", "")
	} else if dbType == "postgresql" || dbType == "mysql" {
		confIni.Set("db_type", dbType)
		confIni.Set("db_user", dbUsername)
		confIni.Set("db_host", dbHost)
		confIni.Set("db_port", dbPort)
		confIni.Set("db_password", dbPassword)
		confIni.Set("db_name", dbName)
	}

	err = confIni.SaveConfigFile(*utils.Dir + "/config.ini")
	if err != nil {
		return "", err
	}

	log.Debug("sqliteDbUrl: %s", sqliteDbUrl)

	go func() {

		configIni, err = confIni.GetSection("default")

		if dbType == "sqlite" && len(sqliteDbUrl) > 0 {
			if utils.DB != nil && utils.DB.DB != nil {
				utils.DB.Close()
				log.Debug("DB CLOSE")
			}
			for i := 0; i < 5; i++ {
				log.Debug("sqliteDbUrl %v", sqliteDbUrl)
				_, err := utils.DownloadToFile(sqliteDbUrl, *utils.Dir+"/litedb.db", 3600, nil, nil, "install")
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
				if err == nil {
					break
				}
			}
			if err != nil {
				panic(err)
				os.Exit(1)
			}
			utils.DB, err = utils.NewDbConnect(configIni)
			log.Debug("DB OPEN")
			log.Debug("%v", utils.DB)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				panic(err)
				os.Exit(1)
			}
		} else {
			utils.DB, err = utils.NewDbConnect(configIni)
		}

		c.DCDB = utils.DB
		if c.DCDB.DB == nil {
			err = fmt.Errorf("utils.DB == nil")
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		if dbType != "sqlite" || len(sqliteDbUrl) == 0 {
			schema_ := &schema.SchemaStruct{}
			schema_.DCDB = c.DCDB
			schema_.DbType = dbType
			schema_.PrefixUserId = 0
			schema_.GetSchema()

		}

		log.Debug("setupPassword: (%s) / (%s)", setupPassword, utils.DSha256(setupPassword))
		if len(setupPassword) > 0 {
			setupPassword = string(utils.DSha256(setupPassword))
		}
		err = c.DCDB.ExecSql("INSERT INTO config (sqlite_db_url, first_load_blockchain, first_load_blockchain_url, setup_password, auto_reload, chat_enabled) VALUES (?, ?, ?, ?, ?, ?)", sqliteDbUrl, firstLoad, url, setupPassword, 259200, 1)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		err = c.DCDB.ExecSql(`INSERT INTO install (progress) VALUES ('complete')`)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		schema.Migration()

		// если есть значит это тестовый запуск с генерацией 1block
		if _, err := os.Stat(*utils.Dir + "/NodePrivateKey"); err == nil {

			NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
			err = c.DCDB.ExecSql(`INSERT INTO my_node_keys (private_key) VALUES (?)`, NodePrivateKey)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				panic(err)
				os.Exit(1)
			}
			err = c.DCDB.ExecSql(`UPDATE config SET dlt_wallet_id = ?`, 1)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				panic(err)
				os.Exit(1)
			}
		}
	} ()


	utils.Sleep(3) // даем время обновиться config.ini, чтобы в content выдался не installStep0, а updatingBlockchain
	TemplateStr, err := makeTemplate("install_step_1", "installStep1", &installStep1Struct{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
