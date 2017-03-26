// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

var (
	isGlobal bool
)

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) NewStateInit() error {

	fields := []map[string]string{{"state_name": "string"}, {"currency_name": "string"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateGlobal(country, currency string) error {
	if !isGlobal {
		list, err := utils.DB.GetAllTables()
		if err != nil {
			return err
		}
		isGlobal = utils.InSliceString(`global_currencies_list`, list) && utils.InSliceString(`global_states_list`, list)
	}
	if isGlobal {
		if id, err := utils.DB.Single(`select id from global_states_list where lower(state_name)=lower(?)`, country).Int64(); err != nil {
			return err
		} else if id > 0 {
			return fmt.Errorf(`State %s already exists`, country)
		}
		if id, err := utils.DB.Single(`select id from global_currencies_list where lower(currency_code)=lower(?)`, currency).Int64(); err != nil {
			return err
		} else if id > 0 {
			return fmt.Errorf(`Currency %s already exists`, currency)
		}
	}
	return nil
}

func (p *Parser) NewStateFront() error {
	err := p.generalCheck(`new_state`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"state_name": "state_name", "currency_name": "currency_name"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["state_name"], p.TxMap["currency_name"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.NewStateGlobal(string(p.TxMap["state_name"]), string(p.TxMap["currency_name"]))
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateMain(country, currency string) (id string, err error) {
	id, err = p.ExecSqlGetLastInsertId(`INSERT INTO system_states DEFAULT VALUES`, "system_states")
	if err != nil {
		return
	}
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "system_states", id)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_state_parameters" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"bytecode" bytea  NOT NULL DEFAULT '',
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_state_parameters" ADD CONSTRAINT "` + id + `_state_parameters_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}
	sid := "ContractConditions(`MainCondition`)" //`$citizen == ` + utils.Int64ToStr(p.TxWalletID) // id + `_citizens.id=` + utils.Int64ToStr(p.TxWalletID)
	psid := sid                                  //fmt.Sprintf(`Eval(StateParam(%s, "main_conditions"))`, id) //id+`_state_parameters.main_conditions`
	err = p.ExecSql(`INSERT INTO "`+id+`_state_parameters" (name, value, bytecode, conditions) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?)`,
		"restore_access_condition", sid, "", psid,
		"new_table", sid, "", psid,
		"new_column", sid, "", psid,
		"changing_tables", sid, "", psid,
		"changing_language", sid, "", psid,
		"changing_signature", sid, "", psid,
		"changing_smart_contracts", sid, "", psid,
		"changing_menu", sid, "", psid,
		"changing_page", sid, "", psid,
		"currency_name", currency, "", psid,
		"gender_list", "male,female", "", psid,
		"money_digit", "0", "", psid,
		"tx_fiat_limit", "10", "", psid,
		"state_name", country, "", psid,
		"gov_account", p.TxWalletID, "", psid,
		"dlt_spending", p.TxWalletID, "", psid,
		"state_flag", "", "", psid,
		"state_coords", ``, "", psid,
		"citizenship_price", "1000000", "", psid)
	if err != nil {
		return
	}
	err = p.ExecSql(`CREATE SEQUENCE "` + id + `_smart_contracts_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_smart_contracts" (
				"id" bigint NOT NULL  default nextval('` + id + `_smart_contracts_id_seq'),
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"wallet_id" bigint  NOT NULL DEFAULT '0',
				"active" character(1) NOT NULL DEFAULT '0',
				"conditions" text  NOT NULL DEFAULT '',
				"variables" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_smart_contracts_id_seq" owned by "` + id + `_smart_contracts".id;
				ALTER TABLE ONLY "` + id + `_smart_contracts" ADD CONSTRAINT "` + id + `_smart_contracts_pkey" PRIMARY KEY (id);
				CREATE INDEX "` + id + `_smart_contracts_index_name" ON "` + id + `_smart_contracts" (name);
				`)
	if err != nil {
		return
	}
	err = p.ExecSql(`INSERT INTO "`+id+`_smart_contracts" (name, value, wallet_id, active) VALUES
		(?, ?, ?, ?)`,
		`MainCondition`, `contract MainCondition {
            data {}
            conditions {
                    if(StateVal("gov_account")!=$citizen)
                    {
                        warning "Sorry, you don't have access to this action."
                    }
            }
            action {}
    }`, p.TxWalletID, 1,
	)

	if err != nil {
		return
	}
	err = p.ExecSql(`UPDATE "`+id+`_smart_contracts" SET conditions = ?`, sid)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_tables" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"columns_and_permissions" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_tables" ADD CONSTRAINT "` + id + `_tables_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_tables" (name, columns_and_permissions, conditions) VALUES
		(?, ?, ?)`,
		id+`_citizens`, `{"general_update":"`+sid+`", "update": {"public_key_0": "`+sid+`"}, "insert": "`+sid+`", "new_column":"`+sid+`"}`, psid)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_pages" (
				"name" varchar(255)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"menu" varchar(255)  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_pages" ADD CONSTRAINT "` + id + `_pages_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_pages" (name, value, menu, conditions) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?)`,
		`dashboard_default`, `FullScreen(1)

If(StateVal(type_office))
Else:
Title : Basic Apps
Divs: col-md-4
		Divs: panel panel-default elastic
			Divs: panel-body text-center fill-area flexbox-item-grow
				Divs: flexbox-item-grow flex-center
					Divs: pv-lg
					Image("/static/img/apps/money.png", Basic, center-block img-responsive img-circle img-thumbnail thumb96 )
					DivsEnd:
					P(h4,Basic Apps)
					P(text-left,"Election and Assign, Polling, Messenger, Simple Money System")
				DivsEnd:
			DivsEnd:
			Divs: panel-footer
				Divs: clearfix
					Divs: pull-right
						BtnPage(app-basic, Install,'',btn btn-primary lang)
					DivsEnd:
				DivsEnd:
			DivsEnd:
		DivsEnd:
	DivsEnd:
IfEnd:
PageEnd:
`, `menu_default`, sid,

		`government`, `Title : 

PageEnd:
`, `government`, sid,
	)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_menu" (
				"name" varchar(255)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_menu" ADD CONSTRAINT "` + id + `_menu_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}
	err = p.ExecSql(`INSERT INTO "`+id+`_menu" (name, value, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?)`,
		`menu_default`, `MenuItem(Dashboard, dashboard_default)
 MenuItem(Government dashboard, government)`, sid,
		`government`, `MenuItem(Citizen dashboard, dashboard_default)
MenuItem(Government dashboard, government)
MenuGroup(Admin tools,admin)
MenuItem(Tables,sys-listOfTables)
MenuItem(Smart contracts, sys-contracts)
MenuItem(Interface, sys-interface)
MenuItem(App List, sys-app_catalog)
MenuItem(Export, sys-export_tpl)
MenuItem(Wallet,  sys-edit_wallet)
MenuItem(Languages, sys-languages)
MenuItem(Signatures, sys-signatures)
MenuEnd:
MenuBack(Welcome)`, sid)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_citizens" (
				"id" bigint NOT NULL DEFAULT '0',
				"public_key_0" bytea  NOT NULL DEFAULT '',				
				"block_id" bigint NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_citizens" ADD CONSTRAINT "` + id + `_citizens_pkey" PRIMARY KEY (id);
				`)
	if err != nil {
		return
	}

	pKey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Bytes()
	if err != nil {
		return
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_citizens" (id,public_key_0) VALUES (?, [hex])`, p.TxWalletID, utils.BinToHex(pKey))
	if err != nil {
		return
	}
	err = p.ExecSql(`CREATE TABLE "` + id + `_languages" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"res" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_languages" ADD CONSTRAINT "` + id + `_languages_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}
	err = p.ExecSql(`INSERT INTO "`+id+`_languages" (name, res, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?),
		(?, ?, ?)`,
		`dateformat`, `{"en": "YYYY-MM-DD", "ru": "DD.MM.YYYY"}`, sid,
		`timeformat`, `{"en": "YYYY-MM-DD HH:MI:SS", "ru": "DD.MM.YYYY HH:MI:SS"}`, sid,
		`Gender`, `{"en": "Gender", "ru": "Пол"}`, sid,
		`male`, `{"en": "Male", "ru": "Мужской"}`, sid,
		`female`, `{"en": "Female", "ru": "Женский"}`, sid)
	if err != nil {
		return
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_signatures" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_signatures" ADD CONSTRAINT "` + id + `_signatures_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return
	}

	err = utils.LoadContract(id)
	return
}

func (p *Parser) NewState() error {
	var pkey string
	country := string(p.TxMap["state_name"])
	currency := string(p.TxMap["currency_name"])
	id, err := p.NewStateMain(country, currency)
	if err != nil {
		return p.ErrInfo(err)
	}
	if isGlobal {
		_, err = p.selectiveLoggingAndUpd([]string{"stateId", "state_name"},
			[]interface{}{id, country}, "global_states_list", nil, nil, true)
		if err != nil {
			return p.ErrInfo(err)
		}
		_, err = p.selectiveLoggingAndUpd([]string{"currency_code", "settings_table"},
			[]interface{}{currency, id + `_state_parameters`}, "global_currencies_list", nil, nil, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	if pkey, err = p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String(); err != nil {
		return p.ErrInfo(err)
	} else if len(p.TxMaps.Bytes["public_key"]) > 30 && len(pkey) == 0 {
		_, err = p.selectiveLoggingAndUpd([]string{"public_key_0"}, []interface{}{utils.HexToBin(p.TxMaps.Bytes["public_key"])}, "dlt_wallets",
			[]string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	return err
}

func (p *Parser) NewStateRollback() error {
	id, err := p.Single(`SELECT table_id FROM rollback_tx WHERE tx_hash = [hex] AND table_name = ?`, p.TxHash, "system_states").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.autoRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	for _, name := range []string{`menu`, `pages`, `citizens`, `languages`, `signatures`, `tables`,
		`smart_contracts`, `state_parameters` /*, `citizenship_requests`*/} {
		err = p.ExecSql(fmt.Sprintf(`DROP TABLE "%d_%s"`, id, name))
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ExecSql(`DELETE FROM rollback_tx WHERE tx_hash = [hex] AND table_name = ?`, p.TxHash, "system_states")
	if err != nil {
		return p.ErrInfo(err)
	}

	maxId, err := p.Single(`SELECT max(id) FROM "system_states"`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// обновляем AI
	err = p.SetAI("system_states", maxId+1)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`DELETE FROM "system_states" WHERE id = ?`, id)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

/*func (p *Parser) NewStateRollbackFront() error {

	return nil
}
*/
