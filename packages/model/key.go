package model

import (
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
)

// Key is model
// TODO rename to Key
type KeySchema struct {
	ID        int64  `json:"id"`
	PublicKey []byte `json:"public_key"`
	Amount    string `json:"amount"`
	Deleted   bool   `json:"deleted"`
	Blocked   bool   `json:"blocked"`
}

func (ks KeySchema) ModelName() string {
	return "keys"
}

func (ks KeySchema) GetIndexes() []types.Index {
	registry := &types.Registry{Name: ks.ModelName()}
	return []types.Index{
		{
			Field:    "amount",
			Registry: registry,
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
			},
		},
	}
}

func (ks KeySchema) CreateFromData(data map[string]interface{}) (types.RegistryModel, error) {
	k := &KeySchema{}
	err := mapstructure.Decode(data, &k)
	return k, err
}

func (ks KeySchema) GetPrimaryKey() string {
	return strconv.FormatInt(ks.ID, 10)
}
