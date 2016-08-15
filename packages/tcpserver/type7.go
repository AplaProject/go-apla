package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/* Выдаем тело указанного блока
 * запрос шлет демон blocksCollection и queue_parser_blocks через p.GetBlocks()
 */

func (t *TcpServer) Type7() {

	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	blockId := utils.BinToDec(buf)
	block, err := t.Single("SELECT data FROM block_chain WHERE id  =  ?", blockId).Bytes()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}

	log.Debug("blockId %d", blockId)
	log.Debug("block %x", block)
	err = utils.WriteSizeAndData(block, t.Conn)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
}
