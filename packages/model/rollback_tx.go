package model

// RollbackTx is model
type RollbackTx struct {
	ID        int64  `gorm:"primary_key;not null"`
	BlockID   int64  `gorm:"not null"`
	TxHash    []byte `gorm:"not null"`
	NameTable string `gorm:"not null;size:255;column:table_name"`
	TableID   string `gorm:"not null;size:255"`
}

// TableName returns name of table
func (RollbackTx) TableName() string {
	return "rollback_tx"
}

// GetRollbackTransactions is returns rollback transactions
func (rt *RollbackTx) GetRollbackTransactions(dbTransaction *DbTransaction, transactionHash []byte) ([]map[string]string, error) {
	return GetAllTx(dbTransaction, "SELECT * from rollback_tx WHERE tx_hash = ?", -1, transactionHash)
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
func (rt *RollbackTx) Get(dbTransaction *DbTransaction, transactionHash []byte, tableName string) error {
	return GetDB(dbTransaction).Where("tx_hash = ? AND table_name = ?", transactionHash, tableName).First(rt).Error
}
