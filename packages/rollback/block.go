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

package rollback

import (
	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// BlockRollback is blocking rollback
func RollbackBlock(blockModel *blockchain.Block, hash []byte) error {
	ldbTx, err := blockchain.DB.OpenTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("starting transaction")
		return err
	}
	txs, err := blockModel.Transactions(ldbTx)
	if err != nil {
		return err
	}
	b, err := block.FromBlockchainBlock(blockModel, txs, hash, ldbTx)
	if err != nil {
		return err
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
		return err
	}

	// metadb := model.MetadataRegistry.Begin(ldbTx)
	//err = rollbackBlock(dbTransaction, ldbTx, metadb, b)
	err = rollbackBlock(dbTransaction, ldbTx, b)
	if err != nil {
		dbTransaction.Rollback()
		ldbTx.Discard()
		// metadb.Rollback()
		return err
	}

	err = dbTransaction.Commit()
	err = ldbTx.Commit()
	return err
}

func rollbackBlock(dbTransaction *model.DbTransaction, ldbTx *leveldb.Transaction, block *block.PlayableBlock) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		t := block.Transactions[i]
		t.DbTransaction = dbTransaction
		t.LdbTx = ldbTx

		if t.TxContract != nil {
			if err := rollbackTransaction(t.TxHash, t.DbTransaction, logger); err != nil {
				return err
			}
		}
	}

	//err := metadb.RollbackBlock(block.Hash)
	//if err != nil {
	//	return err
	//}

	return nil
}
