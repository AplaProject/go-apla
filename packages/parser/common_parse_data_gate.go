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
	"errors"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

/*
Обработка данных (блоков или транзакций), пришедших с гейта. Только проверка.
Data processing (blocks or transactions) gotten from a gate. Just checking.
*/
func (p *Parser) ParseDataGate(onlyTx bool) (*tx.Header, error) {
	var err error
	p.dataPre()
	p.ParseInit()
	transactionBinaryData := p.BinaryData
	p.TxBinaryData = transactionBinaryData
	var transactionBinaryDataFull []byte
	var header *tx.Header

	log.Debug("p.dataType: %d", p.dataType)
	// если это транзакции (type>0), а не блок (type==0)
	// if it's transactions (type>0), but block (type==0)
	if p.dataType > 0 {

		// проверим, есть ли такой тип тр-ий
		// check if the transaction type exist
		if p.dataType < 128 && len(consts.TxTypes[p.dataType]) == 0 {
			return nil, p.ErrInfo("Incorrect tx type " + converter.IntToStr(p.dataType))
		}

		log.Debug("p.dataType: %d", p.dataType)
		transactionBinaryData = append(converter.DecToBin(int64(p.dataType), 1), transactionBinaryData...)
		transactionBinaryDataFull = transactionBinaryData

		// нет ли хэша этой тр-ии у нас в БД?
		// Does the transaction hash exist?
		err = p.CheckLogTx(transactionBinaryDataFull, true, false)
		if err != nil {
			return nil, p.ErrInfo(err)
		}

		hash, err := crypto.Hash(transactionBinaryData)
		if err != nil {
			log.Fatal(err)
		}
		p.TxHash = string(hash)

		// преобразуем бинарные данные транзакции в массив
		// transforming binary data of the transaction to an array
		log.Debug("transactionBinaryData: %x", transactionBinaryData)
		p.TxSlice, header, err = p.ParseTransaction(&transactionBinaryData)
		if err != nil {
			return nil, p.ErrInfo(err)
		}
		log.Debug("p.TxSlice", p.TxSlice)
		if len(p.TxSlice) < 3 {
			return nil, p.ErrInfo(errors.New("len(p.TxSlice) < 3"))
		}

		// время транзакции может быть немного больше, чем время на ноде.
		// у нода может быть просто не настроено время.
		// время транзакции используется только для борьбы с атаками вчерашними транзакциями.
		// А т.к. мы храним хэши в rb_transaction за 36 часов, то боятся нечего.

		// Time of transaction can be slightly longer than time of a node.
		// A node can use wrong time
		// Time of a transaction used only for fighting off attacks of yesterday transactions
		curTime := time.Now().Unix()
		if p.TxContract != nil {
			if int64(p.TxPtr.(*consts.TXHeader).Time)-consts.MAX_TX_FORW > curTime || int64(p.TxPtr.(*consts.TXHeader).Time) < curTime-consts.MAX_TX_BACK {
				return nil, p.ErrInfo(errors.New("incorrect tx time"))
			}
		} else {
			if converter.BytesToInt64(p.TxSlice[2])-consts.MAX_TX_FORW > curTime || converter.BytesToInt64(p.TxSlice[2]) < curTime-consts.MAX_TX_BACK {
				return nil, p.ErrInfo(errors.New("incorrect tx time"))
			}
			// $this->transaction_array[3] могут подсунуть пустой
			if !utils.CheckInputData(p.TxSlice[3], "bigint") {
				return nil, p.ErrInfo(errors.New("incorrect user id"))
			}
		}
	}
	// Оперативные транзакции
	// Operative transactions
	if p.TxContract != nil {
		if err := p.CallContract(smart.CallInit | smart.CallCondition); err != nil {
			return nil, utils.ErrInfo(err)
		}
	} else {
		MethodName := consts.TxTypes[p.dataType]
		parser, err := GetParser(p, MethodName)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}
		log.Debug("MethodName", MethodName+"Init")
		err_ := parser.Init()
		if _, ok := err_.(error); ok {
			log.Error("%v", utils.ErrInfo(err_.(error)))
			return nil, utils.ErrInfo(err_.(error))
		}

		log.Debug("MethodName", MethodName+"Front")
		err_ = parser.Validate()
		if _, ok := err_.(error); ok {
			log.Error("%v", utils.ErrInfo(err_.(error)))
			return nil, utils.ErrInfo(err_.(error))
		}
	}
	return header, nil
}
