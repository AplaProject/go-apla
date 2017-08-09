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
	"bytes"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type updatingBlockchainStruct struct {
	Lang            map[string]string
	WaitText        string
	BlockTime       int64
	BlockID         int64
	StartDaemons    string
	BlockMeter      int64
	CheckTime       string
	LastBlock       int64
	BlockChainSize  int64
	Mobile          bool
	AlertTime       string
	RestartDb       bool
	StandardInstall bool
	SleepTime       int64
	NewVersion      string
}

// UpdatingBlockchain is a controller which displays information about updating blockchain
func (c *Controller) UpdatingBlockchain() (string, error) {

	var blockTime, blockID, blockMeter int64
	var waitText, startDaemons, checkTime string
	var restartDb, standardInstall bool

	if c.dbInit {
		confirmation := &model.Confirmation{}
		err := confirmation.GetMaxGoodBlock()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if confirmation.BlockID == 0 {
			if c.NodeConfig.FirstLoadBlockchain == "file" {
				waitText = c.Lang["loading_blockchain_please_wait"]
			} else {
				waitText = c.Lang["is_synchronized_with_the_dc_network"]
			}
		} else {
			b := &model.Block{}
			LastBlockData, err := b.GetLastBlockData()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			blockTime = LastBlockData["lastBlockTime"]
			blockID = LastBlockData["blockId"]
		}

		blockchainURL := c.NodeConfig.FirstLoadBlockchainURL
		if len(blockchainURL) == 0 {
			blockchainURL = syspar.GetBlockchainURL()
		}

		blockMeter = int64(converter.RoundWithPrecision(float64((blockID/consts.LAST_BLOCK)*100), 0))
		if blockMeter > 0 {
			blockMeter--
		}

	} else {
		waitText = c.Lang["loading_blockchain_please_wait"]
	}

	var mobile bool
	if utils.Mobile() {
		mobile = true
	}

	networkTime, err := utils.GetNetworkTime()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	diff := int64(math.Abs(float64(time.Now().Unix() - networkTime.Unix())))
	var alertTime string
	if c.dbInit && diff > consts.ALERT_ERROR_TIME {
		alertTime = strings.Replace(c.Lang["alert_time"], "[sec]", converter.Int64ToStr(diff), -1)
	}

	sleepTime := int64(1500)
	var newVersion string

	if c.dbInit {
		if strings.HasPrefix(c.r.Host, `localhost`) { //c.NodeAdmin
			if updinfo, err := utils.GetUpdVerAndURL(consts.UPD_AND_VER_URL); err == nil && updinfo != nil {
				newVersion = strings.Replace(c.Lang["new_version"], "[ver]", updinfo.Version, -1)
			}
		}
	}

	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	data, err := static.Asset("static/updating_blockchain.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	b := new(bytes.Buffer)
	standardInstall = config.ConfigIni[`install_type`] == `standard`

	t.Execute(b, &updatingBlockchainStruct{SleepTime: sleepTime, StandardInstall: standardInstall, RestartDb: restartDb, Lang: c.Lang,
		WaitText: waitText, BlockID: blockID, BlockTime: blockTime, StartDaemons: startDaemons,
		BlockMeter: blockMeter, CheckTime: checkTime, LastBlock: consts.LAST_BLOCK,
		BlockChainSize: consts.BLOCKCHAIN_SIZE, Mobile: mobile, AlertTime: alertTime, NewVersion: newVersion})

	return b.String(), nil
}
