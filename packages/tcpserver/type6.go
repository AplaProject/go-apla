package tcpserver

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"math/big"
)

func (t *TcpServer) Type6() {
	/**
	- проверяем, находится ли отправитель на одном с нами уровне
	- получаем  block_id, user_id, mrkl_root, signature
	- если хэш блока меньше того, что есть у нас в табле testblock, то смотртим, есть ли такой же хэш тр-ий,
	- если отличается, то загружаем блок от отправителя
	- если не отличается, то просто обновляем хэш блока у себя
	данные присылает демон testblockDisseminator
	*/
	currentBlockId, err := t.GetBlockId()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	if currentBlockId == 0 {
		log.Debug("%v", utils.ErrInfo("currentBlockId == 0"))
		return
	}
	buf := make([]byte, 4)
	_, err = t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := utils.BinToDec(buf)
	log.Debug("size: %v", size)
	if size < 10485760 {
		binaryData := make([]byte, size)
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("binaryData: %x", binaryData)
		newTestblockBlockId := utils.BinToDecBytesShift(&binaryData, 4)
		newTestblockTime := utils.BinToDecBytesShift(&binaryData, 4)
		newTestblockUserId := utils.BinToDecBytesShift(&binaryData, 4)
		newTestblockMrklRoot := utils.BinToHex(utils.BytesShift(&binaryData, 32))
		newTestblockSignatureHex := utils.BinToHex(utils.BytesShift(&binaryData, utils.DecodeLength(&binaryData)))
		log.Debug("newTestblockBlockId: %v", newTestblockBlockId)
		log.Debug("newTestblockTime: %v", newTestblockTime)
		log.Debug("newTestblockUserId: %v", newTestblockUserId)
		log.Debug("newTestblockMrklRoot: %s", newTestblockMrklRoot)
		log.Debug("newTestblockSignatureHex: %s", newTestblockSignatureHex)
		if !utils.CheckInputData(newTestblockBlockId, "int") {
			log.Debug("%v", utils.ErrInfo("incorrect newTestblockBlockId"))
			return
		}
		if !utils.CheckInputData(newTestblockTime, "int") {
			log.Debug("%v", utils.ErrInfo("incorrect newTestblockTime"))
			return
		}
		if !utils.CheckInputData(newTestblockUserId, "int") {
			log.Debug("%v", utils.ErrInfo("incorrect newTestblockUserId"))
			return
		}
		if !utils.CheckInputData(newTestblockMrklRoot, "sha256") {
			log.Debug("%v", utils.ErrInfo("incorrect newTestblockMrklRoot"))
			return
		}
		/*
		 * Проблема одновременных попыток локнуть. Надо попробовать без локов
		 * */
		//t.DbLockGate("6")
		exists, err := t.Single(`
				SELECT block_id
				FROM testblock
				WHERE status = 'active'
				`).Int64()
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		if exists == 0 {
			t.PrintSleep(utils.ErrInfo("null testblock"), 0)
			return
		}
		//prevBlock, myUserId, myMinerId, currentUserId, level, levelsRange, err := t.TestBlock()
		prevBlock, _, _, _, level, levelsRange, err := t.TestBlock()
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		nodesIds := utils.GetOurLevelNodes(level, levelsRange)
		log.Debug("nodesIds: %v ", nodesIds)
		log.Debug("prevBlock: %v ", prevBlock)
		log.Debug("level: %v ", level)
		log.Debug("levelsRange: %v ", levelsRange)
		log.Debug("newTestblockBlockId: %v ", newTestblockBlockId)
		// проверим, верный ли ID блока
		if newTestblockBlockId != prevBlock.BlockId+1 {
			t.PrintSleep(utils.ErrInfo(fmt.Sprintf("newTestblockBlockId != prevBlock.BlockId+1 %d!=%d+1", newTestblockBlockId, prevBlock.BlockId)), 1)
			return
		}
		// проверим, есть ли такой майнер
		minerId, err := t.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ?", newTestblockUserId).Int64()
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		if minerId == 0 {
			t.PrintSleep(utils.ErrInfo("minerId == 0"), 0)
			return
		}
		log.Debug("minerId: %v ", minerId)
		// проверим, точно ли отправитель с нашего уровня
		if !utils.InSliceInt64(minerId, nodesIds) {
			t.PrintSleep(utils.ErrInfo("!InSliceInt64(minerId, nodesIds)"), 0)
			return
		}
		// допустимая погрешность во времени генерации блока
		maxErrorTime := t.variables.Int64["error_time"]
		// получим значения для сна
		sleep, err := t.GetGenSleep(prevBlock, level)
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		// исключим тех, кто сгенерил блок слишком рано
		if prevBlock.Time+sleep-newTestblockTime > maxErrorTime {
			t.PrintSleep(utils.ErrInfo("prevBlock.Time + sleep - newTestblockTime > maxErrorTime"), 0)
			return
		}
		// исключим тех, кто сгенерил блок с бегущими часами
		if newTestblockTime > utils.Time() {
			t.PrintSleep(utils.ErrInfo("newTestblockTime > Time()"), 0)
			return
		}
		// получим хэш заголовка
		newHeaderHash := utils.DSha256(fmt.Sprintf("%v,%v,%v", newTestblockUserId, newTestblockBlockId, prevBlock.HeadHash))
		myTestblock, err := t.OneRow(`
				SELECT block_id,
							user_id,
							hex(mrkl_root) as mrkl_root,
							hex(signature) as signature
				FROM testblock
				WHERE status = 'active'
				`).String()
		if len(myTestblock) > 0 {
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			// получим хэш заголовка
			myHeaderHash := utils.DSha256(fmt.Sprintf("%v,%v,%v", myTestblock["user_id"], myTestblock["block_id"], prevBlock.HeadHash))
			// у кого меньше хэш, тот и круче
			hash1 := big.NewInt(0)
			hash1.SetString(string(newHeaderHash), 16)
			hash2 := big.NewInt(0)
			hash2.SetString(string(myHeaderHash), 16)
			log.Debug("%v", hash1.Cmp(hash2))
			//if HexToDecBig(newHeaderHash) > string(myHeaderHash) {
			if hash1.Cmp(hash2) == 1 {
				t.PrintSleep(utils.ErrInfo(fmt.Sprintf("newHeaderHash > myHeaderHash (%s > %s)", newHeaderHash, myHeaderHash)), 0)
				return
			}
			/* т.к. на данном этапе в большинстве случаев наш текущий блок будет заменен,
			 * то нужно парсить его, рассылать другим нодам и дождаться окончания проверки
			 */
			err = t.ExecSql("UPDATE testblock SET status = 'pending'")
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
		}
		// если отличается, то загружаем недостающии тр-ии от отправителя
		if string(newTestblockMrklRoot) != myTestblock["mrkl_root"] {
			log.Debug("download new tx")
			sendData := ""
			// получим все имеющиеся у нас тр-ии, которые еще не попали в блоки
			txArray, err := t.GetMap(`SELECT hex(hash) as hash, data FROM transactions`, "hash", "data")
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			for hash, _ := range txArray {
				sendData += hash
			}
			err = utils.WriteSizeAndData([]byte(sendData), t.Conn)
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			/*
				в ответ получаем:
				BLOCK_ID   				       4
				TIME       					       4
				USER_ID                         5
				SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
				Размер всех тр-ий, размер 1 тр-ии, тело тр-ии.
				Хэши три-ий (порядок тр-ий)
			*/
			buf := make([]byte, 4)
			_, err = t.Conn.Read(buf)
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			dataSize := utils.BinToDec(buf)
			log.Debug("dataSize %d", dataSize)
			// и если данных менее 10мб, то получаем их
			if dataSize < 10485760 {
				binaryData := make([]byte, dataSize)
				//binaryData, err = ioutil.ReadAll(t.Conn)
				_, err = io.ReadFull(t.Conn, binaryData)
				if err != nil {
					t.PrintSleep(utils.ErrInfo(err), 0)
					return
				}
				// Разбираем полученные бинарные данные
				newTestblockBlockId := utils.BinToDecBytesShift(&binaryData, 4)
				newTestblockTime := utils.BinToDecBytesShift(&binaryData, 4)
				newTestblockUserId := utils.BinToDecBytesShift(&binaryData, 5)
				newTestblockSignature := utils.BytesShift(&binaryData, utils.DecodeLength(&binaryData))
				log.Debug("newTestblockBlockId %v", newTestblockBlockId)
				log.Debug("newTestblockTime %v", newTestblockTime)
				log.Debug("newTestblockUserId %v", newTestblockUserId)
				log.Debug("newTestblockSignature %x", newTestblockSignature)
				// недостающие тр-ии
				length := utils.DecodeLength(&binaryData) // размер всех тр-ий
				txBinary := utils.BytesShift(&binaryData, length)
				for {
					// берем по одной тр-ии
					length := utils.DecodeLength(&txBinary) // размер всех тр-ий
					if length == 0 {
						break
					}
					log.Debug("length %d", length)
					tx := utils.BytesShift(&txBinary, length)
					log.Debug("tx %x", tx)
					txArray[string(utils.Md5(tx))] = string(tx)
				}
				// порядок тр-ий
				var orderHashArray []string
				for {
					orderHashArray = append(orderHashArray, string(utils.BinToHex(utils.BytesShift(&binaryData, 16))))
					if len(binaryData) == 0 {
						break
					}
				}
				// сортируем и наши и полученные транзакции
				var transactions []byte
				for _, txMd5 := range orderHashArray {
					transactions = append(transactions, utils.EncodeLengthPlusData([]byte(txArray[txMd5]))...)
				}
				// формируем блок, который далее будем тщательно проверять
				/*
					Заголовок (от 143 до 527 байт )
					TYPE (0-блок, 1-тр-я)     1
					BLOCK_ID   				       4
					TIME       					       4
					USER_ID                         5
					LEVEL                              1
					SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
					Далее - тело блока (Тр-ии)
				*/
				newBlockIdBinary := utils.DecToBin(newTestblockBlockId, 4)
				timeBinary := utils.DecToBin(newTestblockTime, 4)
				userIdBinary := utils.DecToBin(newTestblockUserId, 5)
				levelBinary := utils.DecToBin(level, 1)
				newBlockHeader := utils.DecToBin(0, 1) // 0 - это блок
				newBlockHeader = append(newBlockHeader, newBlockIdBinary...)
				newBlockHeader = append(newBlockHeader, timeBinary...)
				newBlockHeader = append(newBlockHeader, userIdBinary...)
				newBlockHeader = append(newBlockHeader, levelBinary...) // $level пишем, чтобы при расчете времени ожидания в следующем блоке не пришлось узнавать, какой был max_miner_id
				newBlockHeader = append(newBlockHeader, utils.EncodeLengthPlusData(newTestblockSignature)...)
				newBlockHex := utils.BinToHex(append(newBlockHeader, transactions...))
				// и передаем блок для обратотки через демон queue_parser_testblock
				// т.к. есть запросы к log_time_, а их можно выполнять только по очереди
				err = t.ExecSql(`DELETE FROM queue_testblock WHERE hex(head_hash) = ?`, newHeaderHash)
				if err != nil {
					t.PrintSleep(utils.ErrInfo(err), 0)
					return
				}
				log.Debug("INSERT INTO queue_testblock  (head_hash, data)  VALUES (%s, %s)", newHeaderHash, newBlockHex)
				err = t.ExecSql(`INSERT INTO queue_testblock (head_hash, data) VALUES ([hex], [hex])`, newHeaderHash, newBlockHex)
				if err != nil {
					t.PrintSleep(utils.ErrInfo(err), 0)
					return
				}
			}
		} else {
			err := t.DbLockGate("type6")
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			// если всё нормально, то пишем в таблу testblock новые данные
			exists, err := t.Single(`SELECT block_id FROM testblock`).Int64()
			if err != nil {
				t.PrintSleep(utils.ErrInfo(err), 0)
				return
			}
			if exists == 0 {
				err = t.ExecSql(`INSERT INTO testblock (block_id, time, level, user_id, header_hash, signature, mrkl_root) VALUES (?, ?, ?, ?, [hex], [hex], [hex])`,
					newTestblockBlockId, newTestblockTime, level, newTestblockUserId, string(newHeaderHash), newTestblockSignatureHex, string(newTestblockMrklRoot))
				if err != nil {
					t.PrintSleep(utils.ErrInfo(err), 0)
					return
				}
			} else {
				err = t.ExecSql(`
						UPDATE testblock
						SET   time = ?,
								user_id = ?,
								header_hash = [hex],
								signature = [hex]
						`, newTestblockTime, newTestblockUserId, string(newHeaderHash), string(newTestblockSignatureHex))
				if err != nil {
					t.PrintSleep(utils.ErrInfo(err), 0)
					return
				}
			}
			t.DbUnlock("type6")
		}

		err = t.DbLockGate("type6")
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		err = t.ExecSql("UPDATE testblock SET status = 'active'")
		if err != nil {
			t.PrintSleep(utils.ErrInfo(err), 0)
			return
		}
		t.DbUnlock("type6")

		//t.DbUnlockGate("6")
	}
}
