package model

// TransactionStatus is model
type TransactionStatus struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Time     int64  `gorm:"not null;"`
	Type     int64  `gorm:"not null"`
	WalletID int64  `gorm:"not null"`
	BlockID  int64  `gorm:"not null"`
	Error    string `gorm:"not null;size 255"`
}

// TableName returns name of table
func (ts *TransactionStatus) TableName() string {
	return "transactions_status"
}

// Create is creating record of model
func (ts *TransactionStatus) Create() error {
	return DBConn.Create(ts).Error
}

// Get is retrieving model from database
func (ts *TransactionStatus) Get(transactionHash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", transactionHash).First(ts))
}

func (ts *TransactionStatus) UpdateBlockID(transaction *DbTransaction, newBlockID int64, transactionHash []byte) error {
	return GetDB(transaction).Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Update("block_id", newBlockID).Error
}

func (ts *TransactionStatus) UpdateBlockMsg(transaction *DbTransaction, newBlockID int64, msg string, transactionHash []byte) error {
	return GetDB(transaction).Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Updates(
		map[string]interface{}{"block_id": newBlockID, "error": msg}).Error
}

func (ts *TransactionStatus) SetError(errorText string, transactionHash []byte) error {
	return DBConn.Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Update("error", errorText).Error
}
