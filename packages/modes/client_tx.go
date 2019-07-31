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

package modes

import (
	"bytes"
	"errors"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils/tx"
	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

var ErrDiffKey = errors.New("Different keys")

type blockchainTxPreprocessor struct{}

func (p blockchainTxPreprocessor) ProcessClientTranstaction(txData []byte, key int64, le *log.Entry) (string, error) {
	rtx := &transaction.RawTransaction{}
	if err := rtx.Unmarshall(bytes.NewBuffer(txData)); err != nil {
		le.WithFields(log.Fields{"error": err}).Error("on unmarshalling to raw tx")
		return "", err
	}

	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(rtx.Payload(), &smartTx); err != nil {
		le.WithFields(log.Fields{"error": err}).Error("on unmarshalling to sc")
		return "", err
	}

	if smartTx.Header.KeyID != key {
		return "", ErrDiffKey
	}

	if err := model.SendTx(rtx, key); err != nil {
		le.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("sending tx")
		return "", err
	}

	return string(converter.BinToHex(rtx.Hash())), nil
}

type ObsTxPreprocessor struct{}

func (p ObsTxPreprocessor) ProcessClientTranstaction(txData []byte, key int64, le *log.Entry) (string, error) {

	tx, err := transaction.UnmarshallTransaction(bytes.NewBuffer(txData), true)
	if err != nil {
		le.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("on unmarshaling user tx")
		return "", err
	}

	ts := &model.TransactionStatus{
		BlockID:  1,
		Hash:     tx.TxHash,
		Time:     time.Now().Unix(),
		WalletID: key,
		Type:     tx.TxType,
	}

	if err := ts.Create(); err != nil {
		le.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on creating tx status")
		return "", err
	}

	res, _, err := tx.CallOBSContract()
	if err != nil {
		le.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("on execution contract")
		return "", err
	}

	if err := ts.UpdateBlockMsg(nil, 1, res, tx.TxHash); err != nil {
		le.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": tx.TxHash}).Error("updating transaction status block id")
		return "", err
	}

	return string(converter.BinToHex(tx.TxHash)), nil
}

func GetClientTxPreprocessor() types.ClientTxPreprocessor {
	if conf.Config.IsSupportingOBS() {
		return ObsTxPreprocessor{}
	}

	return blockchainTxPreprocessor{}
}

// BlockchainSCRunner implementls SmartContractRunner for blockchain mode
type BlockchainSCRunner struct{}

// RunContract runs smart contract on blockchain mode
func (runner BlockchainSCRunner) RunContract(data, hash []byte, keyID int64, le *log.Entry) error {
	if err := tx.CreateTransaction(data, hash, keyID); err != nil {
		le.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		return err
	}

	return nil
}

// OBSSCRunner implementls SmartContractRunner for obs mode
type OBSSCRunner struct{}

// RunContract runs smart contract on obs mode
func (runner OBSSCRunner) RunContract(data, hash []byte, keyID int64, le *log.Entry) error {
	proc := GetClientTxPreprocessor()
	_, err := proc.ProcessClientTranstaction(data, keyID, le)
	if err != nil {
		le.WithFields(log.Fields{"error": consts.ContractError}).Error("on run internal NewUser")
		return err
	}

	return nil
}

// GetSmartContractRunner returns mode boundede implementation of SmartContractRunner
func GetSmartContractRunner() types.SmartContractRunner {
	if !conf.Config.IsSupportingOBS() {
		return BlockchainSCRunner{}
	}

	return OBSSCRunner{}
}
