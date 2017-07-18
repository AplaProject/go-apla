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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

//var err error

// BlockGenerator generates blocks
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
		MonitorDaemonCh <- []string{GoroutineName, converter.Int64ToStr(time.Now().Unix())}

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

		infoBlock := &model.InfoBlock{}
		err = infoBlock.GetInfoBlock()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		newBlockID := infoBlock.BlockID + 1
		logger.Debug("newBlockID: %v", newBlockID)

		config := &model.Config{}
		err = config.GetConfig()
		//logger.Debug("%v", myWalletID)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		myStateID := config.StateID
		myWalletID := config.DltWalletID

		if myStateID > 0 {
			systemState := &model.SystemRecognizedStates{}
			delegate, err := systemState.IsDelegated(myStateID)
			if err != nil {
				d.dbUnlock()
				logger.Error("%v", err)
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue
			}
			// Если мы - государство и у нас указан delegate в system_recognized_states, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
			// If we are the state and we have the record delegate specified at the system_recognized_states (it means we have delegated the authority to maintain the node to another user or state), in that case exit.
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
		fullNodes := &model.FullNodes{}
		err = fullNodes.FindNode(myStateID, myWalletID, myStateID, myWalletID)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		myFullNodeID := fullNodes.ID
		logger.Debug("myFullNodeID %d", myFullNodeID)
		if myFullNodeID == 0 {
			d.dbUnlock()
			logger.Debug("myFullNodeID == 0")
			d.sleepTime = 10
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// если дошли до сюда, значит мы есть в full_nodes. Надо определить в каком месте списка
		// получим state_id, wallet_id и время последнего блока
		// If we have reached here, we are in full_nodes. It is necessary to determine where in the list we
		// will get state_id, wallet_id and the time of the last block
		infoBlock = &model.InfoBlock{}
		err = infoBlock.GetInfoBlock()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("prevBlock %v", infoBlock)

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

		sleepTime, err := d.GetSleepTime(myWalletID, myStateID, infoBlock.StateID, infoBlock.WalletID)
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
		sleep := int64(sleepTime) - (time.Now().Unix() - int64(infoBlock.Time))
		if sleep < 0 {
			sleep = 0
		}

		logger.Debug("time.Now().Unix*() %v / prevBlock[time] %v", time.Now().Unix(), infoBlock.Time)

		logger.Debug("sleep %v", sleep)

		// спим
		// sleep
		for i := 0; i < int(sleep); i++ {
			time.Sleep(time.Second)
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
		infoBlock = &model.InfoBlock{}
		err = infoBlock.GetInfoBlock()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("prevBlock %v", infoBlock)

		logger.Debug("blockID %v", infoBlock.BlockID)

		logger.Debug("blockgeneration begin")
		if infoBlock.BlockID < 1 {
			logger.Debug("continue")
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		newBlockID = infoBlock.BlockID + 1

		// получим наш приватный нодовский ключ
		// Recieve our private node key
		myNodeKeys := &model.MyNodeKeys{}
		err = myNodeKeys.GetNodeWithMaxBlockID()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		if len(myNodeKeys.PrivateKey) < 1 {
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

		if infoBlock.BlockID >= newBlockID {
			logger.Debug("continue %d >= %d", infoBlock.BlockID, newBlockID)
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
		// transfer the transactions into `verified` = 1
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
			var mrklRoot []byte
			var blockDataTx []byte
			// берем все данные из очереди. Они уже были проверены ранее, и можно их не проверять, а просто брать
			// take all the data from the turn. It is tested already, you may not check them again but just take
			transactions, err := model.GetAllUnusedTransactions()
			if err != nil {
				logging.WriteSelectiveLog(err)
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue
			}
			for _, transaction := range *transactions {
				// Check if we need to get out from the cycle
				if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
					break BEGIN
				}

				logging.WriteSelectiveLog("hash: " + string(transaction.Hash))
				logger.Debug("data %v", transaction.Data)
				logger.Debug("hash %v", transaction.Hash)
				transactionType := transaction.Data[1:2]
				logger.Debug("%v", transactionType)
				logger.Debug("%x", transactionType)
				doubleHash, err := crypto.DoubleHash(transaction.Data)
				if err != nil {
					log.Fatal(err)
				}
				doubleHash = converter.BinToHex(doubleHash)
				mrklArray = append(mrklArray, doubleHash)
				logger.Debug("mrklArray %v", mrklArray)

				oneMoreHash, err := crypto.Hash(transaction.Data)
				if err != nil {
					log.Fatal(err)
				}
				logger.Debug("hash: %s", oneMoreHash)

				dataHex := fmt.Sprintf("%x", transaction.Data)
				logger.Debug("dataHex %v", dataHex)

				blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData(transaction.Data)...)
			}

			if len(mrklArray) == 0 {
				mrklArray = append(mrklArray, []byte("0"))
			}
			mrklRoot = utils.MerkleTreeRoot(mrklArray)
			logger.Debug("mrklRoot: %s", mrklRoot)

			// подписываем нашим нод-ключем заголовок блока
			// sign the heading of a block by our node-key
			var forSign string
			forSign = fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", newBlockID, infoBlock.Hash, Time, myWalletID, myStateID, string(mrklRoot))
			//			forSign = fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", newBlockID, prevBlock[`hash`], Time, myWalletID, myStateID, string(mrklRoot))
			logger.Debug("forSign: %v", forSign)
			//		bytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(forSign))
			bytes, err := crypto.Sign(string(myNodeKeys.PrivateKey), forSign)
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
			newBlockIDBinary := converter.DecToBin(newBlockID, 4)
			timeBinary := converter.DecToBin(Time, 4)
			stateIDBinary := converter.DecToBin(myStateID, 1)

			// заголовок
			// heading
			blockHeader := converter.DecToBin(0, 1)
			blockHeader = append(blockHeader, newBlockIDBinary...)
			blockHeader = append(blockHeader, timeBinary...)
			converter.EncodeLenInt64(&blockHeader, myWalletID)
			blockHeader = append(blockHeader, stateIDBinary...)
			blockHeader = append(blockHeader, converter.EncodeLengthPlusData(signatureBin)...)

			// сам блок
			// block itself
			blockBin := append(blockHeader, blockDataTx...)
			logger.Debug("block %x", blockBin)

			// теперь нужно разнести блок по таблицам и после этого мы будем его слать всем нодам демоном disseminator
			// now we have to spread the block into to the tables and then we'll sent it to the all nodes by the 'disseminator' daemon
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
