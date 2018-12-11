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

package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
)

// DefaultBlockchainGap is default value for the number of lagging blocks
const DefaultBlockchainGap int64 = 10

type NodeActualizer struct {
	availableBlockchainGap int64
}

func NewNodeActualizer(availableBlockchainGap int64) NodeActualizer {
	return NodeActualizer{
		availableBlockchainGap: availableBlockchainGap,
	}
}

// Run is starting node monitoring
func (n *NodeActualizer) Run(ctx context.Context) {
	go func() {
		log.Info("Node Actualizer monitoring starting")
		for {
			if ctx.Err() != nil {
				log.WithFields(log.Fields{"error": ctx.Err(), "type": consts.ContextError}).Error("context error")
				return
			}

			actual, err := n.checkBlockchainActuality(ctx)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("checking blockchain actuality")
				return
			}

			if !actual && !IsNodePaused() {
				log.Info("Node Actualizer is pausing node activity")
				n.pauseNodeActivity()
			}

			if actual && IsNodePaused() {
				log.Info("Node Actualizer is resuming node activity")
				n.resumeNodeActivity()
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

func (n *NodeActualizer) checkBlockchainActuality(ctx context.Context) (bool, error) {
	block, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		return false, err
	}

	remoteHosts := syspar.GetRemoteHosts()

	_, maxBlockID, err := tcpclient.HostWithMaxBlock(ctx, remoteHosts)
	if err != nil {
		return false, errors.Wrapf(err, "choosing best host")
	}
	var curBlockID int64
	if found {
		curBlockID = block.Header.BlockID
	}

	// Currently this node is downloading blockchain
	if curBlockID == 0 || curBlockID+n.availableBlockchainGap < maxBlockID {
		return false, nil
	}

	foreignBlock, found, err := blockchain.GetMaxForeignBlock(nil, conf.Config.KeyID)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving last foreign block")
	}

	if !found {
		return false, nil
	}

	// Node did not accept any blocks for an hour
	t := time.Unix(foreignBlock.Header.Time, 0)
	if time.Since(t).Minutes() > 30 && len(remoteHosts) > 1 {
		return false, nil
	}

	return true, nil
}

func (n *NodeActualizer) pauseNodeActivity() {
	np.Set(PauseTypeUpdatingBlockchain)
}

func (n *NodeActualizer) resumeNodeActivity() {
	np.Unset()
}
