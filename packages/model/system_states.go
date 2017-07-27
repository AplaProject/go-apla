package model

type SystemStates struct {
	ID   int64 `gorm:"primary_key;not null"`
	RbID int64 `gorm:"not null"`
}

func GetAllSystemStatesIDs() ([]int64, error) {
	IDs := new([]int64)
	if err := DBConn.Model(&SystemStates{}).Find(IDs).Error; err != nil {
		return nil, err
	}
	return *IDs, nil
}

func (ss *SystemStates) GetLast() error {
	return DBConn.Last(ss).Error
}

func (ss *SystemStates) Delete() error {
	return DBConn.Delete(ss).Error
}

func (ss *SystemStates) Create() error {
	return DBConn.Create(ss).Error
}
