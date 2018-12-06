package querycost

import (
	"github.com/AplaProject/go-apla/packages/model"
)

type QueryCosterType int

const (
	ExplainQueryCosterType        QueryCosterType = iota
	ExplainAnalyzeQueryCosterType QueryCosterType = iota
	FormulaQueryCosterType        QueryCosterType = iota
)

type QueryCoster interface {
	QueryCost(*model.DbTransaction, string, ...interface{}) (int64, error)
}

type ExplainQueryCoster struct {
}

func (*ExplainQueryCoster) QueryCost(transaction *model.DbTransaction, query string, args ...interface{}) (int64, error) {
	return explainQueryCost(transaction, true, query, args...)
}

type ExplainAnalyzeQueryCoster struct {
}

func (*ExplainAnalyzeQueryCoster) QueryCost(transaction *model.DbTransaction, query string, args ...interface{}) (int64, error) {
	return explainQueryCost(transaction, true, query, args...)
}

func GetQueryCoster(tp QueryCosterType) QueryCoster {
	switch tp {
	case ExplainQueryCosterType:
		return &ExplainQueryCoster{}
	case ExplainAnalyzeQueryCosterType:
		return &ExplainAnalyzeQueryCoster{}
	case FormulaQueryCosterType:
		return &FormulaQueryCoster{&DBCountQueryRowCounter{}}
	}
	return nil
}
