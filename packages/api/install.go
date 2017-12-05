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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/AplaProject/go-apla/packages/conf"

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

	if !conf.WebInstall {
		return fmt.Errorf(`E_INSTALLED`)
	}

	if data.logLevel != "DEBUG" {
		data.logLevel = "ERROR"
	}

	conf.Config.LogLevel = data.logLevel

	if len(data.firstLoadBlockchainURL) == 0 {
		log.WithFields(log.Fields{
			"url": syspar.GetBlockchainURL(),
		}).Info("firstLoadBlockchainURL is not set through POST data, setting it to first load blockchain url from syspar")
		data.firstLoadBlockchainURL = syspar.GetBlockchainURL()
	}

	conf.Config.DB.Host = data.dbHost
	conf.Config.DB.Port = converter.StrToInt(data.dbPort)
	conf.Config.DB.Name = data.dbName
	conf.Config.DB.User = data.dbUsername
	conf.Config.DB.Password = data.dbPassword

	if err := model.InitDB(conf.Config.DB); err != nil {
		return err
	}

	fb := *conf.FirstBlockPath
	if data.firstBlockDir != "" {
		fb = filepath.Join(data.firstBlockDir, "1block")
	}
	if _, err = os.Stat(fb); len(fb) > 0 && os.IsNotExist(err) {
		logger.WithFields(log.Fields{"path": fb}).Info("First block does not exists, generating new keys")
		// If there is no key, this is the first run and the need to create them in the working directory.
		pkf := filepath.Join(conf.Config.PrivateDir, "/PrivateKey")
		if _, err = os.Stat(pkf); os.IsNotExist(err) {
			log.WithFields(log.Fields{"path": pkf}).Info("private key is not exists, generating new one")
			if len(*utils.FirstBlockPublicKey) == 0 {
				log.WithFields(log.Fields{"type": consts.EmptyObject}).Info("first block public key is empty")
				priv, pub, err := crypto.GenHexKeys()
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("generating hex keys")
				}
				err = ioutil.WriteFile(pkf, []byte(priv), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating private key file")
					return err
				}
				err = ioutil.WriteFile(filepath.Join(conf.Config.PrivateDir, "/PublicKey"), []byte(pub), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating public key file")
					return err
				}
				*utils.FirstBlockPublicKey = pub
			}
		}
		npkFile := filepath.Join(conf.Config.PrivateDir, "/NodePrivateKey")
		if _, err = os.Stat(npkFile); os.IsNotExist(err) {
			logger.WithFields(log.Fields{"path": npkFile}).Info("NodePrivateKey does not exists, generating new keys")
			if len(*utils.FirstBlockNodePublicKey) == 0 {
				priv, pub, _ := crypto.GenHexKeys()
				err = ioutil.WriteFile(npkFile, []byte(priv), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("creating NodePrivateKey")
					return err
				}
				err = ioutil.WriteFile(filepath.Join(conf.Config.PrivateDir, "/NodePublicKey"), []byte(pub), 0644)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("creating NodePublicKey")
					return err
				}
				*utils.FirstBlockNodePublicKey = pub
			}
		}
		parser.GenerateFirstBlock()
	}

	if conf.KeyID == 0 {
		key, err := parser.GetKeyIDFromPublicKey()
		if err != nil {
			return err
		}
		conf.KeyID = key
	}

	if err := conf.SaveConfig(); err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("saving config")
		return err
	}

	return daemonsctl.RunAllDaemons()
}

func install(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var result installResult

	data.result = &result

	params := installParams{
		installType:            data.params["type"].(string),
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
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("installCommon")
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	conf.WebInstall = false
	result.Success = true
	return nil
}
