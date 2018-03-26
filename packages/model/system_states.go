package model

import (
	"fmt"
)

const ecosysTable = "1_ecosystems"

// Ecosystem is model
type Ecosystem struct {
	ID       int64 `gorm:"primary_key;not null"`
	Name     string
	IsValued bool
}

// TableName returns name of table
// only first ecosystem has this entity
func (sys *Ecosystem) TableName() string {
	return ecosysTable
}

// GetAllSystemStatesIDs is retrieving all ecosystems ids
func GetAllSystemStatesIDs() ([]int64, error) {
	if !IsTable(ecosysTable) {
		return nil, fmt.Errorf("%s does not exists", ecosysTable)
	}

	ecosystems := new([]Ecosystem)
	if err := DBConn.Find(&ecosystems).Order("id").Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(*ecosystems))
	for _, s := range *ecosystems {
		ids = append(ids, s.ID)
	}

	return ids, nil
}

// Get is fill reciever from db
func (sys *Ecosystem) Get(id int64) error {
	return DBConn.First(sys, "id = ?", id).Error
}

// Delete is deleting record
func (sys *Ecosystem) Delete(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(sys).Error
}
