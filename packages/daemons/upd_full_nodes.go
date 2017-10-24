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
	"errors"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// UpdFullNodes sends UpdFullNodes transactions
func UpdFullNodes(d *daemon, ctx context.Context) error {
	d.sleepTime = 60 * time.Second

	DBLock()
	defer DBUnlock()

	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		return err
	}

	if infoBlock.BlockID == 0 {
		d.sleepTime = 10 * time.Second
		return nil
	}

	nodeConfig := &model.Config{}
	found, err := nodeConfig.Get()
	if err != nil {
		return err
	}

	if !found {
		return errors.New("can't find config")
	}

	myStateID := nodeConfig.StateID
	myWalletID := nodeConfig.DltWalletID

	// If we are in the list of those who are able to generate the blocks
	fullNode := &model.FullNode{}
	found, err = fullNode.FindNode(myStateID, myWalletID, myStateID, myWalletID)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("can't find full node with stateID: %d, walletID: %d", myStateID, myWalletID)
	}

	fullNodeID := fullNode.ID
	if fullNodeID == 0 {
		d.sleepTime = 10 * time.Second // because 1s is too small for non-full nodes
		return nil
	}

	curTime := time.Now().Unix()

	// check if the time of the last updating passed
	updFn := &model.UpdFullNode{}
	found, err = updFn.Read(nil)
	if err != nil {
		return err
	}

	if !found {
		return errors.New("can't find update_full_nodes")
	}

	updFullNodes := int64(updFn.Time)
	if curTime-updFullNodes <= syspar.GetUpdFullNodesPeriod() {
		log.Debugf("curTime-adminTime <= consts.UPD_FULL_NODES_PERIO")
		return nil
	}

	myNodeKey := &model.MyNodeKey{}
	err = myNodeKey.GetNodeWithMaxBlockID()
	if err != nil {
		return err
	}
	var (
		hash, data []byte
	)

	contract := smart.GetContract(`@0UpdFullNodes`, 0)
	if contract == nil {
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
		return err
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: smartTx.Header.Time,
			UserID: myWalletID, BinSignatures: converter.EncodeLengthPlusData(signature)},
		Data: make([]byte, 0),
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		return err
	}
	data = append([]byte{128}, serializedData...)
	if hash, err = model.SendTx(int64(info.ID), myWalletID, data); err != nil {
		return err
	}
	p := new(parser.Parser)
	err = p.TxParser(hash, data, true)
	if err != nil {
		return err
	}

	return nil
}
