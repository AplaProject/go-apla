package protocols

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlockTimeCounter(t *testing.T) {
	btc := BlockTimeCounter{
		start:       time.Unix(0, 0),
		duration:    5 * time.Second,
		numberNodes: 3,
	}

	at := time.Unix(13, 0)

	queue, err := btc.Queue(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, queue)

	np, err := btc.NodePosition(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, np)

	nextTime, err := btc.NextTime(at, 2)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(25, 0).Add(1*time.Millisecond), nextTime)

	start, end, err := btc.RangesByTime(at)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(10, 0).Add(1*time.Millisecond), start)
	assert.Equal(t, time.Unix(15, 0), end)
	fmt.Println("ranges:", start.Unix(), end.Unix())
}
