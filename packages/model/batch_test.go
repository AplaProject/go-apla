package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestBatchModel struct {
	ID   int64
	Name string
}

func (m TestBatchModel) TableName() string {
	return "test_batch"
}

func (m TestBatchModel) FieldValue(fieldName string) (interface{}, error) {
	switch fieldName {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	default:
		return nil, fmt.Errorf("Unknown field %s of TestBatchModel", fieldName)
	}
}

func TestPrepareQuery(t *testing.T) {
	slice := []BatchModel{
		TestBatchModel{ID: 1, Name: "first"},
		TestBatchModel{ID: 2, Name: "second"},
	}

	query, args, err := prepareQuery(slice, []string{"id", "name"})
	require.NoError(t, err)

	checkQuery := `INSERT INTO "test_batch" (id,name) VALUES (?,?),(?,?)`
	checkArgs := []interface{}{1, "first", 2, "second"}

	require.Equal(t, checkQuery, query)
	require.Equal(t, checkArgs, args)
}
