// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package block

import (
	"bytes"
	"fmt"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"
	log "github.com/sirupsen/logrus"
)

// MarshallBlock is marshalling block
func MarshallBlock(header *utils.BlockData, trData [][]byte, prev *utils.BlockData, key string) ([]byte, error) {
	var mrklArray [][]byte
	var blockDataTx []byte
	var signed []byte
	logger := log.WithFields(log.Fields{"block_id": header.BlockID, "block_hash": header.Hash, "block_time": header.Time, "block_version": header.Version, "block_wallet_id": header.KeyID, "block_state_id": header.EcosystemID})

	for _, tr := range trData {
		doubleHash, err := crypto.DoubleHash(tr)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing transaction")
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash))
		blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData(tr)...)
	}

	if key != "" {
		if len(mrklArray) == 0 {
			mrklArray = append(mrklArray, []byte("0"))
		}
		mrklRoot := utils.MerkleTreeRoot(mrklArray)

		var err error
		signSource := header.ForSign(prev, mrklRoot)
		signed, err = crypto.SignString(key, signSource)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing block")
			return nil, err
		}
	}

	buf := new(bytes.Buffer)

	// fill header
	buf.Write(converter.DecToBin(header.Version, 2))
	buf.Write(converter.DecToBin(header.BlockID, 4))
	buf.Write(converter.DecToBin(header.Time, 4))
	buf.Write(converter.DecToBin(header.EcosystemID, 4))
	buf.Write(converter.EncodeLenInt64InPlace(header.KeyID))
	buf.Write(converter.DecToBin(header.NodePosition, 1))
	buf.Write(converter.EncodeLengthPlusData(prev.RollbacksHash))

	// fill signature
	buf.Write(converter.EncodeLengthPlusData(signed))

	// data
	buf.Write(blockDataTx)

	return buf.Bytes(), nil
}

func UnmarshallBlock(blockBuffer *bytes.Buffer, fillData bool) (*Block, error) {
	header, prev, err := utils.ParseBlockHeader(blockBuffer)
	if err != nil {
		return nil, err
	}

	logger := log.WithFields(log.Fields{"block_id": header.BlockID, "block_time": header.Time, "block_wallet_id": header.KeyID,
		"block_state_id": header.EcosystemID, "block_hash": header.Hash, "block_version": header.Version})
	transactions := make([]*transaction.Transaction, 0)

	var mrklSlice [][]byte

	// parse transactions
	for blockBuffer.Len() > 0 {
		transactionSize, err := converter.DecodeLengthBuf(blockBuffer)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("transaction size is 0")
			return nil, fmt.Errorf("bad block format (%s)", err)
		}
		if blockBuffer.Len() < int(transactionSize) {
			logger.WithFields(log.Fields{"size": blockBuffer.Len(), "match_size": int(transactionSize), "type": consts.SizeDoesNotMatch}).Error("transaction size does not matches encoded length")
			return nil, fmt.Errorf("bad block format (transaction len is too big: %d)", transactionSize)
		}

		if transactionSize == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("transaction size is 0")
			return nil, fmt.Errorf("transaction size is 0")
		}

		bufTransaction := bytes.NewBuffer(blockBuffer.Next(int(transactionSize)))
		t, err := transaction.UnmarshallTransaction(bufTransaction, fillData)
		if err != nil {
			if t != nil && t.TxHash != nil {
				transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			}
			return nil, fmt.Errorf("parse transaction error(%s)", err)
		}
		t.BlockData = &header

		transactions = append(transactions, t)

		// build merkle tree
		if len(t.TxFullData) > 0 {
			dSha256Hash, err := crypto.DoubleHash(t.TxFullData)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing tx full data")
				return nil, err
			}
			dSha256Hash = converter.BinToHex(dSha256Hash)
			mrklSlice = append(mrklSlice, dSha256Hash)
		}
	}

	if len(mrklSlice) == 0 {
		mrklSlice = append(mrklSlice, []byte("0"))
	}

	return &Block{
		Header:            header,
		PrevRollbacksHash: prev.RollbacksHash,
		Transactions:      transactions,
		MrklRoot:          utils.MerkleTreeRoot(mrklSlice),
	}, nil
}
