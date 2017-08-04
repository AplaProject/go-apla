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

package sql

import (
	"fmt"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// GetEndBlockID returns the end block id
func GetEndBlockID() (int64, error) {

	if _, err := os.Stat(*utils.Dir + "/public/blockchain"); os.IsNotExist(err) {
		return 0, nil
	}

	// размер блока, записанный в 5-и последних байтах файла blockchain
	// size of a block recorded into the last 5 bytes of blockchain file
	fname := *utils.Dir + "/public/blockchain"
	file, err := os.Open(fname)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fi.Size() == 0 {
		return 0, utils.ErrInfo("/public/blockchain size=0")
	}

	// размер блока, записанный в 5-и последних байтах файла blockchain
	// size of a block recorded into the last 5 bytes of blockchain file
	_, err = file.Seek(-5, 2)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	buf := make([]byte, 5)
	_, err = file.Read(buf)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	size := converter.BinToDec(buf)
	if size > SysInt64(MaxBlockSize) {
		return 0, utils.ErrInfo("size > MAX_BLOCK_SIZE")
	}
	// block itself
	_, err = file.Seek(-(size + 5), 2)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	dataBinary := make([]byte, size+5)
	_, err = file.Read(dataBinary)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	// размер (id блока + тело блока)
	// size (block id + body of a block)
	converter.BinToDecBytesShift(&dataBinary, 5)
	return converter.BinToDecBytesShift(&dataBinary, 5), nil
}

// GetMrklroot returns MerkleTreeRoot
func GetMrklroot(binaryData []byte, first bool) ([]byte, error) {
	var mrklSlice [][]byte
	var txSize int64
	// [error] парсим после вызова функции
	// parse [error] after the calling of a function
	if len(binaryData) > 0 {
		for {
			// чтобы исключить атаку на переполнение памяти
			// to exclude an attack on memory overflow
			if !first {
				if txSize > SysInt64(MaxTxSize) {
					return nil, utils.ErrInfoFmt("[error] MAX_TX_SIZE")
				}
			}
			txSize, err := converter.DecodeLength(&binaryData)
			if err != nil {
				log.Fatal(err)
			}

			// отчекрыжим одну транзакцию от списка транзакций
			// separate one transaction from the list of transactions
			if txSize > 0 {
				transactionBinaryData := converter.BytesShift(&binaryData, txSize)
				dSha256Hash, err := crypto.DoubleHash(transactionBinaryData)
				if err != nil {
					log.Fatal(err)
				}
				dSha256Hash = converter.BinToHex(dSha256Hash)
				mrklSlice = append(mrklSlice, dSha256Hash)
				//if len(transactionBinaryData) > 500000 {
				//	ioutil.WriteFile(string(dSha256Hash)+"-"+Int64ToStr(txSize), transactionBinaryData, 0644)
				//}
			}

			// чтобы исключить атаку на переполнение памяти
			// to exclude an attack on memory overflow
			if !first {
				if len(mrklSlice) > SysInt(MaxTxCount) {
					return nil, utils.ErrInfo(fmt.Errorf("[error] MAX_TX_COUNT (%v > %v)", len(mrklSlice), SysInt(MaxTxCount)))
				}
			}
			if len(binaryData) == 0 {
				break
			}
		}
	} else {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	if len(mrklSlice) == 0 {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	return utils.MerkleTreeRoot(mrklSlice), nil
}
