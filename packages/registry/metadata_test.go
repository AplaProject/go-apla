package registry_test

import (
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/registry"
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
		err     bool
	}{
		{
			testname: "insert-good",
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

			err: false,
		},

		{
			testname: "insert-bad-1",
			registry: types.Registry{
				Name:      "key",
				Ecosystem: &types.Ecosystem{ID: 1},
				Type:      types.RegistryTypeMetadata,
			},
			pkValue: "1",
			value:   make(chan int),

			err: true,
		},
	}

	for _, c := range cases {
		bunt, err := buntdb.Open(":memory:")
		require.Nil(t, err, c.testname)

		reg := registry.NewMetadataProvider(bunt)
		metadataTx, err := reg.Begin(true)
		require.Nil(t, err, c.testname)

		err = metadataTx.Insert(&c.registry, c.pkValue, c.value)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.Nil(t, err)

		saved := testModel{}
		err = metadataTx.Get(&c.registry, c.pkValue, &saved)
		require.Nil(t, err)

		assert.Equal(t, c.value, saved, c.testname)
	}
}
