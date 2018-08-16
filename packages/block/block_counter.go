//go:generate sh -c "mockery -inpkg -name intervalBlocksCounter -print > file.tmp && mv file.tmp block_counter_mock.go"

package block

import "github.com/GenesisKernel/go-genesis/packages/blockchain"

const lastNBlocks = 5

type intervalBlocksCounter interface {
	count(state blockGenerationState) (int, error)
}

type blocksCounter struct {
}

func (bc *blocksCounter) count(state blockGenerationState) (int, error) {
	blocks, err := blockchain.GetNLastBlocks(lastNBlocks)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, b := range blocks {
		if b.NodePosition == statePosition && (b.Time >= state.start.Unix() || b.Time <= state.start.Add(state.duration).Unix()) {
			count += 1
		}
	}
	return count, nil
}
