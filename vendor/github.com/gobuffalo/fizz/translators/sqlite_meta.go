package translators

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gobuffalo/fizz"
)

type sqliteIndexListInfo struct {
	Seq     int    `db:"seq"`
	Name    string `db:"name"`
	Unique  bool   `db:"unique"`
	Origin  string `db:"origin"`
	Partial string `db:"partial"`
}

type sqliteIndexInfo struct {
	Seq  int    `db:"seqno"`
	CID  int    `db:"cid"`
	Name string `db:"name"`
}

type sqliteTableInfo struct {
	CID     int         `db:"cid"`
	Name    string      `db:"name"`
	Type    string      `db:"type"`
	NotNull bool        `db:"notnull"`
	Default interface{} `db:"dflt_value"`
	PK      bool        `db:"pk"`
}

func (t sqliteTableInfo) ToColumn() fizz.Column {
	c := fizz.Column{
		Name:    t.Name,
		ColType: t.Type,
		Primary: t.PK,
		Options: fizz.Options{},
	}
	if !t.NotNull {
		c.Options["null"] = true
	}
	if t.Default != nil {
		c.Options["default"] = strings.TrimSuffix(strings.TrimPrefix(fmt.Sprintf("%s", t.Default), "'"), "'")
	}
	return c
}

type sqliteSchema struct {
	Schema
}

func (p *sqliteSchema) Build() error {
	var err error
	db, err := sql.Open("sqlite3", p.URL)
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		table := &fizz.Table{
			Columns: []fizz.Column{},
			Indexes: []fizz.Index{},
		}
		err = res.Scan(&table.Name)
		if err != nil {
			return err
		}
		if table.Name != "sqlite_sequence" {
			err = p.buildTableData(table, db)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (p *sqliteSchema) buildTableData(table *fizz.Table, db *sql.DB) error {
	prag := fmt.Sprintf(`SELECT "cid", "name", "type", "notnull", "dflt_value", "pk" FROM pragma_table_info('%s')`, table.Name)

	res, err := db.Query(prag)
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		ti := sqliteTableInfo{}
		err = res.Scan(&ti.CID, &ti.Name, &ti.Type, &ti.NotNull, &ti.Default, &ti.PK)
		if err != nil {
			return err
		}
		table.Columns = append(table.Columns, ti.ToColumn())
	}
	err = p.buildTableIndexes(table, db)
	if err != nil {
		return err
	}
	p.schema[table.Name] = table
	return nil
}

func (p *sqliteSchema) buildTableIndexes(t *fizz.Table, db *sql.DB) error {
	prag := fmt.Sprintf(`SELECT "seq", "name", "unique", "origin", "partial" FROM pragma_index_list('%s')`, t.Name)
	res, err := db.Query(prag)
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		li := sqliteIndexListInfo{}
		err = res.Scan(&li.Seq, &li.Name, &li.Unique, &li.Origin, &li.Partial)
		if err != nil {
			return err
		}

		i := fizz.Index{
			Name:    li.Name,
			Unique:  li.Unique,
			Columns: []string{},
		}

		prag = fmt.Sprintf(`SELECT "seqno", "cid", "name" FROM pragma_index_info('%s');`, i.Name)
		iires, err := db.Query(prag)
		if err != nil {
			return err
		}
		defer iires.Close()

		for iires.Next() {
			ii := sqliteIndexInfo{}
			err = iires.Scan(&ii.Seq, &ii.CID, &ii.Name)
			if err != nil {
				return err
			}
			i.Columns = append(i.Columns, ii.Name)
		}

		t.Indexes = append(t.Indexes, i)

	}
	return nil
}
