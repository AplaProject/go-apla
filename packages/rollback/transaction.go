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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

func rollbackUpdatedRow(tx map[string]string, where string, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	var rollbackInfo map[string]string
	if err := json.Unmarshal([]byte(tx["data"]), &rollbackInfo); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
		return err
	}
	addSQLUpdate := ""
	for k, v := range rollbackInfo {
		if v == "NULL" {
			addSQLUpdate += k + `=NULL,`
		} else if converter.IsByteColumn(tx["table_name"], k) && len(v) != 0 {
			addSQLUpdate += k + `=decode('` + string(converter.BinToHex([]byte(v))) + `','HEX'),`
		} else {
			addSQLUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
		}
	}
	addSQLUpdate = addSQLUpdate[0 : len(addSQLUpdate)-1]
	if err := model.Update(dbTransaction, tx["table_name"], addSQLUpdate, where); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "query": addSQLUpdate}).Error("updating table")
		return err
	}
	return nil
}

func rollbackInsertedRow(tx map[string]string, where string, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	if err := model.Delete(dbTransaction, tx["table_name"], where); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from table")
		return err
	}
	return nil
}

func rollbackTransaction(txHash []byte, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTransactions(dbTransaction, txHash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback transactions")
		return err
	}
	for _, tx := range txs {
		if tx["table_name"] == smart.SysName {
			var sysData smart.SysRollData
			err := json.Unmarshal([]byte(tx["data"]), &sysData)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
				return err
			}
			switch sysData.Type {
			case "NewTable":
				smart.SysRollbackTable(dbTransaction, sysData)
			case "NewColumn":
				smart.SysRollbackColumn(dbTransaction, sysData)
			case "NewContract":
				smart.SysRollbackNewContract(sysData, tx["table_id"])
			case "EditContract":
				smart.SysRollbackEditContract(dbTransaction, sysData, tx["table_id"])
			case "NewEcosystem":
				smart.SysRollbackEcosystem(dbTransaction, sysData)
			case "ActivateContract":
				smart.SysRollbackActivate(sysData)
			case "DeactivateContract":
				smart.SysRollbackDeactivate(sysData)
			case "DeleteColumn":
				smart.SysRollbackDeleteColumn(dbTransaction, sysData)
			case "DeleteTable":
				smart.SysRollbackDeleteTable(dbTransaction, sysData)
			}
			continue
		}
		where := " WHERE id='" + tx["table_id"] + `'`
		table := tx[`table_name`]
		if under := strings.IndexByte(table, '_'); under > 0 {
			keyName := table[under+1:]
			if v, ok := model.FirstEcosystemTables[keyName]; ok && !v {
				where += fmt.Sprintf(` AND ecosystem='%d'`, converter.StrToInt64(table[:under]))
				tx[`table_name`] = `1_` + keyName
			}
		}
		if len(tx["data"]) > 0 {
			if err := rollbackUpdatedRow(tx, where, dbTransaction, logger); err != nil {
				return err
			}
		} else {
			if err := rollbackInsertedRow(tx, where, dbTransaction, logger); err != nil {
				return err
			}
		}
	}
	txForDelete := &model.RollbackTx{TxHash: txHash}
	err = txForDelete.DeleteByHash(dbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback transaction by hash")
		return err
	}
	return nil
}
