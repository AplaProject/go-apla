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

package daemons

import (
	"context"
	"time"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/nodeban"

	log "github.com/sirupsen/logrus"
)

var tick int

// Confirmations gets and checks blocks from nodes
// Getting amount of nodes, which has the same hash as we do
func Confirmations(ctx context.Context, d *daemon) error {

	// the first 2 minutes we sleep for 10 sec for blocks to be collected
	tick++

	d.sleepTime = 1 * time.Second
	if tick < 12 {
		d.sleepTime = 10 * time.Second
	}

	blocks, err := blockchain.GetUnconfirmedBlocks(nil, consts.MIN_CONFIRMED_NODES)
	if err != nil {
		return err
	}
	if len(blocks) == 0 {
		return nil
	}

	return confirmationsBlocks(ctx, d, blocks)
}

func confirmationsBlocks(ctx context.Context, d *daemon, blocks []*blockchain.BlockWithHash) error {
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": err}).Error("error in context")
			return err
		}

		hashStr := string(converter.BinToHex(block.Hash))
		d.logger.WithFields(log.Fields{"hash": hashStr}).Debug("checking hash")

		hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			return err
		}

		ch := make(chan string)
		for i := 0; i < len(hosts); i++ {
			host, err := tcpclient.NormalizeHostAddress(hosts[i], consts.DEFAULT_TCP_PORT)
			if err != nil {
				d.logger.WithFields(log.Fields{"host": host[i], "type": consts.ParseError, "error": err}).Error("wrong host address")
				continue
			}

			d.logger.WithFields(log.Fields{"host": host, "block_hash": block.Hash}).Debug("checking block id confirmed at node")
			go func() {
				IsReachable(host, block.Hash, ch, d.logger)
			}()
		}
		var answer string
		var st0, st1 int
		for i := 0; i < len(hosts); i++ {
			answer = <-ch
			if answer == hashStr {
				st1++
			} else {
				st0++
			}
		}
		confirmation := &blockchain.Confirmation{}
		confirmation.BlockID = block.Block.Header.BlockID
		confirmation.Good = st1
		confirmation.Bad = st0
		confirmation.Time = time.Now().Unix()
		if err = confirmation.Insert(nil, block.Hash); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving confirmation")
			return err
		}
	}

	return nil
}

// IsReachable checks if there is blockID on the host
func IsReachable(host string, blockHash []byte, ch0 chan string, logger *log.Entry) {
	ch := make(chan string, 1)
	go func() {
		ch <- tcpclient.CheckConfirmation(host, blockHash, logger)
	}()
	select {
	case reachable := <-ch:
		ch0 <- reachable
	case <-time.After(consts.WAIT_CONFIRMED_NODES * time.Second):
		ch0 <- "0"
	}
}
