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

package syspar

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"time"

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
	// MaxBlockFuel is the maximum fuel of the block
	MaxBlockFuel = `max_fuel_block`
	// MaxTxFuel is the maximum fuel of the transaction
	MaxTxFuel = `max_fuel_tx`
	// MaxTxCount is the maximum count of the transactions
	MaxTxCount = `max_tx_count`
	// MaxBlockGenerationTime is the time limit for block generation (in ms)
	MaxBlockGenerationTime = `max_block_generation_time`
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
	// BlockReward value of reward, which is chrged on block generation
	BlockReward = "block_reward"
	// IncorrectBlocksPerDay is value of incorrect blocks per day before global ban
	IncorrectBlocksPerDay = `incorrect_blocks_per_day`
	// NodeBanTime is value of ban time for bad nodes (in ms)
	NodeBanTime = `node_ban_time`
)

var (
	cache = map[string]string{
		BlockchainURL: "https://raw.githubusercontent.com/egaas-blockchain/egaas-blockchain.github.io/master/testnet_blockchain",
	}
	nodes           = make(map[int64]*FullNode)
	nodesByPosition = make([]*FullNode, 0)
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
	if len(cache[FullNodes]) > 0 {
		if err = updateNodes(); err != nil {
			return err
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

func updateNodes() (err error) {
	nodes = make(map[int64]*FullNode)
	nodesByPosition = make([]*FullNode, 0)

	items := make([]*FullNode, 0)
	if len(cache[FullNodes]) > 0 {
		err = json.Unmarshal([]byte(cache[FullNodes]), &items)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "v": cache[FullNodes]}).Error("unmarshalling full nodes from json")
			return err
		}
	}

	nodesByPosition = items
	for _, item := range items {
		nodes[item.KeyID] = item
	}

	return nil
}

// AddFullNodeKeys adds node by keys to list of nodes
func AddFullNodeKeys(keyID int64, publicKey []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	nodesByPosition = append(nodesByPosition, &FullNode{
		KeyID:     keyID,
		PublicKey: publicKey,
	})
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

func GetNodes() map[int64]*FullNode {
	return nodes
}

// GetNodePositionByKeyID is returning node position by key id
func GetNodePositionByKeyID(keyID int64) (int64, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	for i, item := range nodesByPosition {
		if item.KeyID == keyID {
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
	return nodesByPosition[position], nil
}

func GetNodeByHost(host string) (*FullNode, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	for k, n := range nodes {
		if n.TCPAddress == host {
			return nodes[k], nil
		}
	}

	return nil, fmt.Errorf("incorrect host")
}

// GetNodeHostByPosition is retrieving node host by position
func GetNodeHostByPosition(position int64) (string, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	nodeData, err := GetNodeByPosition(position)
	if err != nil {
		return "", err
	}
	return nodeData.TCPAddress, nil
}

// GetNodePublicKeyByPosition is retrieving node public key by position
func GetNodePublicKeyByPosition(position int64) ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodesByPosition)) <= position {
		return nil, fmt.Errorf("incorrect position")
	}
	nodeData, err := GetNodeByPosition(position)
	if err != nil {
		return nil, err
	}
	return nodeData.PublicKey, nil
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

// GetMaxBlockFuel is returns max block fuel
func GetMaxBlockFuel() int64 {
	return converter.StrToInt64(SysString(MaxBlockFuel))
}

// GetMaxTxFuel is returns max tx fuel
func GetMaxTxFuel() int64 {
	return converter.StrToInt64(SysString(MaxTxFuel))
}

// GetMaxBlockGenerationTime is returns max block generation time (in ms)
func GetMaxBlockGenerationTime() int64 {
	return converter.StrToInt64(SysString(MaxBlockGenerationTime))
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

func GetIncorrectBlocksPerDay() int {
	return converter.StrToInt(SysString(IncorrectBlocksPerDay))
}

func GetNodeBanTime() time.Duration {
	return time.Millisecond * time.Duration(converter.StrToInt64(SysString(NodeBanTime)))
}

// GetRemoteHosts returns array of hostnames excluding myself
func GetRemoteHosts() []string {
	ret := make([]string, 0)

	mutex.RLock()
	defer mutex.RUnlock()

	for keyID, item := range nodes {
		if keyID != conf.Config.KeyID {
			ret = append(ret, item.TCPAddress)
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
