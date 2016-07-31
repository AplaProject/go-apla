package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"io/ioutil"
	"os"
)

func (t *TcpServer) Type12() {

	/* Получаем данные от send_promised_amount_to_pool */
	log.Debug("Type12")
	// размер данных
	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := utils.BinToDec(buf)
	log.Debug("size: %d", size)
	if size < 64<<20 {
		// сами данные
		log.Debug("read data")
		binaryData := make([]byte, size)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		//log.Debug("binaryData %x", binaryData)
		userId := utils.BinToDec(utils.BytesShift(&binaryData, 5))
		currencyId := utils.BinToDec(utils.BytesShift(&binaryData, 1))
		log.Debug("userId %d", userId)
		log.Debug("currencyId %d", currencyId)
		// проверим, есть ли такой юзер на пуле
		inPool, err := t.Single(`SELECT user_id FROM community WHERE user_id=?`, userId).Int64()
		if inPool <= 0 {
			log.Error("%v", utils.ErrInfo("inPool<=0"))
			_, err = t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		log.Debug("inPool %d", inPool)
		filesSign := utils.BytesShift(&binaryData, utils.DecodeLength(&binaryData))
		log.Debug("filesSign %x", filesSign)
		size := utils.DecodeLength(&binaryData)
		log.Debug("size %d", size)
		data := utils.BytesShift(&binaryData, size)
		//log.Debug("data %x", data)
		fileType := utils.BinToDec(utils.BytesShift(&data, 1))
		log.Debug("fileType %d", fileType)
		fileName := utils.Int64ToStr(userId) + "_promised_amount_"+utils.Int64ToStr(currencyId)+".mp4"
        forSign := string(utils.DSha256((data)))
		log.Debug("forSign %s", forSign)
		err = ioutil.WriteFile(os.TempDir()+"/"+fileName, data, 0644)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}

		if len(forSign) == 0 {
			log.Error("%v", utils.ErrInfo("len(forSign) == 0"))
			_, err = t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		// проверим подпись
		publicKey, err := t.GetUserPublicKey(userId)
		resultCheckSign, err := utils.CheckSign([][]byte{[]byte(publicKey)}, forSign, utils.HexToBin(filesSign), true)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			_, err = t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		if resultCheckSign {
			utils.CopyFileContents(os.TempDir()+"/"+fileName, *utils.Dir+"/public/"+fileName)
		} else {
			os.Remove(os.TempDir()+"/"+fileName)
		}

		// и возвращаем статус
		_, err = t.Conn.Write(utils.DecToBin(1, 1))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	} else {
		log.Error("%v", utils.ErrInfo("size>32mb"))
	}
}
