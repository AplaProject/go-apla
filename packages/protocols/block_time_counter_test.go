package protocols

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockTimeCounter(t *testing.T) {
	btc := BlockTimeCounter{
		start:       time.Unix(0, 0),
		duration:    5 * time.Second,
		numberNodes: 3,
	}

	at := time.Unix(13, 0)

	queue, err := btc.queue(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, queue)

	np, err := btc.nodePosition(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, np)

	nextTime, err := btc.nextTime(at, 2)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(25, 0).Add(1*time.Millisecond), nextTime)

	start, end, err := btc.RangeByTime(at)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(10, 0).Add(1*time.Millisecond), start)
	assert.Equal(t, time.Unix(15, 0), end)
	fmt.Println("ranges:", start.Unix(), end.Unix())
}

func TestRangeByTime(t *testing.T) {
	btc := BlockTimeCounter{
		start:       time.Unix(1532977623, 0),
		duration:    4 * time.Second,
		numberNodes: 1,
	}

	st, end, err := btc.RangeByTime(time.Unix(1533062723, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	st, end, err = btc.RangeByTime(time.Unix(1533062724, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	// 1532977623
	st, end, err = btc.RangeByTime(time.Unix(1532977624, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	// 1533062719 1533062723
	// 1533062723 1533062727
	// 1532977623 1532977627
}
