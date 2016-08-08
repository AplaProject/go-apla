package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) getJSON(sqlString string) (string, error) {
	rows, err := c.DB.Query(sqlString)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err := json.Marshal(tableData)
	if err != nil {
		return "", err
	}
	fmt.Println(string(jsonData))
	return string(jsonData), nil
}

func (c *Controller) AnonymHistory() (string, error) {

	str, err := c.getJSON(`SELECT id, hex(recipient_wallet_address) as recipient_wallet_address, amount, time, comment, block_id FROM dlt_transactions`);
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(str), nil
}