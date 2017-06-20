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
	"io"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
 * от disseminator
// from disseminator
*/

func (t *TCPServer) Type2() {
	// размер данных
	// data size
	buf := make([]byte, 4)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	size := converter.BinToDec(buf)
	log.Debug("size: %d", size)
	if size < consts.MAX_TX_SIZE {
		// сами данные
		// data size
		binaryData := make([]byte, size)
		//binaryData, err = ioutil.ReadAll(t.Conn)
		_, err = io.ReadFull(t.Conn, binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		/*
					 * Прием тр-ий от простых юзеров, а не нодов. Вызывается демоном disseminator
			// take the transactions from usual users but not nodes. It's called by 'disseminator' daemon
					 * */
		_, _, decryptedBinData, err := t.DecryptData(&binaryData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("decryptedBinData: %x", decryptedBinData)
		// проверим размер
		// check the size
		if int64(len(binaryData)) > consts.MAX_TX_SIZE {
			log.Debug("%v", utils.ErrInfo("len(txBinData) > max_tx_size"))
			return
		}
		if len(binaryData) < 5 {
			log.Debug("%v", utils.ErrInfo("len(binaryData) < 5"))
			return
		}
		decryptedBinDataFull := decryptedBinData
		hash, err := crypto.Hash(decryptedBinDataFull)
		if err != nil {
			log.Fatal(err)
		}
		hash = converter.BinToHex(hash)
		err = t.ExecSQL(`DELETE FROM queue_tx WHERE hex(hash) = ?`, hash)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
		log.Debug("INSERT INTO queue_tx (hash, data) (%s, %s)", hash, converter.BinToHex(decryptedBinDataFull))
		err = t.ExecSQL(`INSERT INTO queue_tx (hash, data) VALUES ([hex], ?, [hex])`, hash, converter.BinToHex(decryptedBinDataFull))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
