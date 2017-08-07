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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// Type2 serves requests from disseminator
func (t *TCPServer) Type2(r *DisRequest) (*DisTrResponse, error) {
	binaryData := r.Data
	// take the transactions from usual users but not nodes.
	_, _, decryptedBinData, err := t.DecryptData(&binaryData)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	if int64(len(binaryData)) > consts.MAX_TX_SIZE {
		return nil, utils.ErrInfo("len(txBinData) > max_tx_size")
	}

	if len(binaryData) < 5 {
		return nil, utils.ErrInfo("len(binaryData) < 5")
	}

	decryptedBinDataFull := decryptedBinData
	hash, err := crypto.Hash(decryptedBinDataFull)
	if err != nil {
		log.Fatal(err)
	}

	hash = converter.BinToHex(hash)
	err = model.DeleteQueuedTransaction(hash)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	hexBinData := converter.BinToHex(decryptedBinDataFull)
	log.Debug("INSERT INTO queue_tx (hash, data) (%s, %s)", hash, hexBinData)
	err = model.InsertIntoQueueTransaction(hash, hexBinData, 0)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	return &DisTrResponse{}, nil
}
