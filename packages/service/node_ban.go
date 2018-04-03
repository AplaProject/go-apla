package service

import (
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
)

type localBannedNode struct {
	FullNode       *syspar.FullNode
	LocalUnBanTime time.Time
}

type NodesBanService struct {
	badBlocksValue int

	localBannedNodes map[int64]localBannedNode
	fullNodes        []*syspar.FullNode

	*sync.Mutex
}

var nbs *NodesBanService

func GetNodesBanService() *NodesBanService {
	return nbs
}

func InitNodesBanService(badBlocksValue int) error {
	nbs = &NodesBanService{badBlocksValue: badBlocksValue}
	return nil
}

func (nbs *NodesBanService) RegisterBadBlock(node syspar.FullNode, badBlockId int64) error {
	if !nbs.IsBanned(node) {
		nbs.localBan(node)
	}

	err := nbs.newBadBlock(node, badBlockId)
	if err != nil {
		return err
	}

	return nil
}

func (nbs *NodesBanService) IsBanned(node syspar.FullNode) bool {
	nbs.Lock()
	defer nbs.Unlock()

	// Searching for local ban
	now := time.Now()
	if fn, ok := nbs.localBannedNodes[node.KeyID]; ok {
		if now.Equal(fn.LocalUnBanTime) || now.After(fn.LocalUnBanTime) {
			delete(nbs.localBannedNodes, node.KeyID)
			return false
		}

		return true
	}

	// Searching for global ban.
	// Here we don't estimating global ban expiration. If ban time doesn't equal zero - we assuming
	// that node is still banned (even if `unban` time has already passed)
	for _, fn := range nbs.fullNodes {
		if fn.KeyID == node.KeyID {
			if fn.GlobalUnBanTime.After(time.Unix(0, 0)) {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func (nbs *NodesBanService) RefreshParams(nodes []*syspar.FullNode, badBlocksValue int) {
	nbs.Lock()
	defer nbs.Unlock()
	nbs.fullNodes = nodes
	nbs.badBlocksValue = badBlocksValue
}

func (nbs *NodesBanService) localBan(node syspar.FullNode) {
	nbs.Lock()
	defer nbs.Unlock()

	nbs.localBannedNodes[node.KeyID] = localBannedNode{
		FullNode:       &node,
		LocalUnBanTime: time.Now().Add(time.Minute * 30),
	}
}

// TODO
func (nbs *NodesBanService) newBadBlock(node syspar.FullNode, blockId int64) error { return nil }
