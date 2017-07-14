package model

type UpdFullNodes struct {
	ID   int64 `gorm:"primary_key;not null"`
	Time int32 `gorm:"not null"`
	RbID int64 `gorm: "not null"`
}

func (ufn *UpdFullNodes) Read() error {
	return DBConn.First(ufn).Error
}
