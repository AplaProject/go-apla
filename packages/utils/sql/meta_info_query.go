package sql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (db *DCDB) GetQueryTotalCost(query string, args ...interface{}) (int64, error) {
	var planStr string
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	err := db.QueryRow(fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", newQuery), newArgs...).Scan(&planStr)
	switch {
	case err == sql.ErrNoRows:
		return 0, errors.New("No rows")
	case err != nil:
		return 0, err
	}
	var queryPlan []map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(planStr))
	dec.UseNumber()
	if err := dec.Decode(&queryPlan); err != nil {
		return 0, err
	}
	if len(queryPlan) == 0 {
		return 0, errors.New("Query plan is empty")
	}
	firstNode := queryPlan[0]
	var plan interface{}
	var ok bool
	if plan, ok = firstNode["Plan"]; !ok {
		return 0, errors.New("No Plan key in result")
	}
	var planMap map[string]interface{}
	if planMap, ok = plan.(map[string]interface{}); !ok {
		return 0, errors.New("Plan is not map[string]interface{}")
	}
	if totalCost, ok := planMap["Total Cost"]; ok {
		if totalCostNum, ok := totalCost.(json.Number); ok {
			if totalCostF64, err := totalCostNum.Float64(); err != nil {
				return 0, err
			} else {
				return int64(totalCostF64), nil
			}
		} else {
			return 0, errors.New("Total cost is not a number")
		}
	} else {
		return 0, errors.New("PlanMap has no TotalCost")
	}
	return 0, nil
}

// GetFirstColumnName returns the name of the first column in the table
func (db *DCDB) GetFirstColumnName(table string) (string, error) {
	rows, err := db.Query(`SELECT * FROM "` + table + `" LIMIT 1`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if len(columns) > 0 {
		return columns[0], nil
	}
	return "", nil
}

// GetAllTables returns the list of the tables
func (db *DCDB) GetAllTables() ([]string, error) {
	var result []string
	sql := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND    table_schema NOT IN ('pg_catalog', 'information_schema')"
	result, err := db.GetList(sql).String()
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetFirstColumnNamesPg returns the first name of the column with PostgreSQL request
func (db *DCDB) GetFirstColumnNamesPg(table string) (string, error) {
	var result []string
	result, err := db.GetList("SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='" + table + "'").String()
	if err != nil {
		return "", err
	}
	return result[0], nil
}

func (db *DCDB) IsIndex(tblname, column string) (bool, error) {
	indexes, err := db.GetAll(`select t.relname as table_name, i.relname as index_name, a.attname as column_name 
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?  and a.attname = ?`, 1, tblname, column)
	if err != nil {
		return false, err
	}
	return len(indexes) > 0, nil
}

// NumIndexes returns the amount of the indexes in the table
func (db *DCDB) NumIndexes(tblname string) (int, error) {
	indexes, err := db.Single(`select count( i.relname) from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?`, tblname).Int64()
	if err != nil {
		return 0, err
	}
	return int(indexes - 1), nil
}

// IsCustomTable checks if the table is created by the users
func (db *DCDB) IsCustomTable(table string) (isCustom bool, err error) {
	if (table[0] >= '0' && table[0] <= '9') || strings.HasPrefix(table, `global_`) {
		if off := strings.IndexByte(table, '_'); off > 0 {
			prefix := table[:off]
			if name, err := db.Single(`select name from "`+prefix+`_tables" where name = ?`, table).String(); err == nil {
				isCustom = name == table
			}
		}
	}
	return
}

// GetColumnType returns the type of the column
func GetColumnType(tblname, column string) (itype string) {
	coltype, _ := DB.OneRow(`select data_type,character_maximum_length from information_schema.columns
where table_name = ? and column_name = ?`, tblname, column).String()
	if len(coltype) > 0 {
		switch {
		case coltype[`data_type`] == "character varying":
			itype = `text`
		case coltype[`data_type`] == "bytea":
			itype = "varchar"
		case coltype[`data_type`] == `bigint`:
			itype = "numbers"
		case strings.HasPrefix(coltype[`data_type`], `timestamp`):
			itype = "date_time"
		case strings.HasPrefix(coltype[`data_type`], `numeric`):
			itype = "money"
		case strings.HasPrefix(coltype[`data_type`], `double`):
			itype = "double"
		}
	}
	return
}

// IsTable checks if there is a table with this name
func (db *DCDB) IsTable(tblname string) bool {
	name, _ := db.Single(`SELECT table_name FROM information_schema.tables 
         WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')
     	AND table_name=?`, tblname).String()
	return name == tblname
}
