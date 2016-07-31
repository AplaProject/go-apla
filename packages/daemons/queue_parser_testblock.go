package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
)

/*
 * Парсим и разносим данные из queue_testblock
 * */

func QueueParserTestblock(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "QueueParserTestblock"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 1800
	} else {
		d.sleepTime = 1
	}
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
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

		data, err := d.OneRow("SELECT * FROM queue_testblock ORDER BY head_hash ASC").String()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		if len(data) == 0 {
			if d.unlockPrintSleepInfo(utils.ErrInfo(errors.New("len(data) == 0")), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		newBlock := []byte(data["data"])
		newHeaderHash := utils.BinToHex([]byte(data["head_hash"]))
		tx := utils.DeleteHeader(newBlock)

		// сразу можно удалять данные из таблы-очереди
		err = d.ExecSql("DELETE FROM queue_testblock WHERE hex(head_hash) = ?", newHeaderHash)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(errors.New("len(data) == 0")), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// прежде всего нужно проверить, а нет ли в этом блоке ошибок с несовметимыми тр-ми
		// при полной проверке, а не только фронтальной проблем с несовместимыми тр-ми не будет, т.к. там даные сразу пишутся в таблицы
		// а тут у нас данные пишутся только в log_time_
		// и сами тр-ии пишем в отдельную таблу
		p := new(dcparser.Parser)
		p.DCDB = d.DCDB
		if len(tx) > 0 {
			logger.Debug("len(tx): %d", len(tx))
			for {
				logger.Debug("tx: %x", tx)
				txSize := utils.DecodeLength(&tx)
				logger.Debug("txSize: %d", txSize)
				// отделим одну транзакцию от списка транзакций
				txBinaryData := utils.BytesShift(&tx, txSize)
				logger.Debug("txBinaryData: %x", txBinaryData)
				// проверим, нет ли несовместимых тр-ий
				fatalError, waitError, _, _, _, _ := p.ClearIncompatibleTx(txBinaryData, false)
				if len(fatalError) > 0 || len(waitError) > 0 {
					if d.unlockPrintSleep(utils.ErrInfo(errors.New(" len(fatalError) > 0 || len(waitError) > 0")), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				if len(tx) == 0 {
					break
				}
			}
		}
		// откатим тр-ии тестблока, но не удаляя их, т.к. далее еще можем их вернуть
		p.RollbackTransactionsTestblock(false)

		// проверим блок, который получился с данными, которые прислал другой нод
		p.BinaryData = newBlock
		err = p.ParseDataGate(false)
		if err != nil {

			logger.Error("%v", err)

			// т.к. мы откатили наши тр-ии из transactions_testblock, то теперь нужно обработать их по новой
			// получим наши транзакции в 1 бинарнике, просто для удобства

			var myTestBlockBody []byte
			transactionsTestblock, err := d.GetAll("SELECT data FROM transactions_testblock ORDER BY id ASC", -1)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for _, data := range transactionsTestblock {
				myTestBlockBody = append(myTestBlockBody, utils.EncodeLengthPlusData([]byte(data["data"]))...)
			}

			if len(myTestBlockBody) > 0 {
				p.BinaryData = append(utils.DecToBin(0, 1), myTestBlockBody...)
				err = p.ParseDataGate(true)
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
		} else {
			// наши тр-ии уже не актуальны, т.к. мы их откатили
			err = d.ExecSql("DELETE FROM transactions_testblock")
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			exists, err := d.Single(`SELECT block_id FROM testblock`).Int64()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				return
			}
			if exists > 0 {
				// если всё нормально, то пишем в таблу testblock новые тр-ии и новые данные по юзеру их сгенерившему
				err = d.ExecSql(`
				UPDATE testblock
				SET  time = ?,
						user_id = ?,
						header_hash = [hex],
						signature = [hex],
						mrkl_root = [hex]
				`, p.BlockData.Time, p.BlockData.UserId, newHeaderHash, utils.BinToHex(p.BlockData.Sign), p.MrklRoot)
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}

			// и сами тр-ии пишем в отдельную таблу
			if len(tx) > 0 {
				for {
					txSize := utils.DecodeLength(&tx)
					// отчекрыжим одну транзакцию от списка транзакций
					txBinaryData := utils.BytesShift(&tx, txSize)
					// получим тип тр-ии и юзера
					txType, userId, toUserId := utils.GetTxTypeAndUserId(txBinaryData)
					md5 := utils.Md5(txBinaryData)
					dataHex := utils.BinToHex(txBinaryData)

					err = d.ExecSql("DELETE FROM transactions_testblock")
					if err != nil {
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}

					err = d.ExecSql("INSERT INTO transactions_testblock (hash, data, type, user_id, third_var) VALUES ([hex], [hex], ?, ?, ?)", md5, dataHex, txType, userId, toUserId)
					if err != nil {
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}

					if len(tx) == 0 {
						break
					}
				}
			}

			// удаляем всё, где хэш больше нашего
			err = d.ExecSql("DELETE FROM queue_testblock WHERE hex(head_hash) > ?", newHeaderHash)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// возможно нужно откатить и тр-ии с verified=1 и used=0 из transactions
			// т.к. в transactions может быть тр-ия на удаление банкноты
			// и в transactions_testblock только что была залита такая же тр-ия
			// выходит, что блок, который будет сгенерен на основе transactions будет ошибочным
			// или при откате transactions будет сделан вычет из log_time_....
			// и выйдет что попавшая в блок тр-я из transactions_testblock попала минуя запись  log_time_....
			err = p.RollbackTransactions()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)

}
