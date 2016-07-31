package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
)

func (t *TcpServer) Type8() {
	/* делаем запрос на указанную ноду, чтобы получить оттуда тело блока
	 * запрос шлет демон blocksCollection и queueParserBlocks через p.GetBlocks()
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
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		blockId := utils.BinToDecBytesShift(&binaryData, 4)
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
		_, err = conn2.Write(utils.DecToBin(7, 2))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// шлем ID блока
		_, err = conn2.Write(utils.DecToBin(blockId, 4))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		// в ответ получаем размер данных, которые нам хочет передать сервер
		buf := make([]byte, 4)
		_, err = conn2.Read(buf)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		dataSize := utils.BinToDec(buf)
		// и если данных менее 10мб, то получаем их
		if dataSize < 10485760 {
			blockBinary := make([]byte, dataSize)
			/*_, err := conn2.Read(blockBinary)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}*/
			//blockBinary, err = ioutil.ReadAll(conn2)
			_, err = io.ReadFull(conn2, binaryData)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
			// шлем тому, кто запросил блок из демона
			_, err = t.Conn.Write(blockBinary)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
				return
			}
		}
		return
	}
}
