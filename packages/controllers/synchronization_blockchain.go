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
	"net/http"
	"os"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

var (
	// data may not be updated yet at the first running
	timeSynchro int64 // Когда первый запуск // When the first running
	lastSBlock  int64 // последний блок // last block
	lastSTime   int64
)

// SynchronizationBlockchain synchronizes the blockchain
func (c *Controller) SynchronizationBlockchain() (string, error) {

	if model.DBConn == nil {
		return "", nil
	}
	ib := &model.InfoBlock{}
	err := ib.GetInfoBlock()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))

		var (
			downloadFile, blockURL string
			fileSize               int64
		)
		downloadFile = *utils.Dir + "/public/blockchain"

		if len(c.NodeConfig.FirstLoadBlockchainURL) == 0 {
			blockURL = syspar.SysString(syspar.BlockchainURL)
		}
		resp, err := http.Get(blockURL)
		if err != nil {
			return "", err
		}
		fileSize = resp.ContentLength
		resp.Body.Close()

		// block is downloading
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
			return `{"download": "` + converter.Int64ToStr(int64(converter.RoundWithPrecision(float64((float64(stat.Size())/float64(fileSize))*100), 0))) + `"}`, nil
		}
		return `{"download": "0"}`, nil
	}
	blockID := ib.BlockID
	blockTime := ib.Time

	wTime := int64(12)
	wTimeReady := int64(1)
	if config.ConfigIni["test_mode"] == "1" {
		wTime = 2 * 365 * 86400
		wTimeReady = 2 * 365 * 86400
	}
	now := time.Now().Unix()
	log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, now, ib.Time)
	// if time differs less than for 12 hours from current time, give not affected but those which are in blockchain
	if now-ib.Time < 3600*wTime {
		b := &model.Block{}
		lastBlockData, err := b.GetLastBlockData()
		if err != nil {
			return "", err
		}
		log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
		log.Debug("time.Now().Unix(): %v", time.Now().Unix())
		// if almost all blocks are collected
		if time.Now().Unix()-lastBlockData["lastBlockTime"] < 600*wTimeReady {
			blockID = -1
			blockTime = -1
		}
	}

	confirmation := &model.Confirmation{}
	err = confirmation.GetMaxGoodBlock()
	if err != nil {
		return "", err
	}
	confirmedBlockID := confirmation.BlockID

	currentLoadBlockchain := "nodes"
	if c.NodeConfig.FirstLoadBlockchain == "file" {
		currentLoadBlockchain = c.NodeConfig.FirstLoadBlockchainURL
	}
	var needReload string
	iBlock := blockID
	if timeSynchro == 0 {
		timeSynchro = time.Now().Unix()
		lastSBlock = iBlock
		lastSTime = time.Now().Unix()
	} else if time.Now().Unix()-timeSynchro > 300 { // Here is possible to set 20 minutes or less
		if lastSBlock != iBlock {
			lastSBlock = iBlock
			lastSTime = time.Now().Unix()
		} else if time.Now().Unix()-lastSTime > 60 { // Set the timeout in 60 seconds on the next block
			// There is a sence to check the last block
			if time.Now().Unix()-blockTime > 3600 {
				needReload = `1`
			}
		}
	}

	result := map[string]string{"block_id": converter.Int64ToStr(blockID), "confirmed_block_id": converter.Int64ToStr(confirmedBlockID),
		"block_time": converter.Int64ToStr(blockTime), "current_load_blockchain": currentLoadBlockchain,
		"need_reload": needReload}
	resultJ, _ := json.Marshal(result)

	return string(resultJ), nil
}
