package registry

import (
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataTx_Fill(t *testing.T) {
	converter := converter{}
	value, err := converter.createFromParams(model.KeySchema{}.ModelName(), map[string]interface{}{
		"id":        1,
		"publickey": []byte{1, 2, 3},
		"amount":    "10",
		"blocked":   true,
	})

	require.Nil(t, err)
	assert.Equal(t, value, &model.KeySchema{
		ID:        1,
		PublicKey: []byte{1, 2, 3},
		Amount:    "10",
		Blocked:   true,
	})
}

func TestMetadataTx_FillUpdate(t *testing.T) {
	converter := converter{}

	key := model.KeySchema{Amount: "999"}
	require.Error(t, converter.updateFromParams(model.KeySchema{}.ModelName(), &key, map[string]interface{}{}))

	err := converter.updateFromParams(model.KeySchema{}.ModelName(), &key, map[string]interface{}{
		"id":        1,
		"publickey": []byte{1, 2, 3},
		"blocked":   true,
	})

	require.Nil(t, err)
	assert.Equal(t, key, model.KeySchema{
		ID:        1,
		PublicKey: []byte{1, 2, 3},
		Amount:    "999",
		Blocked:   true,
	})

	err = converter.updateFromParams(model.KeySchema{}.ModelName(), &key, map[string]interface{}{
		"id":        2,
		"publickey": []byte{3, 2, 1},
		"amount":    "666",
		"blocked":   false,
	})
	require.Nil(t, err)
	assert.Equal(t, key, model.KeySchema{
		ID:        2,
		PublicKey: []byte{3, 2, 1},
		Amount:    "666",
		Blocked:   false,
	})
}
