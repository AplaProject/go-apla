package nodeban

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

import (
	"strconv"
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/queue"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type localBannedNode struct {
	FullNode       *syspar.FullNode
	LocalUnBanTime time.Time
}

type NodesBanService struct {
	localBannedNodes map[int64]localBannedNode
	fullNodes        []syspar.FullNode

	m *sync.Mutex
}

var nbs *NodesBanService

// GetNodesBanService is returning nodes ban service
func GetNodesBanService() *NodesBanService {
	return nbs
}

// InitNodesBanService initializing nodes ban storage
func InitNodesBanService() error {
	nbs = &NodesBanService{
		localBannedNodes: make(map[int64]localBannedNode),
		m:                &sync.Mutex{},
	}

	nbs.refreshNodes()
	return nil
}

// RegisterBadBlock is set node to local ban and saving bad block to global registry
func (nbs *NodesBanService) RegisterBadBlock(node syspar.FullNode, badBlockId, blockTime int64, reason string) error {
	if nbs.IsBanned(node) {
		return nil
	}

	nbs.localBan(node)

	err := nbs.newBadBlock(node, badBlockId, blockTime, reason)
	if err != nil {
		return err
	}

	return nil
}

// IsBanned is allows to check node ban (local or global)
func (nbs *NodesBanService) IsBanned(node syspar.FullNode) bool {
	nbs.refreshNodes()

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
			if !fn.UnbanTime.Equal(time.Unix(0, 0)) {
				return true
			} else {
				break
			}
		}
	}

	return false
}

func (nbs *NodesBanService) refreshNodes() {
	nbs.m.Lock()
	nbs.fullNodes = syspar.GetNodes()
	nbs.m.Unlock()
}

func (nbs *NodesBanService) localBan(node syspar.FullNode) {
	nbs.m.Lock()
	defer nbs.m.Unlock()

	nbs.localBannedNodes[node.KeyID] = localBannedNode{
		FullNode:       &node,
		LocalUnBanTime: time.Now().Add(syspar.GetLocalNodeBanTime()),
	}
}

func (nbs *NodesBanService) newBadBlock(producer syspar.FullNode, blockId, blockTime int64, reason string) error {
	nodePrivateKey, err := utils.GetNodePrivateKey()
	if err != nil || len(nodePrivateKey) < 1 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	var currentNode syspar.FullNode
	nbs.m.Lock()
	for _, fn := range nbs.fullNodes {
		if fn.KeyID == conf.Config.KeyID {
			currentNode = fn
			break
		}
	}
	nbs.m.Unlock()

	if currentNode.KeyID == 0 {
		return errors.New("cant find current node in full nodes list")
	}

	params := map[string]string{
		"ProducerNodeID": strconv.FormatInt(producer.KeyID, 10),
		"ConsumerNodeID": strconv.FormatInt(currentNode.KeyID, 10),
		"BlockID":        strconv.FormatInt(blockId, 10),
		"Timestamp":      strconv.FormatInt(blockTime, 10),
		"Reason":         reason,
	}
	forSign := []string{strconv.FormatInt(producer.KeyID, 10), strconv.FormatInt(currentNode.KeyID, 10), strconv.FormatInt(blockId, 10), strconv.FormatInt(blockTime, 10), reason}
	smartTx, err := smart.CallContract("NewBadBlock", 1, params, forSign)
	if err != nil {
		return err
	}
	return queue.ValidateTxQueue.Enqueue(smartTx)
}

func (nbs *NodesBanService) FilterBannedHosts(hosts []string) ([]string, error) {
	var goodHosts []string
	for _, h := range hosts {
		n, err := syspar.GetNodeByHost(h)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "host": h}).Error("getting node by host")
			return nil, err
		}

		if !nbs.IsBanned(n) {
			goodHosts = append(goodHosts, n.TCPAddress)
		}
	}
	return goodHosts, nil
}
