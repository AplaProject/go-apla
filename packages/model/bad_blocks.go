// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package model

import (
	"time"
)

type BadBlocks struct {
	ID             int64
	ProducerNodeId int64
	BlockId        int64
	ConsumerNodeId int64
	BlockTime      time.Time
	Deleted        bool
}

// TableName returns name of table
func (r BadBlocks) TableName() string {
	return "1_bad_blocks"
}

// BanRequests represents count of unique ban requests for node
type BanRequests struct {
	ProducerNodeId int64
	Count          int64
}

// GetNeedToBanNodes is returns list of ban requests for each node
func (r *BadBlocks) GetNeedToBanNodes(now time.Time, blocksPerNode int) ([]BanRequests, error) {
	var res []BanRequests

	err := DBConn.
		Raw(
			`SELECT
				producer_node_id,
				COUNT(consumer_node_id) as count
			FROM (
				SELECT
					producer_node_id,
					consumer_node_id,
					count(DISTINCT block_id)
				FROM
				"1_bad_blocks"
				WHERE
					block_time > ?::date - interval '24 hours'
					AND deleted = 0
				GROUP BY
					producer_node_id,
					consumer_node_id
				HAVING
					count(DISTINCT block_id) >= ?) AS tbl
			GROUP BY
			producer_node_id`,
			now,
			blocksPerNode,
		).
		Scan(&res).
		Error

	return res, err
}

func (r *BadBlocks) GetNodeBlocks(nodeId int64, now time.Time) ([]BadBlocks, error) {
	var res []BadBlocks
	err := DBConn.
		Table(r.TableName()).
		Model(&BadBlocks{}).
		Where(
			"producer_node_id = ? AND block_time > ?::date - interval '24 hours' AND deleted = ?",
			nodeId,
			now,
			false,
		).
		Scan(&res).
		Error

	return res, err
}
