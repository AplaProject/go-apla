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
