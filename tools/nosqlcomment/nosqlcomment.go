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
