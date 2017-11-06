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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"
)

type installResult struct {
	Success bool `json:"success"`
}

type InstallParams struct {
	GenerateFirstBlock     bool
	InstallType            string
	LogLevel               string
	FirstLoadBlockchainURL string
	FirstBlockDir          string
	DbHost                 string
	DbPort                 string
	DbName                 string
	DbPassword             string
	DbUsername             string
}

func InstallCommon(data *InstallParams) (err error) {
	if data.GenerateFirstBlock {
		*utils.GenerateFirstBlock = 1
	}
	if data.LogLevel == "" {
		data.LogLevel = "INFO"
	}
	if data.InstallType == `PRIVATE_NET` {
		*utils.FirstBlockDir = *utils.Dir
		if len(data.FirstBlockDir) > 0 {
			*utils.FirstBlockDir = data.FirstBlockDir
		}
	}

	if len(data.FirstLoadBlockchainURL) == 0 {
		data.FirstLoadBlockchainURL = syspar.GetBlockchainURL()
	}
	dbConfig := config.DBConfig{
		Type:     `postgresql`,
		User:     data.DbUsername,
		Host:     data.DbHost,
		Port:     data.DbPort,
		Password: data.DbPassword,
		Name:     data.DbName,
	}
	err = config.Save(data.LogLevel, data.InstallType, &dbConfig)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			config.Drop()
		}
	}()
	if err = config.Read(); err != nil {
		return err
	}
	err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
	if err != nil || model.DBConn == nil {
		fmt.Printf("err = %s\n", err)
		return fmt.Errorf(`E_DBNIL`)
	}
	if err = model.DropTables(); err != nil {
		return err
	}
	if err = model.ExecSchema(); err != nil {
		return err
	}

	conf := &model.Config{
		FirstLoadBlockchain:    data.InstallType,
		FirstLoadBlockchainURL: data.FirstLoadBlockchainURL,
		AutoReload:             259200,
	}
	if err = conf.Create(); err != nil {
		return err
	}
	install := &model.Install{Progress: "complete"}
	if err = install.Create(); err != nil {
		return err
	}
	if _, err = os.Stat(*utils.FirstBlockDir + "/1block"); len(*utils.FirstBlockDir) > 0 && os.IsNotExist(err) {
		// If there is no key, this is the first run and the need to create them in the working directory.
		if _, err = os.Stat(*utils.Dir + "/PrivateKey"); os.IsNotExist(err) {
			if len(*utils.FirstBlockPublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				err = ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
				if err != nil {
					return
				}
				*utils.FirstBlockPublicKey = pub
			}
		}
		if _, err = os.Stat(*utils.Dir + "/NodePrivateKey"); os.IsNotExist(err) {
			if len(*utils.FirstBlockNodePublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				err = ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
				if err != nil {
					return err
				}
				*utils.FirstBlockNodePublicKey = pub
			}
		}
		*utils.GenerateFirstBlock = 1
		parser.FirstBlock()
	}

	NodePrivateKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePrivateKey")
	var npubkey []byte
	npubkey, err = crypto.PrivateToPublic(NodePrivateKey)
	if err != nil {
		return err
	}
	nodeKeys := &model.MyNodeKey{PrivateKey: string(NodePrivateKey), PublicKey: npubkey, BlockID: 1}
	err = nodeKeys.Create()
	if err != nil {
		return err
	}
	if *utils.DltWalletID == 0 {
		var key []byte
		key, err = ioutil.ReadFile(*utils.Dir + "/PrivateKey")
		if err != nil {
			return err
		}
		key, err = hex.DecodeString(string(key))
		if err != nil {
			return err
		}
		key, err = crypto.PrivateToPublic(key)
		if err != nil {
			return err
		}
		*utils.DltWalletID = crypto.Address(key)
	}
	err = model.UpdateConfig("dlt_wallet_id", *utils.DltWalletID)
	if err != nil {
		return err
	}
	return nil
}

func install(w http.ResponseWriter, r *http.Request, data *apiData) error {
	if installed || model.DBConn != nil || config.IsExist() {
		return errorAPI(w, "E_INSTALLED", http.StatusInternalServerError)
	}

	var result installResult
	data.result = &result

	params := InstallParams{InstallType: data.params["type"].(string),
		LogLevel:               data.params["log_level"].(string),
		FirstLoadBlockchainURL: data.params["first_load_blockchain_url"].(string),
		DbHost:                 data.params["db_host"].(string),
		DbPort:                 data.params["db_port"].(string),
		DbName:                 data.params["db_name"].(string),
		DbUsername:             data.params["db_user"].(string),
		DbPassword:             data.params["db_pass"].(string),
		FirstBlockDir:          data.params["first_block_dir"].(string),
	}
	if val := data.params["generate_first_block"]; val.(int64) == 1 {
		params.GenerateFirstBlock = true
	}
	err := InstallCommon(&params)
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	err = daemonsctl.RunAllDaemons()
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	result.Success = true
	return nil
}
