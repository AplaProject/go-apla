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

package daemons

import (
	"context"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// UpdFullNodes sends UpdFullNodes transactions
func UpdFullNodes(d *daemon, ctx context.Context) error {
	d.sleepTime = 60 * time.Second

	DBLock()
	defer DBUnlock()

	infoBlock := &model.InfoBlock{}
	err := infoBlock.GetInfoBlock()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		return err
	}

	if infoBlock.BlockID == 0 {
		d.logger.Warning("info block not found, sleeping 10 seconds")
		d.sleepTime = 10 * time.Second
		return nil
	}

	nodeConfig := &model.Config{}
	err = nodeConfig.GetConfig()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting config")
		return err

	}
	myStateID := nodeConfig.StateID
	myWalletID := nodeConfig.DltWalletID
	// If we are in the list of those who are able to generate the blocks
	fullNode := &model.FullNode{}
	err = fullNode.FindNode(myStateID, myWalletID, myStateID, myWalletID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("finding full node")
		return err
	}

	fullNodeID := fullNode.ID
	if fullNodeID == 0 {
		d.logger.Warning("full node not found, sleeping 10 seconds")
		d.sleepTime = 10 * time.Second // because 1s is too small for non-full nodes
		return nil
	}

	curTime := time.Now().Unix()

	// check if the time of the last updating passed
	updFn := &model.UpdFullNode{}
	err = updFn.Read(nil)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("reading upd full node")
		return err
	}

	updFullNodes := int64(updFn.Time)
	if curTime-updFullNodes <= syspar.GetUpdFullNodesPeriod() {
		d.logger.Debug("upd full nodes period is not expired")
		return nil
	}

	myNodeKey := &model.MyNodeKey{}
	err = myNodeKey.GetNodeWithMaxBlockID()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node with max block id")
		return err
	}
	var (
		hash, data []byte
	)

	contract := smart.GetContract(`@0UpdFullNodes`, 0)
	if contract == nil {
		d.logger.WithFields(log.Fields{"contract_name": "@0UpdFullNodes"}).Error("Getting contract")
		return fmt.Errorf(`there is not @0UpdFullNodes contract`)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	var (
		smartTx     tx.SmartContract
		toSerialize interface{}
	)
	smartTx.Header = tx.Header{Type: int(info.ID), Time: time.Now().Unix(), UserID: myWalletID, StateID: 0}
	signature, err := crypto.Sign(myNodeKey.PrivateKey, smartTx.ForSign())
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing smart tx with node private key")
		return err
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: smartTx.Header.Time,
			UserID: myWalletID, BinSignatures: converter.EncodeLengthPlusData(signature)},
		Data: make([]byte, 0),
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smartContract transaction to msgpack")
		return err
	}
	data = append([]byte{128}, serializedData...)
	if hash, err = model.SendTx(int64(info.ID), myWalletID, data); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("sending tx to the queue")
		return err
	}
	p := new(parser.Parser)
	err = p.TxParser(hash, data, true)
	if err != nil {
		d.logger.WithFields(log.Fields{"error": err}).Error("parsing transaction")
		return err
	}

	return nil
}
