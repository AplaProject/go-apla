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

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
)

var updatingEndWhilePaused = make(chan struct{})

type NodeRelevanceService struct {
	availableBlockchainGap int64
	checkingInterval       time.Duration
}

func NewNodeRelevanceService(availableBlockchainGap int64, checkingInterval time.Duration) NodeRelevanceService {
	return NodeRelevanceService{
		availableBlockchainGap: availableBlockchainGap,
		checkingInterval:       checkingInterval,
	}
}

// Run is starting node monitoring
func (n *NodeRelevanceService) Run(ctx context.Context) {
	go func() {
		log.Info("Node relevance monitoring started")
		for {
			relevance, err := n.checkNodeRelevance(ctx)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.BCRelevanceError, "err": err}).Error("checking blockchain relevance")
				return
			}

			if !relevance && !IsNodePaused() {
				log.Info("Node Relevance Service is pausing node activity")
				n.pauseNodeActivity()
			}

			if relevance && IsNodePaused() {
				log.Info("Node Relevance Service is resuming node activity")
				n.resumeNodeActivity()
			}

			select {
			case <-time.After(n.checkingInterval):
			case <-updatingEndWhilePaused:
			}
		}
	}()
}

func NodeDoneUpdatingBlockchain() {
	go func() {
		if IsNodePaused() {
			updatingEndWhilePaused <- struct{}{}
		}
	}()
}

func (n *NodeRelevanceService) checkNodeRelevance(ctx context.Context) (relevant bool, err error) {
	curBlock := &model.InfoBlock{}
	_, err = curBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Error("retrieving info block from db")
		return false, errors.Wrapf(err, "retrieving info block from db")
	}

	nbs := GetNodesBanService()
	remoteHosts, err := nbs.FilterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on filtering banned hosts")
	}
	// Node is single in blockchain network and it can't be irrelevant
	if len(remoteHosts) == 0 {
		return true, nil
	}

	_, maxBlockID, err := tcpclient.HostWithMaxBlock(ctx, remoteHosts)
	if err != nil {
		if err == tcpclient.ErrNodesUnavailable {
			return false, nil
		}
		return false, errors.Wrapf(err, "choosing best host")
	}

	// Node can't connect to others
	if maxBlockID == -1 {
		log.WithFields(log.Fields{"hosts": remoteHosts}).Info("can't connect to others, stopping node relevance")
		return false, nil
	}

	// Node blockchain is stale
	if curBlock.BlockID+n.availableBlockchainGap < maxBlockID {
		log.WithFields(log.Fields{"maxBlockID": maxBlockID, "curBlockID": curBlock.BlockID, "Gap": n.availableBlockchainGap}).Info("blockchain is stale, stopping node relevance")
		return false, nil
	}

	return true, nil
}

func (n *NodeRelevanceService) pauseNodeActivity() {
	np.Set(PauseTypeUpdatingBlockchain)
}

func (n *NodeRelevanceService) resumeNodeActivity() {
	np.Unset()
}
