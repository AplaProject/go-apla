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

package apiv2

import (
	"net/http"
	/*	"encoding/hex"
		"fmt"
		"io/ioutil"
		"os"

		"github.com/EGaaS/go-egaas-mvp/packages/config"
		"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
		"github.com/EGaaS/go-egaas-mvp/packages/crypto"
		"github.com/EGaaS/go-egaas-mvp/packages/model"
		"github.com/EGaaS/go-egaas-mvp/packages/utils"*/)

type installResult struct {
	Success bool `json:"success"`
}

// If State == 0 then APLA has not been installed
// If Wallet == 0 then login is required

func install(w http.ResponseWriter, r *http.Request, data *apiData) error {
	/*	var result installResult

		data.result = &result
		if installed || model.DBConn != nil || config.IsExist() {
			return errorAPI(w, fmt.Sprintf(`Apla is already installed`), http.StatusInternalServerError)
		}

		if val := data.params["dir"]; len(val.(string)) > 0 {
			*utils.Dir = val.(string)
		}
		if val := data.params["generate_first_block"]; val.(int64) == 1 {
			*utils.GenerateFirstBlock = val.(int64)
		}
		installType := data.params["type"].(string)
		if val := data.params["tcp_host"]; len(val.(string)) > 0 {
			*utils.TCPHost = val.(string)
		}
		if val := data.params["http_port"]; len(val.(string)) > 0 {
			*utils.ListenHTTPPort = val.(string)
		}
		logLevel := data.params["log_level"].(string)
		if logLevel != "DEBUG" {
			logLevel = "ERROR"
		}
		firstLoadBlockchainURL := data.params["first_load_blockchain_url"].(string)
		firstLoad := data.params["first_load"].(string)
		dbType := data.params["db_type"].(string)
		dbHost := data.params["host"].(string)
		dbPort := data.params["port"].(string)
		dbName := data.params["db_name"].(string)
		dbUsername := data.params["username"].(string)
		dbPassword := data.params["password"].(string)
		firstBlockDir := data.params["first_block_dir"].(string)

		if firstLoad == `Private-net` {
			*utils.FirstBlockDir = *utils.Dir
			if len(firstBlockDir) > 0 {
				*utils.FirstBlockDir = firstBlockDir
			}
		}

		if len(firstLoadBlockchainURL) == 0 {
			firstLoadBlockchainURL = syspar.GetBlockchainURL()
		}
		dbConfig := config.DBConfig{
			Type:     dbType,
			User:     dbUsername,
			Host:     dbHost,
			Port:     dbPort,
			Password: dbPassword,
			Name:     dbName,
		}
		err := config.Save(logLevel, installType, &dbConfig)
		if err != nil {
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		err = config.Read()
		if err != nil {
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
		if err != nil {
			log.Errorf("db error: %s", err)
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		if model.DBConn == nil {
			err = fmt.Errorf("utils.DB == nil")
			log.Error("%v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		err = model.DropTables()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		err = model.ExecSchema()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		conf := &model.Config{FirstLoadBlockchain: firstLoad, FirstLoadBlockchainURL: firstLoadBlockchainURL, AutoReload: 259200}
		err = conf.Create()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		install := &model.Install{Progress: "complete"}
		err = install.Create()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

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

			log.Debugf("start to generate first block")
			*utils.GenerateFirstBlock = 1
			utils.FirstBlock()
		}

		NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
		npubkey, err := crypto.PrivateToPublic(NodePrivateKey)
		if err != nil {
			log.Fatal(err)
		}
		nodeKeys := &model.MyNodeKey{PrivateKey: NodePrivateKey, PublicKey: npubkey, BlockID: 1}
		err = nodeKeys.Create()
		if err != nil {
			log.Error("my_node_key insert failed: %v", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
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

		err = model.UpdateConfig("dlt_wallet_id", *utils.DltWalletID)
		if err != nil {
			log.Errorf("can't update config: %s", utils.ErrInfo(err))
			config.Drop()
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}

		result.Success = true*/
	return nil
}
