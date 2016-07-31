package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"io/ioutil"
	"os"
)

func (t *TcpServer) Type11() {

	/* Получаем данные от send_to_pool */
	log.Debug("Type11")
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
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		//log.Debug("binaryData %x", binaryData)
		userId := utils.BinToDec(utils.BytesShift(&binaryData, 5))
		log.Debug("userId %d", userId)
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
		forSign := ""
		var files []string
		for i := 0; i < 3; i++ {
			size := utils.DecodeLength(&binaryData)
			log.Debug("size %d", size)
			data := utils.BytesShift(&binaryData, size)
			//log.Debug("data %x", data)
			fileType := utils.BinToDec(utils.BytesShift(&data, 1))
			log.Debug("fileType %d", fileType)
			var name string
			switch fileType {
			case 0:
				name = utils.Int64ToStr(userId) + "_user_face.jpg"
			case 1:
				name = utils.Int64ToStr(userId) + "_user_profile.jpg"
			case 2:
				name = utils.Int64ToStr(userId) + "_user_video.mp4"
				/*case 3:
					name = utils.Int64ToStr(userId)+"_user_video.webm"
				case 4:
					name = utils.Int64ToStr(userId)+"_user_video.ogv"*/
			}
			forSign = forSign + string(utils.DSha256((data))) + ","
			log.Debug("forSign %s", forSign)
			err = ioutil.WriteFile(os.TempDir()+"/"+name, data, 0644)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			files = append(files, name)
			log.Debug("files %d", files)
			if len(binaryData) == 0 {
				break
			}
		}

		if len(forSign) == 0 {
			log.Error("%v", utils.ErrInfo("len(forSign) == 0"))
			_, err = t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		if len(files) == 3 {
			forSign = forSign[:len(forSign)-1]
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
			for i := 0; i < len(files); i++ {
				utils.CopyFileContents(os.TempDir()+"/"+files[i], *utils.Dir+"/public/"+files[i])
			}
		} else {
			for i := 0; i < len(files); i++ {
				os.Remove(os.TempDir()+"/"+files[i])
			}
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
