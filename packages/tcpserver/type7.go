// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package tcpserver

import (
	"github.com/EGaaS/go-mvp/packages/utils"
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
