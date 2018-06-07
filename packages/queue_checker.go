package packages

import (
	"time"
)

// QueueChecker allow check queue to generate current block
type QueueChecker interface {
	TimeToGenerate(position int64) (bool, error)
	ValidateBlock(position int64, at time.Time) (bool, error)
	NextTime(position int64, t time.Time) time.Time
	BlockForTimeExists(t time.Time, nodePosition int) (bool, error)
	RangesByTime(t time.Time) (start, end time.Time)
}
