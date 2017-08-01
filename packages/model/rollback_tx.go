package model

type RollbackTx struct {
	ID        int64  `gorm:"primary_key;not null"`
	BlockID   int64  `gorm:"not null"`
	TxHash    []byte `gorm:"not null"`
	TableName string `gorm:"not null;size:255"`
	TableID   string `gorm:"not null;size:255"`
}

func (rt *RollbackTx) GetRollbackTransactions(transactionHash []byte) ([]RollbackTx, error) {
	transactions := new([]RollbackTx)
	err := DBConn.Where("tx_hash = ", transactionHash).Order("id desc").Find(transactions).Error
	if err != nil {
		return nil, err
	}
	return *transactions, err
}

func (rt *RollbackTx) DeleteByHash() error {
	return DBConn.Where("tx_hash = ?", rt.TxHash).Delete(rt).Error
}

func (rt *RollbackTx) DeleteByHashAndTableName() error {
	return DBConn.Where("tx_hash = ? and table_name = ?", rt.TxHash, rt.TableName).Delete(rt).Error
}

func (rt *RollbackTx) Create() error {
	return DBConn.Create(rt).Error
}

func (rt *RollbackTx) Get(transactionHash []byte, tableName string) error {
	return DBConn.Where("tx_hash = ? AND table_name = ?", transactionHash, tableName).First(rt).Error
}
