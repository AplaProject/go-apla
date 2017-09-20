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

package parser

import (
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

// ParseDataFull checks the condiitions and proceeds of transactions
// frontal check + adding the data from the block to a table and info_block
func (p *Parser) ParseDataFull(blockGenerator bool) error {
	var txType int
	p.dataPre()
	if p.dataType != 0 {
		// parse only blocks
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	if blockGenerator {
		err = p.GetInfoBlock()
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}

	err = p.CheckBlockHeader()
	if err != nil {
		return utils.ErrInfo(err)
	}

	logging.WriteSelectiveLog("DELETE FROM transactions WHERE used = 1")
	afect, err := model.DeleteUsedTransactions()
	if err != nil {
		logging.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	logging.WriteSelectiveLog("afect: " + converter.Int64ToStr(afect))

	txCounter := make(map[int64]int64)
	p.fullTxBinaryData = p.BinaryData
	var txForRollbackTo []byte
	if len(p.BinaryData) > 0 {
		for {
			// transactions processing can take a lot of time
			log.Debugf("parse transaction: block data = %+v, transctions data = %x", p.BlockData, p.BinaryData)
			transactionSize, err := converter.DecodeLength(&p.BinaryData)
			if err != nil {
				log.Fatal(err)
			}
			if len(p.BinaryData) == 0 {
				return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
			}

			// get next transaction
			transactionBinaryData := converter.BytesShift(&p.BinaryData, transactionSize)
			transactionBinaryDataFull := transactionBinaryData

			// add the the transaction in a set of transactions for RollbackTo where we will go in reverse order
			txForRollbackTo = append(txForRollbackTo, converter.EncodeLengthPlusData(transactionBinaryData)...)

			err = p.CheckLogTx(transactionBinaryDataFull, false, false)
			if err != nil {
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("rollback error: %s", err0)
				}
				return utils.ErrInfo(err)
			}

			hashFull, err := crypto.Hash(transactionBinaryDataFull)
			if err != nil {
				log.Fatal(err)
			}
			logging.WriteSelectiveLog("UPDATE transactions SET used=1 WHERE hex(hash) = " + string(converter.BinToHex(hashFull)))
			affect, err := model.MarkTransactionUsed(hashFull)
			if err != nil {
				logging.WriteSelectiveLog(err)
				logging.WriteSelectiveLog("RollbackTo")
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("rollback error: %s", err0)
				}
				return utils.ErrInfo(err)
			}
			logging.WriteSelectiveLog("affected: " + converter.Int64ToStr(affect))

			hash, err := crypto.Hash(transactionBinaryData)
			if err != nil {
				log.Fatal(err)
			}

			p.TxHash = hash
			p.TxBinaryData = transactionBinaryData
			txType = int(converter.BinToDecBytesShift(&p.TxBinaryData, 1))
			p.TxSlice, _, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				err0 := p.RollbackTo(txForRollbackTo, true)
				if err0 != nil {
					log.Error("error: %v", err0)
				}
				return err
			}

			if p.BlockData.BlockID > 1 && p.TxContract == nil {
				var userID int64
				if len(p.TxSlice) > 3 {
					if !utils.CheckInputData(p.TxSlice[3], "int64") {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					}
					userID = converter.BytesToInt64(p.TxSlice[3])
				} else {
					return utils.ErrInfo(fmt.Errorf("empty user_id"))
				}

				// count for each user how many transactions from him are in the block
				txCounter[userID]++

				// to prevent the possibility when 1 user can send a 10-gigabyte dos-block which will fill with his own transactions
				if txCounter[userID] > int64(syspar.GetMaxBlockUserTx()) {
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
				}
			}
			if p.TxContract == nil {
				// time in the transaction cannot be more than MAX_TX_FORW seconds of block time
				// and time in transaction cannot be less than -24 of block time
				if converter.BytesToInt64(p.TxSlice[2])-consts.MAX_TX_FORW > p.BlockData.Time || converter.BytesToInt64(p.TxSlice[2]) < p.BlockData.Time-consts.MAX_TX_BACK {
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
				}

				// check if such type of transaction exists
				_, ok := consts.TxTypes[converter.BytesToInt(p.TxSlice[1])]
				if !ok {
					return utils.ErrInfo(fmt.Errorf("nonexistent type"))
				}
			} else {
				if p.TxSmart.Time-consts.MAX_TX_FORW > p.BlockData.Time || p.TxSmart.Time < p.BlockData.Time-consts.MAX_TX_BACK {
					return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
				}

			}
			p.TxMap = map[string][]byte{}

			p.TxIds++
			p.TxUsedCost = decimal.New(0, 0)
			p.TxCost = 0
			if p.TxContract != nil {
				// check that there are enough money in CallContract
				err := p.CallContract(smart.CallInit | smart.CallCondition | smart.CallAction)
				// pay for CPU resources
				p.payFPrice()
				if err != nil {
					if p.TxContract.Called == smart.CallCondition || p.TxContract.Called == smart.CallAction {
						err0 := p.RollbackTo(txForRollbackTo, false)
						if err0 != nil {
							log.Error("error: %v", err0)
						}
					}
					return utils.ErrInfo(err)
				}
			} else {
				MethodName := consts.TxTypes[txType]
				parser, err := GetParser(p, MethodName)
				if err != nil {
					return utils.ErrInfo(err)
				}
				log.Debug("MethodName", MethodName+"Init")
				err = parser.Init()
				if _, ok := err.(error); ok {
					log.Error("error: %v", err)
					return utils.ErrInfo(err.(error))
				}

				log.Debug("MethodName", MethodName+"Front")
				err = parser.Validate()
				if _, ok := err.(error); ok {
					log.Error("error: %v", err)
					err0 := p.RollbackTo(txForRollbackTo, true)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(err.(error))
				}

				log.Debug("MethodName", MethodName)
				err = parser.Action()
				// pay for CPU resources
				p.payFPrice()
				if _, ok := err.(error); ok {
					log.Error("error: %v", err)
					err0 := p.RollbackTo(txForRollbackTo, false)
					if err0 != nil {
						log.Error("error: %v", err0)
					}
					return utils.ErrInfo(err.(error))
				}
			}
			// let user know that his transaction  is added in the block
			ts := &model.TransactionStatus{}
			ts.UpdateBlockID(p.BlockData.BlockID, hashFull)
			log.Debug("UPDATE transactions_status SET block_id = %d WHERE hex(hash) = %s", p.BlockData.BlockID, hashFull)

			// here was a time(). That means if blocks with the same hashes of transactions were in the chain of blocks, ParseDataFull would return the error
			err = InsertInLogTx(transactionBinaryDataFull, converter.BytesToInt64(p.TxMap["time"]))
			if err != nil {
				return utils.ErrInfo(err)
			}

			if len(p.BinaryData) == 0 {
				break
			}
		}
	}
	if blockGenerator {
		p.UpdBlockInfo()
		p.InsertIntoBlockchain()
	} else {
		p.UpdBlockInfo()
	}
	return nil
}
