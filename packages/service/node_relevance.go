package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"
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
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Error("retrieving info block from db")
		return false, errors.Wrapf(err, "retrieving info block from db")
	}

	remoteHosts := syspar.GetRemoteHosts()
	// Node is single in blockchain network and it can't be irrelevant
	if len(remoteHosts) == 0 {
		return true, nil
	}

	ctx, _ := context.WithCancel(context.Background())
	_, maxBlockID, err := utils.ChooseBestHost(ctx, remoteHosts, &log.Entry{Logger: &log.Logger{}})
	if err != nil {
		if err == utils.ErrNodesUnavailable {
			log.WithFields(log.Fields{"hosts": remoteHosts}).Info("can't connect to others, stopping node relevance")
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
