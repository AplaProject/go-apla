package model

type TransactionStatus struct {
	Hash      []byte `gorm:"primary_key;not null"`
	Time      int64  `gorm:"not null;"`
	Type      int64  `gorm:"not null"`
	WalletID  int64  `gorm:"not null"`
	CitizenID int64  `gorm:"not null"`
	BlockID   int64  `gorm:"not null"`
	Error     string `gorm:"not null;size 255"`
}

func (ts *TransactionStatus) TableName() string {
	return "transactions_status"
}

func (ts *TransactionStatus) Create() error {
	return DBConn.Create(ts).Error
}

func (ts *TransactionStatus) Get(transactionHash []byte) (bool, error) {
	query := DBConn.Where("hash = ?", transactionHash).First(ts)
	return query.RecordNotFound(), query.Error
}

func (ts *TransactionStatus) UpdateBlockID(transaction *DbTransaction, newBlockID int64, transactionHash []byte) error {
	return getDB(transaction).Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Update("block_id", newBlockID).Error
}

func (ts *TransactionStatus) SetError(errorText string, transactionHash []byte) error {
	return DBConn.Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Update("error", errorText).Error
}
