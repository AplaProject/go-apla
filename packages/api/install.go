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

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/daylight/daemonsctl"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type installResult struct {
	Success bool `json:"success"`
}

type installParams struct {
	installType   string
	logLevel      string
	firstBlockDir string

	dbHost     string
	dbPort     string
	dbName     string
	dbPassword string
	dbUsername string

	centrifugoSecret string
	centrifugoURL    string
}

func installCommon(data *installParams, logger *log.Entry) (err error) {

	if conf.Config.Installed {
		return fmt.Errorf(`E_INSTALLED`)
	}

	conf.Config.LogConfig.LogLevel = data.logLevel

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

	conf.Config.Centrifugo = conf.CentrifugoConfig{
		Secret: data.centrifugoSecret,
		URL:    data.centrifugoURL,
	}

	if err := conf.SaveConfig(conf.Config.ConfigPath); err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("saving config")
		return err
	}

	return daemonsctl.RunAllDaemons()
}

func doInstall(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var result installResult

	data.result = &result

	params := installParams{
		installType:      data.ParamString("type"),
		logLevel:         data.ParamString("log_level"),
		dbHost:           data.ParamString("db_host"),
		dbPort:           data.ParamString("db_port"),
		dbName:           data.ParamString("db_name"),
		dbUsername:       data.ParamString("db_user"),
		dbPassword:       data.ParamString("db_pass"),
		centrifugoSecret: data.ParamString("centrifugo_secret"),
		centrifugoURL:    data.ParamString("centrifugo_url"),
	}
	err := installCommon(&params, logger)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "error": err}).Error("installCommon")
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	conf.Config.Installed = true
	result.Success = true
	return nil
}
