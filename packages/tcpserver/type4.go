package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (t *TcpServer) Type4() {
	// данные присылает демон confirmations
	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	blockId := utils.BinToDec(buf)
	// используется для учета кол-ва подвержденных блоков, т.е. тех, которые есть у большинства нодов
	hash, err := t.Single("SELECT hash FROM block_chain WHERE id =  ?", blockId).String()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		t.Conn.Write(utils.DecToBin(0, 1))
		return
	}
	_, err = t.Conn.Write([]byte(hash))
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
}
