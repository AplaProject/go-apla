package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
)

func (t *TcpServer) Type3() {
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
		/*
		 * Пересылаем тр-ию, полученную по локальной сети, конечному ноду, указанному в первых 100 байтах тр-ии
		 * от демона disseminator
		* */
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
		_, err = conn2.Write(utils.DecToBin(2, 2))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		err = utils.WriteSizeAndDataTCPConn(binaryData, conn2)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
