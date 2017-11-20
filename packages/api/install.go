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

package api

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

type installResult struct {
	Success bool `json:"success"`
}

type installParams struct {
	generateFirstBlock     bool
	installType            string
	logLevel               string
	firstLoadBlockchainURL string
	firstBlockDir          string
	dbHost                 string
	dbPort                 string
	dbName                 string
	dbPassword             string
	dbUsername             string
}

func installCommon(data *installParams, logger *log.Entry) (err error) {
	if IsInstalled() || model.DBConn != nil || config.IsExist() {
		return fmt.Errorf(`E_INSTALLED`)
	}
	if data.generateFirstBlock {
		*utils.GenerateFirstBlock = 1
	}
	if data.logLevel != "DEBUG" {
		data.logLevel = "ERROR"
	}
	if data.installType == `PRIVATE_NET` {
		logger.WithFields(log.Fields{"dir": *utils.Dir}).Info("Because install type is PRIVATE NET, first block dir is set to dir")
		*utils.FirstBlockDir = *utils.Dir
		if len(data.firstBlockDir) > 0 && data.firstBlockDir != "undefined" {
			logger.WithFields(log.Fields{"dir": data.firstBlockDir}).Info("first block dir is sent with data, so set first block dir flag to it")
			*utils.FirstBlockDir = data.firstBlockDir
		}
	}
	if len(data.firstLoadBlockchainURL) == 0 {
		log.WithFields(log.Fields{"url": syspar.GetBlockchainURL()}).Info("firstLoadBlockchainURL is not set through POST data, setting it to first load blockchain url from syspar")
		data.firstLoadBlockchainURL = syspar.GetBlockchainURL()
	}
	dbConfig := config.DBConfig{
		Type:     `postgresql`,
		User:     data.dbUsername,
		Host:     data.dbHost,
		Port:     data.dbPort,
		Password: data.dbPassword,
		Name:     data.dbName,
	}
	err = config.Save(data.logLevel, data.installType, &dbConfig)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("saving config")
		return err
	}
	defer func() {
		if err != nil {
			config.Drop()
		}
	}()
	if err = config.Read(); err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("reading config")
		return err
	}
	err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
	if err != nil || model.DBConn == nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("initializing DB")
		err = fmt.Errorf(`E_DBNIL`)
		return err
	}
	if err = model.DropTables(); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping all tables")
		return err
	}
	if err = model.ExecSchema(); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing db schema")
		return err
	}
	conf := &model.Config{FirstLoadBlockchain: data.installType, FirstLoadBlockchainURL: data.firstLoadBlockchainURL, AutoReload: 259200}
	if err = conf.Create(); err != nil {
		return err
	}
	install := &model.Install{Progress: "complete"}
	if err = install.Create(); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating install")
		return err
	}
	if _, err = os.Stat(*utils.FirstBlockDir + "/1block"); len(*utils.FirstBlockDir) > 0 && os.IsNotExist(err) {
		logger.WithFields(log.Fields{"path": *utils.FirstBlockDir + "/1block"}).Info("First block does not exists, generating new keys")
		// If there is no key, this is the first run and the need to create them in the working directory.
		if _, err = os.Stat(*utils.Dir + "/PrivateKey"); os.IsNotExist(err) {
			log.WithFields(log.Fields{"path": *utils.Dir + "/PrivateKey"}).Info("private key is not exists, generating new one")
			if len(*utils.FirstBlockPublicKey) == 0 {
				log.WithFields(log.Fields{"type": consts.EmptyObject}).Info("first block public key is empty")
				priv, pub, err := crypto.GenHexKeys()
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("generating hex keys")
				}
				err = ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating private key file")
					return err
				}
				*utils.FirstBlockPublicKey = pub
			}
		}
		if _, err = os.Stat(*utils.Dir + "/NodePrivateKey"); os.IsNotExist(err) {
			logger.WithFields(log.Fields{"path": *utils.FirstBlockDir + "/NodePrivateKey"}).Info("NodePrivateKey does not exists, generating new keys")
			if len(*utils.FirstBlockNodePublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				err = ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("generating hex keys")
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
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting private key to public")
		return err
	}
	nodeKeys := &model.MyNodeKey{PrivateKey: string(NodePrivateKey), PublicKey: npubkey, BlockID: 1}
	err = nodeKeys.Create()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating MyNodeKey")
		return err
	}
	if *utils.KeyID == 0 {
		logger.Info("dltWallet is not set from command line, retrieving it from private key file")
		var key []byte
		key, err = ioutil.ReadFile(*utils.Dir + "/PrivateKey")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading private key file")
			return err
		}
		key, err = hex.DecodeString(string(key))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
			return err
		}
		key, err = crypto.PrivateToPublic(key)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting private key to public")
			return err
		}
		*utils.KeyID = crypto.Address(key)
	}
	err = model.UpdateConfig("key_id", *utils.KeyID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("setting config.dlt_wallet_id")
		return err
	}

	err = daemonsctl.RunAllDaemons()
	if err != nil {
		return err
	}

	return nil
}

func install(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var result installResult

	data.result = &result

	params := installParams{installType: data.params["type"].(string),
		logLevel:               data.params["log_level"].(string),
		firstLoadBlockchainURL: data.params["first_load_blockchain_url"].(string),
		dbHost:                 data.params["db_host"].(string),
		dbPort:                 data.params["db_port"].(string),
		dbName:                 data.params["db_name"].(string),
		dbUsername:             data.params["db_user"].(string),
		dbPassword:             data.params["db_pass"].(string),
		firstBlockDir:          data.params["first_block_dir"].(string),
	}
	if val := data.params["generate_first_block"]; val.(int64) == 1 {
		params.generateFirstBlock = true
	}
	err := installCommon(&params, logger)
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	result.Success = true
	return nil
}
