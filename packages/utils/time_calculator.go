// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package utils

import (
	"time"

	"github.com/AplaProject/go-apla/packages/model"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
)

type BlockTimeCounter struct {
	start       time.Time
	duration    time.Duration
	numberNodes int
}

// Block returns serial block number for time
func (btc *BlockTimeCounter) Block(t time.Time) int {
	return int((t.Sub(btc.start) - 1) / btc.duration)
}

// NodePosition returns generating node position for time
func (btc *BlockTimeCounter) NodePosition(t time.Time) int {
	return btc.Block(t) % btc.numberNodes
}

// ValidateBlock checks conformity between time and nodePosition
func (btc *BlockTimeCounter) ValidateBlock(t time.Time, nodePosition int) bool {
	return btc.NodePosition(t) == nodePosition
}

func (btc *BlockTimeCounter) BlockForTimeExists(t time.Time, nodePosition int) (bool, error) {
	startInterval, endInterval := btc.RangesByTime(t)

	b := &model.Block{}
	blocks, err := b.GetNodeBlocksAtTime(startInterval, endInterval, int64(nodePosition))
	if err != nil {
		return false, err
	}

	if len(blocks) != 0 {
		return false, DuplicateBlockError
	}

	return true, nil
}

// NextTime returns next generation time for node position at time
func (btc *BlockTimeCounter) NextTime(t time.Time, nodePosition int) time.Time {
	block := btc.Block(t)
	curNodePosition := block % btc.numberNodes

	d := nodePosition - curNodePosition
	if curNodePosition >= nodePosition {
		d += btc.numberNodes
	}

	return btc.start.Add(btc.duration*time.Duration(block+d) + time.Second)
}

// RangesByTime returns start and end of interval by time
func (btc *BlockTimeCounter) RangesByTime(t time.Time) (start, end time.Time) {
	atTimePosition := btc.NodePosition(t)
	end = btc.start.Add(btc.duration*time.Duration(atTimePosition) + 1)
	start = end.Add(-btc.duration)
	return
}

// NewBlockTimeCounter return initialized BlockTimeCounter
func NewBlockTimeCounter() *BlockTimeCounter {
	firstBlock, _ := syspar.GetFirstBlockData()
	blockGenerationDuration := time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
	blocksGapDuration := time.Second * time.Duration(syspar.GetGapsBetweenBlocks())

	return &BlockTimeCounter{
		start:       time.Unix(int64(firstBlock.Time), 0),
		duration:    blockGenerationDuration + blocksGapDuration,
		numberNodes: int(syspar.GetCountOfActiveNodes()),
	}
}
