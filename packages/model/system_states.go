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
	return DBConn.First(ss).Error
}

func (ss *SystemStates) Delete() error {
	return DBConn.Delete(ss).Error
}

func (ss *SystemStates) IsExists(stateID int64) (bool, error) {
	query := DBConn.Where("id = ?", stateID).First(ss)
	return !query.RecordNotFound(), query.Error
}
