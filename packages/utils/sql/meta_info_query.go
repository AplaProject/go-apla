package sql

import (
	"strings"
)

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
