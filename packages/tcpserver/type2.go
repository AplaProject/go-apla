package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (t *TcpServer) Type2() {
	// размер данных
	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := utils.BinToDec(buf)
	log.Debug("size: %d", size)
	if size < 10485760 {
		// сами данные
		binaryData := make([]byte, size)
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		/*
		 * Прием тр-ий от простых юзеров, а не нодов. Вызывается демоном disseminator
		 * */
		_, _, decryptedBinData, err := t.DecryptData(&binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("decryptedBinData: %x", decryptedBinData)
		// проверим размер
		if int64(len(binaryData)) > consts.MAX_TX_SIZE {
			log.Debug("%v", utils.ErrInfo("len(txBinData) > max_tx_size"))
			return
		}
		if len(binaryData) < 5 {
			log.Debug("%v", utils.ErrInfo("len(binaryData) < 5"))
			return
		}
		decryptedBinDataFull := decryptedBinData
		txType := utils.BytesShift(&decryptedBinData, 1) // type
		txTime := utils.BytesShift(&decryptedBinData, 4) // time
		log.Debug("txType: %d", utils.BinToDec(txType))
		log.Debug("txTime: %d", utils.BinToDec(txTime))
		size := utils.DecodeLength(&decryptedBinData)
		log.Debug("size: %d", size)
		if int64(len(decryptedBinData)) < size {
			log.Debug("%v", utils.ErrInfo("len(binaryData) < size"))
			return
		}
		userId := utils.BytesToInt64(utils.BytesShift(&decryptedBinData, size))
		log.Debug("userId: %d", userId)
		highRate := 0
		if userId == 1 {
			highRate = 1
		}
		// заливаем тр-ию в БД
		err = t.ExecSql(`DELETE FROM queue_tx WHERE hex(hash) = ?`, utils.Md5(decryptedBinDataFull))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("INSERT INTO queue_tx (hash, high_rate, data) (%s, %d, %s)", utils.Md5(decryptedBinDataFull), highRate, utils.BinToHex(decryptedBinDataFull))
		err = t.ExecSql(`INSERT INTO queue_tx (hash, high_rate, data) VALUES ([hex], ?, [hex])`, utils.Md5(decryptedBinDataFull), highRate, utils.BinToHex(decryptedBinDataFull))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
