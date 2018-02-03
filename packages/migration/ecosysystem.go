package migration

var (
	SchemaVDE = `DROP TABLE IF EXISTS "%[1]d_vde_languages"; CREATE TABLE "%[1]d_vde_languages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"res" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_languages" ADD CONSTRAINT "%[1]d_vde_languages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_languages_index_name" ON "%[1]d_vde_languages" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_menu"; CREATE TABLE "%[1]d_vde_menu" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "title" character varying(255) NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_menu" ADD CONSTRAINT "%[1]d_vde_menu_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_menu_index_name" ON "%[1]d_vde_menu" (name);


	  INSERT INTO "%[1]d_vde_menu" ("id","name","title","value","conditions") VALUES('2','admin_menu','Admin menu','MenuItem(
    Icon: "icon-screen-desktop",
    Page: "interface",
    Vde: "true",
    Title: "Interface"
)
MenuItem(
    Icon: "icon-docs",
    Page: "tables",
    Vde: "true",
    Title: "Tables"
)
MenuItem(
    Icon: "icon-briefcase",
    Page: "contracts",
    Vde: "true",
    Title: "Smart Contracts"
)
MenuItem(
    Icon: "icon-settings",
    Page: "parameters",
    Vde: "true",
    Title: "Ecosystem parameters"
)
MenuItem(
    Icon: "icon-globe",
    Page: "languages",
    Vde: "true",
    Title: "Language resources"
)
MenuItem(
    Icon: "icon-cloud-upload",
    Page: "import",
    Vde: "true",
    Title: "Import"
)
MenuItem(
    Icon: "icon-cloud-download",
    Page: "export",
    Vde: "true",
    Title: "Export"
)','true');

	  DROP TABLE IF EXISTS "%[1]d_vde_pages"; CREATE TABLE "%[1]d_vde_pages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "menu" character varying(255) NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_pages" ADD CONSTRAINT "%[1]d_vde_pages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_pages_index_name" ON "%[1]d_vde_pages" (name);

	  INSERT INTO "%[1]d_vde_pages" ("id","name","value","menu","conditions") VALUES('2','admin_index','','admin_menu','true');

	  DROP TABLE IF EXISTS "%[1]d_vde_blocks"; CREATE TABLE "%[1]d_vde_blocks" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_blocks" ADD CONSTRAINT "%[1]d_vde_blocks_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_blocks_index_name" ON "%[1]d_vde_blocks" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_signatures"; CREATE TABLE "%[1]d_vde_signatures" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "value" jsonb,
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_signatures" ADD CONSTRAINT "%[1]d_vde_signatures_pkey" PRIMARY KEY (name);
	  
	  CREATE TABLE "%[1]d_vde_contracts" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "value" text  NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_contracts" ADD CONSTRAINT "%[1]d_vde_contracts_pkey" PRIMARY KEY (id);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_parameters";
	  CREATE TABLE "%[1]d_vde_parameters" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(255) UNIQUE NOT NULL DEFAULT '',
	  "value" text NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_parameters" ADD CONSTRAINT "%[1]d_vde_parameters_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_vde_parameters_index_name" ON "%[1]d_vde_parameters" (name);
	  
	  INSERT INTO "%[1]d_vde_parameters" ("id","name", "value", "conditions") VALUES 
	  ('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
	  ('2','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('3','new_column', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('4','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('5','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('6','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('7','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('8','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('9','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('10','stylesheet', 'body { 
		/* You can define your custom styles here or create custom CSS rules */
	  }', 'ContractConditions("MainCondition")');

	  DROP TABLE IF EXISTS "%[1]d_vde_cron";
	  CREATE TABLE "%[1]d_vde_cron" (
		  "id"        bigint NOT NULL DEFAULT '0',
		  "owner"	  bigint NOT NULL DEFAULT '0',
		  "cron"      varchar(255) NOT NULL DEFAULT '',
		  "contract"  varchar(255) NOT NULL DEFAULT '',
		  "counter"   bigint NOT NULL DEFAULT '0',
		  "till"      timestamp NOT NULL DEFAULT timestamp '1970-01-01 00:00:00',
		  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_cron" ADD CONSTRAINT "%[1]d_vde_cron_pkey" PRIMARY KEY ("id");


	  CREATE TABLE "%[1]d_vde_tables" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(100) UNIQUE NOT NULL DEFAULT '',
	  "permissions" jsonb,
	  "columns" jsonb,
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_tables" ADD CONSTRAINT "%[1]d_vde_tables_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_vde_tables_index_name" ON "%[1]d_vde_tables" (name);
	  
	  INSERT INTO "%[1]d_vde_tables" ("id", "name", "permissions","columns", "conditions") VALUES ('1', 'contracts', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"value": "ContractConditions(\"MainCondition\")",
				"conditions": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("EditTable")'),
			  ('2', 'languages', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{ "name": "ContractConditions(\"MainCondition\")",
				"res": "ContractConditions(\"MainCondition\")",
				"conditions": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("EditTable")'),
			  ('3', 'menu', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"name": "ContractConditions(\"MainCondition\")",
		  "value": "ContractConditions(\"MainCondition\")",
		  "conditions": "ContractConditions(\"MainCondition\")"
			  }', 'ContractAccess("EditTable")'),
			  ('4', 'pages', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"name": "ContractConditions(\"MainCondition\")",
		  "value": "ContractConditions(\"MainCondition\")",
		  "menu": "ContractConditions(\"MainCondition\")",
		  "conditions": "ContractConditions(\"MainCondition\")"
			  }', 'ContractAccess("EditTable")'),
			  ('5', 'blocks', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"name": "ContractConditions(\"MainCondition\")",
		  "value": "ContractConditions(\"MainCondition\")",
		  "conditions": "ContractConditions(\"MainCondition\")"
			  }', 'ContractAccess("EditTable")'),
			  ('6', 'signatures', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"name": "ContractConditions(\"MainCondition\")",
		  "value": "ContractConditions(\"MainCondition\")",
		  "conditions": "ContractConditions(\"MainCondition\")"
			  }', 'ContractAccess("EditTable")'),
			  ('7', 'cron',
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")",
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"owner": "ContractConditions(\"MainCondition\")",
				"cron": "ContractConditions(\"MainCondition\")",
				"contract": "ContractConditions(\"MainCondition\")",
				"counter": "ContractConditions(\"MainCondition\")",
				"till": "ContractConditions(\"MainCondition\")",
                  "conditions": "ContractConditions(\"MainCondition\")"
				}', 'ContractConditions(\"MainCondition\")');
	  
	  INSERT INTO "%[1]d_vde_contracts" ("id", "value", "conditions") VALUES 
	  ('1','contract MainCondition {
		conditions {
		  if EcosysParam("founder_account")!=$key_id
		  {
			warning "Sorry, you do not have access to this action."
		  }
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('2','contract VDEFunctions {}
	  
		func DBFind(table string).Columns(columns string).Where(where string, params ...)
			.WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
			return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
		}

		func One(list array, name string) string {
			if list {
				var row map 
				row = list[0]
				return row[name]
			}
			return nil
		}

		func Row(list array) map {
			var ret map
			if list {
				ret = list[0]
			}
			return ret
		}

		func DBRow(table string).Columns(columns string).Where(where string, params ...)
			.WhereId(id int).Order(order string).Ecosystem(ecosystem int) map {

			var result array
			result = DBFind(table).Columns(columns).Where(where, params ...).WhereId(id).Order(order).Ecosystem(ecosystem)

			var row map
			if Len(result) > 0 {
				row = result[0]
			}

			return row
		}

		func ConditionById(table string, validate bool) {
			var row map
			row = DBRow(table).Columns("conditions").WhereId($Id)
			if !row["conditions"] {
				error Sprintf("Item %%d has not been found", $Id)
			}

			Eval(row["conditions"])

			if validate {
				ValidateCondition($Conditions,$ecosystem_id)
			}
		}
	  ', 'ContractConditions("MainCondition")'),
	  ('3','contract NewContract {
		  data {
			  Value      string
			  Conditions string
		  }
		  conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			  var list array
			  list = ContractsList($Value)
			  var i int
			  while i < Len(list) {
				  if IsObject(list[i], $ecosystem_id) {
					  warning Sprintf("Contract or function %%s exists", list[i] )
				  }
				  i = i + 1
			  }
		  }
		  action {
			  var root, id int
			  root = CompileContract($Value, $ecosystem_id, 0, 0)
			  id = DBInsert("contracts", "value,conditions", $Value, $Conditions )
			  FlushContract(root, id, false)
			  $result = id
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('4','contract EditContract {
		  data {
			  Id         int
			  Value      string
			  Conditions string
		  }
		  conditions {
			  RowConditions("contracts", $Id)
			  ValidateCondition($Conditions, $ecosystem_id)

			  var row array
			  row = DBFind("contracts").Columns("id,value,conditions").WhereId($Id)
			  if !Len(row) {
				  error Sprintf("Contract %%d does not exist", $Id)
			  }
			  $cur = row[0]

			  var list, curlist array
			  list = ContractsList($Value)
			  curlist = ContractsList($cur["value"])
			  if Len(list) != Len(curlist) {
				  error "Contracts cannot be removed or inserted"
			  }
			  var i int
			  while i < Len(list) {
				  var j int
				  var ok bool
				  while j < Len(curlist) {
					  if curlist[j] == list[i] {
						  ok = true
						  break
					  }
					  j = j + 1 
				  }
				  if !ok {
					  error "Contracts or functions names cannot be changed"
				  }
				  i = i + 1
			  }
		  }
		  action {
			  var root int
			  root = CompileContract($Value, $ecosystem_id, 0, 0)
			  DBUpdate("contracts", $Id, "value,conditions", $Value, $Conditions)
			  FlushContract(root, $Id, false)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('5','contract NewParameter {
		  data {
			  Name string
			  Value string
			  Conditions string
		  }
		  conditions {
			  var ret array
			  ValidateCondition($Conditions, $ecosystem_id)
			  ret = DBFind("parameters").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Parameter %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("parameters", "name,value,conditions", $Name, $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('6','contract EditParameter {
		  data {
			  Id int
			  Value string
			  Conditions string
		  }
		  conditions {
			  RowConditions("parameters", $Id)
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  action {
			  DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('7', 'contract NewMenu {
		  data {
			  Name       string
			  Value      string
			  Title      string "optional"
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("menu").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Menu %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('8','contract EditMenu {
		  data {
			  Id         int
			  Value      string
			  Title      string "optional"
			  Conditions string
		  }
		  conditions {
			  RowConditions("menu", $Id)
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  action {
			  DBUpdate("menu", $Id, "value,title,conditions", $Value, $Title, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('9','contract AppendMenu {
		data {
			Id     int
			Value  string
		}
		conditions {
			RowConditions("menu", $Id)
		}
		action {
			var row map
			row = DBRow("menu").Columns("value").WhereId($Id)
			DBUpdate("menu", $Id, "value", row["value"] + "\r\n" + $Value)
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('10','contract NewPage {
		  data {
			  Name       string
			  Value      string
			  Menu       string
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("pages").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Page %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("pages", "name,value,menu,conditions", $Name, $Value, $Menu, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('11','contract EditPage {
		  data {
			  Id         int
			  Value      string
			  Menu      string
			  Conditions string
		  }
		  conditions {
			  RowConditions("pages", $Id)
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  action {
			  DBUpdate("pages", $Id, "value,menu,conditions", $Value, $Menu, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('12','contract AppendPage {
		  data {
			  Id         int
			  Value      string
		  }
		  conditions {
			  RowConditions("pages", $Id)
		  }
		  action {
			  var row map
			  row = DBRow("pages").Columns("value").WhereId($Id)
			  DBUpdate("pages", $Id, "value", row["value"] + "\r\n" + $Value)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('13','contract NewBlock {
		  data {
			  Name       string
			  Value      string
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("blocks").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Block %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("blocks", "name,value,conditions", $Name, $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('14','contract EditBlock {
		  data {
			  Id         int
			  Value      string
			  Conditions string
		  }
		  conditions {
			  RowConditions("blocks", $Id)
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  action {
			  DBUpdate("blocks", $Id, "value,conditions", $Value, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('15','contract NewTable {
		  data {
			  Name       string
			  Columns      string
			  Permissions string
		  }
		  conditions {
			  TableConditions($Name, $Columns, $Permissions)
		  }
		  action {
			  CreateTable($Name, $Columns, $Permissions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('16','contract EditTable {
		  data {
			  Name       string
			  Permissions string
		  }
		  conditions {
			  TableConditions($Name, "", $Permissions)
		  }
		  action {
			  PermTable($Name, $Permissions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('17','contract NewColumn {
		  data {
			  TableName   string
			  Name        string
			  Type        string
			  Permissions string
		  }
		  conditions {
			  ColumnCondition($TableName, $Name, $Type, $Permissions)
		  }
		  action {
			  CreateColumn($TableName, $Name, $Type, $Permissions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('18','contract EditColumn {
		  data {
			  TableName   string
			  Name        string
			  Permissions string
		  }
		  conditions {
			  ColumnCondition($TableName, $Name, "", $Permissions)
		  }
		  action {
			  PermColumn($TableName, $Name, $Permissions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('19','contract NewLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
			var row array
			row = DBFind("languages").Columns("name").Where("name=?", $Name).Limit(1)
			if Len(row) > 0 {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res", $Name, $Trans )
			UpdateLang($Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('20','contract EditLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdateExt("languages", "name", $Name, "res", $Trans )
			UpdateLang($Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('21','func ImportList(row array, cnt string) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			idata = row[i]

			if(cnt == "pages"){
				$ret_page = DBFind("pages").Columns("id").Where("name=$", idata["Name"])
				$page_id = One($ret_page, "id") 
				if ($page_id != nil){
					idata["Id"] = Int($page_id) 
					CallContract("EditPage", idata)
				} else {
					CallContract("NewPage", idata)
				}
			}
			if(cnt == "blocks"){
				$ret_block = DBFind("blocks").Columns("id").Where("name=$", idata["Name"])
				$block_id = One($ret_block, "id") 
				if ($block_id != nil){
					idata["Id"] = Int($block_id)
					CallContract("EditBlock", idata)
				} else {
					CallContract("NewBlock", idata)
				}
			}
			if(cnt == "menus"){
				$ret_menu = DBFind("menu").Columns("id,value").Where("name=$", idata["Name"])
				$menu_id = One($ret_menu, "id") 
				$menu_value = One($ret_menu, "value") 
				if ($menu_id != nil){
					idata["Id"] = Int($menu_id)
					idata["Value"] = Str($menu_value) + "\n" + Str(idata["Value"])
					CallContract("EditMenu", idata)
				} else {
					CallContract("NewMenu", idata)
				}
			}
			if(cnt == "parameters"){
				$ret_param = DBFind("parameters").Columns("id").Where("name=$", idata["Name"])
				$param_id = One($ret_param, "id")
				if ($param_id != nil){ 
					idata["Id"] = Int($param_id) 
					CallContract("EditParameter", idata)
				} else {
					CallContract("NewParameter", idata)
				}
			}
			if(cnt == "languages"){
				$ret_lang = DBFind("languages").Columns("id").Where("name=$", idata["Name"])
				$lang_id = One($ret_lang, "id")
				if ($lang_id != nil){
					CallContract("EditLang", idata)
				} else {
					CallContract("NewLang", idata)
				}
			}
			if(cnt == "contracts"){
				if IsObject(idata["Name"], $ecosystem_id){
				} else {
					CallContract("NewContract", idata)
				} 
			}
			if(cnt == "tables"){
				$ret_table = DBFind("tables").Columns("id").Where("name=$", idata["Name"])
				$table_id = One($ret_table, "id")
				if ($table_id != nil){	
				} else {
					CallContract("NewTable", idata)
				}
			}

			i = i + 1
		}
	}
	
	func ImportData(row array) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			var list array
			var tblname, columns string
			idata = row[i]
			i = i + 1
			tblname = idata["Table"]
			columns = Join(idata["Columns"], ",")
			list = idata["Data"] 
			if !list {
				continue
			}
			var j int
			while j < Len(list) {
				var ilist array
				ilist = list[j]
				DBInsert(tblname, columns, ilist)
				j=j+1
			}
		}
	}
	
	contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONToMap($Data)
		}
		action {
			ImportList($list["pages"], "pages")
			ImportList($list["blocks"], "blocks")
			ImportList($list["menus"], "menus")
			ImportList($list["parameters"], "parameters")
			ImportList($list["languages"], "languages")
			ImportList($list["contracts"], "contracts")
			ImportList($list["tables"], "tables")
			ImportData($list["data"])
		}
	}', 'ContractConditions("MainCondition")'),
	('22', 'contract NewCron {
		data {
			Cron       string
			Contract   string
			Limit      int "optional"
			Till       string "optional date"
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			ValidateCron($Cron)
		}
		action {
			if !$Till {
				$Till = "1970-01-01 00:00:00"
			}
			if !HasPrefix($Contract, "@") {
				$Contract = "@" + Str($ecosystem_id) + $Contract
			}
			$result = DBInsert("cron", "owner,cron,contract,counter,till,conditions",
				$key_id, $Cron, $Contract, $Limit, $Till, $Conditions)
			UpdateCron($result)
		}
	}', 'ContractConditions("MainCondition")'),
	('23','contract EditCron {
		data {
			Id         int
			Contract   string
			Cron       string "optional"
			Limit      int "optional"
			Till       string "optional date"
			Conditions string
		}
		conditions {
			ConditionById("cron", true)
			ValidateCron($Cron)
		}
		action {
			if !$Till {
				$Till = "1970-01-01 00:00:00"
			}
			if !HasPrefix($Contract, "@") {
				$Contract = "@" + Str($ecosystem_id) + $Contract
			}
			DBUpdate("cron", $Id, "cron,contract,counter,till,conditions",
				$Cron, $Contract, $Limit, $Till, $Conditions)
			UpdateCron($Id)
		}
	}', 'ContractConditions("MainCondition")');
	`

	SchemaEcosystem = `DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
		"id" bigint  NOT NULL DEFAULT '0',
		"pub" bytea  NOT NULL DEFAULT '',
		"amount" decimal(30) NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "%[1]d_history"; CREATE TABLE "%[1]d_history" (
		"id" bigint NOT NULL  DEFAULT '0',
		"sender_id" bigint NOT NULL DEFAULT '0',
		"recipient_id" bigint NOT NULL DEFAULT '0',
		"amount" decimal(30) NOT NULL DEFAULT '0',
		"comment" text NOT NULL DEFAULT '',
		"block_id" int  NOT NULL DEFAULT '0',
		"txhash" bytea  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_history" ADD CONSTRAINT "%[1]d_history_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_history_index_sender" ON "%[1]d_history" (sender_id);
		CREATE INDEX "%[1]d_history_index_recipient" ON "%[1]d_history" (recipient_id);
		CREATE INDEX "%[1]d_history_index_block" ON "%[1]d_history" (block_id, txhash);
		
		
		DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "res" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_languages_index_name" ON "%[1]d_languages" (name);
		
		DROP TABLE IF EXISTS "%[1]d_sections"; CREATE TABLE "%[1]d_sections" (
		"id" bigint  NOT NULL DEFAULT '0',
		"title" varchar(255)  NOT NULL DEFAULT '',
		"urlname" varchar(255) NOT NULL DEFAULT '',
		"page" varchar(255) NOT NULL DEFAULT '',
		"roles_access" text NOT NULL DEFAULT '',
		"delete" bigint NOT NULL DEFAULT '0'
		);
	  ALTER TABLE ONLY "%[1]d_sections" ADD CONSTRAINT "%[1]d_sections_pkey" PRIMARY KEY (id);

        INSERT INTO "%[1]d_sections" ("id","title","urlname","page","roles_access", "delete") 
	            VALUES('1', 'Home', 'home', 'default_page', '', 0);

		DROP TABLE IF EXISTS "%[1]d_menu"; CREATE TABLE "%[1]d_menu" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"title" character varying(255) NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_menu_index_name" ON "%[1]d_menu" (name);

		INSERT INTO "%[1]d_menu" ("id","name","title","value","conditions") VALUES('2','admin_menu','Admin menu','MenuItem(
    Icon: "icon-screen-desktop",
    Page: "interface",
    Title: "Interface"
)
MenuItem(
    Icon: "icon-docs",
    Page: "tables",
    Title: "Tables"
)
MenuItem(
    Icon: "icon-briefcase",
    Page: "contracts",
    Title: "Smart Contracts"
)
MenuItem(
    Icon: "icon-settings",
    Page: "parameters",
    Title: "Ecosystem parameters"
)
MenuItem(
    Icon: "icon-globe",
    Page: "languages",
    Title: "Language resources"
)
MenuItem(
    Icon: "icon-cloud-upload",
    Page: "import",
    Title: "Import"
)
MenuItem(
    Icon: "icon-cloud-download",
    Page: "export",
    Title: "Export"
)
If("#key_id#" == EcosysParam("founder_account")){
    MenuItem(
        Icon: "icon-lock",
        Page: "vde",
        Title: "Dedicated Ecosystem"
    )
}','true');

		DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"menu" character varying(255) NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);


		INSERT INTO "%[1]d_pages" ("id","name","value","menu","conditions") VALUES('2','admin_index','','admin_menu','true');



		DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_blocks" ADD CONSTRAINT "%[1]d_blocks_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_blocks_index_name" ON "%[1]d_blocks" (name);
		
		DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(100) NOT NULL DEFAULT '',
			"value" jsonb,
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
		
		CREATE TABLE "%[1]d_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '1',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);
		
		INSERT INTO "%[1]d_contracts" ("id", "value", "wallet_id","active", "conditions") VALUES 
		('1','contract MainCondition {
		  conditions {
			if EcosysParam("founder_account")!=$key_id
			{
			  warning "Sorry, you do not have access to this action."
			}
		  }
		}', '%[2]d', '0', 'ContractConditions("MainCondition")');
		
		DROP TABLE IF EXISTS "%[1]d_parameters";
		CREATE TABLE "%[1]d_parameters" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_parameters_index_name" ON "%[1]d_parameters" (name);
		
		INSERT INTO "%[1]d_parameters" ("id","name", "value", "conditions") VALUES 
		('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
		('2','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('3','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('4','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('5','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('6','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('7','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('8','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('9','ecosystem_name', '%[3]s', 'ContractConditions("MainCondition")'),
		('10','max_sum', '1000000', 'ContractConditions("MainCondition")'),
		('11','money_digit', '2', 'ContractConditions("MainCondition")'),
		('12','stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
		}', 'ContractConditions("MainCondition")');
		
		DROP TABLE IF EXISTS "%[1]d_tables";
		CREATE TABLE "%[1]d_tables" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_tables_index_name" ON "%[1]d_tables" (name);
		
		INSERT INTO "%[1]d_tables" ("id", "name", "permissions","columns", "conditions") VALUES 
			('1', 'contracts', '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", "new_column": "ContractConditions(\"MainCondition\")"}', 
			'{"value": "ContractConditions(\"MainCondition\")",
				  "wallet_id": "ContractConditions(\"MainCondition\")",
				  "token_id": "ContractConditions(\"MainCondition\")",
				  "active": "ContractConditions(\"MainCondition\")",
				  "conditions": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("@1EditTable")'),
				('2', 'keys', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"pub": "ContractConditions(\"MainCondition\")",
				  "amount": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("@1EditTable")'),
				('3', 'history', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"sender_id": "ContractConditions(\"MainCondition\")",
				  "recipient_id": "ContractConditions(\"MainCondition\")",
				  "amount":  "ContractConditions(\"MainCondition\")",
				  "comment": "ContractConditions(\"MainCondition\")",
				  "block_id":  "ContractConditions(\"MainCondition\")",
				  "txhash": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("@1EditTable")'),        
				('4', 'languages', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{ "name": "ContractConditions(\"MainCondition\")",
				  "res": "ContractConditions(\"MainCondition\")",
				  "conditions": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("@1EditTable")'),
				('5', 'menu', 
					'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"name": "ContractConditions(\"MainCondition\")",
			"value": "ContractConditions(\"MainCondition\")",
			"conditions": "ContractConditions(\"MainCondition\")"
				}', 'ContractAccess("@1EditTable")'),
				('6', 'pages', 
					'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"name": "ContractConditions(\"MainCondition\")",
			"value": "ContractConditions(\"MainCondition\")",
			"menu": "ContractConditions(\"MainCondition\")",
			"conditions": "ContractConditions(\"MainCondition\")"
				}', 'ContractAccess("@1EditTable")'),
				('7', 'blocks', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"name": "ContractConditions(\"MainCondition\")",
			"value": "ContractConditions(\"MainCondition\")",
			"conditions": "ContractConditions(\"MainCondition\")"
				}', 'ContractAccess("@1EditTable")'),
				('8', 'signatures', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"name": "ContractConditions(\"MainCondition\")",
			"value": "ContractConditions(\"MainCondition\")",
			"conditions": "ContractConditions(\"MainCondition\")"
				}', 'ContractAccess("@1EditTable")'),
				('9', 'member', 
					'{"insert": "ContractAccess(\"Profile_Edit\")", "update": "ContractAccess(\"Profile_Edit\")", 
					  "new_column": "ContractConditions(\"MainCondition\")"}',
					'{"member_name": "ContractAccess(\"Profile_Edit\")",
					  "avatar": "ContractAccess(\"Profile_Edit\")"}', 'ContractConditions(\"MainCondition\")'),
				('10', 'roles_list', 
					'{"insert": "ContractAccess(\"Roles_Create\")", "update": "ContractAccess(\"Roles_Del\")", 
					 "new_column": "ContractConditions(\"MainCondition\")"}',
					'{"default_page": "false",
					  "role_name": "false",
					  "delete": "ContractAccess(\"Roles_Del\")",
					  "role_type": "false",
					  "creator_id": "false",
					  "date_create": "false",
					  "date_delete": "ContractAccess(\"Roles_Del\")",
					  "creator_name": "false",
					  "creator_avatar": "false"}',
					   'ContractConditions(\"MainCondition\")'),
				('11', 'roles_assign', 
					'{"insert": "ContractAccess(\"Roles_Assign\", \"voting_CheckDecision\")", "update": "ContractAccess(\"Roles_Unassign\")", 
					"new_column": "ContractConditions(\"MainCondition\")"}',
					'{"role_id": "false",
						"role_type": "false",
						"role_name": "false",
						"member_id": "false",
						"member_name": "false",
						"member_avatar": "false",
						"appointed_by_id": "false",
						"appointed_by_name": "false",
						"date_start": "false",
						"date_end": "ContractAccess(\"Roles_Unassign\")",
						"delete": "ContractAccess(\"Roles_Unassign\")"}', 
						'ContractConditions(\"MainCondition\")'),
				('12', 'notifications', 
						'{"insert": "ContractAccess(\"Notifications_Single_Send\",\"Notifications_Roles_Send\")", "update": "true", 
						"new_column": "ContractConditions(\"MainCondition\")"}',
						'{"icon": "false",
							"started_processing_time": "ContractAccess(\"Notifications_Roles_Processing\")",
							"date_create": "false",
							"page_params": "ContractAccess(\"Notifications_Single_Send\",\"Notifications_Roles_Send\")",
							"body_text": "false",
							"recipient_id": "false",
							"started_processing_id": "ContractAccess(\"Notifications_Roles_Processing\")",
							"role_id": "false",
							"role_name": "false",
							"recipient_name": "false",
							"closed": "ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Finishing\")", 
							"header_text": "false", 
							"recipient_avatar": "false", 
							"notification_type": "false", 
							"finished_processing_id": "ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Finishing\")", 
							"finished_processing_time": "ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Finishing\")", 
							"page_name": "false"}', 
							'ContractAccess(\"@1EditTable\")'),
				('13', 'sections', 
					'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
					"new_column": "ContractConditions(\"MainCondition\")"}',
					'{"title": "ContractConditions(\"MainCondition\")",
						"urlname": "ContractConditions(\"MainCondition\")",
						"page": "ContractConditions(\"MainCondition\")",
						"roles_access": "ContractConditions(\"MainCondition\")",
						"delete": "ContractConditions(\"MainCondition\")"}', 
						'ContractConditions(\"MainCondition\")');

		DROP TABLE IF EXISTS "%[1]d_notifications";
		CREATE TABLE "%[1]d_notifications" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"icon"	varchar(255) NOT NULL DEFAULT '',
			"closed" bigint NOT NULL DEFAULT '0',
			"notification_type"	bigint NOT NULL DEFAULT '0',
			"started_processing_time" timestamp,
			"page_name"	varchar(255) NOT NULL DEFAULT '',
			"recipient_avatar"	bytea NOT NULL DEFAULT '',
			"date_create"	timestamp,
			"page_params"	text NOT NULL DEFAULT '',
			"recipient_name" varchar(255) NOT NULL DEFAULT '',
			"finished_processing_id" bigint NOT NULL DEFAULT '0',
			"finished_processing_time" timestamp,
			"role_id"	bigint NOT NULL DEFAULT '0',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"recipient_id"	bigint NOT NULL DEFAULT '0',
			"started_processing_id"	bigint NOT NULL DEFAULT '0',
			"body_text"	text NOT NULL DEFAULT '',
			"header_text"	text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_notifications" ADD CONSTRAINT "%[1]d_notifications_pkey" PRIMARY KEY ("id");


		DROP TABLE IF EXISTS "%[1]d_roles_list";
		CREATE TABLE "%[1]d_roles_list" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"default_page"	varchar(255) NOT NULL DEFAULT '',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"delete"    bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"creator_id" bigint NOT NULL DEFAULT '0',
			"date_create" timestamp,
			"date_delete" timestamp,
			"creator_name"	varchar(255) NOT NULL DEFAULT '',
			"creator_avatar" bytea NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_roles_list" ADD CONSTRAINT "%[1]d_roles_list_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_roles_list_index_delete" ON "%[1]d_roles_list" (delete);
		CREATE INDEX "%[1]d_roles_list_index_type" ON "%[1]d_roles_list" (role_type);

		INSERT INTO "%[1]d_roles_list" ("id", "default_page", "role_name", "delete", "role_type",
			"date_create","creator_name") VALUES('1','default_ecosystem_page', 
				'Admin', '0', '3', NOW(), '');


		DROP TABLE IF EXISTS "%[1]d_roles_assign";
		CREATE TABLE "%[1]d_roles_assign" (
			"id" bigint NOT NULL DEFAULT '0',
			"role_id" bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"member_id" bigint NOT NULL DEFAULT '0',
			"member_name" varchar(255) NOT NULL DEFAULT '',
			"member_avatar"	bytea NOT NULL DEFAULT '',
			"appointed_by_id" bigint NOT NULL DEFAULT '0',
			"appointed_by_name"	varchar(255) NOT NULL DEFAULT '',
			"date_start" timestamp,
			"date_end" timestamp,
			"delete" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_roles_assign" ADD CONSTRAINT "%[1]d_roles_assign_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_roles_assign_index_role" ON "%[1]d_roles_assign" (role_id);
		CREATE INDEX "%[1]d_roles_assign_index_type" ON "%[1]d_roles_assign" (role_type);
		CREATE INDEX "%[1]d_roles_assign_index_member" ON "%[1]d_roles_assign" (member_id);

		INSERT INTO "%[1]d_roles_assign" ("id","role_id","role_type","role_name","member_id",
			"member_name","date_start") VALUES('1','1','3','Admin','%[4]d','founder', NOW());


		DROP TABLE IF EXISTS "%[1]d_member";
		CREATE TABLE "%[1]d_member" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"avatar"	bytea NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_member" ADD CONSTRAINT "%[1]d_member_pkey" PRIMARY KEY ("id");

		INSERT INTO "%[1]d_member" ("id", "member_name") VALUES('%[4]d', 'founder');

		`

	SchemaFirstEcosystem = `INSERT INTO "system_states" ("id") VALUES ('1');

	INSERT INTO "1_contracts" ("id","value", "wallet_id", "conditions") VALUES 
	('2','contract SystemFunctions {
	}
	
	func DBFind(table string).Columns(columns string).Where(where string, params ...)
		 .WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
		return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
	}

	func One(list array, name string) string {
		if list {
			var row map 
			row = list[0]
			return row[name]
		}
		return nil
	}
	
	func Row(list array) map {
		var ret map
		if list {
			ret = list[0]
		}
		return ret
	}

	func DBRow(table string).Columns(columns string).Where(where string, params ...)
		.WhereId(id int).Order(order string).Ecosystem(ecosystem int) map {
		
		var result array
		result = DBFind(table).Columns(columns).Where(where, params ...).WhereId(id).Order(order).Ecosystem(ecosystem)

		var row map
		if Len(result) > 0 {
			row = result[0]
		}

		return row
	}
	
	func ConditionById(table string, validate bool) {
		var row map
		row = DBRow(table).Columns("conditions").WhereId($Id)
		if !row["conditions"] {
			error Sprintf("Item %%d has not been found", $Id)
		}

		Eval(row["conditions"])

		if validate {
			ValidateCondition($Conditions,$ecosystem_id)
		}
	}
	
	', '%[1]d','ContractConditions("MainCondition")'),
	('3','contract MoneyTransfer {
		data {
			Recipient string
			Amount    string
			Comment     string "optional"
		}
		conditions {
			$recipient = AddressToId($Recipient)
			if $recipient == 0 {
				error Sprintf("Recipient %%s is invalid", $Recipient)
			}
			var total money
			$amount = Money($Amount) 
			if $amount == 0 {
				error "Amount is zero"
			}
			var row map
			row = DBRow("keys").Columns("amount").WhereId($key_id)
			total = Money(row["amount"])
			if $amount >= total {
				error Sprintf("Money is not enough %%v < %%v",total, $amount)
			}
		}
		action {
			DBUpdate("keys", $key_id,"-amount", $amount)
			DBUpdate("keys", $recipient,"+amount", $amount)
			DBInsert("history", "sender_id,recipient_id,amount,comment,block_id,txhash", 
				$key_id, $recipient, $amount, $Comment, $block, $txhash)
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('4','contract NewContract {
		data {
			Value      string
			Conditions string
			Wallet         string "optional"
			TokenEcosystem int "optional"
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			$walletContract = $key_id
			   if $Wallet {
				$walletContract = AddressToId($Wallet)
				if $walletContract == 0 {
				   error Sprintf("wrong wallet %%s", $Wallet)
				}
			}
			var list array
			list = ContractsList($Value)
			var i int
			while i < Len(list) {
				if IsObject(list[i], $ecosystem_id) {
					warning Sprintf("Contract or function %%s exists", list[i] )
				}
				i = i + 1
			}
			if !$TokenEcosystem {
				$TokenEcosystem = 1
			} else {
				if !SysFuel($TokenEcosystem) {
					warning Sprintf("Ecosystem %%d is not system", $TokenEcosystem )
				}
			}
		}
		action {
			var root, id int
			root = CompileContract($Value, $ecosystem_id, $walletContract, $TokenEcosystem)
			id = DBInsert("contracts", "value,conditions, wallet_id, token_id", 
				   $Value, $Conditions, $walletContract, $TokenEcosystem)
			FlushContract(root, id, false)
			$result = id
		}
		func price() int {
			return  SysParamInt("contract_price")
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('5','contract EditContract {
		data {
			Id         int
			Value      string
			Conditions string
			WalletId   string "optional"
		}
		conditions {
			RowConditions("contracts", $Id)
			ValidateCondition($Conditions, $ecosystem_id)

			$cur = DBRow("contracts").Columns("id,value,conditions,active,wallet_id,token_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}

			var list, curlist array
			list = ContractsList($Value)
			curlist = ContractsList($cur["value"])
			if Len(list) != Len(curlist) {
				error "Contracts cannot be removed or inserted"
			}
			var i int
			while i < Len(list) {
				var j int
				var ok bool
				while j < Len(curlist) {
					if curlist[j] == list[i] {
						ok = true
						break
					}
					j = j + 1 
				}
				if !ok {
					error "Contracts or functions names cannot be changed"
				}
				i = i + 1
			}
			if $WalletId != "" {
				$recipient = AddressToId($WalletId)
				if $recipient == 0 {
					error Sprintf("New contract owner %%s is invalid", $WalletId)
				}
				if Int($cur["active"]) == 1 {
					error "Contract must be deactivated before wallet changing"
				}
			} else {
				$recipient = Int($cur["wallet_id"])
			}
		}
		action {
			var root int
			root = CompileContract($Value, $ecosystem_id, Int($cur["wallet_id"]), Int($cur["token_id"]))
			DBUpdate("contracts", $Id, "value,conditions,wallet_id", $Value, $Conditions, $recipient)
			FlushContract(root, $Id, Int($cur["active"]) == 1)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('6','contract ActivateContract {
		data {
			Id  int
		}
		conditions {
			$cur = DBRow("contracts").Columns("id,conditions,active,wallet_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if Int($cur["active"]) == 1 {
				error Sprintf("The contract %%d has been already activated", $Id)
			}
			Eval($cur["conditions"])
			if $key_id != Int($cur["wallet_id"]) {
				error Sprintf("Wallet %%d cannot activate the contract", $key_id)
			}
		}
		action {
			DBUpdate("contracts", $Id, "active", 1)
			Activate($Id, $ecosystem_id)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('7','contract NewEcosystem {
		data {
			Name  string "optional"
		}
		action {
			$result = CreateEcosystem($key_id, $Name)
		}
		func price() int {
			return  SysParamInt("ecosystem_price")
		}
		func rollback() {
			RollbackEcosystem()
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('8','contract NewParameter {
		data {
			Name string
			Value string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions, $ecosystem_id)

			var row map
			row = DBRow("parameters").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Parameter %%s already exists", $Name)
			}
		}
		action {
			DBInsert("parameters", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('9','contract EditParameter {
		data {
			Id int
			Value string
			Conditions string
		}
		conditions {
			RowConditions("parameters", $Id)
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('10', 'contract NewMenu {
		data {
			Name       string
			Value      string
			Title      string "optional"
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("menu").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Menu %%s already exists", $Name)
			}
		}
		action {
			DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions )
		}
		func price() int {
			return  SysParamInt("menu_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('11','contract EditMenu {
		data {
			Id         int
			Value      string
			Title      string "optional"
			Conditions string
		}
		conditions {
			RowConditions("menu", $Id)
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("menu", $Id, "value,title,conditions", $Value, $Title, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('12','contract AppendMenu {
		data {
			Id     int
			Value      string
		}
		conditions {
			ConditionById("menu", false)
		}
		action {
			var row map
			row = DBRow("menu").Columns("value").WhereId($Id)
			DBUpdate("menu", $Id, "value", row["value"] + "\r\n" + $Value)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('13','contract NewPage {
		data {
			Name       string
			Value      string
			Menu       string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("pages").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Page %%s already exists", $Name)
			}
		}
		action {
			DBInsert("pages", "name,value,menu,conditions", $Name, $Value, $Menu, $Conditions )
		}
		func price() int {
			return  SysParamInt("page_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('14','contract EditPage {
		data {
			Id         int
			Value      string
			Menu      string
			Conditions string
		}
		conditions {
			RowConditions("pages", $Id)
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("pages", $Id, "value,menu,conditions", $Value, $Menu, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('15','contract AppendPage {
		data {
			Id         int
			Value      string
		}
		conditions {
			RowConditions("pages", $Id)
		}
		action {
			var value string
			var row map
			row = DBRow("pages").Columns("value").WhereId($Id)
			value = row["value"]
			if Contains(value, "PageEnd:") {
				value = Replace(value, "PageEnd:", $Value) + "\r\nPageEnd:"
			} else {
				value = value + "\r\n" + $Value
			}
			DBUpdate("pages", $Id, "value",  value )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('16','contract NewLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")

			var row map
			row = DBRow("languages").Columns("id").Where("name = ?", $Name)

			if row {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res", $Name, $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('17','contract EditLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdateExt("languages", "name", $Name, "res", $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('18','contract NewSign {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			var exist string

			var row map
			row = DBRow("signatures").Columns("id").Where("name = ?", $Name)

			if row {
				error Sprintf("The signature %%s already exists", $Name)
			}
		}
		action {
			DBInsert("signatures", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('19','contract EditSign {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			RowConditions("signatures", $Id)
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("signatures", $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('20','contract NewBlock {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("blocks").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Block %%s already exists", $Name)
			}
		}
		action {
			DBInsert("blocks", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('21','contract EditBlock {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			RowConditions("blocks", $Id)
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("blocks", $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('22','contract NewTable {
		data {
			Name       string
			Columns      string
			Permissions string
		}
		conditions {
			TableConditions($Name, $Columns, $Permissions)
		}
		action {
			CreateTable($Name, $Columns, $Permissions)
		}
		func rollback() {
			RollbackTable($Name)
		}
		func price() int {
			return  SysParamInt("table_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('23','contract EditTable {
		data {
			Name       string
			Permissions string
		}
		conditions {
			TableConditions($Name, "", $Permissions)
		}
		action {
			PermTable($Name, $Permissions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('24','contract NewColumn {
		data {
			TableName   string
			Name        string
			Type        string
			Permissions string
		}
		conditions {
			ColumnCondition($TableName, $Name, $Type, $Permissions)
		}
		action {
			CreateColumn($TableName, $Name, $Type, $Permissions)
		}
		func rollback() {
			RollbackColumn($TableName, $Name)
		}
		func price() int {
			return  SysParamInt("column_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('25','contract EditColumn {
		data {
			TableName   string
			Name        string
			Permissions string
		}
		conditions {
			ColumnCondition($TableName, $Name, "", $Permissions)
		}
		action {
			PermColumn($TableName, $Name, $Permissions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('26','func ImportList(row array, cnt string) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			idata = row[i]

			if(cnt == "pages"){
				$ret_page = DBFind("pages").Columns("id").Where("name=$", idata["Name"])
				$page_id = One($ret_page, "id") 
				if ($page_id != nil){
					idata["Id"] = Int($page_id) 
					CallContract("EditPage", idata)
				} else {
					CallContract("NewPage", idata)
				}
			}
			if(cnt == "blocks"){
				$ret_block = DBFind("blocks").Columns("id").Where("name=$", idata["Name"])
				$block_id = One($ret_block, "id") 
				if ($block_id != nil){
					idata["Id"] = Int($block_id)
					CallContract("EditBlock", idata)
				} else {
					CallContract("NewBlock", idata)
				}
			}
			if(cnt == "menus"){
				$ret_menu = DBFind("menu").Columns("id,value").Where("name=$", idata["Name"])
				$menu_id = One($ret_menu, "id") 
				$menu_value = One($ret_menu, "value") 
				if ($menu_id != nil){
					idata["Id"] = Int($menu_id)
					idata["Value"] = Str($menu_value) + "\n" + Str(idata["Value"])
					CallContract("EditMenu", idata)
				} else {
					CallContract("NewMenu", idata)
				}
			}
			if(cnt == "parameters"){
				$ret_param = DBFind("parameters").Columns("id").Where("name=$", idata["Name"])
				$param_id = One($ret_param, "id")
				if ($param_id != nil){ 
					idata["Id"] = Int($param_id) 
					CallContract("EditParameter", idata)
				} else {
					CallContract("NewParameter", idata)
				}
			}
			if(cnt == "languages"){
				$ret_lang = DBFind("languages").Columns("id").Where("name=$", idata["Name"])
				$lang_id = One($ret_lang, "id")
				if ($lang_id != nil){
					CallContract("EditLang", idata)
				} else {
					CallContract("NewLang", idata)
				}
			}
			if(cnt == "contracts"){
				if IsObject(idata["Name"], $ecosystem_id){
				} else {
					CallContract("NewContract", idata)
				} 
			}
			if(cnt == "tables"){
				$ret_table = DBFind("tables").Columns("id").Where("name=$", idata["Name"])
				$table_id = One($ret_table, "id")
				if ($table_id != nil){	
				} else {
					CallContract("NewTable", idata)
				}
			}

			i = i + 1
		}
	}
	
	func ImportData(row array) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			var list array
			var tblname, columns string
			idata = row[i]
			i = i + 1
			tblname = idata["Table"]
			columns = Join(idata["Columns"], ",")
			list = idata["Data"] 
			if !list {
				continue
			}
			var j int
			while j < Len(list) {
				var ilist array
				ilist = list[j]
				DBInsert(tblname, columns, ilist)
				j=j+1
			}
		}
	}
	
	contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONToMap($Data)
		}
		action {
			ImportList($list["pages"], "pages")
			ImportList($list["blocks"], "blocks")
			ImportList($list["menus"], "menus")
			ImportList($list["parameters"], "parameters")
			ImportList($list["languages"], "languages")
			ImportList($list["contracts"], "contracts")
			ImportList($list["tables"], "tables")
			ImportData($list["data"])
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('27','contract DeactivateContract {
		data {
			Id         int
		}
		conditions {
			$cur = DBRow("contracts").Columns("id,conditions,active,wallet_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if Int($cur["active"]) == 0 {
				error Sprintf("The contract %%d has been already deactivated", $Id)
			}
			Eval($cur["conditions"])
			if $key_id != Int($cur["wallet_id"]) {
				error Sprintf("Wallet %%d cannot deactivate the contract", $key_id)
			}
		}
		action {
			DBUpdate("contracts", $Id, "active", 0)
			Deactivate($Id, $ecosystem_id)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('28','contract UpdateSysParam {
		data {
			Name  string
			Value string
			Conditions string "optional"
		}
		action {
			DBUpdateSysParam($Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")');`
)
