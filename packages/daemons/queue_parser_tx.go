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
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
 * Берем тр-ии из очереди и обрабатываем
 * */

func QueueParserTx(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "QueueParserTx"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	d.sleepTime = 1

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
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
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

		blockId, err := d.GetBlockID()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if blockId == 0 {
			if d.unlockPrintSleep(utils.ErrInfo("blockId == 0"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// чистим зацикленные
		utils.WriteSelectiveLog("DELETE FROM transactions WHERE verified = 0 AND used = 0 AND counter > 10")
		affect, err := d.ExecSQLGetAffect("DELETE FROM transactions WHERE verified = 0 AND used = 0 AND counter > 10")
		if err != nil {
			utils.WriteSelectiveLog(err)
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		p := new(parser.Parser)
		p.DCDB = d.DCDB
		err = p.AllTxParser()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
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
