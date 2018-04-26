package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMaxBlockID(t *testing.T) {
	var ret getMaxBlockIDResult
	err := sendGet(`maxblockid`, nil, &ret)
	assert.NoError(t, err)
}

func TestGetBlockInfo(t *testing.T) {
	var ret getBlockInfoResult
	err := sendGet(`block/1`, nil, &ret)
	assert.NoError(t, err)
}
