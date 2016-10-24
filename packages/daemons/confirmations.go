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
	"github.com/EGaaS/go-mvp/packages/consts"
	"github.com/EGaaS/go-mvp/packages/utils"
	"time"
	"net"
)

/*
Getting amount of nodes, which has the same hash as we do
Using it for watching for forks
Получаем кол-во нодов, у которых такой же хэш последнего блока как и у нас
Нужно чтобы следить за вилками
*/


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
		s++

		d.sleepTime = 1

		if s < 12 {
			d.sleepTime = 1
		}

		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		var startBlockId int64
		// если последний проверенный был давно (пропасть более 5 блоков),
		// то начинаем проверку последних 5 блоков
		ConfirmedBlockId, err := d.GetConfirmedBlockId()
		if err != nil {
			logger.Error("%v", err)
		}
		LastBlockId, err := d.GetBlockId()
		if err != nil {
			logger.Error("%v", err)
		}
		if LastBlockId-ConfirmedBlockId > 5 {
			startBlockId = ConfirmedBlockId + 1
			d.sleepTime = 10
			s = 0 // 2 минуты отчитываем с начала
		}
		if startBlockId == 0 {
			startBlockId = LastBlockId - 1
		}
		logger.Debug("startBlockId: %d / LastBlockId: %d", startBlockId, LastBlockId)

		for blockId := LastBlockId; blockId > startBlockId; blockId-- {

			// проверим, не нужно ли нам выйти из цикла
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				break BEGIN
			}

			logger.Debug("blockId: %d", blockId)

			hash, err := d.Single("SELECT hash FROM block_chain WHERE id = ?", blockId).String()
			if err != nil {
				logger.Error("%v", err)
			}
			logger.Info("hash: %x", hash)
			if len(hash) == 0 {
				logger.Debug("len(hash) == 0")
				continue
			}

			var hosts []string
			if d.ConfigIni["test_mode"] == "1" {
				hosts = []string{"localhost:"+consts.TCP_PORT}
			} else {
				hosts, err = d.GetHosts()
				if err != nil {
					logger.Error("%v", err)
				}
			}

			ch := make(chan string)
			for i := 0; i < len(hosts); i++ {
				host := hosts[i]+":"+consts.TCP_PORT
				logger.Info("host %v", host)
				go func() {
					IsReachable(host, blockId, ch)
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
			exists, err := d.Single("SELECT block_id FROM confirmations WHERE block_id= ?", blockId).Int64()
			if exists > 0 {
				logger.Debug("UPDATE confirmations SET good = %v, bad = %v, time = %v WHERE block_id = %v", st1, st0, time.Now().Unix(), blockId)
				err = d.ExecSql("UPDATE confirmations SET good = ?, bad = ?, time = ? WHERE block_id = ?", st1, st0, time.Now().Unix(), blockId)
				if err != nil {
					logger.Error("%v", err)
				}
			} else {
				logger.Debug("INSERT INTO confirmations ( block_id, good, bad, time ) VALUES ( %v, %v, %v, %v )", blockId, st1, st0, time.Now().Unix())
				err = d.ExecSql("INSERT INTO confirmations ( block_id, good, bad, time ) VALUES ( ?, ?, ?, ? )", blockId, st1, st0, time.Now().Unix())
				if err != nil {
					logger.Error("%v", err)
				}
			}
			logger.Debug("blockId > startBlockId && st1 >= consts.MIN_CONFIRMED_NODES %d>%d && %d>=%d\n", blockId, startBlockId, st1, consts.MIN_CONFIRMED_NODES)
			if blockId > startBlockId && st1 >= consts.MIN_CONFIRMED_NODES {
				break
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}

func checkConf(host string, blockId int64) string {

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
	_, err = conn.Write(utils.DecToBin(4, 2))
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}

	// в 4-х байтах пишем ID блока, хэш которого хотим получить
	size := utils.DecToBin(blockId, 4)
	_, err = conn.Write(size)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}

	// ответ всегда 32 байта
	hash := make([]byte, 32)
	_, err = conn.Read(hash)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return "0"
	}
	return string(hash)
}

func IsReachable(host string, blockId int64, ch0 chan string) {
	logger.Info("IsReachable %v", host)
	ch := make(chan string, 1)
	go func() {
		ch <- checkConf(host, blockId)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
