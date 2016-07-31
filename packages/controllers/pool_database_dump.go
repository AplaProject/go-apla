package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"

	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"regexp"
	"time"
)

func (c *Controller) PoolDataBaseDump() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	allTables, err := c.GetAllTables()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	c.r.ParseForm()
	dumpUserId := utils.StrToInt64(c.r.FormValue("dump_user_id"))

	mainMap := make(map[string][]map[string]string)

	if dumpUserId > 0 {
		for _, table := range consts.MyTables {
			data, err := c.GetAll(`SELECT * FROM `+table, -1)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			mainMap[table] = data
		}
	} else {
		for i := 0; i < len(c.CommunityUsers); i++ {
			for _, table := range consts.MyTables {
				table = utils.Int64ToStr(c.CommunityUsers[i]) + "_" + table
				if utils.InSliceString(table, allTables) {
					data, err := c.GetAll(`SELECT * FROM `+table, -1)
					for k, arr := range data {
						for name, value := range arr {
							if ok, _ := regexp.MatchString("(hash_code|public_key|encrypted)", name); ok {
								data[k][name] = string(utils.BinToHex([]byte(value)))
							}
						}
					}
					if err != nil {
						return "", utils.ErrInfo(err)
					}
					mainMap[table] = data
				}
			}
		}
	}

	jsonData, _ := json.Marshal(mainMap)
	log.Debug(string(jsonData))

	c.w.Header().Set("Content-Type", "text/plain")
	c.w.Header().Set("Content-Length", utils.IntToStr(len(jsonData)))
	t := time.Unix(utils.Time(), 0)
	c.w.Header().Set("Content-Disposition", `attachment; filename="dcoin_users_backup-`+t.Format(c.TimeFormat)+`.txt`)
	if _, err := c.w.Write(jsonData); err != nil {
		return "", utils.ErrInfo(errors.New("unable to write text"))
	}

	err = json.Unmarshal(jsonData, &mainMap)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// для теста
	for table, arr := range mainMap {
		log.Debug(table)
		for i, data := range arr {
			log.Debug("%v", i)
			colNames := ""
			values := []string{}
			qq := ""
			for name, value := range data {
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
		}
	}

	return "", nil
}
