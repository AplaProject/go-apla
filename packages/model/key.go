package model

import (
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/tidwall/gjson"
)

// Key is model
// TODO rename to Key
type KeySchema struct {
	ID          int64
	PublicKey   []byte
	Amount      string
	Deleted     int64
	Blocked     int64
	EcosystemID int64
}

func (ks KeySchema) Name() string {
	return "key"
}

func (ks KeySchema) GetIndexes() []types.Index {
	registry := &types.Registry{Name: ks.Name()}
	return []types.Index{
		{
			Name:     "amount",
			Registry: registry,
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
			},
		},
	}
}
