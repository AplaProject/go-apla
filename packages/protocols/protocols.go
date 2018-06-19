package protocols

import "time"

// TimeToGenerate allow check block generation time for nodePosition
func TimeToGenerate(at time.Time, nodePosition int) (bool, error) {
	tbc := NewBlockTimeCounter()
	return tbc.TimeToGenerate(at, nodePosition)
}

func BlockForTimeExists(t time.Time, nodePosition int) (bool, error) {
	btc := NewBlockTimeCounter()
	return btc.BlockForTimeExists(t, nodePosition)
}
