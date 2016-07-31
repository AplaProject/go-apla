package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//	"regexp"
)

func (p *Parser) AdminVariablesInit() error {
	fields := []string{"variables", "sign"}
	TxMap := make(map[string][]byte)
	TxMap, err := p.GetTxMap(fields)
	p.TxMap = TxMap
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) AdminVariablesFront() error {

	var Variables map[string]interface{}
	err := json.Unmarshal(p.TxMap["variables"], &Variables)
	if err != nil {
		return p.ErrInfo(err)
	}

	var VARIABLES_COUNT int
	if p.BlockData != nil && p.BlockData.BlockId < 29047 {
		VARIABLES_COUNT = 71
	} else if p.BlockData != nil && p.BlockData.BlockId < 279102 {
		VARIABLES_COUNT = 72
	} else {
		VARIABLES_COUNT = 73
	}
	err = p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(Variables) != VARIABLES_COUNT {
		return p.ErrInfo(fmt.Sprintf("incorrect variables count (%d != %d)", len(Variables), VARIABLES_COUNT))
	}

	i := 0
	for name, value := range Variables {
		errorText := "incorrect variable %s"
		// проверим допустимые значения в value. Хотя админу и можно доверять, но лучше перестраховаться.
		switch name {
		case "max_pool_users","alert_error_time", "error_time", "promised_amount_points", "promised_amount_votes_0", "promised_amount_votes_1", "promised_amount_votes_period", "holidays_max", "limit_abuses", "limit_abuses_period", "limit_promised_amount", "limit_promised_amount_period", "limit_cash_requests_out", "limit_cash_requests_out_period", "limit_change_geolocation", "limit_change_geolocation_period", "limit_holidays", "limit_holidays_period", "limit_message_to_admin", "limit_message_to_admin_period", "limit_mining", "limit_mining_period", "limit_node_key", "limit_node_key_period", "limit_primary_key", "limit_primary_key_period", "limit_votes_miners", "limit_votes_miners_period", "limit_votes_complex", "limit_votes_complex_period", "limit_commission", "limit_commission_period", "limit_new_miner", "limit_new_miner_period", "limit_new_user", "limit_new_user_period", "max_block_size", "max_block_user_transactions", "max_day_votes", "max_tx_count", "max_tx_size", "max_user_transactions", "miners_keepers", "miner_points", "miner_votes_0", "miner_votes_1", "miner_votes_attempt", "miner_votes_period", "mining_votes_0", "mining_votes_1", "mining_votes_period", "min_miners_keepers", "node_voting", "node_voting_period", "rollback_blocks_1", "rollback_blocks_2", "limit_change_host", "limit_change_host_period", "min_miners_of_voting", "min_hold_time_promise_amount", "min_promised_amount", "points_update_time", "reduction_period", "new_pct_period", "new_max_promised_amount", "new_max_other_currencies", "cash_request_time", "limit_for_repaid_fix", "limit_for_repaid_fix_period", "miner_newbie_time":
			if !utils.CheckInputData(value, "bigint") {
				return p.ErrInfo(fmt.Errorf(errorText, name))
			}
			i++
		case "points_factor":
			if !utils.CheckInputData(value, "float") {
				return p.ErrInfo(fmt.Errorf(errorText, name))
			}
			i++
		case "system_commission":
			if  p.BlockData == nil || (p.BlockData != nil && p.BlockData.BlockId > 281647) {
				if !utils.CheckInputData(value, "system_commission") {
					return p.ErrInfo(fmt.Errorf(errorText, name))
				}
			}
			i++
		case "sleep":
			if !utils.CheckInputData(value, "sleep_var") {
				return p.ErrInfo(fmt.Errorf(errorText, name))
			}
			i++
		case "default":
			return p.ErrInfo(fmt.Errorf(errorText, name))
		}
	}

	if i != VARIABLES_COUNT {
		return p.ErrInfo(fmt.Sprintf("incorrect variables count (%d != %d)", i, VARIABLES_COUNT))
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["variables"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

type sleepVar struct {
	IsReady   []uint32 `json:"is_ready"`
	Generator []uint32 `json:"generator"`
}

func (p *Parser) AdminVariables() error {

	logData, err := p.DCDB.GetMap("SELECT name, value FROM variables", "name", "value")
	if err != nil {
		return p.ErrInfo(err)
	}

	sleepVar := new(sleepVar)
	err = json.Unmarshal([]byte(logData["sleep"]), &sleepVar)
	if err != nil {
		return p.ErrInfo(err)
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.DCDB.ExecSql("INSERT INTO log_variables (data) VALUES (?)", jsonData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var Variables map[string]interface{}
	err = json.Unmarshal(p.TxMap["variables"], &Variables)
	if err != nil {
		return p.ErrInfo(err)
	}
	for name, value := range Variables {
		exists, err := p.DCDB.Single("SELECT name FROM variables WHERE name = ?", name).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if len(exists) > 0 {
			err := p.DCDB.ExecSql("UPDATE variables SET value = ? WHERE name = ?", value, name)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err := p.DCDB.ExecSql("INSERT INTO variables (name, value) VALUES (?, ?)", name, value)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	return nil
}

func (p *Parser) AdminVariablesRollback() error {
	// данные, которые восстановим
	logData, err := p.DCDB.OneRow("SELECT data, log_id FROM log_variables ORDER BY `log_id` DESC").String()

	var Variables map[string]interface{}
	err = json.Unmarshal([]byte(logData["data"]), &Variables)
	if err != nil {
		return p.ErrInfo(err)
	}
	for name, value := range Variables {
		err = p.DCDB.ExecSql("UPDATE variables SET value = ? WHERE name = ?", value, name)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// подчищаем _log
	err = p.DCDB.ExecSql("DELETE FROM log_variables WHERE log_id = ?", logData["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("log_variables", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminVariablesRollbackFront() error {
	return p.limitRequestsRollback("new_user")
}
