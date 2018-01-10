package querycost

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestTableRowCounter struct {
}

const tableRowCount = 10000

func (t *TestTableRowCounter) RowCount(tableName string) (int64, error) {
	if tableName == "small" {
		return tableRowCount, nil
	}
	return 0, errors.New("Unknown table")
}

type QueryCostByFormulaTestSuite struct {
	suite.Suite
	queryCoster QueryCoster
}

func (s *QueryCostByFormulaTestSuite) SetupTest() {
	s.queryCoster = &FormulaQueryCoster{&TestTableRowCounter{}}
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostUnknownQueryType() {
	_, err := s.queryCoster.QueryCost(nil, "UNSELECT * FROM name")
	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, UnknownQueryTypeError)
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromSelectNoTable() {
	tableName, err := getTableNameFromSelectQuery("select 3")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "")
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromSelect() {
	tableName, err := getTableNameFromSelectQuery("select a from keys where 3=1")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "keys")
	tableName, err = getTableNameFromSelectQuery(`select a,  b,  c from "1_keys" where 3=1`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "1_keys")
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromInsertNoInto() {
	_, err := getTableNameFromInsertQuery(`insert "1_keys"(id) values (1)`)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, IntoStatementMissingError)
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromInsert() {
	tableName, err := getTableNameFromInsertQuery("insert into keys(a,b,c) values (1,2,3)")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "keys")
	tableName, err = getTableNameFromInsertQuery(`insert into "1_keys" (a,b,c) values (1,2,3)`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "1_keys")
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromUpdateNoSet() {
	_, err := getTableNameFromUpdateQuery(`update keys a = b where id = 1`)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, SetStatementMissingError)
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromUpdate() {
	tableName, err := getTableNameFromUpdateQuery("update keys set a = 1 where id = 2")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "keys")
	tableName, err = getTableNameFromUpdateQuery(`update "1_keys" set a = 1`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "1_keys")
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromDeleteNoFrom() {
	_, err := getTableNameFromDeleteQuery(`delete table where id = 1`)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, FromStatementMissingError)
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromDeleteNoTable() {
	_, err := getTableNameFromDeleteQuery(`delete from`)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, DeleteMinimumThreeFieldsError)
}

func (s *QueryCostByFormulaTestSuite) TestGetTableNameFromDelete() {
	tableName, err := getTableNameFromDeleteQuery("delete from keys")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "keys")
	tableName, err = getTableNameFromDeleteQuery(`delete from "1_keys" where a = 1`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), tableName, "1_keys")
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostSelect() {
	cost, err := s.queryCoster.QueryCost(nil, "SELECT * FROM small WHERE id = ?", 3)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), cost, calcSelectCost(tableRowCount))
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostUpdate() {
	cost, err := s.queryCoster.QueryCost(nil, "UPDATE small SET a = ?", 3)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), cost, calcUpdateCost(tableRowCount))
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostDelete() {
	cost, err := s.queryCoster.QueryCost(nil, "DELETE FROM small")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), cost, calcDeleteCost(tableRowCount))
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostInsert() {
	cost, err := s.queryCoster.QueryCost(nil, "INSERT INTO small(a,b) VALUES (1,2)")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), cost, calcInsertCost(tableRowCount))
}

func (s *QueryCostByFormulaTestSuite) TestQueryCostInsertWrongTable() {
	_, err := s.queryCoster.QueryCost(nil, "INSERT INTO unknown(a,b) VALUES (1,2)")
	assert.Error(s.T(), err)
}

func TestQueryCostFormula(t *testing.T) {
	suite.Run(t, new(QueryCostByFormulaTestSuite))
}
