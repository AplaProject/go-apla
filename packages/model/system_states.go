package model

type SystemState struct {
	ID   int64 `gorm:"primary_key;not null"`
	RbID int64 `gorm:"not null"`
}

func (ss *SystemState) TableName() string {
	return "system_states"
}

func GetAllSystemStatesIDs() ([]int64, error) {
	IDs := new([]int64)
	if err := DBConn.Model(&SystemState{}).Find(IDs).Error; err != nil {
		return nil, err
	}
	return *IDs, nil
}

func (ss *SystemState) Get(id int64) error {
	return DBConn.Where("id = ?", id).First(ss).Error
}

func (ss *SystemState) GetCount() (int64, error) {
	var count int64
	err := DBConn.Table("system_states").Count(count).Error
	return count, err
}

func (ss *SystemState) GetLast() error {
	return DBConn.First(ss).Error
}

func (ss *SystemState) Delete() error {
	return DBConn.Delete(ss).Error
}

func (ss *SystemState) IsExists(stateID int64) (bool, error) {
	query := DBConn.Where("id = ?", stateID).First(ss)
	return !query.RecordNotFound(), query.Error
}
