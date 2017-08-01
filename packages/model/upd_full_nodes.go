package model

type UpdFullNode struct {
	ID   int64 `gorm:"primary_key;not null"`
	Time int64 `gorm:"not null"`
	RbID int64 `gorm: "not null"`
}

func (ufn *UpdFullNode) Read() error {
	return DBConn.First(ufn).Error
}

func (ufn *UpdFullNode) GetAll() ([]UpdFullNode, error) {
	var result []UpdFullNode
	err := DBConn.Find(result).Error
	return result, err
}

func (ufn *UpdFullNode) ToMap() map[string]string {
	result := make(map[string]string)
	result["id"] = string(ufn.ID)
	result["time"] = string(ufn.Time)
	result["rb_id"] = string(ufn.RbID)
	return result
}
