package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tevino/abool"

	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

// nodePaused is global state represents that node does not generate blocks only collect it
var nodePaused = abool.New()

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
func (n *NodeRelevanceService) Run() {
	go func() {
		log.Info("Node relevance monitoring started")
		for {
			relevance, err := n.checkNodeRelevance()
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

func IsNodePaused() bool {
	return nodePaused.IsSet()
}

func NodeDoneUpdatingBlockchain() {
	go func() {
		if IsNodePaused() {
			updatingEndWhilePaused <- struct{}{}
		}
	}()
}

func (n *NodeRelevanceService) checkNodeRelevance() (relevant bool, err error) {
	curBlock := &model.InfoBlock{}
	_, err = curBlock.Get()
	if err != nil {
		return false, errors.Wrapf(err, "retrieving info block")
	}

	remoteHosts := syspar.GetRemoteHosts()
	// Node is single in blockchain network and it can't be irrelevant
	if len(remoteHosts) == 0 {
		return true, nil
	}

	ctx, _ := context.WithCancel(context.Background())
	_, maxBlockID, err := utils.ChooseBestHost(ctx, remoteHosts, &log.Entry{Logger: &log.Logger{}})
	if err != nil {
		return false, errors.Wrapf(err, "choosing best host")
	}

	// Node can't connect to others
	if maxBlockID == -1 {
		return false, nil
	}

	// Node blockchain is stale
	if curBlock.BlockID+n.availableBlockchainGap < maxBlockID {
		return false, nil
	}

	return true, nil
}

func (n *NodeRelevanceService) pauseNodeActivity() {
	nodePaused.Set()
}

func (n *NodeRelevanceService) resumeNodeActivity() {
	nodePaused.UnSet()
}
