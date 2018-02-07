// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package syspar

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

const (
	// NumberNodes is the number of nodes
	NumberNodes = `number_of_nodes`
	// FuelRate is the rate
	FuelRate = `fuel_rate`
	// FullNodes is the list of nodes
	FullNodes = `full_nodes`
	// GapsBetweenBlocks is the time between blocks
	GapsBetweenBlocks = `gap_between_blocks`
	// BlockchainURL is the address of the blockchain file.  For those who don't want to collect it from nodes
	BlockchainURL = `blockchain_url`
	// MaxBlockSize is the maximum size of the block
	MaxBlockSize = `max_block_size`
	// MaxTxSize is the maximum size of the transaction
	MaxTxSize = `max_tx_size`
	// MaxTxCount is the maximum count of the transactions
	MaxTxCount = `max_tx_count`
	// MaxColumns is the maximum columns in tables
	MaxColumns = `max_columns`
	// MaxIndexes is the maximum indexes in tables
	MaxIndexes = `max_indexes`
	// MaxBlockUserTx is the maximum number of user's transactions in one block
	MaxBlockUserTx = `max_block_user_tx`
	// SizeFuel is the fuel cost of 1024 bytes of the transaction data
	SizeFuel = `size_fuel`
	// CommissionWallet is the address for commissions
	CommissionWallet = `commission_wallet`
	// RbBlocks1 rollback from queue_bocks
	RbBlocks1 = `rb_blocks_1`
)

// FullNode is storing full node data
type FullNode struct {
	Host   string
	Public []byte
}

var (
	cache = map[string]string{
		BlockchainURL: "https://raw.githubusercontent.com/egaas-blockchain/egaas-blockchain.github.io/master/testnet_blockchain",
	}
	nodes           = make(map[int64]*FullNode)
	nodesByPosition = make([][]string, 0)
	fuels           = make(map[int64]string)
	wallets         = make(map[int64]string)
	mutex           = &sync.RWMutex{}
)

// SysUpdate reloads/updates values of system parameters
func SysUpdate(dbTransaction *model.DbTransaction) error {
	var err error
	systemParameters, err := model.GetAllSystemParameters(dbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all system parameters")
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, param := range systemParameters {
		cache[param.Name] = param.Value
	}

	nodes = make(map[int64]*FullNode)
	nodesByPosition = make([][]string, 0)
	if len(cache[FullNodes]) > 0 {
		inodes := make([][]string, 0)
		err = json.Unmarshal([]byte(cache[FullNodes]), &inodes)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling full nodes from json")
			return err
		}
		nodesByPosition = inodes
		for _, item := range inodes {
			if len(item) < 3 {
				continue
			}
			pub, err := hex.DecodeString(item[2])
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": item[2]}).Error("decoding inode from string")
				return err
			}
			nodes[converter.StrToInt64(item[1])] = &FullNode{Host: item[0], Public: pub}
		}
	}
	getParams := func(name string) (map[int64]string, error) {
		res := make(map[int64]string)
		if len(cache[name]) > 0 {
			ifuels := make([][]string, 0)
			err = json.Unmarshal([]byte(cache[name]), &ifuels)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling params from json")
				return res, err
			}
			for _, item := range ifuels {
				if len(item) < 2 {
					continue
				}
				res[converter.StrToInt64(item[0])] = item[1]
			}
		}
		return res, nil
	}
	fuels, err = getParams(FuelRate)
	wallets, err = getParams(CommissionWallet)

	return err
}

// GetNode is retrieving node by wallet
func GetNode(wallet int64) *FullNode {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := nodes[wallet]; ok {
		return ret
	}
	return nil
}

// GetNodePositionByKeyID is returning node position by key id
func GetNodePositionByKeyID(keyID int64) (int64, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	for i, item := range nodesByPosition {
		if len(item) < 3 {
			continue
		}
		if converter.StrToInt64(item[1]) == keyID {
			return int64(i), nil
		}
	}
	return 0, fmt.Errorf("Incorrect keyID")
}

// GetNumberOfNodes is count number of nodes
func GetNumberOfNodes() int64 {
	return int64(len(nodesByPosition))
}

// GetNodeByPosition is retrieving node by position
func GetNodeByPosition(position int64) (*FullNode, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodesByPosition)) <= position {
		return nil, fmt.Errorf("incorrect position")
	}
	return nodes[converter.StrToInt64(nodesByPosition[position][1])], nil
}

// GetNodeHostByPosition is retrieving node host by position
func GetNodeHostByPosition(position int64) (string, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	nodeData, err := GetNodeByPosition(position)
	if err != nil {
		return "", err
	}
	return nodeData.Host, nil
}

// GetNodePublicKeyByPosition is retrieving node public key by position
func GetNodePublicKeyByPosition(position int64) ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodesByPosition)) <= position {
		return nil, fmt.Errorf("incorrect position")
	}
	pkey, err := hex.DecodeString(nodesByPosition[position][2])
	if err != nil {
		return nil, err
	}
	return pkey, nil
}

// GetSleepTimeByKey is returns sleep time by key
func GetSleepTimeByKey(myKeyID, prevBlockNodePosition int64) (int64, error) {

	myPosition, err := GetNodePositionByKeyID(myKeyID)
	if err != nil {
		return 0, err
	}
	sleepTime := int64(0)
	if myPosition == prevBlockNodePosition {
		sleepTime = ((GetNumberOfNodes() + myPosition) - (prevBlockNodePosition)) * GetGapsBetweenBlocks()
	}

	if myPosition > prevBlockNodePosition {
		sleepTime = (myPosition - (prevBlockNodePosition)) * GetGapsBetweenBlocks()
	}

	if myPosition < prevBlockNodePosition {
		sleepTime = (GetNumberOfNodes() - prevBlockNodePosition) * GetGapsBetweenBlocks()
	}

	return int64(sleepTime), nil
}

// GetSleepTimeByPosition is returns sleep time by position
func GetSleepTimeByPosition(CurrentPosition, prevBlockNodePosition int64) (int64, error) {

	sleepTime := int64(0)
	if CurrentPosition == prevBlockNodePosition {
		sleepTime = ((GetNumberOfNodes() + CurrentPosition) - (prevBlockNodePosition)) * GetGapsBetweenBlocks()
	}

	if CurrentPosition > prevBlockNodePosition {
		sleepTime = (CurrentPosition - (prevBlockNodePosition)) * GetGapsBetweenBlocks()
	}

	if CurrentPosition < prevBlockNodePosition {
		sleepTime = (GetNumberOfNodes() - prevBlockNodePosition) * GetGapsBetweenBlocks()
	}

	return int64(sleepTime), nil
}

// SysInt64 is converting sys string to int64
func SysInt64(name string) int64 {
	return converter.StrToInt64(SysString(name))
}

// SysInt is converting sys string to int
func SysInt(name string) int {
	return converter.StrToInt(SysString(name))
}

// GetSizeFuel is returns fuel size
func GetSizeFuel() int64 {
	return SysInt64(SizeFuel)
}

// GetBlockchainURL is retrieving blockchain url
func GetBlockchainURL() string {
	return SysString(BlockchainURL)
}

// GetFuelRate is returning fuel rate
func GetFuelRate(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := fuels[ecosystem]; ok {
		return ret
	}
	return ``
}

// GetCommissionWallet is returns commission wallet
func GetCommissionWallet(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := wallets[ecosystem]; ok {
		return ret
	}
	return wallets[1]
}

// GetMaxBlockSize is returns max block size
func GetMaxBlockSize() int64 {
	return converter.StrToInt64(SysString(MaxBlockSize))
}

// GetMaxTxSize is returns max tx size
func GetMaxTxSize() int64 {
	return converter.StrToInt64(SysString(MaxTxSize))
}

// GetGapsBetweenBlocks is returns gaps between blocks
func GetGapsBetweenBlocks() int64 {
	return converter.StrToInt64(SysString(GapsBetweenBlocks))
}

// GetMaxTxCount is returns max tx count
func GetMaxTxCount() int {
	return converter.StrToInt(SysString(MaxTxCount))
}

// GetMaxColumns is returns max columns
func GetMaxColumns() int {
	return converter.StrToInt(SysString(MaxColumns))
}

// GetMaxIndexes is returns max indexes
func GetMaxIndexes() int {
	return converter.StrToInt(SysString(MaxIndexes))
}

// GetMaxBlockUserTx is returns max tx block user
func GetMaxBlockUserTx() int {
	return converter.StrToInt(SysString(MaxBlockUserTx))
}

// GetRemoteHosts returns array of hostnames excluding myself
func GetRemoteHosts() []string {

	ret := make([]string, 0)

	mutex.RLock()
	defer mutex.RUnlock()

	for nodeID, item := range nodes {
		if nodeID != conf.Config.KeyID {
			ret = append(ret, item.Host)
		}
	}
	return ret
}

// SysString returns string value of the system parameter
func SysString(name string) string {
	mutex.RLock()
	ret := cache[name]
	mutex.RUnlock()
	return ret
}

// GetRbBlocks1 is returns RbBlocks1
func GetRbBlocks1() int64 {
	return SysInt64(RbBlocks1)
}

// HasSys returns boolean whether this system parameter exists
func HasSys(name string) bool {
	mutex.RLock()
	_, ok := cache[name]
	mutex.RUnlock()
	return ok
}
