// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
