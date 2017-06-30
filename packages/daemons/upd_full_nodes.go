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

package daemons

import (
	"fmt"
	"time"

	"log"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// UpdFullNodes sends UpdFullNodes transactions
func UpdFullNodes(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "UpdFullNodes"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	d.sleepTime = 60
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}

BEGIN:
	for {
		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, converter.Int64ToStr(time.Now().Unix())}

		// проверим, не нужно ли нам выйти из цикла
		// check if we have to break the cycle
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		restart, err := d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		blockID, err := d.GetBlockID()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if blockID == 0 {
			if d.unlockPrintSleep(utils.ErrInfo("blockID == 0"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		myStateID, myWalletID, err := d.GetMyStateIDAndWalletID()
		logger.Debug("%v", myWalletID)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// Есть ли мы в списке тех, кто может генерить блоки
		// If we are in the list of those who are able to generate the blocks
		fullNodeID, err := d.FindInFullNodes(myStateID, myWalletID)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("fullNodeID = %d", fullNodeID)
		if fullNodeID == 0 {
			d.dbUnlock()
			logger.Debug("fullNodeID == 0")
			d.sleepTime = 10 // because 1s is too small for non-full nodes
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		curTime := time.Now().Unix()

		// проверим, прошло ли время с момента последнего обновления
		// check if the time of the last updating passed
		updFullNodes, err := d.Single("SELECT time FROM upd_full_nodes").Int64()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if curTime-updFullNodes <= consts.UPD_FULL_NODES_PERIOD {
			if d.unlockPrintSleep(utils.ErrInfo("curTime-adminTime <= consts.UPD_FULL_NODES_PERIO"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		forSign := fmt.Sprintf("%v,%v,%v,%v", utils.TypeInt("UpdFullNodes"), curTime, myWalletID, 0)
		binSign, err := d.GetBinSign(forSign)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		data := converter.DecToBin(utils.TypeInt("UpdFullNodes"), 1)
		data = append(data, converter.DecToBin(curTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(myWalletID)...)
		data = append(data, converter.EncodeLengthPlusData(0)...)
		data = append(data, converter.EncodeLengthPlusData([]byte(binSign))...)

		err = d.InsertReplaceTxInQueue(data)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		p := new(parser.Parser)
		p.DCDB = d.DCDB
		hash, err := crypto.Hash(data)
		if err != nil {
			log.Fatal(err)
		}
		hash = converter.BinToHex(hash)
		err = p.TxParser(converter.HexToBin(hash), data, true)
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
