package querycost

import (
	"errors"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

const (
	Select = "select"
	Insert = "insert"
	Update = "update"
	Delete = "delete"

	Set  = "set"
	From = "from"
	Into = "into"

	Quote  = `"`
	Lparen = "("
)

const (
	SelectCost = 1
	UpdateCost = 1
	InsertCost = 1
	DeleteCost = 1

	SelectRowCoeff = 0.0001
	InsertRowCoeff = 0.0001
	DeleteRowCoeff = 0.0001
	UpdateRowCoeff = 0.0001
)

var FromStatementMissingError = errors.New("FROM statement missing")
var DeleteMinimumThreeFieldsError = errors.New("DELETE query must consist minimum of 3 fields")
var SetStatementMissingError = errors.New("SET statement missing")
var IntoStatementMissingError = errors.New("INTO statement missing")
var UnknownQueryTypeError = errors.New("Unknown query type")

type TableRowCounter interface {
	RowCount(string) (int64, error)
}

type DBCountQueryRowCounter struct {
}

func (d *DBCountQueryRowCounter) RowCount(tableName string) (int64, error) {
	count, err := model.GetRecordsCount(tableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": tableName}).Error("Getting record count from table")
	}
	return count, err
}

type FormulaQueryCoster struct {
	rowCounter TableRowCounter
}

func strSliceIndex(fields []string, fieldToFind string) (index int) {
	for i, field := range fields {
		if field == fieldToFind {
			index = i
			break
		}
	}
	return
}

func calcSelectCost(rowCount int64) int64 {
	return SelectCost + int64(SelectRowCoeff*float64(rowCount))
}

func calcUpdateCost(rowCount int64) int64 {
	return UpdateCost + int64(UpdateRowCoeff*float64(rowCount))
}

func calcDeleteCost(rowCount int64) int64 {
	return DeleteCost + int64(DeleteRowCoeff*float64(rowCount))
}

func calcInsertCost(rowCount int64) int64 {
	return InsertCost
}

func (f *FormulaQueryCoster) QueryCost(transaction *model.DbTransaction, query string, args ...interface{}) (int64, error) {
	cleanedQuery := strings.TrimSpace(strings.ToLower(query))
	switch {
	case strings.HasPrefix(cleanedQuery, Select):
		return selectQueryCost(cleanedQuery, f.rowCounter)
	case strings.HasPrefix(cleanedQuery, Insert):
		return insertQueryCost(cleanedQuery, f.rowCounter)
	case strings.HasPrefix(cleanedQuery, Update):
		return updateQueryCost(cleanedQuery, f.rowCounter)
	case strings.HasPrefix(cleanedQuery, Delete):
		return deleteQueryCost(cleanedQuery, f.rowCounter)
	}
	log.WithFields(log.Fields{"type": consts.ParseError, "query": query}).Error("parsing sql query")
	return 0, UnknownQueryTypeError
}

func selectQueryCost(query string, tableRowCounter TableRowCounter) (int64, error) {
	tableName, err := getTableNameFromSelectQuery(query)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "query": query, "error": err}).Error("getting table name from sql query")
		return 0, err
	}
	rowCount, err := tableRowCounter.RowCount(tableName)
	if err != nil {
		return 0, err
	}
	return calcSelectCost(rowCount), nil
}

func insertQueryCost(query string, tableRowCounter TableRowCounter) (int64, error) {
	tableName, err := getTableNameFromInsertQuery(query)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "query": query, "error": err}).Error("getting table name from sql query")
		return 0, err
	}
	rowCount, err := tableRowCounter.RowCount(tableName)
	if err != nil {
		return 0, err
	}
	return calcInsertCost(rowCount), nil
}

func updateQueryCost(query string, tableRowCounter TableRowCounter) (int64, error) {
	tableName, err := getTableNameFromUpdateQuery(query)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "query": query, "error": err}).Error("getting table name from sql query")
		return 0, err
	}
	rowCount, err := tableRowCounter.RowCount(tableName)
	if err != nil {
		return 0, err
	}
	return calcUpdateCost(rowCount), nil
}

func deleteQueryCost(query string, tableRowCounter TableRowCounter) (int64, error) {
	tableName, err := getTableNameFromDeleteQuery(query)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "query": query, "error": err}).Error("getting table name from sql query")
		return 0, err
	}
	rowCount, err := tableRowCounter.RowCount(tableName)
	if err != nil {
		return 0, err
	}
	return calcDeleteCost(rowCount), nil
}

func getTableNameFromSelectQuery(query string) (string, error) {
	queryFields := strings.Fields(query)
	fromFieldIndex := strSliceIndex(queryFields, From)
	if fromFieldIndex == 0 {
		return "", nil
	}
	return strings.Trim(queryFields[fromFieldIndex+1], Quote), nil
}

func getTableNameFromInsertQuery(query string) (string, error) {
	queryFields := strings.Fields(query)
	intoFieldIndex := strSliceIndex(queryFields, Into)
	if intoFieldIndex == 0 {
		return "", IntoStatementMissingError
	}
	tableNameValuesField := queryFields[intoFieldIndex+1]
	tableName := ""
	lparenIndex := strings.Index(tableNameValuesField, Lparen)
	if lparenIndex > 0 {
		tableName = tableNameValuesField[:lparenIndex]
	} else {
		tableName = tableNameValuesField
	}
	return strings.Trim(tableName, Quote), nil
}

func getTableNameFromUpdateQuery(query string) (string, error) {
	queryFields := strings.Fields(query)
	setFieldIndex := strSliceIndex(queryFields, Set)
	if setFieldIndex == 0 {
		return "", SetStatementMissingError
	}
	return strings.Trim(queryFields[setFieldIndex-1], Quote), nil
}

func getTableNameFromDeleteQuery(query string) (string, error) {
	queryFields := strings.Fields(query)
	fromFieldIndex := strSliceIndex(queryFields, From)
	if fromFieldIndex == 0 {
		return "", FromStatementMissingError
	}
	// DELETE FROM table is minimum
	if len(queryFields) < 3 {
		return "", DeleteMinimumThreeFieldsError
	}
	return strings.Trim(queryFields[fromFieldIndex+1], Quote), nil
}
