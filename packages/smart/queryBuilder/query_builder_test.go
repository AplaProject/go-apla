package smart

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

// query="SELECT ,,,id,amount,\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\"ecosystem\"
// FROM \"1_keys\" \nWHERE  AND id = -6752330173818123413 AND ecosystem = '1'\n"

// fields="[+amount]"
// values="[2912910000000000000]"

// whereF="[id]"
// whereV="[-6752330173818123413]"

type TestKeyTableChecker struct {
	Val bool
}

func (tc TestKeyTableChecker) IsKeyTable(tableName string) bool {
	return tc.Val
}
func TestSqlFields(t *testing.T) {
	qb := smartQueryBuilder{
		Entry:        log.WithFields(log.Fields{"mod": "test"}),
		table:        "1_keys",
		Fields:       []string{"+amount"},
		FieldValues:  []interface{}{2912910000000000000},
		WhereFields:  []string{"id"},
		WhereValues:  []string{"-6752330173818123413"},
		KeyTableChkr: TestKeyTableChecker{true},
	}

	fields, err := qb.GetSQLSelectFieldsExpr()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(fields)
}
