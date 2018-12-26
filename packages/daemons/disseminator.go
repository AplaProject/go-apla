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

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/nodeban"
	"github.com/AplaProject/go-apla/packages/queue"

	log "github.com/sirupsen/logrus"
)

// Disseminator is send to all nodes from nodes_connections the following data
// if we are full node(miner): sends blocks and transactions hashes
// else send the full transactions
func Disseminator(ctx context.Context, d *daemon) error {
	DBLock()
	defer DBUnlock()

	isFullNode := true
	myNodePosition, err := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("finding node")
		isFullNode = false
	}

	if isFullNode {
		// send blocks and transactions hashes
		d.logger.Debug("we are full_node, sending hashes")
		return sendBlockWithTxHashes(ctx, myNodePosition, d.logger)
	}

	// we are not full node for this StateID and WalletID, so just send transactions
	d.logger.Debug("we are full_node, sending transactions")
	return sendTransactions(ctx, d.logger)
}

func sendTransactions(ctx context.Context, logger *log.Entry) error {
	// get unsent transactions
	// form packet to send
	return queue.SendTxQueue.ProcessAllItems(func(txs []*blockchain.Transaction) error {
		if len(txs) > 0 {
			hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("on getting remotes hosts")
				return err
			}

			if err := tcpclient.SendTransacitionsToAll(ctx, hosts, txs); err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending transactions")
				return err
			}
		}
		return nil
	})
}

// send block and transactions hashes
func sendBlockWithTxHashes(ctx context.Context, fullNodeID int64, logger *log.Entry) error {
	return queue.SendTxQueue.ProcessAllItems(func(trs []*blockchain.Transaction) error {
		block, isEmpty, err := queue.SendBlockQueue.Dequeue()
		if err != nil {
			return err
		}
		if isEmpty {
			return nil
		}
		if len(trs) == 0 && block == nil {
			// it's nothing to send
			logger.Debug("nothing to send")
			return nil
		}

		hosts, err := nodeban.GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("on getting remotes hosts")
			return err
		}

		if err := tcpclient.SendFullBlockToAll(ctx, hosts, block, trs, fullNodeID); err != nil {
			log.WithFields(log.Fields{"type": consts.TCPClientError, "error": err}).Warn("on sending block with hashes to all")
			return err
		}
		return nil
	})

}
