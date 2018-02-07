// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/converter"

	"github.com/GenesisCommunity/go-genesis/packages/conf"

	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/daylight/daemonsctl"
	"github.com/GenesisCommunity/go-genesis/packages/install"
	"github.com/GenesisCommunity/go-genesis/packages/model"

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

	dbHost     string
	dbPort     string
	dbName     string
	dbPassword string
	dbUsername string

	centrifugoSecret string
	centrifugoURL    string
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

	conf.Config.Centrifugo = conf.CentrifugoConfig{
		Secret: data.centrifugoSecret,
		URL:    data.centrifugoURL,
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
		installType:            data.ParamString("type"),
		logLevel:               data.ParamString("log_level"),
		firstLoadBlockchainURL: data.ParamString("first_load_blockchain_url"),
		firstBlockDir:          data.ParamString("first_block_dir"),
		dbHost:                 data.ParamString("db_host"),
		dbPort:                 data.ParamString("db_port"),
		dbName:                 data.ParamString("db_name"),
		dbUsername:             data.ParamString("db_user"),
		dbPassword:             data.ParamString("db_pass"),
		centrifugoSecret:       data.ParamString("centrifugo_secret"),
		centrifugoURL:          data.ParamString("centrifugo_url"),
	}
	if data.ParamInt64("generate_first_block") == 1 {
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
