package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type DbInfoNodesPage struct {
	Lang      map[string]string
	NodesData []map[string]string
	Titles    []string
}

func (c *Controller) DbInfoNodes() (string, error) {

	// стата по нодам
	q := ""
	if c.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT DISTINCT ON (http_host) http_host FROM miners_data WHERE miner_id > 0  LIMIT 20"
	} else {
		q = "SELECT http_host FROM miners_data WHERE miner_id > 0  GROUP BY http_host LIMIT 20"
	}
	rows, err := c.Query(c.FormatQuery(q))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	var titles []string
	var nodesData []map[string]string
	for rows.Next() {
		var http_host string
		err = rows.Scan(&http_host)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		fmt.Println(http_host)
		jsonData, err := utils.GetHttpTextAnswer(http_host + "/ajax?controllerName=checkNode")
		if err != nil {
			continue
		}
		var jsonMap map[string]string
		err = json.Unmarshal([]byte(jsonData), &jsonMap)
		if err != nil {
			continue
		}
		if len(titles) == 0 {
			for k, _ := range jsonMap {
				titles = append(titles, k)
			}
		}
		jsonMap["host"] = http_host
		nodesData = append(nodesData, jsonMap)
	}
	fmt.Println("nodesData", nodesData)

	TemplateStr, err := makeTemplate("db_info_nodes", "dbInfoNodes", &DbInfoNodesPage{
		Lang:      c.Lang,
		Titles:    titles,
		NodesData: nodesData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
