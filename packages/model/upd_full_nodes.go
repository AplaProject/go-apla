package model

type UpdFullNodes struct {
	ID   int64 `gorm:"primary_key;not null"`
	Time int32 `gorm:"not null"`
	RbID int64 `gorm: "not null"`
}

func (ufn *UpdFullNodes) Read() error {
	return DBConn.First(ufn).Error
}

func (ufn *UpdFullNodes) GetAll() ([]UpdFullNodes, error) {
	var result []UpdFullNodes
	err := DBConn.Find(result).Error
	return result, err
}

func (ufn *UpdFullNodes) ToMap() map[string]string {
	result := make(map[string]string)
	result["id"] = string(ufn.ID)
	result["time"] = string(ufn.Time)
	result["rb_id"] = string(ufn.RbID)
	return result
}
