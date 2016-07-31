package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
)

func (t *TcpServer) Type1() {
	log.Debug("dataType: 1")
	// размер данных
	buf := make([]byte, 4)
	n, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := utils.BinToDec(buf)
	log.Debug("size: %v / n: %v", size, n)
	if size < 10485760 {
		// сами данные
		binaryData := make([]byte, size)
		log.Debug("ReadAll 0")
		_, err = io.ReadFull(t.Conn, binaryData)
		log.Debug("ReadAll 1")
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("binaryData: %x", binaryData)
		/*
		 * принимаем зашифрованный список тр-ий от демона disseminator, которые есть у отправителя
		 * Блоки не качаем тут, т.к. может быть цепочка блоков, а их качать долго
		 * тр-ии качаем тут, т.к. они мелкие и точно скачаются за 60 сек
		 * */
		key, iv, decryptedBinData, err := t.DecryptData(&binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("key: %v / iv: %v", key, iv)
		/*
		 * структура данных:
		 * user_id - 5 байт
		 * type - 1 байт. 0 - блок, 1 - список тр-ий
		 * {если type==1}:
		 * <любое кол-во следующих наборов>
		 * high_rate - 1 байт
		 * tx_hash - 16 байт
		 * </>
		 * {если type==0}:
		 * block_id - 3 байта
		 * hash - 32 байт
		 * head_hash - 32 байт
		 * <любое кол-во следующих наборов>
		 * high_rate - 1 байт
		 * tx_hash - 16 байт
		 * </>
		 * */
		blockId, err := t.GetBlockId()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("decryptedBinData: %x", decryptedBinData)
		// user_id отправителя, чтобы знать у кого брать данные, когда они будут скачиваться другим скриптом
		newDataUserId := utils.BinToDec(utils.BytesShift(&decryptedBinData, 5))
		log.Debug("newDataUserId: %d", newDataUserId)
		// данные могут быть отправлены юзером, который уже не майнер
		minerId, err := t.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ? AND miner_id > 0", newDataUserId).Int64()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("minerId: %v", minerId)
		if minerId == 0 {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// если 0 - значит вначале идет инфа о блоке, если 1 - значит сразу идет набор хэшей тр-ий
		newDataType := utils.BinToDecBytesShift(&decryptedBinData, 1)
		log.Debug("newDataType: %d", newDataType)
		if newDataType == 0 {
			// ID блока, чтобы не скачать старый блок
			newDataBlockId := utils.BinToDecBytesShift(&decryptedBinData, 3)
			log.Debug("newDataBlockId: %d / blockId: %d", newDataBlockId, blockId)
			// нет смысла принимать старые блоки
			if newDataBlockId >= blockId {
				// Это хэш для соревнования, у кого меньше хэш
				newDataHash := utils.BinToHex(utils.BytesShift(&decryptedBinData, 32))
				// Для доп. соревнования, если head_hash равны (шалит кто-то из майнеров и позже будет за такое забанен)
				newDataHeadHash := utils.BinToHex(utils.BytesShift(&decryptedBinData, 32))
				err = t.ExecSql(`DELETE FROM queue_blocks WHERE hex(hash) = ?`, newDataHash)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				err = t.ExecSql(`
						INSERT INTO queue_blocks (
							hash,
							head_hash,
							user_id,
							block_id
						) VALUES (
							[hex],
							[hex],
							?,
							?
						)`, newDataHash, newDataHeadHash, newDataUserId, newDataBlockId)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
			}
		}
		log.Debug("decryptedBinData: %x", decryptedBinData)
		var needTx []byte
		// Разбираем список транзакций
		if len(decryptedBinData) == 0 {
			log.Debug("%v", utils.ErrInfo("len(decryptedBinData) == 0"))
			return
		}
		for {
			// 1 - это админские тр-ии, 0 - обычные
			newDataHighRate := utils.BinToDecBytesShift(&decryptedBinData, 1)
			if len(decryptedBinData) < 16 {
				log.Debug("%v", utils.ErrInfo("len(decryptedBinData) < 16"))
				return
			}
			log.Debug("newDataHighRate: %v", newDataHighRate)
			newDataTxHash := utils.BinToHex(utils.BytesShift(&decryptedBinData, 16))
			if len(newDataTxHash) == 0 {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			log.Debug("newDataTxHash %s", newDataTxHash)
			// проверим, нет ли у нас такой тр-ии
			exists, err := t.Single("SELECT count(hash) FROM log_transactions WHERE hex(hash) = ?", newDataTxHash).Int64()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			if exists > 0 {
				log.Debug("exists")
				continue
			}
			needTx = append(needTx, utils.HexToBin(newDataTxHash)...)
			if len(decryptedBinData) == 0 {
				break
			}
		}
		if len(needTx) == 0 {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("needTx: %v", needTx)
		// шифруем данные. ключ $key сеансовый, iv тоже
		encData, _, err := utils.EncryptCFB(needTx, key, iv)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// в 4-х байтах пишем размер данных, которые пошлем далее
		size := utils.DecToBin(len(encData), 4)
		_, err = t.Conn.Write(size)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("size: %v", len(encData))
		log.Debug("encData: %x", encData)
		// далее шлем сами данные
		_, err = t.Conn.Write(encData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// в ответ получаем размер данных, которые нам хочет передать сервер
		buf := make([]byte, 4)
		_, err = t.Conn.Read(buf)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		dataSize := utils.BinToDec(buf)
		log.Debug("dataSize %v", dataSize)
		// и если данных менее 10мб, то получаем их
		if dataSize < 10485760 {
			encBinaryTxs := make([]byte, dataSize)
			_, err = io.ReadFull(t.Conn, encBinaryTxs)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			// разбираем полученные данные
			log.Debug("encBinaryTxs %x", encBinaryTxs)
			// уберем IV из начала
			utils.BytesShift(&encBinaryTxs, 16)
			// декриптуем
			binaryTxs, err := utils.DecryptCFB(iv, encBinaryTxs, key)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			log.Debug("binaryTxs %x", binaryTxs)
			for {
				txSize := utils.DecodeLength(&binaryTxs)
				if int64(len(binaryTxs)) < txSize {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				txBinData := utils.BytesShift(&binaryTxs, txSize)
				if len(txBinData) == 0 {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				txHex := utils.BinToHex(txBinData)
				// проверим размер
				if int64(len(txBinData)) > t.variables.Int64["max_tx_size"] {
					log.Debug("%v", utils.ErrInfo("len(txBinData) > max_tx_size"))
					return
				}
				newDataHighRate := 0 // временно для тестов

				log.Debug("INSERT INTO queue_tx (hash, high_rate, data) %s, %d, %s", utils.Md5(txBinData), newDataHighRate, txHex)
				err = t.ExecSql(`INSERT INTO queue_tx (hash, high_rate, data) VALUES ([hex], ?, [hex])`, utils.Md5(txBinData), newDataHighRate, txHex)
				if len(txBinData) == 0 {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
			}
		}
	}
}
