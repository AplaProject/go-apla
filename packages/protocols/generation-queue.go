package protocols

import (
	"errors"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	log "github.com/sirupsen/logrus"
)

// QueueChecker allow check queue to generate current block
type QueueChecker interface {
	TimeToGenerate(position int64) (bool, error)
	NextTime(position int64, t time.Time) (time.Time, error)
	BlockForTimeExists(t time.Time, nodePosition int) (bool, error)
	RangeByTime(t time.Time) (start, end time.Time)
}

var (
	WrongNodePositionError = errors.New("wrong node position")
	TimeError              = errors.New("current time before first block")
	DuplicateBlockError    = errors.New("block for this time interval exists")
)

type BlockTimeCounter struct {
	start       time.Time
	duration    time.Duration
	numberNodes int
}

// Queue returns serial queue number for time
func (btc *BlockTimeCounter) queue(t time.Time) (int, error) {
	if t.Before(btc.start) {
		return -1, TimeError
	}

	return int((t.Sub(btc.start) - 1) / btc.duration), nil
}

// NodePosition returns generating node position for time
func (btc *BlockTimeCounter) NodePosition(t time.Time) (int, error) {
	queue, err := btc.queue(t)
	if err != nil {
		return -1, err
	}

	return queue % btc.numberNodes, nil
}

// BlockForTimeExists checks conformity between time and nodePosition
// changes functionality of ValidateBlock prevent blockTimeCalculator
func (btc *BlockTimeCounter) BlockForTimeExists(t time.Time, nodePosition int) (bool, error) {
	startInterval, endInterval, err := btc.RangeByTime(t)
	if err != nil {
		return false, err
	}

	b := &model.Block{}
	blocks, err := b.GetNodeBlocksAtTime(startInterval, endInterval, int64(nodePosition))
	if err != nil {
		return false, err
	}

	if len(blocks) != 0 {
		log.WithFields(log.Fields{"type": "block_time_counter", "error": DuplicateBlockError, "start": startInterval, "end": endInterval}).Error("")
		return false, DuplicateBlockError
	}

	return true, nil
}

// NextTime returns next generation time for node position at time
func (btc *BlockTimeCounter) NextTime(t time.Time, nodePosition int) (time.Time, error) {
	if nodePosition >= btc.numberNodes {
		return time.Unix(0, 0), WrongNodePositionError
	}

	queue, err := btc.queue(t)
	if err != nil {
		return time.Unix(0, 0), err
	}
	curNodePosition := queue % btc.numberNodes

	d := nodePosition - curNodePosition
	if curNodePosition >= nodePosition {
		d += btc.numberNodes
	}

	return btc.start.Add(btc.duration*time.Duration(queue+d) + time.Millisecond), nil
}

// RangesByTime returns start and end of interval by time
func (btc *BlockTimeCounter) RangeByTime(t time.Time) (start, end time.Time, err error) {
	queue, err := btc.queue(t)
	if err != nil {
		st := time.Unix(0, 0)
		return st, st, err
	}

	end = btc.start.Add(btc.duration * (time.Duration(queue) + 1))
	start = end.Add(-btc.duration).Add(1 * time.Millisecond)
	return
}

func (btc *BlockTimeCounter) TimeToGenerate(at time.Time, nodePosition int) (bool, error) {
	if nodePosition >= btc.numberNodes {
		return false, WrongNodePositionError
	}

	position, err := btc.NodePosition(at)
	return position == nodePosition, err
}

// NewBlockTimeCounter return initialized BlockTimeCounter
func NewBlockTimeCounter() *BlockTimeCounter {
	firstBlock, _ := syspar.GetFirstBlockData()
	blockGenerationDuration := time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
	blocksGapDuration := time.Second * time.Duration(syspar.GetGapsBetweenBlocks())

	return &BlockTimeCounter{
		start:       time.Unix(int64(firstBlock.Time), 0),
		duration:    blockGenerationDuration + blocksGapDuration,
		numberNodes: int(syspar.GetNumberOfNodes()),
	}
}
