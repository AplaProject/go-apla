// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package model

// RollbackTx is model
type RollbackTx struct {
	ID        int64  `gorm:"primary_key;not null" json:"-"`
	BlockID   int64  `gorm:"not null" json:"block_id"`
	TxHash    []byte `gorm:"not null" json:"tx_hash"`
	NameTable string `gorm:"not null;size:255;column:table_name" json:"table_name"`
	TableID   string `gorm:"not null;size:255" json:"table_id"`
	Data      string `gorm:"not null;type:jsonb(PostgreSQL)" json:"data"`
}

// TableName returns name of table
func (RollbackTx) TableName() string {
	return "rollback_tx"
}

// GetRollbackTransactions is returns rollback transactions
func (rt *RollbackTx) GetRollbackTransactions(dbTransaction *DbTransaction, transactionHash []byte) ([]map[string]string, error) {
	return GetAllTx(dbTransaction, "SELECT * from rollback_tx WHERE tx_hash = ?", -1, transactionHash)
}

func (rt *RollbackTx) GetBlockRollbackTransactions(dbTransaction *DbTransaction, blockID int64) ([]RollbackTx, error) {
	var rollbackTransactions []RollbackTx
	err := GetDB(dbTransaction).Where("block_id = ?", blockID).Order("tx_hash asc").Find(&rollbackTransactions).Error
	return rollbackTransactions, err
}

func (rt *RollbackTx) GetRollbackTxsByTableIDAndTableName(tableID, tableName string, limit int) (*[]RollbackTx, error) {
	rollbackTx := new([]RollbackTx)
	if err := DBConn.Where("table_id = ? AND table_name = ?", tableID, tableName).Limit(limit).Find(rollbackTx).Error; err != nil {
		return nil, err
	}
	return rollbackTx, nil
}

// DeleteByHash is deleting rollbackTx by hash
func (rt *RollbackTx) DeleteByHash(dbTransaction *DbTransaction) error {
	return GetDB(dbTransaction).Exec("DELETE FROM rollback_tx WHERE tx_hash = ?", rt.TxHash).Error
}

// DeleteByHashAndTableName is deleting tx by hash and table name
func (rt *RollbackTx) DeleteByHashAndTableName(transaction *DbTransaction) error {
	return GetDB(transaction).Where("tx_hash = ? and table_name = ?", rt.TxHash, rt.NameTable).Delete(rt).Error
}

// Create is creating record of model
func (rt *RollbackTx) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(rt).Error
}

// Get is retrieving model from database
func (rt *RollbackTx) Get(dbTransaction *DbTransaction, transactionHash []byte, tableName string) (bool, error) {
	return isFound(GetDB(dbTransaction).Where("tx_hash = ? AND table_name = ?", transactionHash, tableName).First(rt))
}
