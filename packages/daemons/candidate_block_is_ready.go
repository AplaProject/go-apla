package daemons

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
)

/**
 * Демон, который отсчитывает время, которые необходимо ждать после того,
 * как началось одноуровневое соревнование, у кого хэш меньше.
 * Когда время прошло, то берется блок из таблы candidateBlock и заносится в
 * queue и queue_front для занесение данных к себе и отправки другим
 *
 */

func candidateBlockIsReady(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "candidateBlockIsReady"
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
		d.sleepTime = 1
	}
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}

	err = d.notMinerSetSleepTime(1800)
	if err != nil {
		logger.Error("%v", err)
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

		LocalGateIp, err := d.GetMyLocalGateIp()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		if len(LocalGateIp) > 0 {
			if d.dPrintSleep(utils.ErrInfo(errors.New("len(LocalGateIp) > 0")), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// сколько нужно спать
		prevBlock, myUserId, myMinerId, currentUserId, level, levelsRange, err := d.Candidate_block()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Info("%v", prevBlock, myUserId, myMinerId, currentUserId, level, levelsRange)

		if myMinerId == 0 {
			logger.Debug("myMinerId == 0")
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		sleepData, err := d.GetSleepData()
		sleep := d.GetIsReadySleep(prevBlock.Level, sleepData["is_ready"])
		prevHeadHash := prevBlock.HeadHash

		// Если случится откат или придет новый блок, то candidateBlock станет неактуален
		startSleep := utils.Time()
		for i := 0; i < int(sleep); i++ {
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

			newHeadHash, err := d.Single("SELECT head_hash FROM info_block").String()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue
			}
			d.dbUnlock()
			newHeadHash = string(utils.BinToHex([]byte(newHeadHash)))
			if newHeadHash != prevHeadHash {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			logger.Info("%v", "i", i, "time", utils.Time())
			if utils.Time()-startSleep > sleep {
				break
			}
			utils.Sleep(1) // спим 1 сек. общее время = $sleep
		}

		/*
			Заголовок
			TYPE (0-блок, 1-тр-я)       FF (256)
			BLOCK_ID   				       FF FF FF FF (4 294 967 295)
			TIME       					       FF FF FF FF (4 294 967 295)
			USER_ID                          FF FF FF FF FF (1 099 511 627 775)
			LEVEL                              FF (256)
			SIGN                               от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
			Далее - тело блока (Тр-ии)
		*/

		// нужно блокировать т.к. изменение в info_block и block_chain ведут к изменению подписи в candidateBlock
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
		// за промежуток в main_unlock и main_lock мог прийти новый блок
		prevBlock, myUserId, myMinerId, currentUserId, level, levelsRange, err = d.Candidate_block()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Info("%v", prevBlock, myUserId, myMinerId, currentUserId, level, levelsRange)

		// на всякий случай убедимся, что блок не изменился
		if prevBlock.HeadHash != prevHeadHash {
			if d.unlockPrintSleep(utils.ErrInfo(errors.New("prevBlock.HeadHash != prevHeadHash")), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// составим блок. заголовок + тело + подпись
		candidateBlockData, err := d.OneRow("SELECT * FROM candidateBlock WHERE status  =  'active'").String()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(errors.New("prevBlock.HeadHash != prevHeadHash")), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("candidateBlockData: %v", candidateBlockData)
		if len(candidateBlockData) == 0 {
			if d.unlockPrintSleep(utils.ErrInfo(errors.New("null $candidateBlock_data")), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		// получим транзакции
		var candidateBlockDataTx []byte
		transactionscandidateBlock, err := d.GetList("SELECT data FROM transactions_candidate_block ORDER BY id ASC").String()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for _, data := range transactionscandidateBlock {
			candidateBlockDataTx = append(candidateBlockDataTx, utils.EncodeLengthPlusData([]byte(data))...)
		}

		// в промежутке между тем, как блок был сгенерирован и запуском данного демона может измениться текущий блок
		// поэтому нужно проверять подпись блока из тестблока
		prevBlockHash, err := d.Single("SELECT hash FROM info_block").Bytes()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		prevBlockHash = utils.BinToHex(prevBlockHash)
		nodePublicKey, err := d.GetNodePublicKey(utils.StrToInt64(candidateBlockData["user_id"]))
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		forSign := fmt.Sprintf("0,%v,%s,%v,%v,%v,%s", candidateBlockData["block_id"], prevBlockHash, candidateBlockData["time"], candidateBlockData["user_id"], candidateBlockData["level"], utils.BinToHex([]byte(candidateBlockData["mrkl_root"])))
		logger.Debug("forSign %v", forSign)
		logger.Debug("signature %x", candidateBlockData["signature"])

		p := new(dcparser.Parser)
		p.DCDB = d.DCDB
		// проверяем подпись
		_, err0 := utils.CheckSign([][]byte{nodePublicKey}, forSign, []byte(candidateBlockData["signature"]), true)
		if err0 != nil {
			logger.Error("incorrect signature %v")
			err:=p.RollbackTransactionsCandidateBlock(true)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = d.ExecSql("DELETE FROM candidateBlock")
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = p.RollbackTransactions()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if d.unlockPrintSleep(utils.ErrInfo(err0), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		// БАГ
		if utils.StrToInt64(candidateBlockData["block_id"]) == prevBlock.BlockId {
			logger.Error("candidateBlockData block_id =  prevBlock.BlockId (%v=%v)", candidateBlockData["block_id"], prevBlock.BlockId)

			err = p.RollbackTransactionsCandidateBlock(true)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = d.ExecSql("DELETE FROM candidateBlock")
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = p.RollbackTransactions()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if d.unlockPrintSleep(utils.ErrInfo("candidateBlockData block_id =  prevBlock.BlockId"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// готовим заголовок
		newBlockIdBinary := utils.DecToBin(utils.StrToInt64(candidateBlockData["block_id"]), 4)
		timeBinary := utils.DecToBin(utils.StrToInt64(candidateBlockData["time"]), 4)
		userIdBinary := utils.DecToBin(utils.StrToInt64(candidateBlockData["user_id"]), 5)
		levelBinary := utils.DecToBin(utils.StrToInt64(candidateBlockData["level"]), 1)
		//prevBlockHashBinary := prevBlock.Hash
		//merkleRootBinary := candidateBlockData["mrklRoot"];

		// заголовок
		blockHeader := utils.DecToBin(0, 1)
		blockHeader = append(blockHeader, newBlockIdBinary...)
		blockHeader = append(blockHeader, timeBinary...)
		blockHeader = append(blockHeader, userIdBinary...)
		blockHeader = append(blockHeader, levelBinary...)
		blockHeader = append(blockHeader, utils.EncodeLengthPlusData([]byte(candidateBlockData["signature"]))...)

		// сам блок
		block := append(blockHeader, candidateBlockDataTx...)
		logger.Debug("block %x", block)

		// теперь нужно разнести блок по таблицам и после этого мы будем его слать всем нодам демоном disseminator
		p.BinaryData = block
		err = p.ParseDataFront()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// и можно удалять данные о тестблоке, т.к. они перешли в нормальный блок
		affect, err := d.ExecSqlGetAffect("DELETE FROM transactions_candidate_block")
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("affect %d", affect)
		/*
		Почему-то тут не произошло удаление candidateBlock, в итоге в candidateBlock_generator запрос err = d.ExecSql(`INSERT INTO candidateBlock (block_id, time, level, user_id, header_hash, signature, mrkl_root) VALUES (?, ?, ?, ?, [hex], [hex], [hex])`,
 привел к ошибке
		*/
		affect, err = d.ExecSqlGetAffect("DELETE FROM candidateBlock")
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("affect %d", affect)
		candidateBlockTmp, _ := d.OneRow(`SELECT * FROM candidateBlock`).String()
		logger.Debug("candidateBlockTmp %v", candidateBlockTmp)

		// между candidateBlock_generator и testbock_is_ready
		err = p.RollbackTransactionsCandidateBlock(false)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = p.RollbackTransactions()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
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
