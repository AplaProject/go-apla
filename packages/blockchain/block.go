package blockchain

import (
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

const blockPrefix = "block-"

type Block struct {
	Header   *utils.BlockData
	TrData   [][]byte
	PrevHash []byte
	Key      []byte
}
