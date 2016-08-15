package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

/*
 * от disseminator
 */

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
	if size < consts.MAX_TX_SIZE {
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
		err = t.ExecSql(`DELETE FROM queue_tx WHERE hex(hash) = ?`, utils.Md5(decryptedBinDataFull))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("INSERT INTO queue_tx (hash, data) (%s, %s)", utils.Md5(decryptedBinDataFull), utils.BinToHex(decryptedBinDataFull))
		err = t.ExecSql(`INSERT INTO queue_tx (hash, data) VALUES ([hex], ?, [hex])`, utils.Md5(decryptedBinDataFull), utils.BinToHex(decryptedBinDataFull))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
