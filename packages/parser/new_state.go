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

	"github.com/DayLightProject/go-daylight/packages/utils"
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
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"state_name": "string", "currency_name": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	fPrice, err := p.Single(`SELECT value->'new_state' FROM system_parameters WHERE name = ?`, "op_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	fuelRate, err := p.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	dltPrice := int64(fPrice / fuelRate)

	// есть ли нужная сумма на кошельке
	_, err = p.checkSenderDLT(0, dltPrice)
	if err != nil {
		return p.ErrInfo(err)
	}

	// есть ли нужная сумма на кошельке
	_, err = p.checkSenderDLT(0, dltPrice)
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
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_state_parameters" ADD CONSTRAINT "` + id + `_state_parameters_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}
	sid := id + `_citizens.id=` + utils.Int64ToStr(p.TxWalletID)

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
		(?, ?, ?, ?)`,
		"main_conditions", sid, "", "",
		"new_table", sid, "", id+`_state_parameters.main_conditions`,
		"new_column", sid, "", id+`_state_parameters.main_conditions`,
		"changing_tables", sid, "", id+`_state_parameters.main_conditions`,
		"changing_smart_contracts", sid, "", id+`_state_parameters.main_conditions`,
		"currency_name", p.TxMap["currency_name"], "", id+`_state_parameters.main_conditions`,
		"state_name", p.TxMap["state_name"], "", id+`_state_parameters.main_conditions`,
		"dlt_spending", p.TxWalletID, "", id+`_state_parameters.main_conditions`,
		"state_flag", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGQAAAA8CAMAAACn4e/8AAAACXBIWXMAAC4jAAAuIwF4pT92AAAB41BMVEUAJH0BJH0CJH0CJX0CJX4DJX4DJn4EJn4EJ38FJn4FJ38FKH8GKH8GKYAHKYAJKoEKKoEKK4ELK4ELLIELLIIMLIIMLYINLYINLYMOLYMOLoMhPowiPowiPo0iP40jP40kP40kQI0kQI4kQY4lQY4lQo8mQY8mQo8nQo8nQ48nQ5ApRJApRZEqRZBIX6FJYKBJYKFJYaFLYqJMYqJNY6JNY6NNZKNPZaNPZaRQZaRQZqRQZ6RRZ6VSZ6VSaKVSaKZTaKZUaaZUaqZVaqd5irl6irp7i7p8jLt9jbt9jbx+jbt+jbx+jrx/jrx/j7x/j72Aj7yAj72BkL2BkL6Ckb2Ckb6Dkr6Dkr+stdOsttOsttStttSvuNWvudWwuNWwudWwudayu9eyvNezu9ezvNfPFCvQFy7QGS/QGjDU2unV2unW2unW2+rW3OrX3OrX3OvX3erY3OvY3evZ3evZ3uva3uza3+zb3+zc4O3d4e3d4u7e4+7u8fbv8fbw8vfw8vjx8/jx8/ny9Pnz9Pnz9fn09fn09fr19vr29/r2+Pv3+Pv3+fv4+fv5+vz6+/37+/37/P37/P78/P38/P78/f38/f799PX99ff99/f9+Pn9/f7++fn++vr++/z+/v/+//////8M62m4AAAC+ElEQVRYw+2Y61cSURTF76CUlZVR0xg9rLAMezcZVKT2wDRIy8qiJMzAsgeV2ag9kLTS3jRoD0u4f2ofWtrA7Jm5M7BcrRb743Dv+Q1nnXvn7EPoX02e4omWlqco/dkXm9ftWUrfL9ZcLrS+UwQmVKn+A+XFgNi98ZywuRCauraxYAhXe12mehBKR/2rCoM4gi/zY6ogNBMXOesQm0fKUmMIpV9CLqsQdzgNApIZ8JC+aFltBcJ3JFC0adI8nAXPM/f2cGYhNvFxBoTKDh4mhL8whvAzuTkzhrgjMCnJgIMQQrj6qIx+Tpzm2SHV7UkUQw5vmz88TRJaQfv3l7NB7IfuwwAPvXbFf13XOYkWpUMbWCCbI9No9+tg/jVVdwMVHx3380YQoe0NfMHuLepjtMg3iHMm2vUg+ffUnOINdnjpOC+Po+Wp0CZNCOcKw6JJdlRr3tA7o6jS6ai/agWCVDiCsKYy4TqiozKvBA/Ug6MA8sEnwYP8SCwj+nJ2wZdLUzWEwlJJnBWIsXZFYUECCHqZHpSpCrWWnnxmFTLsWwICko9In61CPsFwhLKLJV1QCwOZZdc3JeSHiY3klgnFFDKzj8QWQCVICVKC/CuQPhNSbjSzj/xi11flBfndxMb/6HuSQpKtQmQYjixTq7L5qVXI0JFKEBD0L9j3MaYrdXUrQ9cltGPfhyCwP3vewhsgbOIAbFOl46hNPTEC29S7ezk9xo4e2HAnzzhww81fhH4zo5Mz4Tz2fZFabetQf3PK0G8qpy4N2PfFvfomqFHDb+4DJqimG9bURNtaIzvn1PCbV9ar5lMT2PfVsBjT7dgHvPKv0ZlPAd+nb7G1/OYdcc6jc64I9n3nBPZhgbNLw2/+SUVVgMX3GY89NP3mSkK82PcNiOYHOJ4h7DcPEkbfxzSK0vKb6HuSDrutDtV296aZPlpZyVPIePDYCMN4MBkocNDJXxozgMhhF1fwyNbdO6UHiXvsRRk+Nz7RhLxtFYo1Rnd2KsfovwEjckTHIQpWtgAAAABJRU5ErkJggg==", "", id+`_state_parameters.main_conditions`,
		"state_coords", `{"center_point":["49.922935","18.391113"], "zoom":"5", "cords":[["49.965356","18.347168"],["50.050085","18.061523"],["49.993615","17.863770"],["50.190968","17.600098"],["50.303376","17.819824"],["50.359480","17.534180"],["50.317408","17.336426"],["50.457504","16.853027"],["50.275299","16.918945"],["50.134664","16.699219"],["50.429518","16.149902"],["50.583237","16.435547"],["50.722547","16.237793"],["50.680797","16.105957"],["50.792047","15.776367"],["50.819818","15.358887"],["51.041394","15.183105"],["51.027576","15.007324"],["50.875311","14.897461"],["50.875311","14.743652"],["51.069017","14.479980"],["51.082822","14.238281"],["50.916887","14.414063"],["50.833698","14.040527"],["50.694718","13.579102"],["50.639010","13.249512"],["50.513427","13.007813"],["50.443513","12.656250"],["50.317408","12.392578"],["50.331436","12.019043"],["50.162824","12.150879"],["49.951220","12.458496"],["49.681847","12.458496"],["49.425267","12.656250"],["49.368066","12.985840"],["49.138597","13.227539"],["49.009051","13.513184"],["48.763431","13.842773"],["48.618385","14.018555"],["48.661943","14.611816"],["48.850258","14.941406"],["49.009051","14.941406"],["48.965794","15.249023"],["48.893615","15.666504"],["48.835797","16.040039"],["48.763431","16.303711"],["48.821333","16.479492"],["48.734455","16.743164"],["48.690960","16.962891"],["48.879167","17.182617"],["48.850258","17.512207"],["48.936935","17.885742"],["49.052270","18.017578"],["49.267805","18.171387"],["49.368066","18.391113"],["49.510944","18.588867"],["49.539469","18.852539"],["49.653405","18.852539"],["49.781264","18.632813"],["49.880478","18.566895"],["49.922935","18.391113"]]}`, "", id+`_state_parameters.main_conditions`,
		"citizenship_price", "1000000", "", id+`_state_parameters.main_conditions`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE SEQUENCE "` + id + `_smart_contracts_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_smart_contracts" (
				"id" bigint NOT NULL  default nextval('` + id + `_smart_contracts_id_seq'),
				"name" varchar(100)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
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
		(?, ?),(?, ?),(?,?),(?,?),(?,?)`,
		`TXCitizenRequest`, `contract TXCitizenRequest {
	tx {
		PublicKey  bytes
		StateId    int
		FullName   string
//		MiddleName string "optional"
//		LastName   string
	}
	func front {
		if Balance($wallet) < Float(StateParam($StateId, "citizenship_price")) {
			error "not enough money"
		}
	}
	func main {
		Println("TXCitizenRequest main")
		DBInsert(Sprintf( "%d_citizenship_requests", $StateId), "dlt_wallet_id,public_key_0,name,block_id", $wallet, $PublicKey, $FullName, $block)
	}
}`, `TXNewCitizen`, `contract TXNewCitizen {
	tx {
        RequestId int
        PublicKey bytes
    }

	func main {
		var citizenId int
		Println("NewCitizen Main", $type, $citizen, $block )
		citizenId = DBInsert(Sprintf( "%d_citizens", $state), "public_key_0,block_id,name", $PublicKey, $block, DBString(Sprintf( "%d_citizenship_requests", $state), "name", $RequestId ) )
        DBUpdate(Sprintf( "%d_citizenship_requests", $state), $RequestId, "approved", citizenId)
	}
}`, `TXRejectCitizen`, `contract TXRejectCitizen {
   tx { 
        RequestId int
   }
   func main { 
  //    Println("TXRejectCitizen main", $RequestId  )
	  DBUpdate(Sprintf( "%d_citizenship_requests", $state), $RequestId, "approved", -1)
   }
}`, `TXEditProfile`, `contract TXEditProfile {
	tx {
		FirstName  string
		Image bytes "image"
	}
	func init {
	}
	func front {

	}
	func main {
	  DBUpdate(Sprintf( "%d_citizens", $state), $citizen, "name,image", $FirstName, $Image)
  	  Println("TXEditProfile main")
	}
}`, `TXTest`, `contract TXTest {
	tx {
		Name string 
		Company string "optional"
		Coordinates string "map"
		Photo string "image"
	}
	func main {
		Println("TXTest main")
	}
}`)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`CREATE TABLE "` + id + `_tables" (
				"name" bytea  NOT NULL DEFAULT '',
				"columns_and_permissions" jsonb,
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + id + `_tables" ADD CONSTRAINT "` + id + `_tables_pkey" PRIMARY KEY (name);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_tables" (name, columns_and_permissions, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?)`,
		id+`_citizens`, `{"general_update":"`+sid+`", "update": {"public_key_0": "`+sid+`"}, "insert": "`+sid+`", "new_column":"`+sid+`"}`, id+`_state_parameters.main_conditions`,
		id+`_accounts`, `{"general_update":"`+sid+`", "update": {"amount": "`+sid+`"}, "insert": "`+sid+`", "new_column":"`+sid+`"}`, id+`_state_parameters.main_conditions`)
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
		`dashboard_default`, `*Title : Best country
Navigation( LiTemplate(goverment),non-link text)
PageTitle : Dashboard
MarkDown : ![Flag](http://davutlarhamami.com/images/indir%20%281%29.jpg)
Table{
    Table: 1_smart_contracts
    Order: name
    Columns: [[ID, #id#], [Name, #name#], [Conditions, #conditions#], [Action, BtnEdit(editContract, #id#)]]
}
TxForm { Contract: TXTest }
PageEnd:
`, `menu_default`, sid,
		`citizen_profile`, `{{Title=Profile}}{{Navigation=[Citizen](Citizen) / Editing profile}}
{{PageTitle=Editing profile}}
{{contract.TXEditProfile}}`, `menu_default`, sid)
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
		(?, ?, ?)`,
		`menu_default`, `[Tables](sys.listOfTables)
[Smart contracts](sys.contracts)
[Interface](sys.interface)
[test](dashboard_default)`, sid)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE TABLE "` + id + `_citizens" (
				"id" bigint NOT NULL DEFAULT '0',
				"public_key_0" bytea  NOT NULL DEFAULT '',				
				"name" varchar(100) NOT NULL DEFAULT '',
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

	err = p.ExecSql(`CREATE SEQUENCE "` + id + `_citizenship_requests_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_citizenship_requests" (
				"id" bigint NOT NULL  default nextval('` + id + `_citizenship_requests_id_seq'),
				"dlt_wallet_id" bigint  NOT NULL DEFAULT '0',
				"public_key_0" bytea  NOT NULL DEFAULT '',				
				"name" varchar(100) NOT NULL DEFAULT '',
				"approved" int  NOT NULL DEFAULT '0',
				"block_id" bigint NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_citizenship_requests_id_seq" owned by "` + id + `_citizenship_requests".id;
				ALTER TABLE ONLY "` + id + `_citizenship_requests" ADD CONSTRAINT "` + id + `_citizenship_requests_pkey" PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE SEQUENCE "` + id + `_accounts_id_seq" START WITH 1;
				CREATE TABLE "` + id + `_accounts" (
				"id" bigint NOT NULL  default nextval('` + id + `_accounts_id_seq'),
				"amount" bigint  NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + id + `_accounts_id_seq" owned by "` + id + `_accounts".id;
				ALTER TABLE ONLY "` + id + `_accounts" ADD CONSTRAINT "` + id + `_accounts_pkey" PRIMARY KEY (id);
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

	id_, err := p.Single(`SELECT table_id FROM rollback_tx WHERE tx_hash = [hex] AND table_name = ?`, p.TxHash, "system_states").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	id := utils.Int64ToStr(id_)

	err = p.ExecSql(`DROP TABLE "` + id + `_accounts"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "` + id + `_citizens"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "` + id + `_tables"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "` + id + `_smart_contracts"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "` + id + `_state_parameters"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "` + id + `_citizenship_requests"`)
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
