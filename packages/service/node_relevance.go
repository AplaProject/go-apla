package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
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
	curBlock, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Error("retrieving last block from db")
		return false, errors.Wrapf(err, "retrieving last block from db")
	}

	if !found {
		return true, nil
	}

	remoteHosts := syspar.GetRemoteHosts()
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
	if curBlock.Header.BlockID+n.availableBlockchainGap < maxBlockID {
		log.WithFields(log.Fields{"maxBlockID": maxBlockID, "curBlockID": curBlock.Header.BlockID, "Gap": n.availableBlockchainGap}).Info("blockchain is stale, stopping node relevance")
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
