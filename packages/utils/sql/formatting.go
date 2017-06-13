package sql

import (
	"fmt"
	"regexp"
	"strings"
)

// ReplQ preprocesses a database query
func ReplQ(q string) string {
	var quote, skip bool
	ind := 1
	in := []rune(q)
	out := make([]rune, 0, len(in)+16)
	for i, ch := range in {
		if skip {
			skip = false
		} else if ch == '\'' {
			if quote {
				if i == len(in)-1 || in[i+1] != '\'' {
					quote = false
				} else {
					skip = true
				}
			} else {
				quote = true
			}
		}
		if ch != '?' || quote {
			out = append(out, ch)
		} else {
			out = append(out, []rune(fmt.Sprintf(`$%d`, ind))...)
			ind++
		}
	}
	return string(out)
}

func FormatQueryArgs(q, dbType string, args ...interface{}) (string, []interface{}) {
	var newArgs []interface{}

	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		newQ = strings.Replace(newQ, "[hex]", "decode(?,'HEX')", -1)
		newQ = strings.Replace(newQ, " authorization", ` "authorization"`, -1)
		newQ = strings.Replace(newQ, "user,", `"user",`, -1)
		newQ = ReplQ(newQ)
		newArgs = args
	}

	r, _ := regexp.Compile(`hex\(([\w]+)\)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		newQ = newQ[:indexArr[i][0]] + `LOWER(encode(` + newQ[indexArr[i][2]:indexArr[i][3]] + `, 'hex'))` + newQ[indexArr[i][1]:]
	}

	return newQ, newArgs
}

// FormatQuery formats the query
func (db *DCDB) FormatQuery(q string) string {
	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		switch db.ConfigIni["db_type"] {
		case "postgresql":
			newQ = strings.Replace(newQ, "[hex]", "decode(?,'HEX')", -1)
			newQ = strings.Replace(newQ, " authorization", ` "authorization"`, -1)
			newQ = strings.Replace(newQ, "user,", `"user",`, -1)
			newQ = strings.Replace(newQ, ", user ", `, "user" `, -1)
			newQ = ReplQ(newQ)
		case "mysql":
			newQ = strings.Replace(newQ, "[hex]", "UNHEX(?)", -1)
		}
	}

	if db.ConfigIni["db_type"] == "postgresql" || db.ConfigIni["db_type"] == "sqlite" {
		r, _ := regexp.Compile(`\s*([0-9]+_[\w]+)(?:\.|\s|\)|$)`)
		indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
		for i := len(indexArr) - 1; i >= 0; i-- {
			newQ = newQ[:indexArr[i][2]] + `"` + newQ[indexArr[i][2]:indexArr[i][3]] + `"` + newQ[indexArr[i][3]:]
		}
	}

	r, _ := regexp.Compile(`hex\(([\w]+)\)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		if db.ConfigIni["db_type"] == "mysql" || db.ConfigIni["db_type"] == "sqlite" {
			newQ = newQ[:indexArr[i][0]] + `LOWER(HEX(` + newQ[indexArr[i][2]:indexArr[i][3]] + `))` + newQ[indexArr[i][1]:]
		} else {
			newQ = newQ[:indexArr[i][0]] + `LOWER(encode(` + newQ[indexArr[i][2]:indexArr[i][3]] + `, 'hex'))` + newQ[indexArr[i][1]:]
		}
	}

	log.Debug("%v", newQ)
	return newQ
}
