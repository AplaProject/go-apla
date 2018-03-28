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
	"fmt"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"

	log "github.com/sirupsen/logrus"
)

const (
	letPreprocess = 0x0001 // checking before generating block
	letGenBlock   = 0x0002 // checking during generating block
	letParsing    = 0x0004 // common checking during parsing block
)

// Limits is used for saving current limit information
type Limits struct {
	Mode     int
	Block    *Block    // it equals nil if checking before generatin block
	Limiters []Limiter // the list of limiters
}

// Limiter describes interface functions for limits
type Limiter interface {
	init(*Block)
	check(*Parser, int) error
}

type limiterModes struct {
	limiter Limiter
	modes   int // combination of letPreprocess letGenBlock letParsing
}

var (
	// ErrLimitSkip returns when tx should be skipped during generating block
	ErrLimitSkip = errors.New(`skip tx`)
	// ErrLimitStop returns when the generation of the block should be stopped
	ErrLimitStop = errors.New(`stop generating block`)
)

// NewLimits initializes Limits structure.
func NewLimits(b *Block) (limits *Limits) {
	limits = &Limits{Block: b, Limiters: make([]Limiter, 0, 8)}
	if b == nil {
		limits.Mode = letPreprocess
	} else if b.GenBlock {
		limits.Mode = letGenBlock
	} else {
		limits.Mode = letParsing
	}
	allLimiters := []limiterModes{
		{&txMaxSize{}, letPreprocess | letParsing},
		{&txUserLimit{}, letPreprocess | letParsing},
		{&txMaxLimit{}, letPreprocess | letParsing},
		{&txUserEcosysLimit{}, letPreprocess | letParsing},
		{&timeBlockLimit{}, letGenBlock},
		{&txMaxFuel{}, letGenBlock | letParsing},
	}
	for _, limiter := range allLimiters {
		if limiter.modes&limits.Mode == 0 {
			continue
		}
		limiter.limiter.init(b)
		limits.Limiters = append(limits.Limiters, limiter.limiter)
	}
	return
}

// CheckLimits calls each limiter
func (limits *Limits) CheckLimit(p *Parser) error {
	for _, limiter := range limits.Limiters {
		if err := limiter.check(p, limits.Mode); err != nil {
			return err
		}
	}
	return nil
}

func limitError(limitName, msg string, args ...interface{}) error {
	err := fmt.Errorf(msg, args...)
	log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error(limitName)
	return script.SetVMError(`panic`, err)
}

// Checking the max tx in the block
type txMaxLimit struct {
	Count int // the current count
	Limit int // max count of tx in the block
}

func (bl *txMaxLimit) init(b *Block) {
	bl.Limit = syspar.GetMaxTxCount()
}

func (bl *txMaxLimit) check(p *Parser, mode int) error {
	bl.Count++
	if bl.Count > bl.Limit {
		if mode == letPreprocess {
			return ErrLimitStop
		}
		return limitError(`txMaxLimit`, `Max tx in the block`)
	}
	return nil
}

// Checking the time of the start of generating block
type timeBlockLimit struct {
	Start time.Time     // the time of the start of generating block
	Limit time.Duration // the maximum time
}

func (bl *timeBlockLimit) init(b *Block) {
	bl.Start = time.Now()
	bl.Limit = time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
}

func (bl *timeBlockLimit) check(p *Parser, mode int) error {
	if time.Since(bl.Start) > bl.Limit {
		return ErrLimitStop
	}
	return nil
}

// Checking the max tx from one user in the block
type txUserLimit struct {
	TxUsers map[int64]int // the counter of tx from one user
	Limit   int           // the value of max tx from one user
}

func (bl *txUserLimit) init(b *Block) {
	bl.TxUsers = make(map[int64]int)
	bl.Limit = syspar.GetMaxBlockUserTx()
}

func (bl *txUserLimit) check(p *Parser, mode int) error {
	var (
		count int
		ok    bool
	)
	keyID := p.TxSmart.KeyID
	if count, ok = bl.TxUsers[keyID]; ok {
		if count+1 > bl.Limit {
			if mode == letPreprocess {
				return ErrLimitSkip
			}
			return limitError(`txUserLimit`, `Max tx from one user %d`, keyID)
		}
	}
	bl.TxUsers[keyID] = count + 1
	return nil
}

// Checking the max tx from one user in the ecosystem contracts
type ecosysLimit struct {
	TxUsers map[int64]int // the counter of tx from one user in the ecosystem
	Limit   int           // the value of max tx from one user in the ecosystem
}

type txUserEcosysLimit struct {
	TxEcosys map[int64]ecosysLimit // the counter of tx from one user in ecosystems
}

func (bl *txUserEcosysLimit) init(b *Block) {
	bl.TxEcosys = make(map[int64]ecosysLimit)
}

func (bl *txUserEcosysLimit) check(p *Parser, mode int) error {
	keyID := p.TxSmart.KeyID
	ecosystemID := p.TxSmart.EcosystemID
	if val, ok := bl.TxEcosys[ecosystemID]; ok {
		if user, ok := val.TxUsers[keyID]; ok {
			if user+1 > val.Limit {
				if mode == letPreprocess {
					return ErrLimitSkip
				}
				return limitError(`txUserEcosysLimit`, `Max tx from one user %d in ecosystem %d`,
					keyID, ecosystemID)
			}
			val.TxUsers[keyID] = user + 1
		} else {
			val.TxUsers[keyID] = 1
		}
	} else {
		limit := syspar.GetMaxBlockUserTx()
		sp := &model.StateParameter{}
		sp.SetTablePrefix(converter.Int64ToStr(ecosystemID))
		found, err := sp.Get(p.DbTransaction, `max_block_user_tx`)
		if err != nil {
			return limitError(`txUserEcosysLimit`, err.Error())
		}
		if found {
			limit = converter.StrToInt(sp.Value)
		}
		bl.TxEcosys[ecosystemID] = ecosysLimit{TxUsers: make(map[int64]int), Limit: limit}
		bl.TxEcosys[ecosystemID].TxUsers[keyID] = 1
	}
	return nil
}

// Checking the max tx & block size
type txMaxSize struct {
	Size       int64 // the current size of the block
	LimitBlock int64 // max size of the block
	LimitTx    int64 // max size of tx
}

func (bl *txMaxSize) init(b *Block) {
	bl.LimitBlock = syspar.GetMaxBlockSize()
	bl.LimitTx = syspar.GetMaxTxSize()
}

func (bl *txMaxSize) check(p *Parser, mode int) error {
	size := int64(len(p.TxFullData))
	if size > bl.LimitTx {
		return limitError(`txMaxSize`, `Max size of tx`)
	}
	bl.Size += size
	if bl.Size > bl.LimitBlock {
		if mode == letPreprocess {
			return ErrLimitStop
		}
		return limitError(`txMaxSize`, `Max size of the block`)
	}
	return nil
}

// Checking the max tx & block size
type txMaxFuel struct {
	Fuel       int64 // the current fuel of the block
	LimitBlock int64 // max fuel of the block
	LimitTx    int64 // max fuel of tx
}

func (bl *txMaxFuel) init(b *Block) {
	bl.LimitBlock = syspar.GetMaxBlockFuel()
	bl.LimitTx = syspar.GetMaxTxFuel()
}

func (bl *txMaxFuel) check(p *Parser, mode int) error {
	fuel := p.TxFuel
	if fuel > bl.LimitTx {
		return limitError(`txMaxFuel`, `Max fuel of tx %d > %d`, fuel, bl.LimitTx)
	}
	bl.Fuel += fuel
	if bl.Fuel > bl.LimitBlock {
		if mode == letGenBlock {
			return ErrLimitStop
		}
		return limitError(`txMaxFuel`, `Max fuel of the block`)
	}
	return nil
}
