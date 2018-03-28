//go:generate sh -c "mockery -inpkg -name intervalBlocksCounter -print > file.tmp && mv file.tmp block_counter_mock.go"
package utils

import "github.com/GenesisKernel/go-genesis/packages/model"

type intervalBlocksCounter interface {
	count(state blockGenerationState) (int, error)
}

type blocksCounter struct {
}

func (bc *blocksCounter) count(state blockGenerationState) (int, error) {
	blockchain := &model.Block{}
	blocks, err := blockchain.GetNodeBlocksAtTime(state.start, state.start.Add(state.duration), state.nodePosition)
	if err != nil {
		return 0, err
	}
	return len(blocks), nil
}
