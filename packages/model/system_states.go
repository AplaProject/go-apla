package model

import (
	"encoding/json"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
)

const ecosysTable = "1_ecosystems"

// Ecosystem is model
type Ecosystem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsValued bool   `json:"is_valued"`
}

// TableName returns name of table
// only first ecosystem has this entity
// TODO REMOVE
func (sys *Ecosystem) TableName() string {
	return "ecosystems"
}

func (sys Ecosystem) ModelName() string {
	return "ecosystems"
}

func (sys Ecosystem) GetPrimaryKey() string {
	return strconv.FormatInt(sys.ID, 10)
}

func (sys Ecosystem) CreateFromData(data map[string]interface{}) (types.RegistryModel, error) {
	k := &Ecosystem{}
	err := mapstructure.Decode(data, &k)
	return k, err
}

func (sys Ecosystem) UpdateFromData(model types.RegistryModel, data map[string]interface{}) error {
	oldStruct := model.(*Ecosystem)
	return mapstructure.Decode(data, oldStruct)
}

func (ks Ecosystem) GetData() map[string]interface{} {
	return structs.Map(ks)
}

func (sys Ecosystem) GetIndexes() []types.Index {
	return []types.Index{
		{
			Name:     "name",
			Registry: &types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}
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

func (sys *Ecosystem) UnmarshalJSON(b []byte) error {
	type schema *Ecosystem
	err := json.Unmarshal(b, schema(sys))
	return err
}
