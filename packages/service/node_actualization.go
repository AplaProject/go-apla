// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
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
	curBlock := &model.InfoBlock{}
	_, err := curBlock.Get()
	if err != nil {
		return false, errors.Wrapf(err, "retrieving info block")
	}

	remoteHosts, err := GetNodesBanService().FilterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		return false, err
	}

	_, maxBlockID, err := tcpclient.HostWithMaxBlock(ctx, remoteHosts)
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

func (n *NodeActualizer) pauseNodeActivity() {
	np.Set(PauseTypeUpdatingBlockchain)
}

func (n *NodeActualizer) resumeNodeActivity() {
	np.Unset()
}
