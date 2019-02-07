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

package syspar

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"

	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// NumberNodes is the number of nodes
	NumberNodes = `number_of_nodes`
	// FuelRate is the rate
	FuelRate = `fuel_rate`
	// FullNodes is the list of nodes
	FullNodes = `full_nodes`
	// GapsBetweenBlocks is the time between blocks
	GapsBetweenBlocks = `gap_between_blocks`
	// BlockchainURL is the address of the blockchain file.  For those who don't want to collect it from nodes
	BlockchainURL = `blockchain_url`
	// MaxBlockSize is the maximum size of the block
	MaxBlockSize = `max_block_size`
	// MaxTxSize is the maximum size of the transaction
	MaxTxSize = `max_tx_size`
	// MaxForsignSize is the maximum size of the forsign of transaction
	MaxForsignSize = `max_forsign_size`
	// MaxBlockFuel is the maximum fuel of the block
	MaxBlockFuel = `max_fuel_block`
	// MaxTxFuel is the maximum fuel of the transaction
	MaxTxFuel = `max_fuel_tx`
	// MaxTxCount is the maximum count of the transactions
	MaxTxCount = `max_tx_block`
	// MaxBlockGenerationTime is the time limit for block generation (in ms)
	MaxBlockGenerationTime = `max_block_generation_time`
	// MaxColumns is the maximum columns in tables
	MaxColumns = `max_columns`
	// MaxIndexes is the maximum indexes in tables
	MaxIndexes = `max_indexes`
	// MaxBlockUserTx is the maximum number of user's transactions in one block
	MaxBlockUserTx = `max_tx_block_per_user`
	// SizeFuel is the fuel cost of 1024 bytes of the transaction data
	SizeFuel = `price_tx_data`
	// CommissionWallet is the address for commissions
	CommissionWallet = `commission_wallet`
	// RbBlocks1 rollback from queue_bocks
	RbBlocks1 = `rollback_blocks`
	// BlockReward value of reward, which is chrged on block generation
	BlockReward = "block_reward"
	// IncorrectBlocksPerDay is value of incorrect blocks per day before global ban
	IncorrectBlocksPerDay = `incorrect_blocks_per_day`
	// NodeBanTime is value of ban time for bad nodes (in ms)
	NodeBanTime = `node_ban_time`
	// LocalNodeBanTime is value of local ban time for bad nodes (in ms)
	LocalNodeBanTime = `node_ban_time_local`
	// CommissionSize is the value of the commission
	CommissionSize = `commission_size`
	// Test equals true or 1 if we have a test blockchain
	Test = `test`
	// PrivateBlockchain is value defining blockchain mode
	PrivateBlockchain = `private_blockchain`

	// CostDefault is the default maximum cost of F
	CostDefault = int64(20000000)
)

var (
	cache = map[string]string{
		BlockchainURL: "https://raw.githubusercontent.com/egaas-blockchain/egaas-blockchain.github.io/master/testnet_blockchain",
	}
	nodes             = make(map[int64]*FullNode)
	nodesByPosition   = make([]*FullNode, 0)
	fuels             = make(map[int64]string)
	wallets           = make(map[int64]string)
	mutex             = &sync.RWMutex{}
	activeNodes       = make(map[*FullNode]struct{})
	firstBlockData    *consts.FirstBlock
	errFirstBlockData = errors.New("Failed to get data of the first block")
	errNodeDisabled   = errors.New("node is disabled")
)

// SysUpdate reloads/updates values of system parameters
func SysUpdate(dbTransaction *model.DbTransaction) error {
	var err error
	systemParameters, err := model.GetAllSystemParameters(dbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all system parameters")
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, param := range systemParameters {
		cache[param.Name] = param.Value
	}
	if len(cache[FullNodes]) > 0 {
		if err = updateNodes(); err != nil {
			return err
		}
	}
	getParams := func(name string) (map[int64]string, error) {
		res := make(map[int64]string)
		if len(cache[name]) > 0 {
			ifuels := make([][]string, 0)
			err = json.Unmarshal([]byte(cache[name]), &ifuels)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling params from json")
				return res, err
			}
			for _, item := range ifuels {
				if len(item) < 2 {
					continue
				}
				res[converter.StrToInt64(item[0])] = item[1]
			}
		}
		return res, nil
	}
	fuels, err = getParams(FuelRate)
	wallets, err = getParams(CommissionWallet)

	return err
}

func updateNodes() (err error) {
	nodes = make(map[int64]*FullNode)
	nodesByPosition = make([]*FullNode, 0)

	items := make([]*FullNode, 0)
	if len(cache[FullNodes]) > 0 {
		err = json.Unmarshal([]byte(cache[FullNodes]), &items)

		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "v": cache[FullNodes]}).Error("unmarshalling full nodes from json")
			return err
		}
	}

	activeNodes = make(map[*FullNode]struct{})
	nodesByPosition = items
	for _, item := range items {
		nodes[item.KeyID] = item

		if !item.Stopped {
			activeNodes[item] = struct{}{}
		}
	}

	return nil
}

// addFullNodeKeys adds node by keys to list of nodes
func addFullNodeKeys(keyID int64, publicKey []byte) {
	nodesByPosition = append(nodesByPosition, &FullNode{
		KeyID:     keyID,
		PublicKey: publicKey,
	})
}

// GetNode is retrieving node by wallet
func GetNode(wallet int64) *FullNode {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := nodes[wallet]; ok {
		return ret
	}
	return nil
}

func GetNodes() []FullNode {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]FullNode, 0, len(nodesByPosition))
	for _, node := range nodesByPosition {
		result = append(result, *node)
	}

	return result
}

// GetNodePositionByKeyID is returning node position by key id
func GetNodePositionByKeyID(keyID int64) (int64, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	var counter int64
	for _, item := range nodesByPosition {
		if item.Stopped {
			if item.KeyID == keyID {
				return 0, errNodeDisabled
			}
			continue
		}

		if item.KeyID == keyID {
			return counter, nil
		}

		counter++
	}

	return 0, fmt.Errorf("Incorrect keyID")
}

// GetCountOfActiveNodes is count of nodes with stopped = false
func GetCountOfActiveNodes() int64 {
	return int64(len(activeNodes))
}

// GetNumberOfNodes is count number of nodes
func GetNumberOfNodes() int64 {
	return int64(len(nodesByPosition))
}

func GetNumberOfNodesFromDB(transaction *model.DbTransaction) int64 {
	sp := &model.SystemParameter{}
	sp.GetTransaction(transaction, FullNodes)
	var fullNodes []map[string]interface{}
	if len(sp.Value) > 0 {
		if err := json.Unmarshal([]byte(sp.Value), &fullNodes); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "value": sp.Value}).Error("unmarshalling fullnodes from JSON")
		}
	}
	if len(fullNodes) == 0 {
		return 1
	}
	return int64(len(fullNodes))
}

// GetNodeByPosition is retrieving node by position
func GetNodeByPosition(position int64) (*FullNode, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodesByPosition)) <= position {
		return nil, fmt.Errorf("incorrect position")
	}
	return nodesByPosition[position], nil
}

func GetNodeByHost(host string) (FullNode, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	for _, n := range nodes {
		if n.TCPAddress == host {
			return *n, nil
		}
	}

	return FullNode{}, fmt.Errorf("incorrect host")
}

// GetNodeHostByPosition is retrieving node host by position
func GetNodeHostByPosition(position int64) (string, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	nodeData, err := GetNodeByPosition(position)
	if err != nil {
		return "", err
	}
	return nodeData.TCPAddress, nil
}

// GetNodePublicKeyByPosition is retrieving node public key by position
func GetNodePublicKeyByPosition(position int64) ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if int64(len(nodesByPosition)) <= position {
		return nil, fmt.Errorf("incorrect position")
	}
	nodeData, err := GetNodeByPosition(position)
	if err != nil {
		return nil, err
	}
	return nodeData.PublicKey, nil
}

// SysInt64 is converting sys string to int64
func SysInt64(name string) int64 {
	return converter.StrToInt64(SysString(name))
}

// SysInt is converting sys string to int
func SysInt(name string) int {
	return converter.StrToInt(SysString(name))
}

// GetSizeFuel is returns fuel size
func GetSizeFuel() int64 {
	return SysInt64(SizeFuel)
}

// GetBlockchainURL is retrieving blockchain url
func GetBlockchainURL() string {
	return SysString(BlockchainURL)
}

// GetFuelRate is returning fuel rate
func GetFuelRate(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := fuels[ecosystem]; ok {
		return ret
	}
	return ``
}

// GetCommissionWallet is returns commission wallet
func GetCommissionWallet(ecosystem int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ret, ok := wallets[ecosystem]; ok {
		return ret
	}
	return wallets[1]
}

// GetMaxBlockSize is returns max block size
func GetMaxBlockSize() int64 {
	return converter.StrToInt64(SysString(MaxBlockSize))
}

// GetMaxBlockFuel is returns max block fuel
func GetMaxBlockFuel() int64 {
	return converter.StrToInt64(SysString(MaxBlockFuel))
}

// GetMaxTxFuel is returns max tx fuel
func GetMaxTxFuel() int64 {
	return converter.StrToInt64(SysString(MaxTxFuel))
}

// GetMaxBlockGenerationTime is returns max block generation time (in ms)
func GetMaxBlockGenerationTime() int64 {
	return converter.StrToInt64(SysString(MaxBlockGenerationTime))
}

// GetMaxTxSize is returns max tx size
func GetMaxTxSize() int64 {
	return converter.StrToInt64(SysString(MaxTxSize))
}

// GetMaxTxTextSize is returns max tx text size
func GetMaxForsignSize() int64 {
	return converter.StrToInt64(SysString(MaxForsignSize))
}

// GetGapsBetweenBlocks is returns gaps between blocks
func GetGapsBetweenBlocks() int64 {
	return converter.StrToInt64(SysString(GapsBetweenBlocks))
}

// GetMaxTxCount is returns max tx count
func GetMaxTxCount() int {
	return converter.StrToInt(SysString(MaxTxCount))
}

// GetMaxColumns is returns max columns
func GetMaxColumns() int {
	return converter.StrToInt(SysString(MaxColumns))
}

// GetMaxIndexes is returns max indexes
func GetMaxIndexes() int {
	return converter.StrToInt(SysString(MaxIndexes))
}

// GetMaxBlockUserTx is returns max tx block user
func GetMaxBlockUserTx() int {
	return converter.StrToInt(SysString(MaxBlockUserTx))
}

func IsTestMode() bool {
	return SysString(Test) == `true` || SysString(Test) == `1`
}

func GetIncorrectBlocksPerDay() int {
	return converter.StrToInt(SysString(IncorrectBlocksPerDay))
}

func GetNodeBanTime() time.Duration {
	return time.Millisecond * time.Duration(converter.StrToInt64(SysString(NodeBanTime)))
}

func GetLocalNodeBanTime() time.Duration {
	return time.Millisecond * time.Duration(converter.StrToInt64(SysString(LocalNodeBanTime)))
}

// GetRemoteHosts returns array of hostnames excluding myself
func GetRemoteHosts() []string {
	ret := make([]string, 0)

	mutex.RLock()
	defer mutex.RUnlock()

	for keyID, item := range nodes {
		if keyID != conf.Config.KeyID && !item.Stopped {
			ret = append(ret, item.TCPAddress)
		}
	}
	return ret
}

// SysString returns string value of the system parameter
func SysString(name string) string {
	mutex.RLock()
	ret := cache[name]
	mutex.RUnlock()
	return ret
}

// GetRbBlocks1 is returns RbBlocks1
func GetRbBlocks1() int64 {
	return SysInt64(RbBlocks1)
}

// HasSys returns boolean whether this system parameter exists
func HasSys(name string) bool {
	mutex.RLock()
	_, ok := cache[name]
	mutex.RUnlock()
	return ok
}

// SetFirstBlockData sets data of first block to global variable
func SetFirstBlockData(data *consts.FirstBlock) {
	mutex.Lock()
	defer mutex.Unlock()

	firstBlockData = data

	// If list of nodes is empty, then used node from the first block
	if len(nodesByPosition) == 0 {
		keyID := crypto.Address(firstBlockData.PublicKey)
		addFullNodeKeys(keyID, firstBlockData.NodePublicKey)

		activeNodes[&FullNode{
			KeyID:     keyID,
			PublicKey: firstBlockData.NodePublicKey,
			Stopped:   false,
		}] = struct{}{}
	}
}

// GetFirstBlockData gets data of first block from global variable
func GetFirstBlockData() (*consts.FirstBlock, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if firstBlockData == nil {
		return nil, errFirstBlockData
	}

	return firstBlockData, nil
}

// IsPrivateBlockchain returns the value of private_blockchain system parameter or true
func IsPrivateBlockchain() bool {
	par := SysString(PrivateBlockchain)
	return len(par) > 0 && par != `0` && par != `false`
}

func GetMaxCost() int64 {
	cost := GetMaxTxFuel()
	if cost == 0 {
		cost = CostDefault
	}
	return cost
}
