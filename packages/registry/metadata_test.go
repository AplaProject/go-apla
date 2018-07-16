package registry_test

import (
	"testing"

	registry2 "github.com/GenesisKernel/go-genesis/packages/registry"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/buntdb"
)

type testModel struct {
	Id     int
	Field  string
	Field2 []byte
}

func TestMetadataTx_Insert(t *testing.T) {
	cases := []struct {
		testname string

		registry types.Registry
		pkValue  string
		value    interface{}

		expJson string
		err     error
	}{
		{
			testname: "insert",
			registry: types.Registry{
				Name:      "key",
				Ecosystem: &types.Ecosystem{ID: 1},
				Type:      types.RegistryTypeMetadata,
			},
			pkValue: "1",
			value: testModel{
				Id:     1,
				Field:  "testfield",
				Field2: make([]byte, 10),
			},

			err: nil,
		},
	}

	for _, c := range cases {
		bunt, err := buntdb.Open(":memory:")
		require.Nil(t, err)

		registry := registry2.NewMetadataProvider(bunt)
		metadataTx, err := registry.Begin(true)
		require.Nil(t, err)

		require.Nil(t, metadataTx.Insert(&c.registry, c.pkValue, c.value))

		saved := testModel{}
		require.Nil(t, metadataTx.Get(&c.registry, c.pkValue, &saved))

		assert.Equal(t, c.value, saved)
	}
}
