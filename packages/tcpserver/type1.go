package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"github.com/DayLightProject/go-daylight/packages/consts"
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
		 * принимаем список тр-ий от демона disseminator, которые есть у отправителя
		 * Блоки не качаем тут, т.к. может быть цепочка блоков, а их качать долго
		 * тр-ии качаем тут, т.к. они мелкие и точно скачаются за 60 сек
		 * */

		/*
		 * структура данных:
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
		log.Debug("binaryData: %x", binaryData)
		// host отправителя, чтобы знать у кого брать данные, когда они будут скачиваться другим демоном
		size_ := utils.DecodeLength(&binaryData)
		newDataHost := string(utils.BytesShift(&binaryData, size_))
		log.Debug("newDataHost: %d", newDataHost)

		// если 0 - значит вначале идет инфа о блоке, если 1 - значит сразу идет набор хэшей тр-ий
		newDataType := utils.BinToDecBytesShift(&binaryData, 1)
		log.Debug("newDataType: %d", newDataType)
		if newDataType == 0 {
			// ID блока, чтобы не скачать старый блок
			newDataBlockId := utils.BinToDecBytesShift(&binaryData, 3)
			log.Debug("newDataBlockId: %d / blockId: %d", newDataBlockId, blockId)
			// нет смысла принимать старые блоки
			if newDataBlockId >= blockId {
				// Это хэш для соревнования, у кого меньше хэш
				newDataHash := utils.BinToHex(utils.BytesShift(&binaryData, 32))
				err = t.ExecSql(`DELETE FROM queue_blocks WHERE hex(hash) = ?`, newDataHash)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				err = t.ExecSql(`
						INSERT INTO queue_blocks (
							hash,
							host,
							block_id
						) VALUES (
							[hex],
							?,
							?
						)`, newDataHash, newDataHost, newDataBlockId)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
			}
		}
		log.Debug("binaryData: %x", binaryData)
		var needTx []byte
		// Разбираем список транзакций
		if len(binaryData) == 0 {
			log.Debug("%v", utils.ErrInfo("len(binaryData) == 0"))
			return
		}
		for {
			newDataTxHash := utils.BinToHex(utils.BytesShift(&binaryData, 16))
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
			if len(binaryData) == 0 {
				break
			}
		}
		if len(needTx) == 0 {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("needTx: %v", needTx)

		// в 4-х байтах пишем размер данных, которые пошлем далее
		size := utils.DecToBin(len(needTx), 4)
		_, err = t.Conn.Write(size)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("size: %v", len(needTx))
		log.Debug("encData: %x", needTx)
		// далее шлем сами данные
		_, err = t.Conn.Write(needTx)
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

			binaryTxs := make([]byte, dataSize)
			_, err = io.ReadFull(t.Conn, binaryTxs)
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
				if int64(len(txBinData)) > consts.MAX_TX_SIZE {
					log.Debug("%v", utils.ErrInfo("len(txBinData) > max_tx_size"))
					return
				}

				log.Debug("INSERT INTO queue_tx (hash, data) %s, %d, %s", utils.Md5(txBinData), txHex)
				err = t.ExecSql(`INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])`, utils.Md5(txBinData), txHex)
				if len(txBinData) == 0 {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
			}
		}
	}
}
