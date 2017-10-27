package model

type UpdFullNode struct {
	ID   int64 `gorm:"primary_key;not null"`
	Time int64 `gorm:"not null"`
	RbID int64 `gorm:"not null"`
}

func (ufn *UpdFullNode) Get(transaction *DbTransaction) (bool, error) {
	return isFound(GetDB(transaction).First(ufn))
}
