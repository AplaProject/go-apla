package packages

// QueueChecker allow check queue to generate current block
type QueueChecker interface {
	TimeToGenerate(position int64) (bool, error)
}
