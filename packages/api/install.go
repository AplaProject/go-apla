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
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/daylight/daemonsctl"
	"github.com/AplaProject/go-apla/packages/install"
	"github.com/AplaProject/go-apla/packages/model"

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

	if conf.Installed {
		return fmt.Errorf(`E_INSTALLED`)
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
		if err == model.ErrDBConn {
			return fmt.Errorf(`E_DBNIL`)
		}
		return err
	}

	if !install.IsExistFirstBlock() {
		err = install.GenerateFirstBlock()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("GenerateFirstBlock")
			return err
		}
	}

	if err := conf.SaveConfig(); err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("saving config")
		return err
	}

	return daemonsctl.RunAllDaemons()
}

func doInstall(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
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

	conf.Installed = true
	result.Success = true
	return nil
}
