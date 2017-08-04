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

package tcpserver

import (
	"io"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

/*
 * от disseminator
 */

func (t *TCPServer) Type1() {
	log.Debug("dataType: 1")
	// размер данных
	// data size
	buf := make([]byte, 4)
	n, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := converter.BinToDec(buf)
	log.Debug("size: %v / n: %v", size, n)
	if size < 10485760 {
		// сами данные
		// data itself
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
		 * принимаем список тр-ий от а disseminator, которые есть у отправителя
		 * Блоки не качаем тут, т.к. может быть цепочка блоков, а их качать долго
		 * тр-ии качаем тут, т.к. они мелкие и точно скачаются за 60 сек
		 * */
		// get the list of transactions which belong to the sender from 'disseminator' daemon
		// do not load the blocks here because here could be the chain of blocks that are loaded for a long time
		// download the transactions here, because they are small and definitely will be downloaded in 60 sec

		/*
		 * структура данных: // data structure
		 * type - 1 байт. 0 - блок, 1 - список тр-ий // type - 1 byte. 0 - block, 1 - list of transactions
		 * {если type==1}: // {if type==1}:
		 * <любое кол-во следующих наборов> // <any number of the next sets>
		 * high_rate - 1 байт // high_rate - 1 byte
		 * tx_hash - 16 байт // tx_hash - 16 bytes
		 * </>
		 * {если type==0}: // {if type==0}:
		 * block_id - 3 байта // block_id - 3 bytes
		 * hash - 32 байт // hash - 32 bytes
		 * <любое кол-во следующих наборов> // <any number of the next sets>
		 * high_rate - 1 байт // high_rate - 1 byte
		 * tx_hash - 16 байт // tx_hash - 16 bytes
		 * </>
		 * */
		blockID, err := t.GetBlockID()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("binaryData: %x", binaryData)
		// full_node_id отправителя, чтобы знать у кого брать данные, когда они будут скачиваться другим демоном
		// full_node_id of the sender to know where to take a data when it will be downloaded by another daemon
		fullNodeID := converter.BinToDecBytesShift(&binaryData, 2)
		log.Debug("fullNodeID: %d", fullNodeID)
		// если 0 - значит вначале идет инфа о блоке, если 1 - значит сразу идет набор хэшей тр-ий
		// if 0, it means information about the block goes initially, if 1, it means the set of transactions hashes goes at first
		newDataType := converter.BinToDecBytesShift(&binaryData, 1)
		log.Debug("newDataType: %d", newDataType)
		if newDataType == 0 {
			// ID блока, чтобы не скачать старый блок
			// block id for not to upload the old block
			newDataBlockID := converter.BinToDecBytesShift(&binaryData, 3)
			log.Debug("newDataBlockID: %d / blockID: %d", newDataBlockID, blockID)
			// нет смысла принимать старые блоки
			// there is no reason to accept the old blocks
			if newDataBlockID >= blockID {
				newDataHash := converter.BinToHex(converter.BytesShift(&binaryData, 32))
				err = t.ExecSQL(`
						INSERT INTO queue_blocks (
							hash,
							full_node_id,
							block_id
						) VALUES (
							[hex],
							?,
							?
						) ON CONFLICT DO NOTHING`, newDataHash, fullNodeID, newDataBlockID)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				log.Debug("INSERT INTO queue_blocks")
			} else {
				// просто удалим хэш блока, что бы далее проверить тр-ии
				// just delete the hash of the block to check transactions further
				converter.BinToHex(converter.BytesShift(&binaryData, 32))
			}
		}
		log.Debug("binaryData: %x", binaryData)
		var needTx []byte
		// Разбираем список транзакций, но их может и не быть
		// Parse the list of transactions, but they could be absent
		if len(binaryData) == 0 {
			log.Debug("%v", utils.ErrInfo("len(binaryData) == 0"))
			log.Debug("%x", converter.Int64ToByte(int64(0)))
			_, err = t.Conn.Write(converter.DecToBin(0, 4))
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			return
		}
		for {
			if len(binaryData) == 0 { // если пришли сюда из continue, то binaryData может уже быть пустым
				// if we came here from 'continue', then binaryData could already be empty
				break
			}
			newDataTxHash := converter.BinToHex(converter.BytesShift(&binaryData, 16))
			if len(newDataTxHash) == 0 {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			log.Debug("newDataTxHash %s", newDataTxHash)
			// проверим, нет ли у нас такой тр-ии
			// check if we have such a transaction
			exists, err := t.Single("SELECT count(hash) FROM log_transactions WHERE hex(hash) = ?", newDataTxHash).Int64()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			if exists > 0 {
				log.Debug("exists")
				continue
			}

			// проверим, нет ли у нас такой тр-ии
			// check if we have such a transaction
			exists, err = t.Single("SELECT count(hash) FROM transactions WHERE hex(hash) = ?", newDataTxHash).Int64()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			if exists > 0 {
				log.Debug("exists")
				continue
			}

			// проверим, нет ли у нас такой тр-ии
			// check if we have such a transaction
			exists, err = t.Single("SELECT count(hash) FROM queue_tx WHERE hex(hash) = ?", newDataTxHash).Int64()
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			if exists > 0 {
				log.Debug("exists")
				continue
			}
			needTx = append(needTx, converter.HexToBin(newDataTxHash)...)
			if len(binaryData) == 0 {
				break
			}
		}
		if len(needTx) == 0 {
			log.Debug("len(needTx) == 0")
			_, err = t.Conn.Write(converter.DecToBin(0, 4))
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			return
		}
		log.Debug("needTx: %v", needTx)

		// в 4-х байтах пишем размер данных, которые пошлем далее
		// record the size of data in 4 bytes, we will send this data further
		size := converter.DecToBin(len(needTx), 4)
		_, err = t.Conn.Write(size)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("size: %v", len(needTx))
		log.Debug("encData: %x", needTx)
		// далее шлем сами данные
		// then send the data itself
		_, err = t.Conn.Write(needTx)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// в ответ получаем размер данных, которые нам хочет передать сервер
		// as a response we obtain the size of data, which the server wants to transfer to us
		buf := make([]byte, 4)
		_, err = t.Conn.Read(buf)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		dataSize := converter.BinToDec(buf)
		log.Debug("dataSize %v", dataSize)
		// и если данных менее 10мб, то получаем их
		// if the size of data is less than 10mb, we receive them
		if dataSize < 10485760 {

			binaryTxs := make([]byte, dataSize)
			_, err = io.ReadFull(t.Conn, binaryTxs)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}

			log.Debug("binaryTxs %x", binaryTxs)
			for {
				txSize, err := converter.DecodeLength(&binaryTxs)
				if err != nil {
					log.Fatal(err)
				}
				if int64(len(binaryTxs)) < txSize {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				txBinData := converter.BytesShift(&binaryTxs, txSize)
				if len(txBinData) == 0 {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
				txHex := converter.BinToHex(txBinData)
				// проверим размер
				// check the size
				if int64(len(txBinData)) > sql.SysInt64(sql.MaxTxSize) {
					log.Debug("%v", utils.ErrInfo("len(txBinData) > max_tx_size"))
					return
				}

				hash, err := crypto.Hash(txBinData)
				if err != nil {
					log.Fatal(err)
				}
				hash = converter.BinToHex(hash)
				log.Debug("INSERT INTO queue_tx (hash, data, from_gate) %s, %s, 1", hash, txHex)
				err = t.ExecSQL(`INSERT INTO queue_tx (hash, data, from_gate) VALUES ([hex], [hex], 1)`, hash, txHex)
				if len(txBinData) == 0 {
					log.Error("%v", utils.ErrInfo(err))
					return
				}
			}
		}
	}
}
