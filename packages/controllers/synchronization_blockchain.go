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
	"encoding/json"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"net/http"
	"os"
	"time"
)

var (
	// при запуске данные могут еще не успеть обновиться
	// data may not be updated yet at the first running
	timeSynchro int64 // Когда первый запуск // When the first running
	lastSBlock  int64 // последний блок // last block
	lastSTime   int64
)

func (c *Controller) SynchronizationBlockchain() (string, error) {

	if c.DCDB == nil || c.DCDB.DB == nil {
		return "", nil
	}
	blockData, err := c.DCDB.GetInfoBlock()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))

		var (
			downloadFile, blockUrl string
			fileSize               int64
		)
		downloadFile = *utils.Dir + "/public/blockchain"
		nodeConfig, err := c.GetNodeConfig()
		if err != nil {
			return "", err
		}
		blockUrl = nodeConfig["first_load_blockchain_url"]
		if len(blockUrl) == 0 {
			blockUrl = consts.BLOCKCHAIN_URL
		}
		resp, err := http.Get(blockUrl)
		if err != nil {
			return "", err
		}
		fileSize = resp.ContentLength
		resp.Body.Close()

		// качается блок // block is downloading
		file, err := os.Open(downloadFile)
		if err != nil {
			return "", err
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			return "", err
		}
		if stat.Size() > 0 {
			log.Debug("stat.Size(): %v", int(stat.Size()))
			return `{"download": "` + utils.Int64ToStr(int64(utils.Round(float64((float64(stat.Size())/float64(fileSize))*100), 0))) + `"}`, nil
		} else {
			return `{"download": "0"}`, nil
		}
	}
	blockId := blockData["block_id"]
	blockTime := blockData["time"]
	if len(blockId) == 0 {
		blockId = "0"
	}
	if len(blockTime) == 0 {
		blockTime = "0"
	}

	wTime := int64(12)
	wTimeReady := int64(1)
	if c.ConfigIni["test_mode"] == "1" {
		wTime = 2 * 365 * 86400
		wTimeReady = 2 * 365 * 86400
	}
	log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, utils.Time(), utils.StrToInt64(blockData["time"]))
	// если время менее 12 часов от текущего, то выдаем не подвержденные, а просто те, что есть в блокчейне
	// if time differs less than for 12 hours from current time, give not affected but those which are in blockchain
	if utils.Time()-utils.StrToInt64(blockData["time"]) < 3600*wTime {
		lastBlockData, err := c.DCDB.GetLastBlockData()
		if err != nil {
			return "", err
		}
		log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
		log.Debug("time.Now().Unix(): %v", time.Now().Unix())
		// если уже почти собрали все блоки
		if time.Now().Unix()-lastBlockData["lastBlockTime"] < 600*wTimeReady {
			blockId = "-1"
			blockTime = "-1"
		}
	}

	confirmedBlockId, err := c.GetConfirmedBlockID()
	if err != nil {
		return "", err
	}

	currentLoadBlockchain := "nodes"
	if c.NodeConfig["current_load_blockchain"] == "file" {
		currentLoadBlockchain = c.NodeConfig["first_load_blockchain_url"]
	}
	var needReload string
	iBlock := utils.StrToInt64(blockId)
	if timeSynchro == 0 {
		timeSynchro = utils.Time()
		lastSBlock = iBlock
		lastSTime = utils.Time()
	} else if utils.Time()-timeSynchro > 300 { // Тут можно поставить минут 20 или меньше // Here is possible to set 20 minutes or less
		if lastSBlock != iBlock {
			lastSBlock = iBlock
			lastSTime = utils.Time()
		} else if utils.Time()-lastSTime > 60 { // Ставим timeout на очередной блок в 60 секунд // Set the timeout in 60 seconds on the next block
			// Имеет смысл проверять последний блок // There is a sence to check the last block
			if utils.Time()-utils.StrToInt64(blockTime) > 3600 {
				needReload = `1`
			}
		}
	}

	result := map[string]string{"block_id": blockId, "confirmed_block_id": utils.Int64ToStr(confirmedBlockId),
		"block_time": blockTime, "current_load_blockchain": currentLoadBlockchain,
		"need_reload": needReload}
	resultJ, _ := json.Marshal(result)

	return string(resultJ), nil
}
