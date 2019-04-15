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

package daemons

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/rollback"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// ErrNodesUnavailable is returned when all nodes is unavailable
var ErrNodesUnavailable = errors.New("All nodes unavailable")

// BlocksCollection collects and parses blocks
func BlocksCollection(ctx context.Context, d *daemon) error {
	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	return blocksCollection(ctx, d)
}

func InitialLoad(logger *log.Entry) error {

	// check for initial load
	toLoad, err := needLoad(logger)
	if err != nil {
		return err
	}

	if toLoad {
		logger.Debug("start first block loading")

		if err := firstLoad(logger); err != nil {
			return err
		}

		model.UpdateSchema()
	}

	return nil
}

var bcOnRun uint32

func blocksCollection(ctx context.Context, d *daemon) (err error) {
	if !atomic.CompareAndSwapUint32(&bcOnRun, 0, 1) {
		return nil
	}
	defer func() {
		atomic.StoreUint32(&bcOnRun, 0)
	}()
	host, maxBlockID, err := getHostWithMaxID(ctx, d.logger)
	if err != nil {
		d.logger.WithFields(log.Fields{"error": err}).Warn("on checking best host")
		return err
	}

	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting cur blockID")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("Info block not found")
		return errors.New("Info block not found")
	}

	if infoBlock.BlockID >= maxBlockID {
		log.WithFields(log.Fields{"blockID": infoBlock.BlockID, "maxBlockID": maxBlockID}).Debug("Max block is already in the host")
		return nil
	}

	DBLock()
	defer func() {
		service.NodeDoneUpdatingBlockchain()
		DBUnlock()
	}()

	// update our chain till maxBlockID from the host
	return UpdateChain(ctx, d, host, maxBlockID)
}

// UpdateChain load from host all blocks from our last block to maxBlockID
func UpdateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {
	var (
		err error
	)
	maxBlockReached := false
	// get current block id from our blockchain
	curBlock := &model.InfoBlock{}
	if _, err = curBlock.Get(); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting info block")
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	playRawBlock := func(rb []byte) error {

		bl, err := block.ProcessBlockWherePrevFromBlockchainTable(rb, true)
		defer func() {
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("retrieving blockchain from node")
				banNode(host, bl, err)
			}
		}()

		if err != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("processing block")
			return err
		}

		// hash compare could be failed in the case of fork
		hashMatched, errCheck := bl.CheckHash()
		if errCheck != nil {
			d.logger.WithFields(log.Fields{"error": errCheck, "type": consts.BlockError}).Error("checking block hash")
		}

		if !hashMatched {
			transaction.CleanCache()

			rollbackBlockID := bl.Header.BlockID - 1
			if errCheck == block.ErrIncorrectRollbackHash {
				rollbackBlockID--
			}
			limit := bl.Header.BlockID - rollbackBlockID

			//it should be fork, replace our previous blocks to ones from the host
			err = GetBlocks(ctx, rollbackBlockID, limit, host)
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
				return err
			}
		}

		bl.PrevHeader, err = block.GetBlockDataFromBlockChain(bl.Header.BlockID - 1)
		if err != nil {
			return utils.ErrInfo(fmt.Errorf("can't get block %d", bl.Header.BlockID-1))
		}

		if err = bl.Check(); err != nil {
			return err
		}

		if err = bl.PlaySafe(); err != nil {
			return err
		}
		if maxBlockID == bl.Header.BlockID {
			maxBlockReached = true
		}
		return nil
	}

	var count int
	st := time.Now()

	defer func() {
		if maxBlockReached {
			return
		}
		nodePosition, _ := syspar.GetNodePositionByKeyID(conf.Config.KeyID)
		header := utils.BlockData{
			BlockID:      maxBlockID,
			Time:         time.Now().Unix(),
			NodePosition: nodePosition,
		}
		block := &block.Block{Header: header}
		banNode(host, block, fmt.Errorf("host returns max block, but max block not reached"))
		time.Sleep(1000 * time.Millisecond)
	}()
	d.logger.WithFields(log.Fields{"min_block": curBlock.BlockID, "max_block": maxBlockID, "count": maxBlockID - curBlock.BlockID}).Info("starting downloading blocks")

	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID += int64(network.BlocksPerRequest) {

		if loopErr := func() error {
			ctxDone, cancel := context.WithCancel(ctx)
			defer func() {
				cancel()
				d.logger.WithFields(log.Fields{"count": count, "time": time.Since(st).String()}).Info("blocks downloaded")
			}()

			rawBlocksChan, err := tcpclient.GetBlocksBodies(ctxDone, host, blockID, false)
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("getting block body")
				return err
			}

			for rawBlock := range rawBlocksChan {
				if err = playRawBlock(rawBlock); err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("playing raw block")
					return err
				}
				count++
			}

			return nil
		}(); loopErr != nil {
			return loopErr
		}
	}
	return nil
}

// init first block from file or from embedded value
func loadFirstBlock(logger *log.Entry) error {
	newBlock, err := ioutil.ReadFile(conf.Config.FirstBlockPath)
	if err != nil {
		logger.WithFields(log.Fields{
			"type": consts.IOError, "error": err, "path": conf.Config.FirstBlockPath,
		}).Error("reading first block from file")
	}

	if err = block.InsertBlockWOForks(newBlock, false, true); err != nil {
		logger.WithFields(log.Fields{"type": consts.ParserError, "error": err}).Error("inserting new block")
		return err
	}

	return nil
}

func firstLoad(logger *log.Entry) error {
	DBLock()
	defer DBUnlock()

	return loadFirstBlock(logger)
}

func needLoad(logger *log.Entry) (bool, error) {
	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting info block")
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if infoBlock.BlockID == 0 {
		logger.Debug("blockchain should be loaded")
		return true, nil
	}
	return false, nil
}

func banNode(host string, block *block.Block, err error) {
	var (
		reason             string
		blockId, blockTime int64
	)
	if err != nil {
		if err == transaction.ErrDuplicatedTx {
			return
		}
		reason = err.Error()
	}

	if block != nil {
		blockId, blockTime = block.Header.BlockID, block.Header.Time
	} else {
		blockId, blockTime = -1, time.Now().Unix()
	}

	log.WithFields(log.Fields{"reason": reason, "host": host, "block_id": blockId, "block_time": blockTime}).Debug("ban node")

	n, err := syspar.GetNodeByHost(host)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("getting node by host")
		return
	}

	err = service.GetNodesBanService().RegisterBadBlock(n, blockId, blockTime, reason)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "node": n.KeyID, "block": blockId}).Error("registering bad block from node")
	}
}

// GetHostWithMaxID returns host with maxBlockID
func getHostWithMaxID(ctx context.Context, logger *log.Entry) (host string, maxBlockID int64, err error) {

	nbs := service.GetNodesBanService()
	hosts, err := nbs.FilterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("on filtering banned hosts")
	}

	host, maxBlockID, err = tcpclient.HostWithMaxBlock(ctx, hosts)
	if len(hosts) == 0 || err == tcpclient.ErrNodesUnavailable {
		hosts = conf.GetNodesAddr()
		return tcpclient.HostWithMaxBlock(ctx, hosts)
	}

	return
}

// GetBlocks is returning blocks
func GetBlocks(ctx context.Context, blockID, limit int64, host string) error {
	blocks, err := getBlocks(ctx, blockID+1, limit, host)
	if err != nil {
		return err
	}
	transaction.CleanCache()

	// mark all transaction as unverified
	_, err = model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"type":  consts.DBError,
		}).Error("marking verified and not used transactions unverified")
		return utils.ErrInfo(err)
	}

	// get starting blockID from slice of blocks
	if len(blocks) > 0 {
		blockID = blocks[len(blocks)-1].Header.BlockID
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	block := &model.Block{}
	myRollbackBlocks, err := block.GetBlocksFrom(blockID-1, "desc", 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		err := rollback.RollbackBlock(block.Data, true)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return processBlocks(blocks)
}

func getBlocks(ctx context.Context, blockID, limit int64, host string) ([]*block.Block, error) {
	rollback := syspar.GetRbBlocks1()

	badBlocks := make(map[int64]string)

	blocks := make([]*block.Block, 0)
	var count int64

	// load the block bodies from the host
	blocksCh, err := tcpclient.GetBlocksBodies(ctx, host, blockID, true)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	for binaryBlock := range blocksCh {
		if blockID < 2 {
			break
		}

		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			break
		}

		if count >= limit {
			break
		}

		block, err := block.ProcessBlockWherePrevFromBlockchainTable(binaryBlock, true)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}

		if badBlocks[block.Header.BlockID] == string(converter.BinToHex(block.Header.Sign)) {
			log.WithFields(log.Fields{"block_id": block.Header.BlockID, "type": consts.InvalidObject}).Error("block is bad")
			return nil, utils.ErrInfo(errors.New("bad block"))
		}
		if block.Header.BlockID != blockID {
			log.WithFields(log.Fields{"header_block_id": block.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// save the block
		blocks = append(blocks, block)
		blockID--
		count++
	}

	return blocks, nil
}

func processBlocks(blocks []*block.Block) error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*block.Block, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]

		if prevBlocks[b.Header.BlockID-1] != nil {
			b.PrevHeader.Hash = prevBlocks[b.Header.BlockID-1].Header.Hash
			b.PrevHeader.RollbacksHash, err = block.GetRollbacksHash(dbTransaction, b.Header.BlockID-1)
			if err != nil {
				dbTransaction.Rollback()
				return err
			}
			b.PrevHeader.Time = prevBlocks[b.Header.BlockID-1].Header.Time
			b.PrevHeader.BlockID = prevBlocks[b.Header.BlockID-1].Header.BlockID
			b.PrevHeader.EcosystemID = prevBlocks[b.Header.BlockID-1].Header.EcosystemID
			b.PrevHeader.KeyID = prevBlocks[b.Header.BlockID-1].Header.KeyID
			b.PrevHeader.NodePosition = prevBlocks[b.Header.BlockID-1].Header.NodePosition
		}

		hash, err := crypto.DoubleHash([]byte(b.Header.ForSha(b.PrevHeader, b.MrklRoot)))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
		}
		b.Header.Hash = hash

		if err := b.Check(); err != nil {
			dbTransaction.Rollback()
			return err
		}

		if err := b.Play(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[b.Header.BlockID] = b

		// for last block we should update block info
		if i == 0 {
			err := block.UpdBlockInfo(dbTransaction, b)
			if err != nil {
				dbTransaction.Rollback()
				return utils.ErrInfo(err)
			}
		}
		if b.SysUpdate {
			if err := syspar.SysUpdate(dbTransaction); err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				return utils.ErrInfo(err)
			}
		}
	}

	// If all right we can delete old blockchain and write new
	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]
		// Delete old blocks from blockchain
		bl := &model.Block{}
		err = bl.DeleteById(dbTransaction, b.Header.BlockID)
		if err != nil {
			dbTransaction.Rollback()
			return err
		}
		// insert new blocks into blockchain
		if err := block.InsertIntoBlockchain(dbTransaction, b); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	return dbTransaction.Commit()
}
