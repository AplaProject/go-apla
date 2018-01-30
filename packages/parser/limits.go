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
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
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
	Block    *Block    // it equals nil if checking before generatin block
	Limiters []Limiter // the list of limiters
}

// Limiter describes interface functions for limits
type Limiter interface {
	initLimit(*Block) error
	preLimit(*Parser) error
	postLimit(*Parser) error
}

type initLimiter struct {
	limiter Limiter
	modes   int // combination of letPreprocess letGenBlock letParsing
}

var (
	allLimiters = []initLimiter{
		{&txUserLimit{}, letPreprocess | letParsing},
		{&txMaxLimit{}, letPreprocess | letParsing},
		{&txUserEcosysLimit{}, letPreprocess | letParsing},
		{&timeBlockLimit{}, letGenBlock},
	}
)

// newLimits initializes Limits structure.
func newLimits(b *Block) (limits *Limits, err error) {
	var mode int
	limits = &Limits{Block: b, Limiters: make([]Limiter, 0, 8)}
	if b == nil {
		mode = letPreprocess
	} else if b.GenBlock {
		mode = letGenBlock
	} else {
		mode = letParsing
	}
	for _, limiter := range allLimiters {
		if limiter.modes&mode == 0 {
			continue
		}
		limits.Limiters = append(limits.Limiters, limiter.limiter)
		if err = limiter.initLimit(b); err != nil {
			return nil, err
		}
	}
	return
}

func (limits *Limits) preProcess(p *Parser) error {
	for _, limiter := range limits.Limiters {
		if err := limiter.preLimit(p); err != nil {
			return err
		}
	}
	return nil
}

func (limits *Limits) postProcess(p *Parser) error {
	for _, limiter := range limits.Limiters {
		if err := limiter.postLimit(p); err != nil {
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

// Checking the time of the start of generating block
type timeBlockLimit struct {
	Block *Block
	Start int64 // the time of the start of generating block
	Limit int64 // the maximum time
}

func (bl *timeBlockLimit) initLimit(b *Block) error {
	bl.Block = b
	bl.Start = time.Now().Unix()
	bl.Limit = 1000 // It should be taken from time limit system parameter
	return nil
}

func (bl *timeBlockLimit) preLimit(p *Parser) error {
	return nil
}

func (bl *timeBlockLimit) postLimit(p *Parser) error {
	if time.Now().Unix() > bl.Start+bl.Limit {
		bl.Block.StopBlock = true
		return limitError(`timeBlockLimit`, `Time limitation of generating block`)
	}
	return nil
}

// Checking the max tx from one user in the block
type txUserLimit struct {
	TxUsers map[int64]int // the counter of tx from one user
	Limit   int           // the value of max tx from one user
}

func (bl *txUserLimit) initLimit(b *Block) error {
	bl.TxUsers = make(map[int64]int)
	bl.Limit = converter.StrToInt(syspar.SysString(syspar.MaxBlockUserTx))
	return nil
}

func (bl *txUserLimit) preLimit(p *Parser) error {
	var (
		count int
		ok    bool
	)
	keyID := p.TxSmart.KeyID
	if count, ok = bl.TxUsers[keyID]; ok {
		if count+1 > bl.Limit {
			return limitError(`txUserLimit`, `Max tx from one user %d`, keyID)
		}
	}
	bl.TxUsers[keyID] = count + 1
	return nil
}

func (bl *txUserLimit) postLimit(p *Parser) error {
	return nil
}

// Checking the max tx in the block
type txMaxLimit struct {
	Count int // the current count
	Limit int // max count of tx in the block
}

func (bl *txMaxLimit) initLimit(b *Block) error {
	bl.Limit = syspar.GetMaxTxCount()
	return nil
}

func (bl *txMaxLimit) preLimit(p *Parser) error {
	bl.Count++
	if bl.Count > bl.Limit {
		return limitError(`txMaxLimit`, `Max tx in the block`)
	}
	return nil
}

func (bl *txMaxLimit) postLimit(p *Parser) error {
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

func (bl *txUserEcosysLimit) initLimit(b *Block) error {
	bl.TxEcosys = make(map[int64]ecosysLimit)
	return nil
}

func (bl *txUserEcosysLimit) preLimit(p *Parser) error {
	keyID := p.TxSmart.KeyID
	ecosystemID := p.TxSmart.EcosystemID
	if val, ok := bl.TxEcosys[ecosystemID]; ok {
		if user, ok := val.TxUsers[keyID]; ok {
			if user+1 > val.Limit {
				return limitError(`txUserEcosysLimit`, `Max tx from one user %d in ecosystem %d`,
					keyID, ecosystemID)
			}
			val.TxUsers[keyID] = user + 1
		} else {
			val.TxUsers[keyID] = 1
		}
	} else {
		limit := 20 // This limit should be taken from ecosys params table of ecosystemID
		bl.TxEcosys[ecosystemID] = ecosysLimit{TxUsers: make(map[int64]int), Limit: limit}
		bl.TxEcosys[ecosystemID].TxUsers[keyID] = 1
	}
	return nil
}

func (bl *txUserEcosysLimit) postLimit(p *Parser) error {
	return nil
}
