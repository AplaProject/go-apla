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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// The program deletes COMMENT in sql file. It creates filename-new.sql file.
func main() {
	sql := flag.String("sql", `sql.sql`, "Initial sql file.")

	flag.Parse()
	outfile := filepath.Base(*sql)
	ext := filepath.Ext(outfile)
	outfile = filepath.Dir(*sql) + `/` + outfile[:len(outfile)-len(ext)] + `-new` + ext
	if sqlText, err := ioutil.ReadFile(*sql); err != nil {
		fmt.Println(err.Error())
	} else {
		tmp := strings.Replace(string(sqlText), ` COMMENT=`, ` COMMENT =`, -1)
		parts := strings.Split(strings.Replace(tmp, ` comment=`, ` COMMENT =`, -1), ` COMMENT`)
		pattern := regexp.MustCompile(`(?i)^[=\s]+'[^']*'`) //`^\s*"[^"]*"\s*,`)
		output := ``
		for _, item := range parts {
			found := pattern.FindStringIndex(item)
			if found == nil {
				output += item
			} else {
				output += item[found[1]:]
			}
		}
		if err := ioutil.WriteFile(outfile, []byte(output), 0644); err != nil {
			fmt.Println(err.Error())
		}
	}
}
