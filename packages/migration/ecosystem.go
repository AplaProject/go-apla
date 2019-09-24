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

package migration

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

type SqlData struct {
	Ecosystem int
	Wallet    int64
	Name      string
	Founder   int64
	AppID     int64
	Account   string
}

var _ fizz.Translator = (*translators.Postgres)(nil)
var pgt = translators.NewPostgres()
var tblName string

const (
	sqlPrimary = "primary"
	sqlUnique  = "unique"
	sqlIndex   = "index"
)

func sqlHead(name string) string {
	tblName = name
	return fmt.Sprintf(`sql("DROP TABLE IF EXISTS \"%[1]s\";")
	create_table("%[1]s") {`, name)
}

func sqlEnd(options ...string) (ret string) {
	ret = `t.DisableTimestamps()
	}`
	for _, opt := range options {
		var cname string
		if strings.HasPrefix(opt, sqlPrimary) {
			opt = strings.Replace(opt, sqlPrimary, `PRIMARY KEY (id)`, 1)
			cname = "pkey"
		}
		if strings.HasPrefix(opt, sqlUnique) {
			pars := strings.Split(strings.Trim(opt[len(sqlUnique):], `() `), `,`)
			opt = strings.Replace(opt, sqlUnique, `UNIQUE `, 1)
			for i, val := range pars {
				pars[i] = strings.TrimSpace(val)
			}
			cname = strings.Join(pars, `_`)
		}
		if strings.HasPrefix(opt, sqlIndex) {
			pars := strings.Split(strings.Trim(opt[len(sqlIndex):], `() `), `,`)
			for i, val := range pars {
				pars[i] = strings.TrimSpace(val)
			}
			if len(pars) == 1 {
				ret += fmt.Sprintf(`
		add_index("%s", "%s", {})`, tblName, pars[0])
			} else {
				ret += fmt.Sprintf(`
		add_index("%s", ["%s"], {})`, tblName, strings.Join(pars, `", "`))
			}
			continue
		}
		ret += fmt.Sprintf(`
	sql("ALTER TABLE ONLY \"%[1]s\" ADD CONSTRAINT \"%[1]s_%[3]s\" %[2]s;")`, tblName, opt, cname)
	}
	return
}

func sqlConvert(in []string) (ret string, err error) {
	var item string
	funcs := template.FuncMap{
		"head":   sqlHead,
		"footer": sqlEnd,
	}
	sqlTmpl := template.New("sql").Funcs(funcs)
	for _, sql := range in {
		var (
			tmpl *template.Template
			out  bytes.Buffer
		)

		if tmpl, err = sqlTmpl.Parse(sql); err != nil {
			return
		}
		if err = tmpl.Execute(io.Writer(&out), nil); err != nil {
			return
		}
		item, err = fizz.AString(out.String(), pgt)
		if err != nil {
			return
		}
		ret += item + "\r\n"
	}
	return
}

func sqlTemplate(input []string, data interface{}) (ret string, err error) {
	for _, item := range input {
		var (
			out  bytes.Buffer
			tmpl *template.Template
		)
		tmpl, err = template.New("sql").Parse(item)
		if err != nil {
			return
		}
		if err = tmpl.Execute(io.Writer(&out), data); err != nil {
			return
		}
		ret += out.String() + "\r\n"
	}
	return
}

// GetEcosystemScript returns script to create ecosystem
func GetEcosystemScript(id int, wallet int64, name string, founder,
	appID int64) (string, error) {
	data := SqlData{
		Ecosystem: id,
		Wallet:    wallet,
		Name:      name,
		Founder:   founder,
		AppID:     appID,
		Account:   converter.AddressToString(wallet),
	}
	return sqlTemplate([]string{
		contractsDataSQL,
		menuDataSQL,
		pagesDataSQL,
		parametersDataSQL,
		membersDataSQL,
		sectionsDataSQL,
		keysDataSQL,
	}, data)
}

// GetFirstEcosystemScript returns script to update with additional data for first ecosystem
func GetFirstEcosystemScript(wallet int64) (ret string, err error) {
	ret, err = sqlConvert([]string{
		sqlFirstEcosystemSchema,
	})
	if err != nil {
		return
	}
	var out string
	out, err = sqlTemplate([]string{firstDelayedContractsDataSQL}, SqlData{Wallet: wallet})
	ret += out

	scripts := []string{
		firstEcosystemContractsSQL,
		firstEcosystemPagesDataSQL,
		firstEcosystemBlocksDataSQL,
		firstEcosystemDataSQL,
		firstSystemParametersDataSQL,
		firstTablesDataSQL,
	}
	ret += strings.Join(scripts, "\r\n")
	return
}

// GetFirstTableScript returns script to update _tables for first ecosystem
func GetFirstTableScript() string {
	scripts := []string{
		tablesDataSQL,
	}
	return strings.Join(scripts, "\r\n")
}

// GetCommonEcosystemScript returns script with common tables
func GetCommonEcosystemScript() (string, error) {
	sql, err := sqlConvert([]string{
		sqlFirstEcosystemCommon,
		sqlTimeZonesSQL,
	})
	if err != nil {
		return ``, err
	}
	return sql + "\r\n" + timeZonesSQL, nil
}
