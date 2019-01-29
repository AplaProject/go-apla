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
	"time"

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/network/tcpserver"
	"github.com/AplaProject/go-apla/packages/nodeban"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
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

		model.UpdateSchema()
	}

	return nil
}

func blocksCollection(ctx context.Context, d *daemon) (err error) {
	host, maxBlockID, err := getHostWithMaxID(ctx, d.logger)
	if err != nil {
		d.logger.WithFields(log.Fields{"error": err}).Warn("on checking best host")
		return err
	}

	lastBlock, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("Getting last block")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Info("last block not found")
		return nil
	}

	if lastBlock.Header.BlockID >= maxBlockID {
		log.WithFields(log.Fields{"blockID": lastBlock.Header.BlockID, "maxBlockID": maxBlockID}).Debug("Max block is already in the host")
		return nil
	}

	DBLock()
	defer func() {
		DBUnlock()
		service.NodeDoneUpdatingBlockchain()
	}()

	// update our chain till maxBlockID from the host
	return UpdateChain(ctx, d, host, maxBlockID)
}

// UpdateChain load from host all blocks from our last block to maxBlockID
func UpdateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {
	var (
		err   error
		count int
	)

	// get current block id from our blockchain
	curBlock, curBlockHash, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting last block")
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	playRawBlock := func(rb []byte) error {
		blck := &blockchain.Block{}
		if err := blck.Unmarshal(rb); err != nil {
			return err
		}
		txs, err := blck.Transactions(nil)
		if err != nil {
			return err
		}

		bl, err := block.ProcessBlockWherePrevFromBlockchainTable(blck, txs, true, nil)
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
		hashMatched, thisErrIsOk := bl.CheckHash()
		if thisErrIsOk != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("checking block hash")
		}

		if !hashMatched {
			transaction.CleanCache()
			//it should be fork, replace our previous blocks to ones from the host
			err = GetBlocks(ctx, bl.PrevHash, host)
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.ParserError}).Error("processing block")
				return err
			}
		}

		bl.PrevHeader, err = block.GetBlockDataFromBlockChain(nil, bl.PrevHash)
		if err != nil {
			return utils.ErrInfo(fmt.Errorf("can't get block %d", bl.Header.BlockID-1))
		}

		if err = bl.Check(); err != nil {
			return err
		}

		if err = bl.PlaySafe(txs); err != nil {
			return err
		}

		return nil
	}

	st := time.Now()
	d.logger.Infof("starting downloading blocks from %d to %d (%d) \n", curBlock.Header.BlockID, maxBlockID, maxBlockID-curBlock.Header.BlockID)

	count = 0
	curBlockID := curBlock.Header.BlockID + 1
	nextBlock, found, err := blockchain.GetNextBlock(nil, curBlockHash)
	blockHash := nextBlock.Hash
	if err != nil {
		return err
	}
	if !found {
		blockHash = curBlockHash
		curBlockID = curBlock.Header.BlockID
	}
	for blockID := curBlockID; blockID <= maxBlockID; blockID += int64(tcpserver.BlocksPerRequest) {
		ctxDone, cancel := context.WithCancel(ctx)
		if loopErr := func() error {
			defer func() {
				cancel()
				d.logger.WithFields(log.Fields{"count": count, "time": time.Since(st).String()}).Info("blocks downloaded")
			}()

			rawBlocksChan, err := tcpclient.GetBlocksBodies(ctxDone, host, blockHash, false)
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
			blocks, err := blockchain.GetNBlocksFrom(nil, blockHash, tcpserver.BlocksPerRequest, 1)
			if err != nil {
				return err
			}
			blockHash = blocks[len(blocks)-1].Hash
			return nil
		}(); loopErr != nil {
			return loopErr
		}
	}
	return nil
}

func needLoad(logger *log.Entry) (bool, error) {
	_, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting info block")
		return false, err
	}
	// we have empty blockchain, we need to load blockchain from file or other source
	if !found {
		logger.Debug("blockchain should be loaded")
		return true, nil
	}
	return false, nil
}

func banNode(host string, block *block.PlayableBlock, err error) {
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

	err = nodeban.GetNodesBanService().RegisterBadBlock(n, blockId, blockTime, reason)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "node": n.KeyID, "block": blockId}).Error("registering bad block from node")
	}
}

// GetHostWithMaxID returns host with maxBlockID
func getHostWithMaxID(ctx context.Context, logger *log.Entry) (host string, maxBlockID int64, err error) {

	nbs := nodeban.GetNodesBanService()
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
func GetBlocks(ctx context.Context, blockHash []byte, host string) error {
	blocks, err := getBlocks(ctx, blockHash, host)
	if err != nil {
		return err
	}
	transaction.CleanCache()

	// get starting blockID from slice of blocks
	if len(blocks) > 0 {
		blockHash = blocks[len(blocks)-1].Hash
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	myRollbackBlocks, err := blockchain.DeleteBlocksFrom(nil, blockHash)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, block := range myRollbackBlocks {
		err := rollback.RollbackBlock(block.Block, block.Hash)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return processBlocks(blocks)
}

func getBlocks(ctx context.Context, blockHash []byte, host string) ([]*block.PlayableBlock, error) {
	rollback := syspar.GetRbBlocks1()

	blocks := make([]*block.PlayableBlock, 0)
	var count int64

	// load the block bodies from the host
	blocksCh, err := tcpclient.GetBlocksBodies(ctx, host, blockHash, true)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	for binaryBlock := range blocksCh {
		// if the limit of blocks received from the node was exaggerated
		if count > int64(rollback) {
			break
		}
		bl := &blockchain.Block{}
		if err := bl.Unmarshal(binaryBlock); err != nil {
			break
		}
		txs, err := bl.Transactions(nil)
		if err != nil {
			break
		}

		block, err := block.ProcessBlockWherePrevFromBlockchainTable(bl, txs, true, nil)
		if err != nil {
			return nil, utils.ErrInfo(err)
		}

		if string(block.Hash) != string(blockHash) {
			log.WithFields(log.Fields{"header_block_hash": block.Hash, "block_id": blockHash, "type": consts.InvalidObject}).Error("block hashes does not match")
			return nil, utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// TODO: add checking for MAX_BLOCK_SIZE

		// the public key of the one who has generated this block
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(block.Header.NodePosition)
		if err != nil {
			log.WithFields(log.Fields{"header_block_hash": block.Hash, "block_id": blockHash, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.ErrInfo(err)
		}

		// save the block
		blocks = append(blocks, block)
		count++

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, []byte(block.ForSign()), block.Header.Sign, true)
		if okSignErr == nil {
			break
		}
	}

	return blocks, nil
}

func processBlocks(blocks []*block.PlayableBlock) error {
	ldbTx, err := blockchain.DB.OpenTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.LevelDBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}
	metaTx := model.MetaStorage.Begin(true)

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*block.PlayableBlock, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]

		if prevBlocks[b.Header.BlockID-1] != nil {
			b.PrevHash = prevBlocks[b.Header.BlockID-1].Hash
			b.PrevHeader.Time = prevBlocks[b.Header.BlockID-1].Header.Time
			b.PrevHeader.BlockID = prevBlocks[b.Header.BlockID-1].Header.BlockID
			b.PrevHeader.EcosystemID = prevBlocks[b.Header.BlockID-1].Header.EcosystemID
			b.PrevHeader.KeyID = prevBlocks[b.Header.BlockID-1].Header.KeyID
			b.PrevHeader.NodePosition = prevBlocks[b.Header.BlockID-1].Header.NodePosition
		}

		hash, err := crypto.DoubleHash([]byte(b.ForSha()))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
		}
		b.Hash = hash

		if err := b.Check(); err != nil {
			ldbTx.Discard()
			dbTransaction.Rollback()
			return err
		}
		_, txs, err := b.ToBlockchainBlock()
		if err != nil {
			ldbTx.Discard()
			dbTransaction.Rollback()
			return err
		}

		if err := b.Play(dbTransaction, txs, ldbTx, metaTx); err != nil {
			ldbTx.Discard()
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[b.Header.BlockID] = b

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
		// insert new blocks into blockchain
		bBlock, transactions, err := b.ToBlockchainBlock()
		if err != nil {
			return err
		}
		if err := bBlock.Insert(ldbTx, transactions); err != nil {
			ldbTx.Discard()
			dbTransaction.Rollback()
			metaTx.Rollback()
			return err
		}
	}

	// TODO double phase commit
	err = dbTransaction.Commit()
	err = ldbTx.Commit()
	err = metaTx.Commit()

	return err
}
