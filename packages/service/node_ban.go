package service

import (
	"sync"
	"time"

	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type localBannedNode struct {
	FullNode       *syspar.FullNode
	LocalUnBanTime time.Time
}

type NodesBanService struct {
	localBannedNodes map[int64]localBannedNode
	fullNodes        map[int64]*syspar.FullNode

	m *sync.Mutex
}

var nbs *NodesBanService

func GetNodesBanService() *NodesBanService {
	return nbs
}

func InitNodesBanService(fullNodes map[int64]*syspar.FullNode) error {
	nbs = &NodesBanService{
		localBannedNodes: make(map[int64]localBannedNode),
		fullNodes:        fullNodes,
		m:                &sync.Mutex{},
	}
	return nil
}

func (nbs *NodesBanService) RegisterBadBlock(node syspar.FullNode, badBlockId, blockTime int64) error {
	if !nbs.IsBanned(node) {
		nbs.localBan(node)
	}

	err := nbs.newBadBlock(node, badBlockId, blockTime)
	if err != nil {
		return err
	}

	return nil
}

func (nbs *NodesBanService) IsBanned(node syspar.FullNode) bool {
	nbs.m.Lock()
	defer nbs.m.Unlock()

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
			if fn.UnbanTime.Equal(time.Unix(0, 0)) {
				return false
			} else {
				return true
			}
		}
	}

	return false
}

func (nbs *NodesBanService) localBan(node syspar.FullNode) {
	nbs.m.Lock()
	defer nbs.m.Unlock()

	nbs.localBannedNodes[node.KeyID] = localBannedNode{
		FullNode:       &node,
		LocalUnBanTime: time.Now().Add(time.Minute * 30),
	}
}

func (nbs *NodesBanService) newBadBlock(producer syspar.FullNode, blockId, blockTime int64) error {
	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	var cn syspar.FullNode
	nbs.m.Lock()
	for _, fn := range nbs.fullNodes {
		if fn.KeyID == conf.Config.KeyID {
			cn = *fn
			break
		}
	}
	nbs.m.Unlock()

	if cn.KeyID == 0 {
		return errors.New("cant find current node in full nodes list")
	}

	params := make([]byte, 0)
	for _, p := range []int64{producer.KeyID, cn.KeyID, blockId, blockTime} {
		converter.EncodeLenInt64(&params, p)
	}

	vm := smart.GetVM(false, 0)
	contract := smart.VMGetContract(vm, "NewBadBlock", 1)
	info := contract.Block.Info.(*script.ContractInfo)

	err = tx.BuildTransaction(tx.SmartContract{
		Header: tx.Header{
			Type:        int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: 1,
			KeyID:       conf.Config.KeyID,
		},
		SignedBy: smart.PubToID(NodePublicKey),
		Data:     params,
	},
		NodePrivateKey,
		NodePublicKey,
		strconv.FormatInt(producer.KeyID, 10),
		strconv.FormatInt(cn.KeyID, 10),
		strconv.FormatInt(blockId, 10),
		strconv.FormatInt(blockTime, 10),
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		return err
	}

	return nil
}
