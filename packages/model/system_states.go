package model

// Ecosystem is model
type Ecosystem struct {
	ID       int64 `gorm:"primary_key;not null"`
	Name     string
	IsValued bool
}

// TableName returns name of table
// only first ecosystem has this entity
func (sys *Ecosystem) TableName() string {
	return "1_ecosystems"
}

// GetAllSystemStatesIDs is retrieving all system states ids
func GetAllSystemStatesIDs() ([]int64, error) {
	if !IsTable("1_ecosystems") {
		return nil, nil
	}

	states := new([]Ecosystem)
	if err := DBConn.Find(&states).Order("id").Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(*states))
	for _, s := range *states {
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
