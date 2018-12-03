// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
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
	"bytes"
	"fmt"

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// BlockRollback is blocking rollback
func RollbackBlock(data []byte, deleteBlock bool) error {
	buf := bytes.NewBuffer(data)
	if buf.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty buffer")
		return fmt.Errorf("empty buffer")
	}

	block, err := block.UnmarshallBlock(buf, false, true)
	if err != nil {
		return err
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
		return err
	}

	err = rollbackBlock(dbTransaction, block)

	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	if deleteBlock {
		b := &model.Block{}
		err = b.DeleteById(dbTransaction, block.Header.BlockID)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
			dbTransaction.Rollback()
			return err
		}
	}

	err = dbTransaction.Commit()
	return err
}

func rollbackBlock(dbTransaction *model.DbTransaction, block *block.Block) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		t := block.Transactions[i]
		t.DbTransaction = dbTransaction

		_, err := model.MarkTransactionUnusedAndUnverified(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
			return err
		}
		_, err = model.DeleteLogTransactionsByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting log transactions by hash")
			return err
		}

		ts := &model.TransactionStatus{}
		err = ts.UpdateBlockID(dbTransaction, 0, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating block id in transaction status")
			return err
		}

		_, err = model.DeleteQueueTxByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transacion from queue by hash")
			return err
		}

		if t.TxContract != nil {
			if err = rollbackTransaction(t.TxHash, t.DbTransaction, logger); err != nil {
				return err
			}
		} else {
			MethodName := consts.TxTypes[t.TxType]
			txParser, err := transaction.GetTransaction(t, MethodName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			result := txParser.Init()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
			result = txParser.Rollback()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
		}
	}

	return nil
}
