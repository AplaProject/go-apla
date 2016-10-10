// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/astaxie/beego/config"
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
	firstLoad := c.r.FormValue("first_load")
	dbType := c.r.FormValue("db_type")
	dbHost := c.r.FormValue("host")
	dbPort := c.r.FormValue("port")
	dbName := c.r.FormValue("db_name")
	dbUsername := c.r.FormValue("username")
	dbPassword := c.r.FormValue("password")

	if len(url) == 0 {
		url = consts.BLOCKCHAIN_URL
	}

	if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
		ioutil.WriteFile(*utils.Dir+"/config.ini", []byte(``), 0644)
	}
	confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	confIni.Set("log_level", "ERROR")
	confIni.Set("install_type", installType)
	confIni.Set("db_type", dbType)
	confIni.Set("db_user", dbUsername)
	confIni.Set("db_host", dbHost)
	confIni.Set("db_port", dbPort)
	confIni.Set("db_password", dbPassword)
	confIni.Set("db_name", dbName)

	err = confIni.SaveConfigFile(*utils.Dir + "/config.ini")
	if err != nil {
		return "", err
	}

	go func() {

		configIni, err = confIni.GetSection("default")

		utils.DB, err = utils.NewDbConnect(configIni)

		c.DCDB = utils.DB
		if c.DCDB.DB == nil {
			err = fmt.Errorf("utils.DB == nil")
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		schema, err := static.Asset("static/schema.sql")
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		err = c.DCDB.ExecSql(string(schema))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			panic(err)
			os.Exit(1)
		}

		err = c.DCDB.ExecSql("INSERT INTO config (first_load_blockchain, first_load_blockchain_url, auto_reload) VALUES (?, ?, ?)", firstLoad, url, 259200)
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


		// если есть значит это тестовый запуск с генерацией 1block
		if _, err := os.Stat(*utils.Dir + "/NodePrivateKey"); err == nil {

			NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
			err = c.DCDB.ExecSql(`INSERT INTO my_node_keys (private_key, block_id) VALUES (?, ?)`, NodePrivateKey, 1)
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
	}()

	utils.Sleep(3) // даем время обновиться config.ini, чтобы в content выдался не installStep0, а updatingBlockchain
	TemplateStr, err := makeTemplate("install_step_1", "installStep1", &installStep1Struct{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
