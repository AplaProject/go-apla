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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type updatingBlockchainStruct struct {
	Lang            map[string]string
	WaitText        string
	BlockTime       int64
	BlockId         int64
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

func (c *Controller) UpdatingBlockchain() (string, error) {

	var blockTime, blockId, blockMeter int64
	var waitText, startDaemons, checkTime string
	var restartDb, standardInstall bool

	if c.dbInit {
		ConfirmedBlockId, err := c.DCDB.GetConfirmedBlockID()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if ConfirmedBlockId == 0 {
			firstLoadBlockchain, err := c.DCDB.Single("SELECT first_load_blockchain FROM config").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if firstLoadBlockchain == "file" {
				waitText = c.Lang["loading_blockchain_please_wait"]
			} else {
				waitText = c.Lang["is_synchronized_with_the_dc_network"]
			}
		} else {
			LastBlockData, err := c.DCDB.GetLastBlockData()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			blockTime = LastBlockData["lastBlockTime"]
			blockId = LastBlockData["blockId"]
		}

		nodeConfig, err := c.GetNodeConfig()
		blockchain_url := nodeConfig["first_load_blockchain_url"]
		if len(blockchain_url) == 0 {
			blockchain_url = consts.BLOCKCHAIN_URL
		}
		/*resp, err := http.Get(blockchain_url)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		blockChainSize := resp.ContentLength
		if blockChainSize == 0 {
			blockChainSize = consts.BLOCKCHAIN_SIZE
		}
		defer resp.Body.Close()*/

		blockMeter = int64(utils.Round(float64((blockId/consts.LAST_BLOCK)*100), 0))
		if blockMeter > 0 {
			blockMeter -= 1
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
	diff := int64(math.Abs(float64(utils.Time() - networkTime.Unix())))
	var alertTime string
	if c.dbInit && diff > consts.ALERT_ERROR_TIME {
		alertTime = strings.Replace(c.Lang["alert_time"], "[sec]", utils.Int64ToStr(diff), -1)
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
	standardInstall = configIni[`install_type`] == `standard`

	t.Execute(b, &updatingBlockchainStruct{SleepTime: sleepTime, StandardInstall: standardInstall, RestartDb: restartDb, Lang: c.Lang,
		WaitText: waitText, BlockId: blockId, BlockTime: blockTime, StartDaemons: startDaemons,
		BlockMeter: blockMeter, CheckTime: checkTime, LastBlock: consts.LAST_BLOCK,
		BlockChainSize: consts.BLOCKCHAIN_SIZE, Mobile: mobile, AlertTime: alertTime, NewVersion: newVersion})

	return b.String(), nil
}
