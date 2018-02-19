package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tevino/abool"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

// DefaultBlockchainGap is default value for the number of lagging blocks
const DefaultBlockchainGap int64 = 10

// NodePaused is global flag represents that node is not generate blocks, only collect it
var NodePaused = abool.New()

type NodeActualizer struct {
	availableBlockchainGap int64

	activityPaused bool
}

func NewNodeActualizer(availableBlockchainGap int64) NodeActualizer {
	return NodeActualizer{
		availableBlockchainGap: availableBlockchainGap,
	}
}

// Run is starting node monitoring
func (n *NodeActualizer) Run() {
	go func() {
		log.Info("Node Actualizer monitoring starting")
		for {
			actual, err := n.checkBlockchainActuality()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("checking blockchain actuality")
				return
			}

			if !actual && !n.activityPaused {
				log.Info("Node Actualizer is pausing node activity")

				err := n.pauseNodeActivity()
				if err != nil {
					log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("pausing blockchain activity")
					return
				}
			}

			if actual && n.activityPaused {
				log.Info("Node Actualizer is resuming node activity")

				err := n.resumeNodeActivity()
				if err != nil {
					log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("resuming blockchain activity")
					return
				}
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

func (n *NodeActualizer) checkBlockchainActuality() (bool, error) {
	curBlock := &model.InfoBlock{}
	_, err := curBlock.Get()
	if err != nil {
		return false, errors.Wrapf(err, "retrieving info block")
	}

	remoteHosts := syspar.GetRemoteHosts()
	ctx, _ := context.WithCancel(context.Background())
	_, maxBlockID, err := utils.ChooseBestHost(ctx, remoteHosts, &log.Entry{Logger: &log.Logger{}})
	if err != nil {
		return false, errors.Wrapf(err, "choosing best host")
	}

	// Currently this node is downloading blockchain
	if curBlock.BlockID == 0 || curBlock.BlockID+n.availableBlockchainGap < maxBlockID {
		return false, nil
	}

	foreignBlock := &model.Block{}
	_, err = foreignBlock.GetMaxForeignBlock(conf.Config.KeyID)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving last foreign block")
	}

	// Node did not accept any blocks for an hour
	t := time.Unix(foreignBlock.Time, 0)
	if time.Since(t).Minutes() > 30 && len(remoteHosts) > 1 {
		return false, nil
	}

	return true, nil
}

func (n *NodeActualizer) pauseNodeActivity() error {
	n.activityPaused = true
	NodePaused.Set()

	return nil
}

func (n *NodeActualizer) resumeNodeActivity() error {
	n.activityPaused = false
	NodePaused.UnSet()
	return nil
}
