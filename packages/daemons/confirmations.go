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
	"net"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
Getting amount of nodes, which has the same hash as we do
Using it for watching for forks
Получаем кол-во нодов, у которых такой же хэш последнего блока как и у нас
Нужно чтобы следить за вилками
*/

// Confirmations gets and checks blocks from nodes
func Confirmations(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "Confirmations"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}

	var s int

BEGIN:
	for {
		// первые 2 минуты спим по 10 сек, чтобы блоки успели собраться
		// the first 2 minutes we sleep for 10 sec for blocks to be collected
		s++

		d.sleepTime = 1

		if s < 12 {
			d.sleepTime = 1
		}

		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, converter.Int64ToStr(time.Now().Unix())}

		// проверим, не нужно ли нам выйти из цикла
		// check if we have to break the cycle
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		var startBlockID int64
		// если последний проверенный был давно (пропасть более 5 блоков),
		// то начинаем проверку последних 5 блоков
		// if the last one checked was long ago (interval is more than 5 blocks)
		confirmations := &model.Confirmations{}
		err := confirmations.GetGoodBlock(consts.MIN_CONFIRMED_NODES)
		if err != nil {
			logger.Error("%v", err)
		}
		ConfirmedBlockID := confirmations.BlockID
		infoBlock := &model.InfoBlock{}
		err = infoBlock.GetInfoBlock()
		if err != nil {
			logger.Error("%v", err)
		}
		LastBlockID := infoBlock.BlockID
		if LastBlockID-ConfirmedBlockID > 5 {
			startBlockID = ConfirmedBlockID + 1
			d.sleepTime = 10
			s = 0 // 2 минуты отчитываем с начала
			// count 2 minutes from the beginning
		}
		if startBlockID == 0 {
			startBlockID = LastBlockID - 1
		}
		logger.Debug("startBlockID: %d / LastBlockID: %d", startBlockID, LastBlockID)

		for blockID := LastBlockID; blockID > startBlockID; blockID-- {

			// проверим, не нужно ли нам выйти из цикла
			// check if we have to break the cycle
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				break BEGIN
			}

			logger.Debug("blockID: %d", blockID)

			block := model.Block{}
			err := block.GetBlock(blockID)
			if err != nil {
				logger.Error("%v", err)
			}
			hash := string(block.Hash)
			logger.Info("hash: %x", hash)
			if len(hash) == 0 {
				logger.Debug("len(hash) == 0")
				continue
			}

			var hosts []string
			if d.ConfigIni["test_mode"] == "1" {
				hosts = []string{"localhost:" + consts.TCP_PORT}
			} else {
				hosts, err = model.GetFullNodesHosts()
				if err != nil {
					logger.Error("%v", err)
				}
			}

			ch := make(chan string)
			for i := 0; i < len(hosts); i++ {
				host := hosts[i] + ":" + consts.TCP_PORT
				logger.Info("host %v", host)
				go func() {
					IsReachable(host, blockID, ch)
				}()
			}
			var answer string
			var st0, st1 int64
			for i := 0; i < len(hosts); i++ {
				answer = <-ch
				logger.Info("answer == hash (%x = %x)", answer, hash)
				logger.Info("answer == hash (%s = %s)", answer, hash)
				if answer == hash {
					st1++
				} else {
					st0++
				}
				logger.Info("st0 %v  st1 %v", st0, st1)
			}
			confirmation := &model.Confirmations{}
			err = confirmation.GetConfirmation(blockID)
			if err == nil {
				logger.Debug("UPDATE confirmations SET good = %v, bad = %v, time = %v WHERE block_id = %v", st1, st0, time.Now().Unix(), blockID)
				confirmation.Good = int32(st1)
				confirmation.Bad = int32(st0)
				confirmation.Time = int32(time.Now().Unix())
				err = confirmation.Save()
				if err != nil {
					logger.Error("%v", err)
				}
			} else {
				confirmation.Good = int32(st1)
				confirmation.Bad = int32(st0)
				confirmation.Time = int32(time.Now().Unix())
				logger.Debug("INSERT INTO confirmations ( block_id, good, bad, time ) VALUES ( %v, %v, %v, %v )", blockID, st1, st0, time.Now().Unix())
				err = confirmation.Save()
				if err != nil {
					logger.Error("%v", err)
				}
			}
			logger.Debug("blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES %d>%d && %d>=%d\n", blockID, startBlockID, st1, consts.MIN_CONFIRMED_NODES)
			if blockID > startBlockID && st1 >= consts.MIN_CONFIRMED_NODES {
				break
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}

func checkConf(host string, blockID int64) string {

	logger.Debug("host: %v", host)
	/*tcpAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return "0"
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)*/
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		logger.Debug("%v", utils.ErrInfo(err))
		return "0"
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))

	// вначале шлем тип данных, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные
	// firstly send a data type for the receiving party could understand how exacetly to process the data sent
	_, err = conn.Write(converter.DecToBin(4, 2))
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}

	// в 4-х байтах пишем ID блока, хэш которого хотим получить
	// record the block ID that we want to recive in 4 bytes
	size := converter.DecToBin(blockID, 4)
	_, err = conn.Write(size)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}

	// ответ всегда 32 байта
	// the response is always 32 bytes
	hash := make([]byte, 32)
	_, err = conn.Read(hash)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}
	return string(hash)
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockID int64, ch0 chan string) {
	logger.Info("IsReachable %v", host)
	ch := make(chan string, 1)
	go func() {
		ch <- checkConf(host, blockID)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
