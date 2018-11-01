package model

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
func GetAllSystemStatesIDs() ([]int64, []string, error) {
	if !IsTable(ecosysTable) {
		//return nil, fmt.Errorf("%s does not exists", ecosysTable)
		return nil, nil, nil
	}

	ecosystems := new([]Ecosystem)
	if err := DBConn.Find(&ecosystems).Order("id").Error; err != nil {
		return nil, nil, err
	}

	ids := make([]int64, len(*ecosystems))
	names := make([]string, len(*ecosystems))
	for i, s := range *ecosystems {
		ids[i] = s.ID
		names[i] = s.Name
	}

	return ids, names, nil
}

// Get is fill reciever from db
func (sys *Ecosystem) Get(id int64) (bool, error) {
	return isFound(DBConn.First(sys, "id = ?", id))
}

// Delete is deleting record
func (sys *Ecosystem) Delete(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(sys).Error
}
