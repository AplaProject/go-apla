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
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"fmt"
)

const (
	// NumberNodes is the number of nodes
	NumberNodes = `number_of_dlt_nodes`
	// FuelRate is the rate
	FuelRate = `fuel_rate`
	// FullNodes is the list of nodes
	FullNodes = `full_nodes`
	// OpPrice is the costs of operations
	OpPrice = `op_price`
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
	// SysCurrencies is the list of system currencies
	SysCurrencies = `sys_currencies`
	// UpdFullNodesPeriod is the maximum number of user's transactions in one block
	UpdFullNodesPeriod = `upd_full_nodes_period`
	// RecoveryAddress is the recovery address
	RecoveryAddress = `recovery_address`
	// CommissionWallet is the address for commissions
	CommissionWallet = `commission_wallet`
)

type FullNode struct {
	Host   string
	Public []byte
}

var (
	cache = map[string]string{
		BlockchainURL: "https://raw.githubusercontent.com/egaas-blockchain/egaas-blockchain.github.io/master/testnet_blockchain",
	}
	cost    = make(map[string]int64)
	nodes   = make(map[int64]*FullNode)
	nodesByPosition   = make([][]string, 0)
	fuels   = make(map[int64]string)
	wallets = make(map[int64]string)
	mutex   = &sync.RWMutex{}
)

// SysUpdate reloads/updates values of system parameters
func SysUpdate() error {
	var err error
	systemParameters, err := model.GetAllSystemParametersV2()
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, param := range systemParameters {
		cache[param.Name] = param.Value
	}

	cost = make(map[string]int64)
	json.Unmarshal([]byte(cache[OpPrice]), &cost)

	nodes = make(map[int64]*FullNode)
	nodesByPosition = make([][]string, 0)
	if len(cache[FullNodes]) > 0 {
		inodes := make([][]string, 0)
		err = json.Unmarshal([]byte(cache[FullNodes]), &inodes)
		if err != nil {
			return err
		}
		nodesByPosition = inodes
		for _, item := range inodes {
			if len(item) < 3 {
				continue
			}
			pub, err := hex.DecodeString(item[2])
			if err != nil {
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

func GetNode(wallet int64) *FullNode {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := nodes[wallet]; ok {
		return ret
	}
	return nil
}

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

func GetNumberOfNodes() (int64) {
	return int64(len(nodesByPosition))
}

func GetNodeByPosition(position int64) (*FullNode, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodes))<=position {
		return nil, fmt.Errorf("incorrect position")
	}
	return nodes[position], nil
}

func GetNodePublicKeyByPosition(position int64) ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodes))<=position {
		return nil, fmt.Errorf("incorrect position")
	}
	pkey, err := hex.DecodeString(nodesByPosition[position][2])
	if err != nil {
		return nil, err
	}
	return pkey, nil
}
func GetSleepTimeByKey(myKeyID, prevBlockNodePosition int64) (int64, error) {

	myPosition, err := GetNodePositionByKeyID(myKeyID)
	if err!=nil {
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

func SysInt64(name string) int64 {
	return converter.StrToInt64(SysString(name))
}

func GetSizeFuel() int64 {
	return SysInt64(SizeFuel)
}

func GetBlockchainURL() string {
	return SysString(BlockchainURL)
}

func GetFuelRate(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := fuels[ecosystem]; ok {
		return ret
	}
	return ``
}

func GetCommissionWallet(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := wallets[ecosystem]; ok {
		return ret
	}
	return wallets[1]
}

func GetUpdFullNodesPeriod() int64 {
	return converter.StrToInt64(SysString(UpdFullNodesPeriod))
}

func GetMaxBlockSize() int64 {
	return converter.StrToInt64(SysString(MaxBlockSize))
}

func GetMaxTxSize() int64 {
	return converter.StrToInt64(SysString(MaxTxSize))
}

func GetRecoveryAddress() int64 {
	return converter.StrToInt64(SysString(RecoveryAddress))
}

func GetGapsBetweenBlocks() int64 {
	return converter.StrToInt64(SysString(GapsBetweenBlocks))
}

func GetMaxTxCount() int {
	return converter.StrToInt(SysString(MaxTxCount))
}

func GetMaxColumns() int {
	return converter.StrToInt(SysString(MaxColumns))
}

func GetMaxIndexes() int {
	return converter.StrToInt(SysString(MaxIndexes))
}

func GetMaxBlockUserTx() int {
	return converter.StrToInt(SysString(MaxBlockUserTx))
}

func GetHosts() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	ret := make([]string, 0)
	for _, item := range nodes {
		ret = append(ret, item.Host)
	}
	return ret
}

// SysCost returns the cost of the transaction
func SysCost(name string) int64 {
	return cost[name]
}

// SysString returns string value of the system parameter
func SysString(name string) string {
	mutex.RLock()
	ret := cache[name]
	mutex.RUnlock()
	return ret
}
