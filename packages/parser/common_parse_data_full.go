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
	"bytes"
	"fmt"

	"encoding/hex"
	"strings"

	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"github.com/shopspring/decimal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Block struct {
	Header     utils.BlockData
	PrevHeader *utils.BlockData
	MrklRoot   []byte
	BinData    []byte
	Parsers    []*Parser
}

func InsertBlock(data []byte) error {
	block, err := ProcessBlock(data)
	if err != nil {
		log.Errorf("process block error: %s", err)
		return err
	}

	if err := block.CheckBlock(); err != nil {
		log.Errorf("check block error: %s", err)
		return err
	}

	err = block.PlayBlockSafe()
	if err != nil {
		log.Errorf("play block failed: %s", err)
		return err
	}

	log.Debugf("block %d was inserted successfully", block.Header.BlockID)
	return nil
}

func (block *Block) PlayBlockSafe() error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		return err
	}

	err = block.playBlock(dbTransaction)
	if err != nil {
		log.Errorf("play block error: %s (start rollback)", err)
		dbTransaction.Rollback()
		return err
	}

	if err := UpdBlockInfo(dbTransaction, block); err != nil {
		dbTransaction.Rollback()
		return err
	}

	if err := InsertIntoBlockchain(dbTransaction, block); err != nil {
		dbTransaction.Rollback()
		return err
	}

	dbTransaction.Commit()
	return nil
}

func ProcessBlock(data []byte) (*Block, error) {
	if int64(len(data)) > syspar.GetMaxBlockSize() {
		return nil, utils.ErrInfo(fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]`))
	}

	buf := bytes.NewBuffer(data)
	if buf.Len() == 0 {
		return nil, fmt.Errorf("empty buffer")
	}

	block, err := parseBlock(buf)
	if err != nil {
		return nil, err
	}
	block.BinData = data

	if err := block.readPreviousBlock(); err != nil {
		return nil, err
	}

	return block, nil
}

func getAllTables() (map[string]string, error) {
	allTables, err := model.GetAllTables()
	if err != nil {
		return nil, utils.ErrInfo(err)
	}
	AllPkeys := make(map[string]string)
	for _, table := range allTables {
		col, err := model.GetFirstColumnName(table)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}
		AllPkeys[table] = col
	}
	return AllPkeys, nil
}

func parseBlock(blockBuffer *bytes.Buffer) (*Block, error) {
	header, err := ParseBlockHeader(blockBuffer)
	if err != nil {
		return nil, err
	}

	allKeys, err := getAllTables()
	if err != nil {
		return nil, err
	}
	parsers := make([]*Parser, 0)

	var mrklSlice [][]byte

	// parse transactions
	for blockBuffer.Len() > 0 {
		transactionSize, err := converter.DecodeLengthBuf(blockBuffer)
		if err != nil {
			return nil, fmt.Errorf("bad block format (%s)", err)
		}
		if blockBuffer.Len() < int(transactionSize) {
			return nil, fmt.Errorf("bad block format (transaction len is too big: %d)", transactionSize)
		}

		if transactionSize == 0 {
			return nil, fmt.Errorf("transaction size is 0")
		}

		bufTransaction := bytes.NewBuffer(blockBuffer.Next(int(transactionSize)))
		p, err := ParseTransaction(bufTransaction)
		if err != nil {
			return nil, fmt.Errorf("parse transaction error(%s)", err)
		}
		p.BlockData = &header
		p.AllPkeys = allKeys

		parsers = append(parsers, p)

		// build merkle tree
		if len(p.TxFullData) > 0 {
			dSha256Hash, err := crypto.DoubleHash(p.TxFullData)
			if err != nil {
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
		Header:   header,
		Parsers:  parsers,
		MrklRoot: utils.MerkleTreeRoot(mrklSlice),
	}, nil
}

func ParseBlockHeader(binaryBlock *bytes.Buffer) (utils.BlockData, error) {
	var block utils.BlockData
	var err error

	if binaryBlock.Len() < 9 {
		return utils.BlockData{}, fmt.Errorf("bad binary block length")
	}

	blockVersion := int(converter.BinToDec(binaryBlock.Next(1)))

	if int64(binaryBlock.Len()) > syspar.GetMaxBlockSize() {
		err = fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]  %v > %v`,
			binaryBlock.Len(), syspar.GetMaxBlockSize())

		return utils.BlockData{}, err
	}

	block.BlockID = converter.BinToDec(binaryBlock.Next(4))
	block.Time = converter.BinToDec(binaryBlock.Next(4))
	block.Version = blockVersion

	block.WalletID, err = converter.DecodeLenInt64Buf(binaryBlock)
	if err != nil {
		return utils.BlockData{}, err
	}

	if binaryBlock.Len() < 1 {
		return utils.BlockData{}, fmt.Errorf("bad block format")
	}
	block.StateID = converter.BinToDec(binaryBlock.Next(1))

	if block.BlockID > 1 {
		signSize, err := converter.DecodeLengthBuf(binaryBlock)
		if err != nil {
			return utils.BlockData{}, err
		}
		if binaryBlock.Len() < signSize {
			return utils.BlockData{}, fmt.Errorf("bad block format (no sign)")
		}
		block.Sign = binaryBlock.Next(int(signSize))
	} else {
		binaryBlock.Next(1)
	}

	return block, nil
}

func ParseTransaction(buffer *bytes.Buffer) (*Parser, error) {
	if buffer.Len() == 0 {
		return nil, fmt.Errorf("empty transaction buffer")
	}

	hash, err := crypto.Hash(buffer.Bytes())
	// or DoubleHash ?
	if err != nil {
		return nil, err
	}

	p := new(Parser)
	p.TxHash = hash
	p.TxUsedCost = decimal.New(0, 0)
	p.TxFullData = buffer.Bytes()

	txType := int64(buffer.Bytes()[0])
	p.dataType = int(txType)

	log.Debugf("parse transaction %s", consts.TxTypes[int(txType)])

	// smart contract transaction
	if IsContractTransaction(int(txType)) {
		// skip byte with transaction type
		buffer.Next(1)
		p.TxBinaryData = buffer.Bytes()
		if err := parseContractTransaction(p, buffer); err != nil {
			return nil, err
		}
		if err := p.CallContract(smart.CallInit | smart.CallCondition); err != nil {
			return nil, err
		}

		// struct transaction (only first block transaction for now)
	} else if consts.IsStruct(int(txType)) {
		p.TxBinaryData = buffer.Bytes()
		if err := parseStructTransaction(p, buffer, txType); err != nil {
			return nil, err
		}

		// all other transactions
	} else {
		// skip byte with transaction type
		buffer.Next(1)
		p.TxBinaryData = buffer.Bytes()
		if err := parseRegularTransaction(p, buffer, txType); err != nil {
			return p, err
		}
	}

	return p, nil
}

func IsContractTransaction(txType int) bool {
	return txType > 127
}

func parseContractTransaction(p *Parser, buf *bytes.Buffer) error {
	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(buf.Bytes(), &smartTx); err != nil {
		return err
	}
	p.TxPtr = nil
	p.TxSmart = &smartTx
	p.TxTime = smartTx.Time
	p.TxStateID = uint32(smartTx.StateID)
	p.TxStateIDStr = converter.UInt32ToStr(p.TxStateID)
	if p.TxStateID > 0 {
		p.TxCitizenID = smartTx.UserID
		p.TxWalletID = 0
	} else {
		p.TxCitizenID = 0
		p.TxWalletID = smartTx.UserID
	}

	contract := smart.GetContractByID(int32(smartTx.Type))
	if contract == nil {
		return fmt.Errorf(`unknown contract %d`, smartTx.Type)
	}
	forsign := smartTx.ForSign()

	p.TxContract = contract
	p.TxHeader = &smartTx.Header

	input := smartTx.Data
	p.TxData = make(map[string]interface{})

	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
			var err error
			var v interface{}
			var forv string
			var isforv bool
			switch fitem.Type.String() {
			case `uint64`:
				var val uint64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `float64`:
				var val float64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `int64`:
				v, err = converter.DecodeLenInt64(&input)
			case script.Decimal:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v, err = decimal.NewFromString(s)
			case `string`:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v = s
			case `[]uint8`:
				var b []byte
				if err := converter.BinUnmarshal(&input, &b); err != nil {
					return err
				}
				v = hex.EncodeToString(b)
			case `[]interface {}`:
				count, err := converter.DecodeLength(&input)
				if err != nil {
					return err
				}
				isforv = true
				list := make([]interface{}, 0)
				for count > 0 {
					length, err := converter.DecodeLength(&input)
					if err != nil {
						return err
					}
					if len(input) < int(length) {
						return fmt.Errorf(`input slice is short`)
					}
					list = append(list, string(input[:length]))
					input = input[length:]
					count--
				}
				if len(list) > 0 {
					slist := make([]string, len(list))
					for j, lval := range list {
						slist[j] = lval.(string)
					}
					forv = strings.Join(slist, `,`)
				}
				v = list
			}
			p.TxData[fitem.Name] = v
			if err != nil {
				return err
			}
			if strings.Index(fitem.Tags, `image`) >= 0 {
				continue
			}
			if isforv {
				v = forv
			}
			forsign += fmt.Sprintf(",%v", v)
		}
	}
	p.TxData[`forsign`] = forsign

	return nil
}

func parseStructTransaction(p *Parser, buf *bytes.Buffer, txType int64) error {
	trParser, err := GetParser(p, consts.TxTypes[int(txType)])
	if err != nil {
		return err
	}
	p.txParser = trParser

	p.TxPtr = consts.MakeStruct(consts.TxTypes[int(txType)])
	input := buf.Bytes()
	if err := converter.BinUnmarshal(&input, p.TxPtr); err != nil {
		return err
	}

	head := consts.Header(p.TxPtr)
	p.TxCitizenID = head.CitizenID
	p.TxWalletID = head.WalletID
	p.TxTime = int64(head.Time)
	p.TxType = txType
	p.TxWalletID = head.WalletID
	p.TxCitizenID = head.CitizenID
	return nil
}

func parseRegularTransaction(p *Parser, buf *bytes.Buffer, txType int64) error {
	trParser, err := GetParser(p, consts.TxTypes[int(txType)])
	if err != nil {
		return err
	}
	p.txParser = trParser

	log.Debugf("parse regular transaction: %s", consts.TxTypes[int(txType)])
	err = trParser.Init()
	if err != nil {
		log.Errorf("parser init failed: %s", err)
		return err
	}
	header := trParser.Header()
	if header == nil {
		return fmt.Errorf("tx header is nil")
	}

	p.TxHeader = header
	p.TxTime = header.Time
	p.TxType = txType
	p.TxStateID = uint32(header.StateID)
	p.TxUserID = header.UserID

	log.Debugf("transaction header: %+v", header)

	err = trParser.Validate()
	if _, ok := err.(error); ok {
		log.Errorf("transaction validate failed: %s", err)
		return utils.ErrInfo(err.(error))
	}

	return nil
}

func checkTransaction(p *Parser, checkTime int64, checkForDupTr bool) error {
	err := CheckLogTx(p.TxFullData, checkForDupTr, false)
	if err != nil {
		return utils.ErrInfo(err)
	}

	// time in the transaction cannot be more than MAX_TX_FORW seconds of block time
	if p.TxTime-consts.MAX_TX_FORW > checkTime {
		return utils.ErrInfo(fmt.Errorf("transaction time is too big"))
	}

	// time in transaction cannot be less than -24 of block time
	if p.TxTime < checkTime-consts.MAX_TX_BACK {
		return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
	}

	if p.TxContract == nil {
		if p.BlockData != nil && p.BlockData.BlockID != 1 {
			if p.TxUserID == 0 {
				return utils.ErrInfo(fmt.Errorf("emtpy user id"))
			}
		}
	}

	return nil
}

func CheckTransaction(data []byte) (*tx.Header, error) {
	trBuff := bytes.NewBuffer(data)
	p, err := ParseTransaction(trBuff)
	if err != nil {
		return nil, err
	}

	err = checkTransaction(p, time.Now().Unix(), true)
	if err != nil {
		return nil, err
	}

	return p.TxHeader, nil
}

func (block *Block) readPreviousBlock() error {
	if block.Header.BlockID == 1 {
		block.PrevHeader = &utils.BlockData{}
		return nil
	}

	var err error
	block.PrevHeader, err = GetBlockDataFromBlockChain(block.Header.BlockID - 1)
	if err != nil {
		return utils.ErrInfo(fmt.Errorf("can't get block %d", block.Header.BlockID-1))
	}

	return nil
}

func playTransaction(p *Parser) error {
	log.Debugf("play transaction: %s", consts.TxTypes[int(p.TxType)])
	// smart-contract
	if p.TxContract != nil {
		// check that there are enough money in CallContract
		if err := p.CallContract(smart.CallInit | smart.CallCondition | smart.CallAction); err != nil {
			return utils.ErrInfo(err)
		}

	} else {
		if p.txParser == nil {
			return utils.ErrInfo(fmt.Errorf("can't find parser for %d", p.TxType))
		}

		err := p.txParser.Action()
		if _, ok := err.(error); ok {
			return utils.ErrInfo(err.(error))
		}
	}
	log.Debugf("play transaction %s - ok", consts.TxTypes[int(p.TxType)])
	return nil
}

func (block *Block) playBlock(dbTransaction *model.DbTransaction) error {

	log.Debugf("start play block")
	if _, err := model.DeleteUsedTransactions(dbTransaction); err != nil {
		return err
	}

	for _, p := range block.Parsers {
		p.DbTransaction = dbTransaction

		if err := playTransaction(p); err != nil {
			// skip this transaction
			log.Errorf("play transaction error: %s", err)
			model.MarkTransactionUsed(nil, p.TxHash)
			p.processBadTransaction(p.TxHash, err.Error())
			continue
		}

		if _, err := model.MarkTransactionUsed(p.DbTransaction, p.TxHash); err != nil {
			return err
		}

		// update status
		ts := &model.TransactionStatus{}
		if err := ts.UpdateBlockID(p.DbTransaction, block.Header.BlockID, p.TxHash); err != nil {
			return err
		}
		if err := InsertInLogTx(p.DbTransaction, p.TxFullData, p.TxTime); err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

func (block *Block) CheckBlock() error {
	// exclude blocks from future
	if block.Header.Time > time.Now().Unix() {
		utils.ErrInfo(fmt.Errorf("incorrect block time"))
	}
	// is this block too early? Allowable error = error_time
	if block.PrevHeader != nil {
		if block.Header.BlockID != block.PrevHeader.BlockID+1 {
			return utils.ErrInfo(fmt.Errorf("incorrect block_id %d != %d +1", block.Header.BlockID, block.PrevHeader.BlockID))
		}
		// check time interval between blocks
		sleepTime, err := model.GetSleepTime(block.Header.WalletID, block.Header.StateID, block.PrevHeader.StateID, block.PrevHeader.WalletID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if block.PrevHeader.Time+sleepTime-block.Header.Time > consts.ERROR_TIME {
			return utils.ErrInfo(fmt.Errorf("incorrect block time"))
		}
	}

	// check each transaction
	txCounter := make(map[int64]int)
	txHashes := make(map[string]struct{})
	for _, p := range block.Parsers {
		hexHash := string(converter.BinToHex(p.TxHash))
		// check for duplicate transactions
		if _, ok := txHashes[hexHash]; ok {
			return utils.ErrInfo(fmt.Errorf("duplicate transaction %s", hexHash))
		}
		txHashes[hexHash] = struct{}{}

		// check for max transaction per user in one block
		txCounter[p.TxUserID]++
		if txCounter[p.TxUserID] > syspar.GetMaxBlockUserTx() {
			return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
		}

		if err := checkTransaction(p, block.Header.Time, false); err != nil {
			return utils.ErrInfo(err)
		}

	}

	result, err := block.CheckHash()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if !result {
		return fmt.Errorf("incorrect signature / p.PrevBlock.BlockId: %d", block.PrevHeader.BlockID)
	}
	return nil
}

func (block *Block) CheckHash() (bool, error) {
	if block.Header.BlockID == 1 {
		return true, nil
	}
	// check block signature
	if block.PrevHeader != nil {
		nodePublicKey, err := GetNodePublicKeyWalletOrCB(block.Header.WalletID, block.Header.StateID)
		if err != nil {
			return false, utils.ErrInfo(err)
		}
		if len(nodePublicKey) == 0 {
			return false, utils.ErrInfo(fmt.Errorf("empty nodePublicKey"))
		}
		// check the signature
		forSign := fmt.Sprintf("0,%d,%s,%d,%d,%d,%s", block.Header.BlockID, block.PrevHeader.Hash,
			block.Header.Time, block.Header.WalletID, block.Header.StateID, block.MrklRoot)

		log.Debugf("check block for sign: %s, key: %x", forSign, nodePublicKey)

		resultCheckSign, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, block.Header.Sign, true)
		if err != nil {
			return false, utils.ErrInfo(fmt.Errorf("err: %v / p.PrevBlock.BlockId: %d", err, block.PrevHeader.BlockID))
		}

		return resultCheckSign, nil
	}

	return true, nil
}

func MarshallBlock(header *utils.BlockData, trData [][]byte, prevHash []byte, key string) ([]byte, error) {
	var mrklArray [][]byte
	var blockDataTx []byte
	var signed []byte

	for _, tr := range trData {
		doubleHash, err := crypto.DoubleHash(tr)
		if err != nil {
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash))
		blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData([]byte(tr))...)
	}

	if key != "" {
		if len(mrklArray) == 0 {
			mrklArray = append(mrklArray, []byte("0"))
		}
		mrklRoot := utils.MerkleTreeRoot(mrklArray)

		forSign := fmt.Sprintf("0,%d,%s,%d,%d,%d,%s",
			header.BlockID, prevHash, header.Time, header.WalletID, header.StateID, mrklRoot)

		var err error
		signed, err = crypto.Sign(key, forSign)
		if err != nil {
			return nil, err
		}
		log.Debugf("generate block for sign: %s, key: %x, signed: %x", forSign, key, signed)
	}

	var buf bytes.Buffer
	// fill header
	buf.Write(converter.DecToBin(header.Version, 1))
	buf.Write(converter.DecToBin(header.BlockID, 4))
	buf.Write(converter.DecToBin(header.Time, 4))
	buf.Write(converter.EncodeLenInt64InPlace(header.WalletID))
	buf.Write(converter.DecToBin(header.StateID, 1))
	buf.Write(converter.EncodeLengthPlusData(signed))
	// data
	buf.Write(blockDataTx)

	return buf.Bytes(), nil
}
