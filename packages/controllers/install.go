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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/astaxie/beego/config"
)

// Install is a controller for the installation
func (c *Controller) Install() (string, error) {
	c.r.ParseForm()
	dir := c.r.FormValue("dir")
	if dir != "" {
		*utils.Dir = dir
	}
	generateFirstBlock := c.r.FormValue("generate_first_block")
	if generateFirstBlock != "" {
		*utils.GenerateFirstBlock = converter.StrToInt64(generateFirstBlock)
	}

	installType := c.r.FormValue("type")
	tcpHost := c.r.FormValue("tcp_host")
	if tcpHost != "" {
		*utils.TCPHost = tcpHost
	}
	httpPort := c.r.FormValue("http_port")
	if httpPort != "" {
		*utils.ListenHTTPPort = httpPort
	}
	logLevel := c.r.FormValue("log_level")
	if logLevel != "DEBUG" {
		logLevel = "ERROR"
	}
	firstLoadBlockchainURL := c.r.FormValue("first_load_blockchain_url")
	firstLoad := c.r.FormValue("first_load")
	dbType := c.r.FormValue("db_type")
	dbHost := c.r.FormValue("host")
	dbPort := c.r.FormValue("port")
	dbName := c.r.FormValue("db_name")
	dbUsername := c.r.FormValue("username")
	dbPassword := c.r.FormValue("password")
	firstBlockDir := c.r.FormValue("first_block_dir")

	if firstLoad == `Private-net` {
		*utils.FirstBlockDir = *utils.Dir
		if firstBlockDir != "" {
			*utils.FirstBlockDir = firstBlockDir
		}
	}

	if len(firstLoadBlockchainURL) == 0 {
		firstLoadBlockchainURL = consts.BLOCKCHAIN_URL
	}

	if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
		ioutil.WriteFile(*utils.Dir+"/config.ini", []byte(``), 0644)
	}
	confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	confIni.Set("log_level", logLevel)
	confIni.Set("install_type", installType)
	confIni.Set("dir", *utils.Dir)
	confIni.Set("tcp_host", *utils.TCPHost)
	confIni.Set("http_port", *utils.ListenHTTPPort)
	confIni.Set("first_block_dir", *utils.FirstBlockDir)
	confIni.Set("db_type", dbType)
	confIni.Set("db_user", dbUsername)
	confIni.Set("db_host", dbHost)
	confIni.Set("db_port", dbPort)
	confIni.Set("db_password", dbPassword)
	confIni.Set("db_name", dbName)
	confIni.Set("node_state_id", `*`)

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
	var DB *sql.DCDB
	DB, err = sql.NewDbConnect(configIni)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}
	c.DCDB = DB
	if c.DCDB.DB == nil {
		err = fmt.Errorf("utils.DB == nil")
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = model.DropTables()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	err = model.ExecSchema()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	config := &model.Config{FirstLoadBlockchain: firstLoad, FirstLoadBlockchainURL: firstLoadBlockchainURL, AutoReload: 259200}
	err = config.Create()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	install := &model.Install{Progress: "complete"}
	err = install.Create()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	log.Debug("GenerateFirstBlock", *utils.GenerateFirstBlock)

	if _, err := os.Stat(*utils.FirstBlockDir + "/1block"); len(*utils.FirstBlockDir) > 0 && os.IsNotExist(err) {

		// If there is no key, this is the first run and the need to create them in the working directory.
		if _, err := os.Stat(*utils.Dir + "/PrivateKey"); os.IsNotExist(err) {

			if len(*utils.FirstBlockPublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				err := ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
				*utils.FirstBlockPublicKey = pub
			}
		}

		if _, err := os.Stat(*utils.Dir + "/NodePrivateKey"); os.IsNotExist(err) {
			if len(*utils.FirstBlockNodePublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				fmt.Println("WriteFile " + *utils.Dir + "/NodePrivateKey")
				err := ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
				}
				*utils.FirstBlockNodePublicKey = pub
			}
		}

		*utils.GenerateFirstBlock = 1
		utils.FirstBlock(false)
	}
	log.Debug("1block")

	NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
	npubkey, err := crypto.PrivateToPublic(NodePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	nodeKeys := &model.MyNodeKey{PrivateKey: NodePrivateKey, PublicKey: npubkey, BlockID: 1}
	err = nodeKeys.Create()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	if *utils.DltWalletID == 0 {
		PrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/PrivateKey")
		PrivateHex, _ := hex.DecodeString(string(PrivateKey))
		PublicKeyBytes2, err := crypto.PrivateToPublic(PrivateHex)
		if err != nil {
			log.Fatal(err)
		}
		log.Debug("dlt_wallet_id %d", crypto.Address(PublicKeyBytes2))
		*utils.DltWalletID = crypto.Address(PublicKeyBytes2)
	}

	config.DltWalletID = *utils.DltWalletID
	err = config.Save()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		dropConfig()
		return "", utils.ErrInfo(err)
	}

	return `{"success":1}`, nil
}

func dropConfig() {
	os.Remove(*utils.Dir + "/config.ini")
}
