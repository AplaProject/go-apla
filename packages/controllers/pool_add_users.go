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

/*type JsonBackup struct {
	Community []string `json:"community"`
	Data map[string][]map[string]string `json:"data"`
}*/

func (c *Controller) PoolAddUsers() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	if !c.Community {
		return "", utils.ErrInfo(errors.New("Single mode"))
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

	community := make(map[int64]int64)
	re := regexp.MustCompile(`^([0-9]+)_`)
	for table, _ := range mainMap {
		log.Debug("table %v", table)
		match := re.FindStringSubmatch(table)
		if len(match) != 0 {
			user_id := utils.StrToInt64(match[1])
			community[user_id] = 1
		}
	}

	for user_id, _ := range community {
		schema_.PrefixUserId = int(user_id)
		schema_.GetSchema()
		c.ExecSql(`INSERT INTO community (user_id) VALUES (?)`, user_id)
		log.Debug("mainMap.Community[i] %d", user_id)
	}

	allTables, err := c.GetAllTables()

	for table, arr := range mainMap {
		log.Debug("table %v", table)
		if !utils.InSliceString(table, allTables) {
			continue
		}
		//_ = c.ExecSql(`DROP TABLE `+table)
		//if err != nil {
		//	return "", utils.ErrInfo(err)
		//}
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
				if ok, _ := regexp.MatchString("my_table", table); ok {
					if name == "host" {
						name = "http_host"
					}
				}
				if name == "show_progressbar" {
					name = "show_progress_bar"
				}

				colNames += name + ","
				values = append(values, value)
				if ok, _ := regexp.MatchString("(hash_code|public_key|encrypted)", name); ok {
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
				log.Error("%s", err)
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
