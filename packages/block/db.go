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

package block

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/protocols"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// UpdBlockInfo updates info_block table
func UpdBlockInfo(dbTransaction *model.DbTransaction, block *Block) error {
	blockID := block.Header.BlockID
	// for the local tests
	forSha := block.Header.ForSha(block.PrevHeader, block.MrklRoot)

	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
	}

	block.Header.Hash = hash
	if block.Header.BlockID == 1 {
		ib := &model.InfoBlock{
			Hash:           hash,
			BlockID:        blockID,
			Time:           block.Header.Time,
			EcosystemID:    block.Header.EcosystemID,
			KeyID:          block.Header.KeyID,
			NodePosition:   converter.Int64ToStr(block.Header.NodePosition),
			CurrentVersion: fmt.Sprintf("%d", block.Header.Version),
		}
		err := ib.Create(dbTransaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating info block")
			return fmt.Errorf("error insert into info_block %s", err)
		}
	} else {
		ibUpdate := &model.InfoBlock{
			Hash:         hash,
			BlockID:      blockID,
			Time:         block.Header.Time,
			EcosystemID:  block.Header.EcosystemID,
			KeyID:        block.Header.KeyID,
			NodePosition: converter.Int64ToStr(block.Header.NodePosition),
			Sent:         0,
		}
		if err := ibUpdate.Update(dbTransaction); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating info block")
			return fmt.Errorf("error while updating info_block: %s", err)
		}
	}

	return nil
}

// InsertIntoBlockchain inserts a block into the blockchain
func InsertIntoBlockchain(transaction *model.DbTransaction, block *Block) error {
	// for local tests
	blockID := block.Header.BlockID

	// record into the block chain
	bl := &model.Block{}
	err := bl.DeleteById(transaction, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
		return err
	}
	rollbackTx := &model.RollbackTx{}
	blockRollbackTxs, err := rollbackTx.GetBlockRollbackTransactions(transaction, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block rollback txs")
		return err
	}
	buffer := bytes.Buffer{}
	for _, rollbackTx := range blockRollbackTxs {
		rollbackTxBytes, err := json.Marshal(rollbackTx)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback_tx to json")
			return err
		}

		buffer.Write(rollbackTxBytes)
	}
	rollbackTxsHash, err := crypto.Hash(buffer.Bytes())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing block rollback_txs")
		return err
	}
	b := &model.Block{
		ID:            blockID,
		Hash:          block.Header.Hash,
		Data:          block.BinData,
		EcosystemID:   block.Header.EcosystemID,
		KeyID:         block.Header.KeyID,
		NodePosition:  block.Header.NodePosition,
		Time:          block.Header.Time,
		RollbacksHash: rollbackTxsHash,
		Tx:            int32(len(block.Transactions)),
	}
	validBlockTime := true
	if blockID > 1 {
		exists, err := protocols.NewBlockTimeCounter().BlockForTimeExists(time.Unix(b.Time, 0), int(b.NodePosition))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("block validation")
			return err
		}

		validBlockTime = !exists
	}
	if validBlockTime {
		if err = b.Create(transaction); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating block")
			return err
		}
		if err = model.UpdRollbackHash(transaction, rollbackTxsHash); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating info block")
			return err
		}
	} else {
		err := fmt.Errorf("Invalid block time: %d", block.Header.Time)
		log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("invalid block time")
		return err
	}

	return nil
}

// GetBlockDataFromBlockChain is retrieving block data from blockchain
func GetBlockDataFromBlockChain(blockID int64) (*utils.BlockData, error) {
	BlockData := new(utils.BlockData)
	block := &model.Block{}
	_, err := block.Get(blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting block by ID")
		return BlockData, utils.ErrInfo(err)
	}

	header, err := utils.ParseBlockHeader(bytes.NewBuffer(block.Data), false)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	BlockData = &header
	BlockData.Hash = block.Hash
	BlockData.RollbacksHash = block.RollbacksHash
	return BlockData, nil
}

// GetDataFromFirstBlock returns data of first block
func GetDataFromFirstBlock() (data *consts.FirstBlock, ok bool) {
	block := &model.Block{}
	isFound, err := block.Get(1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting record of first block")
		return
	}

	if !isFound {
		return
	}

	pb, err := UnmarshallBlock(bytes.NewBuffer(block.Data), true, true)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParserError, "error": err}).Error("parsing data of first block")
		return
	}

	if len(pb.Transactions) == 0 {
		log.WithFields(log.Fields{"type": consts.ParserError}).Error("list of parsers is empty")
		return
	}

	t := pb.Transactions[0]
	data, ok = t.TxPtr.(*consts.FirstBlock)
	if !ok {
		log.WithFields(log.Fields{"type": consts.ParserError}).Error("getting data of first block")
		return
	}
	syspar.SysUpdate(nil)
	return
}
