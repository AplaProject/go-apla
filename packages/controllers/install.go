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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
)

func (c *Controller) Install() (string, error) {

	c.r.ParseForm()
	dir := c.r.FormValue("dir")
	if dir != "" {
		*utils.Dir = dir
	}
	generateFirstBlock := c.r.FormValue("generate_first_block")
	if generateFirstBlock != "" {
		*utils.GenerateFirstBlock = utils.StrToInt64(generateFirstBlock)
	}
	firstBlockDir := c.r.FormValue("first_block_dir")
	*utils.FirstBlockDir = *utils.Dir
	if firstBlockDir != "" {
		*utils.FirstBlockDir = firstBlockDir
	}
	installType := c.r.FormValue("type")
	tcpHost := c.r.FormValue("tcp_host")
	if tcpHost != "" {
		*utils.TcpHost = tcpHost
	}
	httpPort := c.r.FormValue("http_port")
	if httpPort != "" {
		*utils.ListenHttpPort = httpPort
	}
	logLevel := c.r.FormValue("log_level")
	if logLevel != "DEBUG" {
		logLevel = "ERROR"
	}
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
	confIni.Set("log_level", logLevel)
	confIni.Set("install_type", installType)
	confIni.Set("dir", *utils.Dir)
	confIni.Set("tcp_host", *utils.TcpHost)
	confIni.Set("http_port", *utils.ListenHttpPort)
	confIni.Set("first_block_dir", *utils.FirstBlockDir)
	confIni.Set("db_type", dbType)
	confIni.Set("db_user", dbUsername)
	confIni.Set("db_host", dbHost)
	confIni.Set("db_port", dbPort)
	confIni.Set("db_password", dbPassword)
	confIni.Set("db_name", dbName)

	err = confIni.SaveConfigFile(*utils.Dir + "/config.ini")
	if err != nil {
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	configIni, err = confIni.GetSection("default")
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}
	utils.DB, err = utils.NewDbConnect(configIni)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}
	c.DCDB = utils.DB
	if c.DCDB.DB == nil {
		err = fmt.Errorf("utils.DB == nil")
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = c.DCDB.ExecSql(`
	DO $$ DECLARE
	    r RECORD;
	BEGIN
	    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
		EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
	    END LOOP;
	END $$;
	`)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	/*err = c.DCDB.ExecSql(`CREATE SCHEMA public`)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}*/

	schema, err := static.Asset("static/schema.sql")
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = c.DCDB.ExecSql(string(schema))
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = c.DCDB.ExecSql("INSERT INTO config (first_load_blockchain, first_load_blockchain_url, auto_reload) VALUES (?, ?, ?)", firstLoad, url, 259200)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = c.DCDB.ExecSql(`INSERT INTO install (progress) VALUES ('complete')`)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	log.Debug("GenerateFirstBlock", *utils.GenerateFirstBlock)

	if _, err := os.Stat(*utils.FirstBlockDir + "/1block"); os.IsNotExist(err) {

		// If there is no key, this is the first run and the need to create them in the working directory.
		if _, err := os.Stat(*utils.Dir + "/PrivateKey"); os.IsNotExist(err) {

			if len(*utils.FirstBlockPublicKey) == 0 {
				priv, pub := lib.GenKeys()
				err := ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
				*utils.FirstBlockPublicKey = pub
			}
			if len(*utils.FirstBlockNodePublicKey) == 0 {
				priv, pub := lib.GenKeys()
				err := ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
				*utils.FirstBlockNodePublicKey = pub
			}
		}

		utils.FirstBlock(false)

		log.Debug("1block")

		NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
		err = c.DCDB.ExecSql(`INSERT INTO my_node_keys (private_key, block_id) VALUES (?, ?)`, NodePrivateKey, 1)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			dropConfig()
			return "", utils.ErrInfo(err)
		}
		PrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/PrivateKey")
		PrivateHex, _ := hex.DecodeString(string(PrivateKey))
		PublicKeyBytes2 := lib.PrivateToPublic(PrivateHex)
		log.Debug("dlt_wallet_id %d", int64(lib.Address(PublicKeyBytes2)))

		err = c.DCDB.ExecSql(`UPDATE config SET dlt_wallet_id = ?`, int64(lib.Address(PublicKeyBytes2)))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			dropConfig()
			return "", utils.ErrInfo(err)
		}
		/*		err = utils.DaylightRestart()
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					dropConfig()
					return "", utils.ErrInfo(err)
				}*/

	}
	return `{"success":1}`, nil
}

func dropConfig() {
	os.Remove(*utils.Dir + "/config.ini")
}
