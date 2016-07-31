package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
)

func (t *TcpServer) Type9() {
	/* Делаем запрос на указанную ноду, чтобы получить оттуда номер макс. блока
	 * запрос шлет демон blocksCollection
	 */
	// размер данных
	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := utils.BinToDec(buf)
	if size < 10485760 {
		// сами данные
		binaryData := make([]byte, size)
		/*_, err = t.Conn.Read(binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}*/
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		//blockId := utils.BinToDecBytesShift(&binaryData, 4)
		host, err := utils.ProtectedCheckRemoteAddrAndGetHost(&binaryData, t.Conn)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// шлем данные указанному хосту
		conn2, err := utils.TcpConn(host)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		defer conn2.Close()
		// шлем тип данных
		_, err = conn2.Write(utils.DecToBin(10, 2))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// в ответ получаем номер блока
		blockIdBin := make([]byte, 4)
		_, err = conn2.Read(blockIdBin)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// и возвращаем номер блока демону, который этот запрос прислал
		_, err = t.Conn.Write(blockIdBin)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
