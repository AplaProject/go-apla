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

package model

import (
	"fmt"
)

// QueueTx is model
type QueueTx struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Data     []byte `gorm:"not null"`
	FromGate int    `gorm:"not null"`
}

// TableName returns name of table
func (qt *QueueTx) TableName() string {
	return "queue_tx"
}

// DeleteTx is deleting tx
func (qt *QueueTx) DeleteTx(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(qt).Error
}

// Save is saving model
func (qt *QueueTx) Save(transaction *DbTransaction) error {
	return GetDB(transaction).Save(qt).Error
}

// Create is creating record of model
func (qt *QueueTx) Create() error {
	return DBConn.Create(qt).Error
}

// GetByHash is retrieving model from database by hash
func (qt *QueueTx) GetByHash(transaction *DbTransaction, hash []byte) (bool, error) {
	return isFound(GetDB(transaction).Where("hash = ?", hash).First(qt))
}

// DeleteQueueTxByHash is deleting queue tx by hash
func DeleteQueueTxByHash(transaction *DbTransaction, hash []byte) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM queue_tx WHERE hash = ?", hash)
	return query.RowsAffected, query.Error
}

// GetQueuedTransactionsCount counting queued transactions
func GetQueuedTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	err := DBConn.Table("queue_tx").Where("hash = ?", hash).Count(&rowsCount).Error
	return rowsCount, err
}

// GetAllUnverifiedAndUnusedTransactions is returns all unverified and unused transaction
func GetAllUnverifiedAndUnusedTransactions() ([]*QueueTx, error) {
	query := `SELECT *
		  FROM (
	              SELECT data,
	                     hash
	              FROM queue_tx
		      UNION
		      SELECT data,
			     hash
		      FROM transactions
		      WHERE verified = 0 AND used = 0
			)  AS x`
	rows, err := DBConn.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data, hash []byte
	result := []*QueueTx{}
	for rows.Next() {
		if err := rows.Scan(&data, &hash); err != nil {
			return nil, err
		}
		result = append(result, &QueueTx{Data: data, Hash: hash})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// FieldValue implementing BatchModel interface
func (qt QueueTx) FieldValue(fieldName string) (interface{}, error) {
	switch fieldName {
	case "hash":
		return qt.Hash, nil
	case "data":
		return qt.Data, nil
	case "from_gate":
		return qt.FromGate, nil
	default:
		return nil, fmt.Errorf("Unknown field '%s' for QueueTx", fieldName)
	}
}
