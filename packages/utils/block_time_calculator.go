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

package utils

import (
	"time"

	"github.com/pkg/errors"
)

// BlockTimeCalculator calculating block generation time
type BlockTimeCalculator struct {
	clock         Clock
	blocksCounter intervalBlocksCounter

	firstBlockTime      time.Time
	blockGenerationTime time.Duration
	blocksGap           time.Duration

	nodesCount int64
}

type blockGenerationState struct {
	start    time.Time
	duration time.Duration

	nodePosition int64
}

var TimeError = errors.New("current time before first block")
var DuplicateBlockError = errors.New("block with that time interval already exists in db")

func NewBlockTimeCalculator(firstBlockTime time.Time, generationTime, blocksGap time.Duration, nodesCount int64) BlockTimeCalculator {
	return BlockTimeCalculator{
		clock:         &ClockWrapper{},
		blocksCounter: &blocksCounter{},

		firstBlockTime:      firstBlockTime,
		blockGenerationTime: generationTime,
		blocksGap:           blocksGap,
		nodesCount:          nodesCount,
	}
}

func (btc *BlockTimeCalculator) TimeToGenerate(nodePosition int64) (bool, error) {
	bgs, err := btc.countBlockTime(btc.clock.Now())
	if err != nil {
		return false, err
	}

	blocks, err := btc.blocksCounter.count(bgs)
	if err != nil {
		return false, err
	}

	if blocks != 0 {
		return false, DuplicateBlockError
	}

	return bgs.nodePosition == nodePosition, nil
}

func (btc *BlockTimeCalculator) ValidateBlock(nodePosition int64, at time.Time) (bool, error) {
	bgs, err := btc.countBlockTime(at)
	if err != nil {
		return false, err
	}

	blocks, err := btc.blocksCounter.count(bgs)
	if err != nil {
		return false, err
	}

	if blocks != 0 {
		return false, DuplicateBlockError
	}

	return bgs.nodePosition == nodePosition, nil
}

func (btc *BlockTimeCalculator) SetClock(clock Clock) *BlockTimeCalculator {
	btc.clock = clock
	return btc
}

func (btc *BlockTimeCalculator) setBlockCounter(counter intervalBlocksCounter) *BlockTimeCalculator {
	btc.blocksCounter = counter
	return btc
}

func (btc *BlockTimeCalculator) countBlockTime(blockTime time.Time) (blockGenerationState, error) {
	bgs := blockGenerationState{}
	nextBlockStart := btc.firstBlockTime
	var curNodeIndex int64

	if blockTime.Before(nextBlockStart) {
		return blockGenerationState{}, TimeError
	}

	for {
		curBlockStart := nextBlockStart
		curBlockEnd := curBlockStart.Add(btc.blocksGap + btc.blockGenerationTime)
		nextBlockStart = curBlockEnd.Add(time.Second)

		if blockTime.Equal(curBlockStart) || blockTime.After(curBlockStart) && blockTime.Before(nextBlockStart) {
			bgs.start = curBlockStart
			bgs.duration = btc.blocksGap + btc.blockGenerationTime
			bgs.nodePosition = curNodeIndex
			return bgs, nil
		}

		if btc.nodesCount > 0 {
			curNodeIndex = (curNodeIndex + 1) % btc.nodesCount
		}
	}
}
