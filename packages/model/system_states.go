package model

import "github.com/jinzhu/gorm"

type SystemState struct {
	ID   int64 `gorm:"primary_key;not null"`
	RbID int64 `gorm:"not null"`
}

func (ss *SystemState) TableName() string {
	return "system_states"
}

func GetAllSystemStatesIDs() ([]int64, error) {
	states := new([]SystemState)
	if err := DBConn.Find(&states).Order("id").Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(*states))
	for _, s := range *states {
		ids = append(ids, s.ID)
	}
	return ids, nil
}

func (ss *SystemState) Get(id int64) error {
	return DBConn.Where("id = ?", id).First(ss).Error
}

func (ss *SystemState) GetCount() (int64, error) {
	count := int64(0)
	err := DBConn.Table("system_states").Count(&count).Error
	return count, err
}

func (ss *SystemState) GetLast(transaction *DbTransaction) (bool, error) {
	last := getDB(transaction).Last(ss)
	if last.RecordNotFound() {
		return true, nil
	}
	return false, last.Error
}

func (ss *SystemState) Delete(transaction *DbTransaction) error {
	return getDB(transaction).Delete(ss).Error
}

func (ss *SystemState) IsExists(stateID int64) (bool, error) {
	query := DBConn.Where("id = ?", stateID).First(ss)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return !query.RecordNotFound(), query.Error
}

func (ss *SystemState) Create(transaction *DbTransaction) error {
	return getDB(transaction).Create(ss).Error
}
