//go:generate sh -c "mockery -inpkg -name Clock -print > file.tmp && mv file.tmp clock_mock.go"

package utils

import (
	"time"

	"github.com/pkg/errors"
)

// BlockTimeCalculator calculating block generation time
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

func NewBlockTimeCalculator(firstBlockTime time.Time,
	blockGenerationTime, blocksGap time.Duration,
	nodesCount int64,
) BlockTimeCalculator {
	return BlockTimeCalculator{
		firstBlockTime:      firstBlockTime,
		blockGenerationTime: blockGenerationTime,
		blocksGap:           blocksGap,
		nodesCount:          nodesCount,
	}
}

func (btc *BlockTimeCalculator) TimeToGenerate(nodePosition int64) (bool, error) {
	bgs, err := btc.countBlockTime(btc.clock.Now())
	if err != nil {
		return false, err
	}

	return int64(bgs.nodePosition) == nodePosition, nil
}

func (btc *BlockTimeCalculator) ValidateBlock(nodePosition int64, at time.Time) (bool, error) {
	bgs, err := btc.countBlockTime(at)
	if err != nil {
		return false, err
	}

	return int64(bgs.nodePosition) == nodePosition, nil
}

func (btc *BlockTimeCalculator) SetClock(clock Clock) *BlockTimeCalculator {
	btc.clock = clock
	return btc
}

func (btc *BlockTimeCalculator) countBlockTime(blockTime time.Time) (blockGenerationState, error) {
	bgs := blockGenerationState{}
	nextBlockStart := btc.firstBlockTime
	curNodeIndex := 0

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

		if curNodeIndex == int(btc.nodesCount-1) {
			curNodeIndex = 0
		} else {
			curNodeIndex++
		}
	}
}