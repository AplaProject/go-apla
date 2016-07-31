package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
	"time"
)

func (c *Controller) CheckNode() (string, error) {

	blockData, err := c.GetInfoBlock()
	if err != nil {
		return "", err
	}

	c.r.ParseForm()
	block_id := utils.StrToInt64(c.r.FormValue("block_id"))
	nodes := c.r.FormValue("nodes")
	//col := c.r.FormValue("col")
	//row := c.r.FormValue("row")
	//table := c.r.FormValue("table")

	if block_id > 0 {
		// используется для учета кол-ва подтвержденных блоков, т.е. тех, которые есть у большинства нодов
		hash, err := c.Single(c.FormatQuery("SELECT hash FROM block_chain WHERE id =  ?"), block_id).String()
		if err != nil {
			return "", err
		}
		return hash, nil
	} else if len(nodes) > 0 {
		nodes, err := c.GetAll("SELECT tcp_host, count(user_id) as count FROM miners_data WHERE status =  'miner' GROUP BY tcp_host", 100)
		if err != nil {
			return "", err
		}
		json, _ := json.Marshal(nodes)
		if err != nil {
			return string(json), err
		}
		/*} else if len(col) > 0 && len(row) > 0 && len(table) > 0 {
			allTables, err := c.GetAllTables()
			if err != nil {
				return "", err
			}
			if !utils.InSliceString(table, allTables) {
				return "incorrect table", err
			}
			if ok, _ := regexp.MatchString(`(?i)(my_|_my|config|_refs)`, table); ok{
				return "incorrect table", err
			}
			if ok, _ := regexp.MatchString(`^\-?[0-9]{1,10}$`, row); !ok{
				return "incorrect table", err
			}
			if ok, _ := regexp.MatchString(`^[0-9]{1,15}$`, col); !ok{
				return "incorrect table", err
			}
		}*/
	} else {
		data, err := c.OneRow("SELECT current_version, block_id FROM info_block").String()
		if err != nil {
			return "", err
		}
		var allCounts []map[string]interface{}
		allCounts = append(allCounts, map[string]interface{}{"time": time.Now().Unix()})
		allCounts = append(allCounts, map[string]interface{}{"block_id": data["block_id"]})
		allCounts = append(allCounts, map[string]interface{}{"db_version": data["current_version"]})
		data, err = c.OneRow("SELECT sum(amount) as amount, sum(tdc_amount) as tdc_amount FROM promised_amount WHERE del_block_id = 0").String()
		if err != nil {
			return "", err
		}
		allCounts = append(allCounts, map[string]interface{}{"sum_promised_amount": data["amount"]})
		allCounts = append(allCounts, map[string]interface{}{"sum_promised_tdc_amount": data["tdc_amount"]})
		sum_wallets_amount, err := c.Single("SELECT sum(amount) FROM wallets").String()
		allCounts = append(allCounts, map[string]interface{}{"sum_wallets_amount": sum_wallets_amount})
		sum_forex_amount, err := c.Single("SELECT sum(amount) FROM forex_orders").String()
		allCounts = append(allCounts, map[string]interface{}{"sum_forex_amount": sum_forex_amount})
		if err != nil {
			return "", err
		}
		allTables, err := c.GetAllTables()
		if err != nil {
			return "", err
		}
		log.Debug("%s", allTables)

		for _, table := range allTables {
			vars, err := c.GetAllVariables()
			if err != nil {
				return "", err
			}
			if ok, _ := regexp.MatchString(`^[0-9_]*my_|^e_|^_my|^x_|authorization|^config|chat`, table); ok {
				continue
			}
			sqlWhere := ""
			orderBy := ""
			r, _ := regexp.Compile("(?i)log_time_(.*)")
			lTable := r.FindStringSubmatch(table)
			blockDataTime := utils.StrToInt64(blockData["time"])
			blockDataBlockId := utils.StrToInt64(blockData["block_id"])
			if len(lTable) > 1 && table != "log_time_money_orders" {
				sqlWhere = fmt.Sprintf(" WHERE time > %d ", blockDataTime-vars.Int64["limit_"+lTable[1]+"_period"])
				orderBy = "user_id, time"
				log.Debug("lTable", lTable[1])
				log.Debug("blockDataTime", blockDataTime)
				log.Debug("limit_"+lTable[1]+"_period", vars.Int64["limit_"+lTable[1]+"_period"])
			} else if ok, _ := regexp.MatchString(`(?i)^(log_transactions)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE time > %d ", blockDataTime-86400*3)
			} else if ok, _ := regexp.MatchString(`(?i)^(log_votes)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE del_block_id > %d ", blockDataBlockId-vars.Int64["rollback_blocks_2"])
				orderBy = "user_id, voting_id"
			} else if ok, _ := regexp.MatchString(`(?i)^(log_time_money_orders)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE del_block_id > %d ", blockDataBlockId-vars.Int64["rollback_blocks_2"])
			} else if ok, _ := regexp.MatchString(`(?i)^(log_forex_orders|log_forex_orders_main)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE block_id > %d ", blockDataBlockId-vars.Int64["rollback_blocks_2"])
			} else if ok, _ := regexp.MatchString(`(?i)^(log_commission|log_faces|log_miners|log_miners_data|log_points|log_promised_amount|log_recycle_bin|log_spots_compatibility|log_users|log_votes_max_other_currencies|log_votes_max_promised_amount|log_votes_miner_pct|log_votes_reduction|log_votes_user_pct|log_wallets)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE block_id > %d ", blockDataBlockId-vars.Int64["rollback_blocks_2"])
				orderBy = "log_id"
			} else if ok, _ := regexp.MatchString(`(?i)^(votes_miners|cf_comments|cf_currency|cf_funding|cf_lang|cf_projects|cf_projects_data)$`, table); ok {
				orderBy = "id"
			} else if ok, _ := regexp.MatchString(`(?i)^(wallets_buffer)$`, table); ok {
				sqlWhere = fmt.Sprintf(" WHERE del_block_id > %d ", blockDataBlockId-vars.Int64["rollback_blocks_2"])
				orderBy = "user_id, del_block_id"
			} else if ok, _ := regexp.MatchString(`(?i)^(arbitration_trust_list)$`, table); ok {
				orderBy = "user_id, arbitrator_user_id"
			} else if ok, _ := regexp.MatchString(`(?i)^(points_status)$`, table); ok {
				orderBy = "block_id, time_start"
			}
			count, err := c.Single(c.FormatQuery("SELECT count(*) FROM " + table + " " + sqlWhere)).Int64()
			if err != nil {
				return "", err
			}
			allCounts = append(allCounts, map[string]interface{}{table: count})
			/*if c.ConfigIni["db_type"] != "sqlite" {
				hash, err := c.HashTableData(table, sqlWhere, orderBy)
				if err != nil {
					return "", err
				}
				if len(hash) > 6 {
					hash = hash[:6]
				}
				allCounts = append(allCounts, map[string]interface{}{"_hash_" + table: hash})
			}*/
			log.Debug("%v", orderBy)

		}
		log.Debug("allCounts", allCounts)
		json := ""
		for i := 0; i < len(allCounts); i++ {
			for k, v := range allCounts[i] {
				//log.Debug("k", k)
				//log.Debug("v", v)
				json += fmt.Sprintf(`"%v":"%v",`, k, v)
			}
		}
		return string("{" + json[:len(json)-1] + "}"), nil
	}

	return "", nil
}
