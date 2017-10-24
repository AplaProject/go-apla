package model

import "strconv"

type UpdFullNode struct {
	ID   int64 `gorm:"primary_key;not null"`
	Time int64 `gorm:"not null"`
	RbID int64 `gorm:"not null"`
}

func (ufn *UpdFullNode) Read(transaction *DbTransaction) (bool, error) {
	return isFound(GetDB(transaction).First(ufn))
}

func (ufn *UpdFullNode) GetAll() ([]UpdFullNode, error) {
	result := make([]UpdFullNode, 0)
	err := DBConn.Find(&result).Error
	return result, err
}

func (ufn *UpdFullNode) ToMap() map[string]string {
	result := make(map[string]string)
	result["id"] = strconv.FormatInt(ufn.ID, 10)
	result["time"] = strconv.FormatInt(ufn.Time, 10)
	result["rb_id"] = strconv.FormatInt(ufn.RbID, 10)
	return result
}
