//go:generate sh -c "mockery -inpkg -name Clock -print > file.tmp && mv file.tmp clock_mock.go"

package utils

import (
	"time"

	"github.com/pkg/errors"
)

type Clock interface {
	Now() time.Time
}

type BlockTimeCalculator struct {
	clock Clock

	firstBlockTime      time.Time
	blockGenerationTime time.Duration
	blocksGap           time.Duration

	nodesCount int64
}

type blockGenerationState struct {
	start    time.Time
	duration time.Duration

	nodePosition int
}

var TimeError = errors.New("current time before first block")

func NewBlockTimeCalculator(clock Clock,
	firstBlockTime time.Time,
	blockGenerationTime, blocksGap time.Duration,
	nodesCount int64,
) BlockTimeCalculator {
	return BlockTimeCalculator{
		clock:               clock,
		firstBlockTime:      firstBlockTime,
		blockGenerationTime: blockGenerationTime,
		blocksGap:           blocksGap,
		nodesCount:          nodesCount,
	}
}

func (btc *BlockTimeCalculator) TimeToGenerate(nodePosition int64) (bool, error) {
	bgs, err := btc.countBlockTime(false, time.Time{})
	if err != nil {
		return false, err
	}

	return int64(bgs.nodePosition) == nodePosition, nil
}

func (btc *BlockTimeCalculator) ValidateBlock(nodePosition int64, at time.Time) (bool, error) {
	bgs, err := btc.countBlockTime(true, at)
	if err != nil {
		return false, err
	}

	return int64(bgs.nodePosition) == nodePosition, nil
}

func (btc *BlockTimeCalculator) countBlockTime(past bool, pastTime time.Time) (blockGenerationState, error) {
	curTime := btc.clock.Now()
	if past {
		curTime = pastTime
	}

	bgs := blockGenerationState{}
	nextBlockStart := btc.firstBlockTime
	curNodeIndex := 0

	if curTime.Before(nextBlockStart) {
		return blockGenerationState{}, TimeError
	}

	for {
		curBlockStart := nextBlockStart
		curBlockEnd := curBlockStart.Add(btc.blocksGap + btc.blockGenerationTime)
		nextBlockStart = curBlockEnd.Add(time.Second)

		if curTime.Equal(curBlockStart) || curTime.After(curBlockStart) && curTime.Before(nextBlockStart) {
			bgs.start = curBlockStart
			bgs.duration = btc.blocksGap + btc.blockGenerationTime
			bgs.nodePosition = curNodeIndex
			return bgs, nil
		}

		if curNodeIndex == int(btc.nodesCount-1) {
			curNodeIndex = 0
		} else {
			curNodeIndex++
		}
	}
}
