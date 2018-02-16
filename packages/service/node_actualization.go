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

// NodePaused is global flag represents that node is not generate blocks, only collect it
var NodePaused = abool.New()

type NodeActualizer struct {
	AvailableBlockchainGap int64

	// is local state of node
	activityPaused bool

	startDaemonsCh chan []string
}

// Run is starting node monitoring
func (n *NodeActualizer) Run() <-chan []string {
	// Waiting until daemons started
	time.Sleep(time.Second * 5)
	n.startDaemonsCh = make(chan []string)
	go func() {
		defer close(n.startDaemonsCh)
		for {
			actual, err := n.checkBlockchainActuality()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("checking blockchain actuality")
				return
			}

			if !actual && !n.activityPaused {
				err := n.pauseNodeActivity()
				if err != nil {
					log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("pausing blockchain activity")
					return
				}
			}

			if actual && n.activityPaused {
				err := n.resumeNodeActivity()
				if err != nil {
					log.WithFields(log.Fields{"type": consts.BCActualizationError, "err": err}).Error("resuming blockchain activity")
					return
				}
			}

			time.Sleep(time.Minute * 5)
		}
	}()

	return n.startDaemonsCh
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
	if curBlock.BlockID == 0 || curBlock.BlockID+n.AvailableBlockchainGap < maxBlockID {
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
	err := model.SetStopNow()
	if err != nil {
		return errors.Wrapf(err, "stopping daemons through database")
	}
	utils.DaemonsCount = 0

	n.activityPaused = true
	NodePaused.Set()
	return nil
}

func (n *NodeActualizer) resumeNodeActivity() error {
	n.activityPaused = false
	NodePaused.UnSet()
	n.startDaemonsCh <- []string{"BlocksCollection", "Confirmations", "Notificator", "Scheduler"}
	return nil
}
