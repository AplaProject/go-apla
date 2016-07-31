package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
)

type nodeConfigPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Config       map[string]string
	WaitingList  []map[string]string
	ConfigIni    string
	UserId       int64
	Lang         map[string]string
	EConfig      map[string]string
	Users        []map[int64]map[string]string
	MyStatus     string
}

func (c *Controller) NodeConfigControl() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	log.Debug("c.Parameters", c.Parameters)
	if _, ok := c.Parameters["save_config"]; ok {
		err := c.ExecSql("UPDATE config SET pool_email = ?, in_connections_ip_limit = ?, in_connections = ?, out_connections = ?, cf_url = ?, pool_url = ?, pool_admin_user_id = ?, exchange_api_url = ?, auto_reload = ?, http_host = ?, chat_enabled = ?, analytics_disabled = ?, auto_update = ?, auto_update_url = ?, stat_host = ?, getpool_host = ?", 
			c.Parameters["pool_email"], c.Parameters["in_connections_ip_limit"], c.Parameters["in_connections"], c.Parameters["out_connections"], c.Parameters["cf_url"], c.Parameters["pool_url"], c.Parameters["pool_admin_user_id"], c.Parameters["exchange_api_url"], c.Parameters["auto_reload"], c.Parameters["http_host"], c.Parameters["chat_enabled"], c.Parameters["analytics_disabled"], c.Parameters["auto_update"], c.Parameters["auto_update_url"], c.Parameters["stat_host"], c.Parameters["getpool_host"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET tcp_listening = ?", c.Parameters["tcp_listening"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	if _, ok := c.Parameters["save_e_config"]; ok {
		err := c.ExecSql("DELETE FROM e_config")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "enable", c.Parameters["e_enable"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(c.Parameters["e_domain"]) > 0 {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "domain", c.Parameters["e_domain"])
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		} else {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "catalog", c.Parameters["e_catalog"])
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		params := []string{"commission", "ps", "pm_s_key", "cp_s_key", "payeer_s_key", "pm_id", "cp_id", "payeer_id", 
				"static_file", "static_file_path", "main_dc_account", "dc_commission", "pm_commission", "cp_commission", "email"}
		for _, data := range params {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, data, c.Parameters["e_"+data])
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}

	tcp_listening, err := c.Single(`SELECT tcp_listening FROM ` + c.MyPrefix + `my_table`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	config, err := c.GetNodeConfig()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	config["tcp_listening"] = tcp_listening



	configIni, err := ioutil.ReadFile(*utils.Dir + "/config.ini")
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	eConfig, err := c.GetMap(`SELECT * FROM e_config`, "name", "value")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	scriptName, err := c.Single("SELECT script_name FROM main_lock").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	myStatus := "ON"
	if scriptName == "my_lock" {
		myStatus = "OFF"
	}
	TemplateStr, err := makeTemplate("node_config", "nodeConfig", &nodeConfigPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		Config:       config,
		UserId:       c.SessUserId,
		ConfigIni:    string(configIni),
		EConfig:      eConfig,
		MyStatus:     myStatus,
		CountSignArr: c.CountSignArr})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
