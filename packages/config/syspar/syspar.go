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
	"strconv"
	"sync"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	log "github.com/sirupsen/logrus"
)

const (
	// NumberNodes is the number of nodes
	NumberNodes = `number_of_dlt_nodes`
	// FuelRate is the rate
	FuelRate = `fuel_rate`
	// OpPrice is the costs of operations
	OpPrice = `op_price`
	// GapsBetweenBlocks is the time between blocks
	GapsBetweenBlocks = `gaps_between_blocks`
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
	// UpdFullNodesPeriod is the maximum number of user's transactions in one block
	UpdFullNodesPeriod = `upd_full_nodes_period`
	// RecoveryAddress is the recovery address
	RecoveryAddress = `recovery_address`
	// CommissionWallet is the address for commissions
	CommissionWallet = `commission_wallet`
)

var (
	cache = map[string]string{
		BlockchainURL: "https://raw.githubusercontent.com/egaas-blockchain/egaas-blockchain.github.io/master/testnet_blockchain",
		// For compatible of develop versions
		// Remove later
		GapsBetweenBlocks:  `3`,
		MaxBlockSize:       `67108864`,
		MaxTxSize:          `33554432`,
		MaxTxCount:         `100000`,
		MaxColumns:         `50`,
		MaxIndexes:         `10`,
		MaxBlockUserTx:     `100`,
		UpdFullNodesPeriod: `3600`, // 3600 is for the test time, then we have to put 86400`
		RecoveryAddress:    `8275283526439353759`,
		CommissionWallet:   `8275283526439353759`,
	}
	cost  = make(map[string]int64)
	mutex = &sync.RWMutex{}
)

// SysUpdate reloads/updates values of system parameters
func SysUpdate() error {
	systemParameters, err := model.GetAllSystemParameters()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all system parameters")
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, param := range systemParameters {
		cache[param.Name] = param.Value
	}

	cost = make(map[string]int64)
	if err := json.Unmarshal([]byte(cache[OpPrice]), &cost); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "value": cache[OpPrice]}).Error("unmarshalling opPrice from json")
	}
	return err
}

func SysInt64(name string) int64 {
	strVal := SysString(name)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetBlockchainURL() string {
	return SysString(BlockchainURL)
}

func GetUpdFullNodesPeriod() int64 {
	strVal := SysString(UpdFullNodesPeriod)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxBlockSize() int64 {
	strVal := SysString(MaxBlockSize)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxTxSize() int64 {
	strVal := SysString(MaxTxSize)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetRecoveryAddress() int64 {
	strVal := SysString(RecoveryAddress)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetCommissionWallet() int64 {
	strVal := SysString(CommissionWallet)
	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetGapsBetweenBlocks() int {
	strVal := SysString(GapsBetweenBlocks)
	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxTxCount() int {
	strVal := SysString(MaxTxCount)
	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxColumns() int {
	strVal := SysString(MaxColumns)
	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxIndexes() int {
	strVal := SysString(MaxIndexes)
	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
}

func GetMaxBlockUserTx() int {
	strVal := SysString(MaxBlockUserTx)
	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strVal}).Error("converting str to int")
	}
	return val
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
