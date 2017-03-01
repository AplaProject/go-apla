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
	//"encoding/json"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) NewStateInit() error {

	fields := []map[string]string{{"state_name": "string"}, {"currency_name": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
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

	return nil
}

func (p *Parser) NewState() error {

	id_, err := p.ExecSqlGetLastInsertId(`INSERT INTO system_states DEFAULT VALUES`, "system_states")
	if err != nil {
		return p.ErrInfo(err)
	}
	id := id_
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "system_states", id)
	if err != nil {
		return err
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
		return p.ErrInfo(err)
	}
	sid := `IsGovAccount($citizen)` //`$citizen == ` + utils.Int64ToStr(p.TxWalletID) // id + `_citizens.id=` + utils.Int64ToStr(p.TxWalletID)
	psid := sid                     //fmt.Sprintf(`Eval(StateParam(%s, "main_conditions"))`, id) //id+`_state_parameters.main_conditions`
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
		(?, ?, ?, ?)`,
		"restore_access_condition", sid, "", psid,
		"new_table", sid, "", psid,
		"new_column", sid, "", psid,
		"changing_tables", sid, "", psid,
		"changing_language", sid, "", psid,
		"changing_signature", sid, "", psid,
		"changing_smart_contracts", sid, "", psid,
		"currency_name", p.TxMap["currency_name"], "", psid,
		"gender_list", "male,female", "", psid,
		"money_digit", "0", "", psid,
		"state_name", p.TxMap["state_name"], "", psid,
		"gov_account", p.TxWalletID, "", psid,
		"dlt_spending", p.TxWalletID, "", psid,
		"state_flag", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGQAAAAyCAYAAACqNX6+AAAAwElEQVR4Xu3TQREAAAiEQK9/aWvsAxMw4O06ysAommCuINgTFKQgmAEMp4UUBDOA4bSQgmAGMJwWUhDMAIbTQgqCGcBwWkhBMAMYTgspCGYAw2khBcEMYDgtpCCYAQynhRQEM4DhtJCCYAYwnBZSEMwAhtNCCoIZwHBaSEEwAxhOCykIZgDDaSEFwQxgOC2kIJgBDKeFFAQzgOG0kIJgBjCcFlIQzACG00IKghnAcFpIQTADGE4LKQhmAMNpIViQBxv1ADO4LcKOAAAAAElFTkSuQmCC", "", psid,
		"state_coords", ``, "", psid,
		"citizenship_price", "1000000", "", psid)
	if err != nil {
		return p.ErrInfo(err)
	}
	/*{"center_point":["49.922935","18.391113"], "zoom":"5", "cords":[["49.965356","18.347168"],["50.050085","18.061523"],["49.993615","17.863770"],["50.190968","17.600098"],["50.303376","17.819824"],["50.359480","17.534180"],["50.317408","17.336426"],["50.457504","16.853027"],["50.275299","16.918945"],["50.134664","16.699219"],["50.429518","16.149902"],["50.583237","16.435547"],["50.722547","16.237793"],["50.680797","16.105957"],["50.792047","15.776367"],["50.819818","15.358887"],["51.041394","15.183105"],["51.027576","15.007324"],["50.875311","14.897461"],["50.875311","14.743652"],["51.069017","14.479980"],["51.082822","14.238281"],["50.916887","14.414063"],["50.833698","14.040527"],["50.694718","13.579102"],["50.639010","13.249512"],["50.513427","13.007813"],["50.443513","12.656250"],["50.317408","12.392578"],["50.331436","12.019043"],["50.162824","12.150879"],["49.951220","12.458496"],["49.681847","12.458496"],["49.425267","12.656250"],["49.368066","12.985840"],["49.138597","13.227539"],["49.009051","13.513184"],["48.763431","13.842773"],["48.618385","14.018555"],["48.661943","14.611816"],["48.850258","14.941406"],["49.009051","14.941406"],["48.965794","15.249023"],["48.893615","15.666504"],["48.835797","16.040039"],["48.763431","16.303711"],["48.821333","16.479492"],["48.734455","16.743164"],["48.690960","16.962891"],["48.879167","17.182617"],["48.850258","17.512207"],["48.936935","17.885742"],["49.052270","18.017578"],["49.267805","18.171387"],["49.368066","18.391113"],["49.510944","18.588867"],["49.539469","18.852539"],["49.653405","18.852539"],["49.781264","18.632813"],["49.880478","18.566895"],["49.922935","18.391113"]]}*/
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
				`)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`INSERT INTO "`+id+`_smart_contracts" (name, value) VALUES
		(?, ?)`,
		`MainCondition`, `contract MainCondition {
            data {}
            conditions {
                    if(StateValue("gov_account")!=$citizen)
                    {
                        warning "Sorry, you don't have access to this action."
                    }
            }
            action {}
    }`,
	)

	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`UPDATE "`+id+`_smart_contracts" SET conditions = ?`, sid)
	if err != nil {
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_tables" (name, columns_and_permissions, conditions) VALUES
		(?, ?, ?)`,
		id+`_citizens`, `{"general_update":"`+sid+`", "update": {"public_key_0": "`+sid+`"}, "insert": "`+sid+`", "new_column":"`+sid+`"}`, psid)
	if err != nil {
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_pages" (name, value, menu, conditions) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?)`,
		`dashboard_default`, `FullScreen(1)
Title : My country
Navigation( Dashboard )

Divs(md-6)
GetRow(my, #state_id#_citizens, "id", #citizen#)
WiCitizen( #my_name#, #my_id#, #my_avatar#, StateValue(state_flag) )
DivsEnd:

PageEnd:
`, `menu_default`, sid,

		`government`, `Title : My country
Navigation( LiTemplate(dashboard_default, citizen),goverment)
Divs(md-8, panel panel-default panel-body)

MarkDown : # Welcome, government!
BtnTemplate(sys-listOfTables, Tables, '', btn btn-pill-left btn-primary)
BtnTemplate(sys-contracts, Contracts)
BtnTemplate(sys-app_catalog, App catalog)
BtnTemplate(sys-interface, Interface, '', btn btn-pill-right btn-primary) BR() BR()
DivsEnd:

Divs(md-4, panel panel-default panel-body)
Ring(GetOne(count(*),#state_id#_citizens), 30)
DivsEnd:

Divs(md-12, panel panel-default panel-body)
     MarkDown : ## Citizens
     Table{
         Table:  `+id+`_citizens
         Order: id
         Columns: [[Avatar,Image(#avatar#)], [ID, Address(#id#)], [Name, #name#]]
     }
DivsEnd:

PageEnd:
`, `government`, sid,
	)
	if err != nil {
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`INSERT INTO "`+id+`_menu" (name, value, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?)`,
		`menu_default`, `MenuItem(Dashboard, dashboard_default)
 MenuItem(Government dashboard, government)`, sid,
		`government`, `MenuItem(Citizen dashboard, dashboard_default)
MenuItem(Government dashboard, government)
MenuItem(Tables,sys-listOfTables)
MenuItem(Smart contracts, sys-contracts)
MenuItem(App List, sys-app_catalog)
MenuItem(Wallet,  sys-edit_wallet)
MenuItem(Languages, sys-languages)
MenuItem(Signatures, sys-signatures)
MenuItem(Export, sys-export_tpl)
MenuItem(Interface, sys-interface)`, sid)
	if err != nil {
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
	}

	pKey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_citizens" (id,public_key_0) VALUES (?, [hex])`, p.TxWalletID, utils.BinToHex(pKey))
	if err != nil {
		return p.ErrInfo(err)
	}
/*
	err = p.ExecSql(`CREATE SEQUENCE "` + id + `_citizenship_requests_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_citizenship_requests" (
				"id" bigint NOT NULL  default nextval('` + id + `_citizenship_requests_id_seq'),
				"dlt_wallet_id" bigint  NOT NULL DEFAULT '0',
				"public_key_0" bytea  NOT NULL DEFAULT '',				
				"name" varchar(100) NOT NULL DEFAULT '',
				"mname" varchar(100) NOT NULL DEFAULT '',
				"gender" varchar(16) NOT NULL DEFAULT '',
				"birthday" varchar(16) NOT NULL DEFAULT '',
				"birthplace" varchar(300) NOT NULL DEFAULT '',
				"approved" bigint  NOT NULL DEFAULT '0',
				"block_id" bigint NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_citizenship_requests_id_seq" owned by "` + id + `_citizenship_requests".id;
				ALTER TABLE ONLY "` + id + `_citizenship_requests" ADD CONSTRAINT "` + id + `_citizenship_requests_pkey" PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}
*/
	/*err = p.ExecSql(`CREATE SEQUENCE "` + id + `_accounts_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_accounts" (
				"id" bigint NOT NULL  default nextval('` + id + `_accounts_id_seq'),
				"amount" decimal(30)  NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_accounts_id_seq" owned by "` + id + `_accounts".id;
				ALTER TABLE ONLY "` + id + `_accounts" ADD CONSTRAINT "` + id + `_accounts_pkey" PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}*/
	err = p.ExecSql(`CREATE TABLE "` + id + `_languages" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"res" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_languages" ADD CONSTRAINT "` + id + `_languages_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
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
		return p.ErrInfo(err)
	}

	if err = utils.LoadContract(id); err != nil {
		return p.ErrInfo(err)
	}
	return nil
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
		`smart_contracts`, `state_parameters`/*, `citizenship_requests`*/} {
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
