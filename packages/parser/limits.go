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

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"

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
	initLimit(*Block)
	checkLimit(*Parser, int) error
}

type initLimiter struct {
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
	allLimiters := []initLimiter{
		{&txMaxSize{}, letPreprocess | letParsing},
		{&txUserLimit{}, letPreprocess | letParsing},
		{&txMaxLimit{}, letPreprocess | letParsing},
		{&txUserEcosysLimit{}, letPreprocess | letParsing},
		{&timeBlockLimit{}, letGenBlock},
	}
	for _, limiter := range allLimiters {
		if limiter.modes&limits.Mode == 0 {
			continue
		}
		limiter.limiter.initLimit(b)
		limits.Limiters = append(limits.Limiters, limiter.limiter)
	}
	return
}

// CheckLimits calls each limiter
func (limits *Limits) CheckLimit(p *Parser) error {
	for _, limiter := range limits.Limiters {
		if err := limiter.checkLimit(p, limits.Mode); err != nil {
			return err
		}
	}
	return nil
}

func limitError(msg, limitName string, args ...interface{}) error {
	err := fmt.Errorf(msg, args)
	log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error(limitName)
	return script.SetVMError(`panic`, err)
}

// Checking the max tx in the block
type txMaxLimit struct {
	Count int // the current count
	Limit int // max count of tx in the block
}

func (bl *txMaxLimit) initLimit(b *Block) {
	bl.Limit = syspar.GetMaxTxCount()
}

func (bl *txMaxLimit) checkLimit(p *Parser, mode int) error {
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
	Start int64 // the time of the start of generating block
	Limit int64 // the maximum time
}

func (bl *timeBlockLimit) initLimit(b *Block) {
	bl.Start = time.Now().Unix()
	bl.Limit = syspar.GetMaxBlockGenerationTime()
}

func (bl *timeBlockLimit) checkLimit(p *Parser, mode int) error {
	if time.Now().Unix() > bl.Start+bl.Limit {
		return ErrLimitStop
	}
	return nil
}

// Checking the max tx from one user in the block
type txUserLimit struct {
	TxUsers map[int64]int // the counter of tx from one user
	Limit   int           // the value of max tx from one user
}

func (bl *txUserLimit) initLimit(b *Block) {
	bl.TxUsers = make(map[int64]int)
	bl.Limit = syspar.GetMaxBlockUserTx()
}

func (bl *txUserLimit) checkLimit(p *Parser, mode int) error {
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

func (bl *txUserEcosysLimit) initLimit(b *Block) {
	bl.TxEcosys = make(map[int64]ecosysLimit)
}

func (bl *txUserEcosysLimit) checkLimit(p *Parser, mode int) error {
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
		found, err := sp.Get(nil, `max_block_user_tx`)
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

func (bl *txMaxSize) initLimit(b *Block) {
	bl.LimitBlock = syspar.GetMaxBlockSize()
	bl.LimitTx = syspar.GetMaxTxSize()
}

func (bl *txMaxSize) checkLimit(p *Parser, mode int) error {
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
