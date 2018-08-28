package model

// RollbackTx is model
type RollbackTx struct {
	ID        int64  `gorm:"primary_key;not null" json:"-"`
	BlockID   int64  `gorm:"not null" json:"block_id"`
	BlockHash []byte `gorm:"not null" json:"block_hash"`
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
	return GetAllTx(dbTransaction, "SELECT * from rollback_tx WHERE tx_hash = ? ORDER BY ID DESC", -1, transactionHash)
}

// GetBlockRollbackTransactions returns records of rollback by blockID
func (rt *RollbackTx) GetBlockRollbackTransactions(dbTransaction *DbTransaction, blockID int64) ([]RollbackTx, error) {
	var rollbackTransactions []RollbackTx
	err := GetDB(dbTransaction).Where("block_id = ?", blockID).Order("tx_hash asc").Find(&rollbackTransactions).Error
	return rollbackTransactions, err
}

// GetRollbackTxsByTableIDAndTableName returns records of rollback by table name and id
func (rt *RollbackTx) GetRollbackTxsByTableIDAndTableName(tableID, tableName string, limit int) (*[]RollbackTx, error) {
	rollbackTx := new([]RollbackTx)
	if err := DBConn.Where("table_id = ? AND table_name = ?", tableID, tableName).
		Order("id desc").Limit(limit).Find(rollbackTx).Error; err != nil {
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
	var err error
	if rt.ID, err = GetNextID(transaction, (*rt).TableName()); err != nil {
		return err
	}
	return GetDB(transaction).Create(rt).Error
}

// Get is retrieving model from database
func (rt *RollbackTx) Get(dbTransaction *DbTransaction, transactionHash []byte, tableName string) (bool, error) {
	return isFound(GetDB(dbTransaction).Where("tx_hash = ? AND table_name = ?", transactionHash, tableName).First(rt))
}
