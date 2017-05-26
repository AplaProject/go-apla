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
	//	"crypto"
	//	"crypto/rand"
	//	"crypto/rsa"
	//	"crypto/x509"
	//	"encoding/pem"
	"fmt"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	_ "github.com/lib/pq"
)

var err error

func BlockGenerator(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "BlockGenerator"
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

BEGIN:
	for {

		// full_node_id == 0 приводит к установке d.sleepTime = 10 в daemons/upd_full_nodes.go, тут надо обнулить, т.к. может быть первичная установка
		// full_node_id == 0 leads to the installation of  d.sleepTime = 10 в daemons/upd_full_nodes.go, here it is necessary to reset, because it could happen a primary installation

		d.sleepTime = 1

		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		// Check, whether we need to get out of the cycle
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
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		newBlockId := blockId + 1
		logger.Debug("newBlockId: %v", newBlockId)

		myStateID, myWalletId, err := d.GetMyStateIDAndWalletID()
		logger.Debug("%v", myWalletId)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		if myStateID > 0 {
			delegate, err := d.CheckDelegateCB(myStateID)
			if err != nil {
				d.dbUnlock()
				logger.Error("%v", err)
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue
			}
			// Если мы - государство и у нас указан delegate в system_recognized_states, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
			// If we are the state and we have the record delegate specified at the system_recognized_states (it means we have delegated the authority to maintain the node to another user or state), in that case go out.
			if delegate {
				d.dbUnlock()
				logger.Debug("delegate > 0")
				d.sleepTime = 3600
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue
			}
		}

		// Есть ли мы в списке тех, кто может генерить блоки
		// If we are in the list of those who can generate blocks
		my_full_node_id, err := d.FindInFullNodes(myStateID, myWalletId)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("my_full_node_id %d", my_full_node_id)
		if my_full_node_id == 0 {
			d.dbUnlock()
			logger.Debug("my_full_node_id == 0")
			d.sleepTime = 10
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// если дошли до сюда, значит мы есть в full_nodes. Надо определить в каком месте списка
		// If we have reached here, we are in full_nodes. It is necessary to determine where in the list we are
		// получим state_id, wallet_id и время последнего блока
		// We will get state_id, wallet_id and the time of the last block
		prevBlock, err := d.OneRow("SELECT state_id, wallet_id, block_id, time, hex(hash) as hash FROM info_block").Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("prevBlock %v", prevBlock)

		/*		prevBlockHash, err := d.Single("SELECT hex(hash) as hash FROM info_block").String()
				if err != nil {
					d.dbUnlock()
					logger.Error("%v", err)
					if d.dSleep(d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				logger.Debug("prevBlockHash %s", prevBlockHash)*/

		sleepTime, err := d.GetSleepTime(myWalletId, myStateID, prevBlock["state_id"], prevBlock["wallet_id"])
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		d.dbUnlock()

		// учтем прошедшее время
		// take into account the passed time
		sleep := int64(sleepTime) - (utils.Time() - prevBlock["time"])
		if sleep < 0 {
			sleep = 0
		}

		logger.Debug("utils.Time() %v / prevBlock[time] %v", utils.Time(), prevBlock["time"])

		logger.Debug("sleep %v", sleep)

		// спим
		// sleep
		for i := 0; i < int(sleep); i++ {
			utils.Sleep(1)
		}

		// пока мы спали последний блок, скорее всего, изменился. Но с большой вероятностью наше место в очереди не изменилось. А если изменилось, то ничего страшного не прозойдет.
		// While we slept, most likely the last block has been changed. But probably our turn is not changed. Even if it is, dont't worry, nothing bad will happen.
		restart, err = d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		prevBlock, err = d.OneRow("SELECT state_id, wallet_id, block_id, time, hex(hash) as hash FROM info_block").Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("prevBlock %v", prevBlock)
		prevBlockHash, err := d.Single("SELECT hex(hash) as hash FROM info_block").String()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		logger.Debug("blockId %v", blockId)

		logger.Debug("blockgeneration begin")
		if blockId < 1 {
			logger.Debug("continue")
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		newBlockId = prevBlock["block_id"] + 1

		// получим наш приватный нодовский ключ
		// Recieve our private node key
		nodePrivateKey, err := d.GetNodePrivateKey()
		if len(nodePrivateKey) < 1 {
			logger.Debug("continue")
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		//#####################################
		//##		 Формируем блок
		//#####################################
		//#####################################
		//##		 Form the block
		//#####################################

		if prevBlock["block_id"] >= newBlockId {
			logger.Debug("continue %d >= %d", prevBlock["block_id"], newBlockId)
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		p := new(parser.Parser)
		p.DCDB = d.DCDB

		//Time := time.Now().Unix()

		// переведем тр-ии в `verified` = 1
		// Transfer the territories into `verified` = 1
		err = p.AllTxParser()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		okBlock := false
		for !okBlock {

			Time := time.Now().Unix()
			var mrklArray [][]byte
			var usedTransactions string
			var mrklRoot []byte
			var blockDataTx []byte
			// берем все данные из очереди. Они уже были проверены ранее, и можно их не проверять, а просто брать
			// Take all the data from the turn. It is tested already, you may not check them again but just take
			rows, err := d.Query(d.FormatQuery("SELECT data, hex(hash), type, wallet_id, citizen_id, third_var FROM transactions WHERE used = 0 AND verified = 1"))
			if err != nil {
				utils.WriteSelectiveLog(err)
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue
			}
			for rows.Next() {
				// проверим, не нужно ли нам выйти из цикла
				// Check if we need to get out from the cycle
				if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
					break BEGIN
				}
				var data []byte
				var hash string
				var txType string
				var txWalletId string
				var txCitizenId string
				var thirdVar string
				err = rows.Scan(&data, &hash, &txType, &txWalletId, &txCitizenId, &thirdVar)
				if err != nil {
					rows.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				utils.WriteSelectiveLog("hash: " + string(hash))
				logger.Debug("data %v", data)
				logger.Debug("hash %v", hash)
				transactionType := data[1:2]
				logger.Debug("%v", transactionType)
				logger.Debug("%x", transactionType)
				mrklArray = append(mrklArray, utils.DSha256(data))
				logger.Debug("mrklArray %v", mrklArray)

				hashMd5 := utils.Md5(data)
				logger.Debug("hashMd5: %s", hashMd5)

				dataHex := fmt.Sprintf("%x", data)
				logger.Debug("dataHex %v", dataHex)

				blockDataTx = append(blockDataTx, utils.EncodeLengthPlusData([]byte(data))...)

				if configIni["db_type"] == "postgresql" {
					usedTransactions += "decode('" + hash + "', 'hex'),"
				} else {
					usedTransactions += "x'" + hash + "',"
				}
			}
			rows.Close()

			if len(mrklArray) == 0 {
				mrklArray = append(mrklArray, []byte("0"))
			}
			mrklRoot = utils.MerkleTreeRoot(mrklArray)
			logger.Debug("mrklRoot: %s", mrklRoot)

			// подписываем нашим нод-ключем заголовок блока
			// Sign the heading of a block by our node-key
			var forSign string
			forSign = fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", newBlockId, prevBlockHash, Time, myWalletId, myStateID, string(mrklRoot))
			//			forSign = fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", newBlockId, prevBlock[`hash`], Time, myWalletId, myStateID, string(mrklRoot))
			logger.Debug("forSign: %v", forSign)
			//		bytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(forSign))
			bytes, err := lib.SignECDSA(nodePrivateKey, forSign)
			if err != nil {
				if d.dPrintSleep(fmt.Sprintf("err %v %v", err, utils.GetParent()), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			logger.Debug("SignECDSA %x", bytes)

			signatureBin := bytes

			// готовим заголовок
			// Prepare the heading
			newBlockIdBinary := utils.DecToBin(newBlockId, 4)
			timeBinary := utils.DecToBin(Time, 4)
			stateIdBinary := utils.DecToBin(myStateID, 1)

			// заголовок
			// heading
			blockHeader := utils.DecToBin(0, 1)
			blockHeader = append(blockHeader, newBlockIdBinary...)
			blockHeader = append(blockHeader, timeBinary...)
			lib.EncodeLenInt64(&blockHeader, myWalletId)
			blockHeader = append(blockHeader, stateIdBinary...)
			blockHeader = append(blockHeader, utils.EncodeLengthPlusData(signatureBin)...)

			// сам блок
			// block itself
			blockBin := append(blockHeader, blockDataTx...)
			logger.Debug("block %x", blockBin)

			// теперь нужно разнести блок по таблицам и после этого мы будем его слать всем нодам демоном disseminator
			// Now we have to spread the block into to the tables and then we'll sent it to the all nodes by the daemon disseminator
			p.BinaryData = blockBin
			err = p.ParseDataFull(true)
			if err != nil {
				p.BlockError(err)
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue
			}
			okBlock = true
		}
		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
