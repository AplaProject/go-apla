package daemons

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/parser"
	"github.com/DayLightProject/go-daylight/packages/utils"
	_ "github.com/lib/pq"
	"time"
	"github.com/DayLightProject/go-daylight/packages/consts"
)



/*
Задержки во времени генерации из-за main_lock во время sleep

Delays during generation because of main_lock currently sleep
*/

var err error

func FindNodePos (fullNodesList []map[string]string, prevBlockFullNodeId int64) int {
	logger.Debug("%v %v", fullNodesList, prevBlockFullNodeId)
	for i, full_nodes := range fullNodesList {
		if utils.StrToInt64(full_nodes["full_node_id"]) == prevBlockFullNodeId {
			return i
		}
	}
	return -1
}

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
	if utils.Mobile() {
		d.sleepTime = 3600
	} else {
		d.sleepTime = 10
	}
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

		err, restart := d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		blockId, err := d.GetBlockId()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		newBlockId := blockId + 1
		logger.Debug("newBlockId: %v", newBlockId)
		candidateBlockId, err := d.GetcandidateBlockId()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		logger.Debug("candidateBlockId %v", candidateBlockId)

		if x, err := d.GetMyLocalGateIp(); x != "" {
			if err != nil {
				logger.Error("%v", err)
			}
			logger.Info("%v", "continue")
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		if candidateBlockId == newBlockId {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		myCBID, err := d.GetMyCBID();
		myWalletId, err := d.GetMyWalletId();
		logger.Debug("%v", myWalletId)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		// Если мы - ЦБ и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
		if myCBID > 0 {
			delegate, err:= d.OneRow("SELECT delegate_wallet_id, delegate_cb_id FROM central_banks WHERE cb_id = ?", myCBID).Int64()
			if err != nil {
				d.dbUnlock()
				logger.Error("%v", err)
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue
			}
			if delegate["delegate_wallet_id"] > 0 || delegate["delegate_cb_id"] > 0  {
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
		full_node_id, err:= d.Single("SELECT full_node_id FROM full_nodes WHERE final_delegate_cb_id = ? OR final_delegate_wallet_id = ? OR cb_id = ? OR wallet_id = ?", myCBID, myWalletId, myCBID, myWalletId).Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		if full_node_id == 0 {
			d.dbUnlock()
			logger.Debug("full_node_id == 0")
			d.sleepTime = 10
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// если дошли до сюда, значит мы есть в full_nodes. надо определить в каком месте списка
		// получим cb_id, wallet_id и время последнего блока
		prevBlock, err := d.OneRow("SELECT cb_id, wallet_id, time FROM info_block").Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// возьмем список всех full_nodes
		fullNodesList, err := d.GetAll("SELECT full_node_id, wallet_id, cb_id FROM full_nodes", -1)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// определим full_node_id того, кто должен был генерить блок (но мог это делегировать)
		prevBlockFullNodeId, err := d.Single("SELECT full_node_id FROM full_nodes WHERE cb_id = ? OR wallet_id = ?", prevBlock["cb_id"], prevBlock["wallet_id"]).Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		prevBlockFullNodePosition := FindNodePos (fullNodesList, prevBlockFullNodeId)

		// определим свое место (в том числе в delegate)
		myPosition := func (fullNodesList []map[string]string, prevBlockFullNodeId int64) int {
			logger.Debug("%v %v", fullNodesList, prevBlockFullNodeId)
			for i, full_nodes := range fullNodesList {
				if utils.StrToInt64(full_nodes["cb_id"]) == myCBID || utils.StrToInt64(full_nodes["wallet_id"]) == myWalletId || utils.StrToInt64(full_nodes["final_delegate_cb_id"]) == myWalletId || utils.StrToInt64(full_nodes["final_delegate_wallet_id"]) == myWalletId {
					return i
				}
			}
			return -1
		} (fullNodesList, full_node_id)

		// имея время предыдущего блока и позицию определяем время сна
		if myPosition < prevBlockFullNodePosition {
			myPosition += len(fullNodesList)
		}
		sleepTime := (myPosition - prevBlockFullNodePosition) * consts.DELAY

		logger.Debug("sleepTime %v / myPosition %v / prevBlockFullNodePosition %v / consts.DELAY %v", sleepTime, myPosition, prevBlockFullNodePosition, consts.DELAY)

		d.dbUnlock()
		// учтем прошедшее время
		sleep := int64(sleepTime) - utils.Time() - prevBlock["time"]
		if sleep < 0 {
			sleep = 0
		}

		logger.Debug("utils.Time() %v / prevBlock[time] %v", utils.Time(), prevBlock["time"])

		logger.Debug("sleep %v", sleep)

		// спим
		for i := 0; i < int(sleep); i++ {
			utils.Sleep(1)
		}

		// пока мы спали последний блок, скорее всего, изменился. Но с большой вероятностью наше место в очереди не изменилось. А если изменилось, то ничего страшного не прозойдет.
		err, restart = d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		prevBlock, err = d.OneRow("SELECT cb_id, wallet_id, block_id, time FROM info_block").Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		blockId = prevBlock["block_id"]
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

		newBlockId = blockId + 1

		// получим наш приватный нодовский ключ
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

		if prevBlock["block_id"] >= newBlockId {
			logger.Debug("continue")
			d.dbUnlock()
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		// откатим transactions_candidate_block
		p := new(parser.Parser)
		p.DCDB = d.DCDB
		p.RollbackTransactionsCandidateBlock(true)

		Time := time.Now().Unix()

		// переведем тр-ии в `verified` = 1
		err = p.AllTxParser()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		var mrklArray [][]byte
		var usedTransactions string
		var mrklRoot []byte
		var blockDataTx []byte
		// берем все данные из очереди. Они уже были проверены ранее, и можно их не проверять, а просто брать
		rows, err := d.Query(d.FormatQuery("SELECT data, hex(hash), type, wallet_id, citizen_id, third_var FROM transactions WHERE used = 0 AND verified = 1"))
		if err != nil {
			utils.WriteSelectiveLog(err)
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		for rows.Next() {
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
		block, _ := pem.Decode([]byte(nodePrivateKey))
		if block == nil {
			logger.Error("bad key data %v ", utils.GetParent())
			utils.Sleep(1)
			continue BEGIN
		}
		if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
			if d.dPrintSleep(fmt.Sprintf("unknown key type %v, want %v / %v ", got, want, utils.GetParent()), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			if d.dPrintSleep(fmt.Sprintf("err %v %v", err, utils.GetParent()), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		var forSign string
		forSign = fmt.Sprintf("0,%v,%v,%v,%v", newBlockId, prevBlock["hash"], Time, string(mrklRoot))
		logger.Debug("forSign: %v", forSign)
		bytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(forSign))
		if err != nil {
			if d.dPrintSleep(fmt.Sprintf("err %v %v", err, utils.GetParent()), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		signatureBin := bytes

		// готовим заголовок
		newBlockIdBinary := utils.DecToBin(newBlockId, 4)
		timeBinary := utils.DecToBin(Time, 4)
		walletIdBinary := utils.EncodeLengthPlusData(myWalletId)
		cbIdBinary := utils.DecToBin(myCBID, 1)

		// заголовок
		blockHeader := utils.DecToBin(0, 1)
		blockHeader = append(blockHeader, newBlockIdBinary...)
		blockHeader = append(blockHeader, timeBinary...)
		blockHeader = append(blockHeader, walletIdBinary...)
		blockHeader = append(blockHeader, cbIdBinary...)
		blockHeader = append(blockHeader, utils.EncodeLengthPlusData(signatureBin)...)

		// сам блок
		blockBin := append(blockHeader, blockDataTx...)
		logger.Debug("block %x", blockBin)

		// теперь нужно разнести блок по таблицам и после этого мы будем его слать всем нодам демоном disseminator
		p.BinaryData = blockBin
		err = p.ParseDataFront()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		/// #######################################
		// Отмечаем транзакции, которые попали в transactions_candidate_block
		// Пока для эксперимента
		// если не отмечать, то получается, что и в transactions_candidate_block и в transactions будут провернные тр-ии, которые откатятся дважды
		if len(usedTransactions) > 0 {
			usedTransactions := usedTransactions[:len(usedTransactions)-1]
			logger.Debug("usedTransactions %v", usedTransactions)
			utils.WriteSelectiveLog("UPDATE transactions SET used=1 WHERE hash IN (" + usedTransactions + ")")
			affect, err := d.ExecSqlGetAffect("UPDATE transactions SET used=1 WHERE hash IN (" + usedTransactions + ")")
			if err != nil {
				utils.WriteSelectiveLog(err)
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			// для теста удаляем, т.к. она уже есть в transactions_candidate_block
			/*  $db->query( __FILE__, __LINE__,  __FUNCTION__,  __CLASS__, __METHOD__, "
			DELETE FROM `".DB_PREFIX."transactions`
			WHERE `hash` IN ({$used_transactions})
			");*/
		}
		// ############################################

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
