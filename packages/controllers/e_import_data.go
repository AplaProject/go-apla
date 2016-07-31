package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/schema"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"regexp"
)

func (c *Controller) EImportData() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseMultipartForm(32 << 20)
	file, _, err := c.r.FormFile("file")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer file.Close()
	//log.Debug("", buffer.String())

	var mainMap map[string][]map[string]string
	err = json.Unmarshal(buffer.Bytes(), &mainMap)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("mainMap %v", mainMap)

	log.Debug("Unmarshal ok")

	schema_ := &schema.SchemaStruct{}
	schema_.DCDB = c.DCDB
	schema_.DbType = c.ConfigIni["db_type"]

	for table, arr := range mainMap {

		log.Debug("table %v", table)

		_ = c.ExecSql(`DELETE FROM  ` + table)

		log.Debug(table)
		var id bool
		for i, data := range arr {
			log.Debug("%v", i)
			colNames := ""
			values := []interface{}{}
			qq := ""
			for name, value := range data {
				if name == "id" {
					id = true
				}
				if name == "lock" && c.ConfigIni["db_type"] == "mysql" {
					name = "`lock`"
				}
				colNames += name + ","
				values = append(values, value)
				if ok, _ := regexp.MatchString("(tx_hash)", name); ok {
					qq += "[hex],"
				} else {
					qq += "?,"
				}
			}
			colNames = colNames[0 : len(colNames)-1]
			qq = qq[0 : len(qq)-1]
			query := `INSERT INTO ` + table + ` (` + colNames + `) VALUES (` + qq + `)`
			log.Debug("%v", query)
			log.Debug("%v", values)
			err = c.ExecSql(query, values...)
			if err != nil {
				if table == "e_authorization" {
					log.Error("%v", values)
				} else {
					return "", utils.ErrInfo(err)
				}
			}
		}
		if id {
			maxId, err := c.Single(`SELECT max(id) FROM ` + table).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			err = c.SetAI(table, maxId+1)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}

	return "", nil
}
