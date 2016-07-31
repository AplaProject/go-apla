package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (t *TcpServer) Type10() {

	/* Выдаем номер макс. блока
	 * запрос шлет демон blocksCollection
	 */
	blockId, err := t.Single("SELECT block_id FROM info_block").Int64()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	_, err = t.Conn.Write(utils.DecToBin(blockId, 4))
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
}
