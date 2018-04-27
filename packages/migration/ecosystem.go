package migration

var (
	SchemaVDE = `
		DROP TABLE IF EXISTS "%[1]d_vde_members";
		CREATE TABLE "%[1]d_vde_members" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"image_id"	bigint,
			"member_info" jsonb
		);
		ALTER TABLE ONLY "%[1]d_vde_members" ADD CONSTRAINT "%[1]d_vde_members_pkey" PRIMARY KEY ("id");

		INSERT INTO "%[1]d_vde_members" ("id", "member_name") VALUES('%[2]d', 'founder');
		INSERT INTO "%[1]d_vde_members" ("id", "member_name") VALUES('4544233900443112470', 'guest');

		DROP TABLE IF EXISTS "%[1]d_vde_languages"; CREATE TABLE "%[1]d_vde_languages" (
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
		  "conditions" text NOT NULL DEFAULT '',
		  "validate_count" bigint NOT NULL DEFAULT '1',
		  "app_id" bigint NOT NULL DEFAULT '0',
		  "validate_mode" character(1) NOT NULL DEFAULT '0'
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
	  "name" text NOT NULL DEFAULT '',
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
	  }', 'ContractConditions("MainCondition")'),
	  ('11','changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")');

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

		DROP TABLE IF EXISTS "%[1]d_vde_binaries";
		CREATE TABLE "%[1]d_vde_binaries" (
			"id" bigint NOT NULL DEFAULT '0',
			"app_id" bigint NOT NULL DEFAULT '1',
			"member_id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"data" bytea NOT NULL DEFAULT '',
			"hash" varchar(32) NOT NULL DEFAULT '',
			"mime_type" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_vde_binaries" ADD CONSTRAINT "%[1]d_vde_binaries_pkey" PRIMARY KEY (id);
		CREATE UNIQUE INDEX "%[1]d_vde_binaries_index_app_id_member_id_name" ON "%[1]d_vde_binaries" (app_id, member_id, name);

	  CREATE TABLE "%[1]d_vde_tables" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(100) UNIQUE NOT NULL DEFAULT '',
	  "permissions" jsonb,
	  "columns" jsonb,
	  "conditions" text  NOT NULL DEFAULT '',
	  "app_id" bigint NOT NULL DEFAULT '1'
	  );
	  ALTER TABLE ONLY "%[1]d_vde_tables" ADD CONSTRAINT "%[1]d_vde_tables_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_vde_tables_index_name" ON "%[1]d_vde_tables" (name);
	  
	  INSERT INTO "%[1]d_vde_tables" ("id", "name", "permissions","columns", "conditions") VALUES ('1', 'contracts', 
			  '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				"new_column": "ContractConditions(\"MainCondition\")"}',
			  '{"name": "false",
				"value": "ContractConditions(\"MainCondition\")",
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
		  "conditions": "ContractConditions(\"MainCondition\")",
		  "validate_count": "ContractConditions(\"MainCondition\")",
		  "validate_mode": "ContractConditions(\"MainCondition\")",
		  "app_id": "ContractConditions(\"MainCondition\")"
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
				}', 'ContractConditions(\"MainCondition\")'),
			  ('8', 'binaries',
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")",
					"new_column": "ContractConditions(\"MainCondition\")"}',
				'{"app_id": "ContractConditions(\"MainCondition\")",
					"member_id": "ContractConditions(\"MainCondition\")",
					"name": "ContractConditions(\"MainCondition\")",
					"data": "ContractConditions(\"MainCondition\")",
					"hash": "ContractConditions(\"MainCondition\")",
					"mime_type": "ContractConditions(\"MainCondition\")"}',
					'ContractConditions(\"MainCondition\")');
	  
	  INSERT INTO "%[1]d_vde_contracts" ("id", "name", "value", "conditions") VALUES 
	  ('1','MainCondition','contract MainCondition {
		conditions {
		  if EcosysParam("founder_account")!=$key_id
		  {
			warning "Sorry, you do not have access to this action."
		  }
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('2','NewContract','contract NewContract {
		data {
			Value      string
			Conditions string
			Wallet         string "optional"
			TokenEcosystem int "optional"
			ApplicationId int "optional"
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

			if Len(list) == 0 {
				error "must be the name"
			}

			var i int
			while i < Len(list) {
				if IsObject(list[i], $ecosystem_id) {
					warning Sprintf("Contract or function %%s exists", list[i] )
				}
				i = i + 1
			}

			$contract_name = list[0]
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
			id = DBInsert("contracts", "name,value,conditions, wallet_id, token_id,app_id",
				   $contract_name, $Value, $Conditions, $walletContract, $TokenEcosystem, $ApplicationId)
			FlushContract(root, id, false)
			$result = id
		}
		func rollback() {
			var list array
    		list = ContractsList($Value)
			var i int
			while i < Len(list) {
				RollbackContract(list[i])
				i = i + 1
			}
		}
		func price() int {
			return  SysParamInt("contract_price")
		}
	}', 'ContractConditions("MainCondition")'),
	  ('3','EditContract','contract EditContract {
		  data {
			  Id         int
			  Value      string "optional"
			  Conditions string "optional"
		  }

		  func onlyConditions() bool {
        	return $Conditions && !$Value
		  }
		  conditions {
			RowConditions("contracts", $Id, onlyConditions())
			if $Conditions {
	    		ValidateCondition($Conditions, $ecosystem_id)
			}

			var row array
			row = DBFind("contracts").Columns("id,value,conditions").WhereId($Id)
			if !Len(row) {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			$cur = row[0]
			if $Value {
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
		  }
		  action {
			var root int
			var pars, vals array

			if $Value {
				root = CompileContract($Value, $ecosystem_id, 0, 0)
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if Len(vals) > 0 {
				DBUpdate("contracts", $Id, Join(pars, ","), vals...)
			}
			if $Value {
			   FlushContract(root, $Id, false)
			}
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('4','NewParameter','contract NewParameter {
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
	  ('5','EditParameter','contract EditParameter {
		  data {
			  Id int
			  Value string
			  Conditions string
		  }
		  func onlyConditions() bool {
            	return $Conditions && !$Value
		  }
		  conditions {
			  RowConditions("parameters", $Id, onlyConditions())
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  action {
			  DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('6', 'NewMenu','contract NewMenu {
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
	}', 'ContractConditions("MainCondition")'),
	  ('7','EditMenu','contract EditMenu {
		  data {
			  Id         int
			  Value      string "optional"
			  Title      string "optional"
			  Conditions string "optional"
		  }
		  
		func onlyConditions() bool {
        	return $Conditions && !$Value && !$Title
		}
	  	conditions {
		  RowConditions("menu", $Id, onlyConditions())
		  if $Conditions {
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
	  	}
	  	action {
		  var pars, vals array
		  if $Value {
			  pars[0] = "value"
			  vals[0] = $Value
		  }
		  if $Title {
			  pars[Len(pars)] = "title"
			  vals[Len(vals)] = $Title
		  }
		  if $Conditions {
			  pars[Len(pars)] = "conditions"
			  vals[Len(vals)] = $Conditions
		  }
		  if Len(vals) > 0 {
			  DBUpdate("menu", $Id, Join(pars, ","), vals...)
		  }			
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('8','AppendMenu','contract AppendMenu {
		data {
			Id     int
			Value  string
		}
		conditions {
			RowConditions("menu", $Id, false)
		}
		action {
			var row map
			row = DBRow("menu").Columns("value").WhereId($Id)
			DBUpdate("menu", $Id, "value", row["value"] + "\r\n" + $Value)
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('9','NewPage','contract NewPage {
		data {
			Name       string
			Value      string
			Menu       string
			Conditions string
			ValidateCount int "optional"
			ApplicationId int "optional"
			ValidateMode int "optional"
		}
		func preparePageValidateCount(count int) int {
			var min, max int
			min = Int(EcosysParam("min_page_validate_count"))
			max = Int(EcosysParam("max_page_validate_count"))

			if count < min {
				count = min
			} else {
				if count > max {
					count = max
				}
			}

			return count
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("pages").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Page %%s already exists", $Name)
			}

			$ValidateCount = preparePageValidateCount($ValidateCount)
		}
		action {
			DBInsert("pages", "name,value,menu,validate_count,conditions,app_id,validate_mode", 
				$Name, $Value, $Menu, $ValidateCount, $Conditions, $ApplicationId, $ValidateMode)
		}
		func price() int {
			return  SysParamInt("page_price")
		}
	}', 'ContractConditions("MainCondition")'),
	  ('10','EditPage','contract EditPage {
		  data {
			Id         int
			Value      string "optional"
			Menu      string "optional"
			Conditions string "optional"
			ValidateCount int "optional"
			ValidateMode  string "optional"
		  }
		  func onlyConditions() bool {
        	return $Conditions && !$Value && !$Menu
		  }
		  func preparePageValidateCount(count int) int {
			  var min, max int
			  min = Int(EcosysParam("min_page_validate_count"))
			  max = Int(EcosysParam("max_page_validate_count"))
	  
			  if count < min {
				  count = min
			  } else {
				  if count > max {
					  count = max
				  }
			  }
	  
			  return count
		  }					  
	  	conditions {
		  RowConditions("pages", $Id, onlyConditions())
		  if $Conditions {
			  ValidateCondition($Conditions, $ecosystem_id)
		  }
		  $ValidateCount = preparePageValidateCount($ValidateCount)
	  	}
	  	action {
			var pars, vals array
			if $Value {
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Menu {
				pars[Len(pars)] = "menu"
				vals[Len(vals)] = $Menu
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if $ValidateCount {
				pars[Len(pars)] = "validate_count"
				vals[Len(vals)] = $ValidateCount
			}
			if $ValidateMode {
				if $ValidateMode != "1" {
					$ValidateMode = "0"
				}
				pars[Len(pars)] = "validate_mode"
				vals[Len(vals)] = $ValidateMode
			}
			if Len(vals) > 0 {
				DBUpdate("pages", $Id, Join(pars, ","), vals...)
			}
	  	}		  
	  }', 'ContractConditions("MainCondition")'),
	  ('11','AppendPage','contract AppendPage {
		  data {
			  Id         int
			  Value      string
		  }
		  conditions {
			  RowConditions("pages", $Id, false)
		  }
		  action {
			  var row map
			  row = DBRow("pages").Columns("value").WhereId($Id)
			  DBUpdate("pages", $Id, "value", row["value"] + "\r\n" + $Value)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('12','NewBlock','contract NewBlock {
		data {
			Name       string
			Value      string
			Conditions string
			ApplicationId int "optional"
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
			DBInsert("blocks", "name,value,conditions,app_id", $Name, $Value, $Conditions, $ApplicationId )
		}
	 }', 'ContractConditions("MainCondition")'),
	  ('13','EditBlock','contract EditBlock {
		data {
			Id         int
			Value      string "optional"
		  	Conditions string "optional"
		}
		
		func onlyConditions() bool {
			return $Conditions && !$Value
		}

	  	conditions {
			RowConditions("blocks", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
	  	}
	  	action {
		  var pars, vals array
		  if $Value {
			  pars[0] = "value"
			  vals[0] = $Value
		  }
		  if $Conditions {
			  pars[Len(pars)] = "conditions"
			  vals[Len(vals)] = $Conditions
		  }
		  if Len(vals) > 0 {
			  DBUpdate("blocks", $Id, Join(pars, ","), vals...)
		  }
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('14','NewTable','contract NewTable {
		data {
			Name       string
			Columns      string
			Permissions string
			ApplicationId int "optional"
		}
		conditions {
			TableConditions($Name, $Columns, $Permissions)
		}
		action {
			CreateTable($Name, $Columns, $Permissions, $ApplicationId)
		}
		func rollback() {
			RollbackTable($Name)
		}
		func price() int {
			return  SysParamInt("table_price")
		}
	}', 'ContractConditions("MainCondition")'),
	  ('15','EditTable','contract EditTable {
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
	  ('16','NewColumn','contract NewColumn {
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
	  ('17','EditColumn','contract EditColumn {
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
	  ('18','NewLang','contract NewLang {
		data {
			Name  string
			Trans string
			AppID int
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
			var row array
			row = DBFind("languages").Columns("name").Where("name=? AND app_id=?", $Name, $AppID).Limit(1)
			if Len(row) > 0 {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res,app_id", $Name, $Trans, $AppID)
			UpdateLang($AppID, $Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('19','EditLang','contract EditLang {
		data {
			Id    int
			Name  string
			Trans string
			AppID int
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdate("languages", $Id, "name,res,app_id", $Name, $Trans, $AppID)
			UpdateLang($AppID, $Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('20','Import','contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONDecode($Data)
		}
		func ImportList(row array, cnt string) {
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
	('21', 'NewCron','contract NewCron {
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
	('22','EditCron','contract EditCron {
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
	}', 'ContractConditions("MainCondition")'),
	('23', 'UploadBinary', contract UploadBinary {
		data {
			Name  string
			Data  bytes "file"
			AppID int
			DataMimeType string "optional"
			MemberID int "optional"
		}
		conditions {
			$Id = Int(DBFind("binaries").Columns("id").Where("app_id = ? AND member_id = ? AND name = ?", $AppID, $MemberID, $Name).One("id"))
		}
		action {
			var hash string
			hash = MD5($Data)

			if $DataMimeType == "" {
				$DataMimeType = "application/octet-stream"
			}

			if $Id != 0 {
				DBUpdate("binaries", $Id, "data,hash,mime_type", $Data, hash, $DataMimeType)
			} else {
				$Id = DBInsert("binaries", "app_id,member_id,name,data,hash,mime_type", $AppID, $MemberID, $Name, $Data, hash, $DataMimeType)
			}

			$result = $Id
		}
	}', 'ContractConditions("MainCondition")');
	`
	// SchemaEcosystem contains SQL queries for creating ecosystem
	SchemaEcosystem = `DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
		"id" bigint  NOT NULL DEFAULT '0',
		"pub" bytea  NOT NULL DEFAULT '',
		"amount" decimal(30) NOT NULL DEFAULT '0' CHECK (amount >= 0),
		"multi" bigint NOT NULL DEFAULT '0',
		"deleted" bigint NOT NULL DEFAULT '0',
		"blocked" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "%[1]d_history"; CREATE TABLE "%[1]d_history" (
		"id" bigint NOT NULL  DEFAULT '0',
		"sender_id" bigint NOT NULL DEFAULT '0',
		"recipient_id" bigint NOT NULL DEFAULT '0',
		"amount" decimal(30) NOT NULL DEFAULT '0',
		"comment" text NOT NULL DEFAULT '',
		"block_id" int  NOT NULL DEFAULT '0',
		"txhash" bytea  NOT NULL DEFAULT '',
		"created_at" timestamp DEFAULT NOW()
		);
		ALTER TABLE ONLY "%[1]d_history" ADD CONSTRAINT "%[1]d_history_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_history_index_sender" ON "%[1]d_history" (sender_id);
		CREATE INDEX "%[1]d_history_index_recipient" ON "%[1]d_history" (recipient_id);
		CREATE INDEX "%[1]d_history_index_block" ON "%[1]d_history" (block_id, txhash);
		
		
		DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "res" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT '',
		  "app_id" bigint NOT NULL DEFAULT '1'
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

		DROP TABLE IF EXISTS "%[1]d_menu";
		CREATE TABLE "%[1]d_menu" (
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
','true');

		DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"menu" character varying(255) NOT NULL DEFAULT '',
			"validate_count" bigint NOT NULL DEFAULT '1',
			"conditions" text NOT NULL DEFAULT '',
			"app_id" bigint NOT NULL DEFAULT '1',
			"validate_mode" character(1) NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);


		INSERT INTO "%[1]d_pages" ("id","name","value","menu","conditions") VALUES
			('2','admin_index','','admin_menu','true'),
			('3','notifications',$$DBFind(Name: notifications, Source: notifications_members).Columns("id,page_name,notification->icon,notification->header,notification->body").Where("closed=0 and notification->type='1' and recipient->member_id='#key_id#'")
			ForList(notifications_members){
				Div(Class: list-group-item){
					LinkPage(Page: #page_name#, PageParams: "notific_id=#id#"){
						Div(media-box){
							Div(Class: pull-left){
								Em(Class: fa #notification.icon# fa-1x text-primary)
							}
							Div(media-box-body clearfix){
								Div(Class: m0 text-normal, Body: #notification.header#)
								Div(Class: m0 text-muted h6, Body: #notification.body#)
							}
						}
					}
				}
			}

			DBFind(Name: notifications, Source: notifications_roles).Columns("id,page_name,notification->icon,notification->header,notification->body,recipient->role_id").Where("closed=0 and notification->type='2' and (date_start_processing is null or processing_info->member_id='#key_id#')")
			ForList(notifications_roles){
			    DBFind(Name: roles_participants, Source: src_roles).Columns("id").Where("member->member_id='#key_id#' and role->id='#recipient.role_id#' and deleted=0").Vars(prefix)
			    If(#prefix_id# > 0){
					Div(Class: list-group-item){
						LinkPage(Page: #page_name#, PageParams: "notific_id=#id#"){
							Div(media-box){
								Div(Class: pull-left){
									Em(Class: fa #notification.icon# fa-1x text-primary)
								}
								Div(media-box-body clearfix){
									Div(Class: m0 text-normal, Body: #notification.header#)
									Div(Class: m0 text-muted h6, Body: #notification.body#)
								}
							}
						}
					}
				}
			}$$,'default_menu','ContractAccess("@1EditPage")');

		DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT '',
			"app_id" bigint NOT NULL DEFAULT '1'
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
		"name" text NOT NULL DEFAULT '',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '1',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);
		
		INSERT INTO "%[1]d_contracts" ("id", "name", "value", "wallet_id","active", "conditions") VALUES 
		('1','MainCondition','contract MainCondition {
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
		('9','max_sum', '1000000', 'ContractConditions("MainCondition")'),
		('10','money_digit', '2', 'ContractConditions("MainCondition")'),
		('11','stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
		}', 'ContractConditions("MainCondition")'),
		('13','max_block_user_tx', '100', 'ContractConditions("MainCondition")'),
		('14','min_page_validate_count', '1', 'ContractConditions("MainCondition")'),
		('15','max_page_validate_count', '6', 'ContractConditions("MainCondition")'),
		('16','changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")');

		DROP TABLE IF EXISTS "%[1]d_app_params";
		CREATE TABLE "%[1]d_app_params" (
		"id" bigint NOT NULL  DEFAULT '0',
		"app_id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_app_params" ADD CONSTRAINT "%[1]d_app_params_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_app_params_index_name" ON "%[1]d_app_params" (name);
		CREATE INDEX "%[1]d_app_params_index_app" ON "%[1]d_app_params" (app_id);
		
		DROP TABLE IF EXISTS "%[1]d_tables";
		CREATE TABLE "%[1]d_tables" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_tables_index_name" ON "%[1]d_tables" (name);
		
		INSERT INTO "%[1]d_tables" ("id", "name", "permissions","columns", "conditions") VALUES 
			('1', 'contracts', '{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", "new_column": "ContractConditions(\"MainCondition\")"}', 
			'{"name": "false", 
				"value": "ContractConditions(\"MainCondition\")",
				  "wallet_id": "ContractConditions(\"MainCondition\")",
				  "token_id": "ContractConditions(\"MainCondition\")",
				  "active": "ContractConditions(\"MainCondition\")",
				  "conditions": "ContractConditions(\"MainCondition\")"}', 'ContractAccess("@1EditTable")'),
				('2', 'keys', 
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
				  "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"pub": "ContractConditions(\"MainCondition\")",
				  "amount": "ContractConditions(\"MainCondition\")",
				  "deleted": "ContractConditions(\"MainCondition\")",
				  "blocked": "ContractConditions(\"MainCondition\")",
				  "multi": "ContractConditions(\"MainCondition\")"}', 
				'ContractAccess("@1EditTable")'),
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
				'{"app_id": "ContractConditions(\"MainCondition\")",
				  "name": "ContractConditions(\"MainCondition\")",
				  "res": "ContractConditions(\"MainCondition\")",
				  "conditions": "ContractConditions(\"MainCondition\")",
				  "app_id": "ContractConditions(\"MainConditions\")"}', 'ContractAccess("@1EditTable")'),
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
			"validate_count": "ContractConditions(\"MainCondition\")",
			"validate_mode": "ContractConditions(\"MainCondition\")",
			"app_id": "ContractConditions(\"MainCondition\")",
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
				('9', 'members', 
					'{"insert":"ContractAccess(\"Profile_Edit\")","update":"ContractConditions(\"MainCondition\")","new_column":"ContractConditions(\"MainCondition\")"}',
					'{"image_id":"ContractAccess(\"Profile_Edit\")","member_info":"ContractAccess(\"Profile_Edit\")","member_name":"false"}', 
					'ContractConditions(\"MainCondition\")'),
				('10', 'roles',
					'{"insert":"ContractAccess(\"Roles_Create\")",
						"update":"ContractConditions(\"MainCondition\")",
						"new_column":"ContractConditions(\"MainCondition\")"}', 
					'{"default_page":"false",
						"creator":"false",
						"deleted":"ContractAccess(\"Roles_Del\")",
						"company_id":"false",
						"date_deleted":"ContractAccess(\"Roles_Del\")",
						"image_id":"ContractAccess(\"Roles_Create\")",
						"role_name":"false",
						"date_created":"false",
						"role_type":"false"}',
					'ContractConditions(\"MainCondition\")'),
				('11', 'roles_participants',
					'{"insert":"ContractAccess(\"Roles_Assign\",\"voting_CheckDecision\")",
						"update":"ContractConditions(\"MainCondition\")",
						"new_column":"ContractConditions(\"MainCondition\")"}',
					'{"deleted":"ContractAccess(\"Roles_Unassign\")",
						"date_deleted":"ContractAccess(\"Roles_Unassign\")",
						"member":"false",
						"role":"false",
						"date_created":"false",
						"appointed":"false"}', 
					'ContractConditions(\"MainCondition\")'),
				('12', 'notifications',
					'{"insert":"ContractAccess(\"Notifications_Single_Send_map\",\"Notifications_Roles_Send_map\")",
						"update":"ContractConditions(\"MainCondition\")",
						"new_column":"ContractConditions(\"MainCondition\")"}',
					'{"date_closed":"ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Close\")",
						"sender":"false",
						"processing_info":"ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Processing\")",
						"date_start_processing":"ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Processing\")",
						"notification":"false",
						"page_name":"false",
						"page_params":"false",
						"closed":"ContractAccess(\"Notifications_Single_Close\",\"Notifications_Roles_Close\")",
						"date_created":"false",
						"recipient":"false"}',
					'ContractAccess(\"@1EditTable\")'),
				('13', 'sections', 
					'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", 
					"new_column": "ContractConditions(\"MainCondition\")"}',
					'{"title": "ContractConditions(\"MainCondition\")",
						"urlname": "ContractConditions(\"MainCondition\")",
						"page": "ContractConditions(\"MainCondition\")",
						"roles_access": "ContractConditions(\"MainCondition\")",
						"delete": "ContractConditions(\"MainCondition\")"}', 
						'ContractConditions(\"MainCondition\")'),
				('14', 'applications',
					'{"insert": "ContractConditions(\"MainCondition\")",
						 "update": "ContractConditions(\"MainCondition\")", 
						 "new_column": "ContractConditions(\"MainCondition\")"}',
					'{"name": "ContractConditions(\"MainCondition\")",
					  "uuid": "false",
					  "conditions": "ContractConditions(\"MainCondition\")",
					  "deleted": "ContractConditions(\"MainCondition\")"}',
					'ContractConditions(\"MainCondition\")'),
				('15', 'binaries',
					'{"insert":"ContractAccess(\"UploadBinary\")",
						"update":"ContractConditions(\"MainCondition\")",
						"new_column":"ContractConditions(\"MainCondition\")"}',
					'{"hash":"ContractAccess(\"UploadBinary\")",
						"member_id":"false",
						"data":"ContractAccess(\"UploadBinary\")",
						"name":"false",
						"app_id":"false"}',
					'ContractConditions(\"MainCondition\")');

		DROP TABLE IF EXISTS "%[1]d_notifications";
		CREATE TABLE "%[1]d_notifications" (
			"id"    bigint NOT NULL DEFAULT '0',
			"recipient" jsonb,
			"sender" jsonb,
			"notification" jsonb,
			"page_params"	jsonb,
			"processing_info" jsonb,
			"page_name"	varchar(255) NOT NULL DEFAULT '',
			"date_created"	timestamp,
			"date_start_processing" timestamp,
			"date_closed" timestamp,
			"closed" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_notifications" ADD CONSTRAINT "%[1]d_notifications_pkey" PRIMARY KEY ("id");


		DROP TABLE IF EXISTS "%[1]d_roles";
		CREATE TABLE "%[1]d_roles" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"default_page"	varchar(255) NOT NULL DEFAULT '',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"deleted"    bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"creator" jsonb NOT NULL DEFAULT '{}',
			"date_created" timestamp,
			"date_deleted" timestamp,
			"company_id" bigint NOT NULL DEFAULT '0',
			"image_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_roles" ADD CONSTRAINT "%[1]d_roles_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_roles_index_deleted" ON "%[1]d_roles" (deleted);
		CREATE INDEX "%[1]d_roles_index_type" ON "%[1]d_roles" (role_type);

		INSERT INTO "%[1]d_roles" ("id", "default_page", "role_name", "deleted", "role_type",
			"date_created","creator") VALUES
			('1','default_ecosystem_page', 'Admin', '0', '3', NOW(), '{}'),
			('2','', 'Candidate for validators', '0', '3', NOW(), '{}'),
			('3','', 'Validator', '0', '3', NOW(), '{}'),
			('4','', 'Investor with voting rights', '0', '3', NOW(), '{}'),
			('5','', 'Delegate', '0', '3', NOW(), '{}'),
			('6','', 'Developer', '0', '3', NOW(), '{}');


		DROP TABLE IF EXISTS "%[1]d_roles_participants";
		CREATE TABLE "%[1]d_roles_participants" (
			"id" bigint NOT NULL DEFAULT '0',
			"role" jsonb,
			"member" jsonb,
			"appointed" jsonb,
			"date_created" timestamp,
			"date_deleted" timestamp,
			"deleted" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_roles_participants" ADD CONSTRAINT "%[1]d_roles_participants_pkey" PRIMARY KEY ("id");

		INSERT INTO "%[1]d_roles_participants" ("id","role" ,"member", "date_created")
		VALUES ('1', '{"id": "1", "type": "3", "name": "Admin", "image_id":"0"}', '{"member_id": "%[4]d", "member_name": "founder", "image_id": "0"}', NOW()),
		('2', '{"id": "6", "type": "3", "name": "Developer", "image_id":"0"}', '{"member_id": "%[4]d", "member_name": "founder", "image_id": "0"}', NOW());

		DROP TABLE IF EXISTS "%[1]d_members";
		CREATE TABLE "%[1]d_members" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"image_id"	bigint NOT NULL DEFAULT '0',
			"member_info"   jsonb
		);
		ALTER TABLE ONLY "%[1]d_members" ADD CONSTRAINT "%[1]d_members_pkey" PRIMARY KEY ("id");

		INSERT INTO "%[1]d_members" ("id", "member_name") VALUES('%[4]d', 'founder');

		DROP TABLE IF EXISTS "%[1]d_applications";
		CREATE TABLE "%[1]d_applications" (
			"id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"uuid" uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
			"conditions" text NOT NULL DEFAULT '',
			"deleted" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_applications" ADD CONSTRAINT "%[1]d_application_pkey" PRIMARY KEY ("id");

		DROP TABLE IF EXISTS "%[1]d_binaries";
		CREATE TABLE "%[1]d_binaries" (
			"id" bigint NOT NULL DEFAULT '0',
			"app_id" bigint NOT NULL DEFAULT '1',
			"member_id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"data" bytea NOT NULL DEFAULT '',
			"hash" varchar(32) NOT NULL DEFAULT '',
			"mime_type" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_binaries" ADD CONSTRAINT "%[1]d_binaries_pkey" PRIMARY KEY (id);
		CREATE UNIQUE INDEX "%[1]d_binaries_index_app_id_member_id_name" ON "%[1]d_binaries" (app_id, member_id, name);
		`

	// SchemaFirstEcosystem contains SQL queries for creating first ecosystem
	SchemaFirstEcosystem = `
	DROP TABLE IF EXISTS "1_ecosystems";
	CREATE TABLE "1_ecosystems" (
			"id" bigint NOT NULL DEFAULT '0',
			"name"	varchar(255) NOT NULL DEFAULT '',
			"is_valued" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "1_ecosystems" ADD CONSTRAINT "1_ecosystems_pkey" PRIMARY KEY ("id");

	INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

	DROP TABLE IF EXISTS "1_system_parameters";
		CREATE TABLE "1_system_parameters" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255)  NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "1_system_parameters" ADD CONSTRAINT "1_system_parameters_pkey" PRIMARY KEY (id);
		CREATE INDEX "1_system_parameters_index_name" ON "1_system_parameters" (name);
		
		INSERT INTO "1_system_parameters" ("id","name", "value", "conditions") VALUES 
		('1','default_ecosystem_page', '', 'true'),
		('2','default_ecosystem_menu', '', 'true'),
		('3','default_ecosystem_contract', '', 'true'),
		('4','gap_between_blocks', '2', 'true'),
		('5','rb_blocks_1', '60', 'true'),
		('7','new_version_url', 'upd.apla.io', 'true'),
		('8','full_nodes', '', 'true'),
		('9','number_of_nodes', '101', 'true'),
		('10','ecosystem_price', '1000', 'true'),
		('11','contract_price', '200', 'true'),
		('12','column_price', '200', 'true'),
		('13','table_price', '200', 'true'),
		('14','menu_price', '100', 'true'),
		('15','page_price', '100', 'true'),
		('16','blockchain_url', '', 'true'),
		('17','max_block_size', '67108864', 'true'),
		('18','max_tx_size', '33554432', 'true'),
		('19','max_tx_count', '1000', 'true'),
		('20','max_columns', '50', 'true'),
		('21','max_indexes', '5', 'true'),
		('22','max_block_user_tx', '100', 'true'),
		('23','max_fuel_tx', '20000', 'true'),
		('24','max_fuel_block', '100000', 'true'),
		('25','commission_size', '3', 'true'),
		('26','commission_wallet', '', 'true'),
		('27','fuel_rate', '[["1","1000000000000000"]]', 'true'),
		('28','extend_cost_address_to_id', '10', 'true'),
		('29','extend_cost_id_to_address', '10', 'true'),
		('30','extend_cost_new_state', '1000', 'true'), -- What cost must be?
		('31','extend_cost_sha256', '50', 'true'),
		('32','extend_cost_pub_to_id', '10', 'true'),
		('33','extend_cost_ecosys_param', '10', 'true'),
		('34','extend_cost_sys_param_string', '10', 'true'),
		('35','extend_cost_sys_param_int', '10', 'true'),
		('36','extend_cost_sys_fuel', '10', 'true'),
		('37','extend_cost_validate_condition', '30', 'true'),
		('38','extend_cost_eval_condition', '20', 'true'),
		('39','extend_cost_has_prefix', '10', 'true'),
		('40','extend_cost_contains', '10', 'true'),
		('41','extend_cost_replace', '10', 'true'),
		('42','extend_cost_join', '10', 'true'),
		('43','extend_cost_update_lang', '10', 'true'),
		('44','extend_cost_size', '10', 'true'),
		('45','extend_cost_substr', '10', 'true'),
		('46','extend_cost_contracts_list', '10', 'true'),
		('47','extend_cost_is_object', '10', 'true'),
		('48','extend_cost_compile_contract', '100', 'true'),
		('49','extend_cost_flush_contract', '50', 'true'),
		('50','extend_cost_eval', '10', 'true'),
		('51','extend_cost_len', '5', 'true'),
		('52','extend_cost_activate', '10', 'true'),
		('53','extend_cost_deactivate', '10', 'true'),
		('54','extend_cost_create_ecosystem', '100', 'true'),
		('55','extend_cost_table_conditions', '100', 'true'),
		('56','extend_cost_create_table', '100', 'true'),
		('57','extend_cost_perm_table', '100', 'true'),
		('58','extend_cost_column_condition', '50', 'true'),
		('59','extend_cost_create_column', '50', 'true'),
		('60','extend_cost_perm_column', '50', 'true'),
		('61','extend_cost_json_to_map', '50', 'true'),
		('62','max_block_generation_time', '2000', 'true'),
		('63','block_reward','1000','true');
		
		DROP TABLE IF EXISTS "1_delayed_contracts";
		CREATE TABLE "1_delayed_contracts" (
			"id" int NOT NULL default 0,
			"contract" varchar(255) NOT NULL DEFAULT '',
			"key_id" bigint NOT NULL DEFAULT '0',
			"block_id" int NOT NULL DEFAULT '0',
			"every_block" int NOT NULL DEFAULT '0',
			"counter" int NOT NULL DEFAULT '0',
			"limit" int NOT NULL DEFAULT '0',
			"deleted" boolean NOT NULL DEFAULT 'false',
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "1_delayed_contracts" ADD CONSTRAINT "1_delayed_contracts_pkey" PRIMARY KEY ("id");
		CREATE INDEX "1_delayed_contracts_index_block_id" ON "1_delayed_contracts" ("block_id");


		INSERT INTO "1_delayed_contracts"
			("id", "contract", "key_id", "block_id", "every_block", "conditions")
		VALUES
			(1, '@1UpdateMetrics', '%[1]d', '100', '100', 'ContractConditions("MainCondition")');

		DROP TABLE IF EXISTS "1_metrics";
		CREATE TABLE "1_metrics" (
			"id" int NOT NULL default 0,
			"time" bigint NOT NULL DEFAULT '0',
			"metric" varchar(255) NOT NULL,
			"key" varchar(255) NOT NULL,
			"value" bigint NOT NULL
		);
		ALTER TABLE ONLY "1_metrics" ADD CONSTRAINT "1_metrics_pkey" PRIMARY KEY (id);
		CREATE INDEX "1_metrics_unique_index" ON "1_metrics" (metric, time, "key");

		INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions") VALUES
			('16', 'delayed_contracts',
			'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")",
			"new_column": "ContractConditions(\"MainCondition\")"}',
			'{"contract": "ContractConditions(\"MainCondition\")",
				"key_id": "ContractConditions(\"MainCondition\")",
				"block_id": "ContractConditions(\"MainCondition\")",
				"every_block": "ContractConditions(\"MainCondition\")",
				"counter": "ContractConditions(\"MainCondition\")",
				"limit": "ContractConditions(\"MainCondition\")",
				"deleted": "ContractConditions(\"MainCondition\")",
				"conditions": "ContractConditions(\"MainCondition\")"}',
				'ContractConditions(\"MainCondition\")'
			),
			(
				'17',
				'ecosystems',
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", "new_column": "ContractConditions(\"MainCondition\")"}',
				'{"name": "ContractConditions(\"MainCondition\")"}',
				'ContractConditions(\"MainCondition\")'
			),
			(
				'18',
				'metrics',
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")","new_column": "ContractConditions(\"MainCondition\")"}',
				'{"time": "ContractConditions(\"MainCondition\")",
					"metric": "ContractConditions(\"MainCondition\")","key": "ContractConditions(\"MainCondition\")",
					"value": "ContractConditions(\"MainCondition\")"}',
				'ContractConditions(\"MainCondition\")'
			),
			(
				'19',
				'system_parameters',
				'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")","new_column": "ContractConditions(\"MainCondition\")"}',
				'{"value": "ContractConditions(\"MainCondition\")"}',
				'ContractConditions(\"MainCondition\")'
			);


	INSERT INTO "1_contracts" ("id", "name","value", "wallet_id", "conditions", "app_id") VALUES
	('2','MoneyTransfer','contract MoneyTransfer {
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
			if $amount <= 0 {
				error "Amount must be greater then zero"
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
            if DBFind("keys").Columns("id").WhereId($recipient).One("id") == nil {
                DBInsert("keys", "id,amount",  $recipient, $amount)
            } else {
                DBUpdate("keys", $recipient,"+amount", $amount)
            }
            DBInsert("history", "sender_id,recipient_id,amount,comment,block_id,txhash",
                    $key_id, $recipient, $amount, $Comment, $block, $txhash)
		}
	}', '%[1]d', 'ContractConditions("MainCondition")', 1),
	('3','NewContract','contract NewContract {
		data {
			Value      string
			Conditions string
			Wallet         string "optional"
			TokenEcosystem int "optional"
			ApplicationId int "optional"
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
			
			if Len(list) == 0 {
				error "must be the name"
			}

			var i int
			while i < Len(list) {
				if IsObject(list[i], $ecosystem_id) {
					warning Sprintf("Contract or function %%s exists", list[i] )
				}
				i = i + 1
			}

			$contract_name = list[0]
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
			id = DBInsert("contracts", "name,value,conditions, wallet_id, token_id,app_id",
				   $contract_name, $Value, $Conditions, $walletContract, $TokenEcosystem, $ApplicationId)
			FlushContract(root, id, false)
			$result = id
		}
		func rollback() {
			var list array
    		list = ContractsList($Value)
			var i int
			while i < Len(list) {
				RollbackContract(list[i])
				i = i + 1
			}
		}
		func price() int {
			return  SysParamInt("contract_price")
		}
	}', '%[1]d', 'ContractConditions("MainCondition")', 1),
	('4','EditContract','contract EditContract {
		data {
			Id         int
			Value      string "optional"
			Conditions string "optional"
			WalletId   string "optional"
		}

		func onlyConditions() bool {
			return $Conditions && !$Value && !$WalletId
		}

		conditions {
			RowConditions("contracts", $Id, onlyConditions())
			if $Conditions {
			    ValidateCondition($Conditions, $ecosystem_id)
			}
			$cur = DBRow("contracts").Columns("id,value,conditions,active,wallet_id,token_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if $Value {
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
			var pars, vals array
			if $Value {
				root = CompileContract($Value, $ecosystem_id, $recipient, Int($cur["token_id"]))
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if $WalletId != "" {
				pars[Len(pars)] = "wallet_id"
				vals[Len(vals)] = $recipient
			}
			if Len(vals) > 0 {
				DBUpdate("contracts", $Id, Join(pars, ","), vals...)
			}		
			if $Value {
				FlushContract(root, $Id, Int($cur["active"]) == 1)
			} else {
				if $WalletId != "" {
					SetContractWallet($Id, $ecosystem_id, $recipient)
				}
			}
		}
		func rollback() {
			RollbackEditContract()
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('5','ActivateContract','contract ActivateContract {
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
		func rollback() {
			Deactivate($Id, $ecosystem_id)
		}

	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('6','NewEcosystem','contract NewEcosystem {
		data {
			Name  string
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('7','NewParameter','contract NewParameter {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('8','EditParameter','contract EditParameter {
		data {
			Id int
			Value string
			Conditions string
		}

		func onlyConditions() bool {
			return $Conditions && !$Value
		}

		conditions {
			RowConditions("parameters", $Id, onlyConditions())
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('9', 'NewMenu','contract NewMenu {
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
			DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions)
		}
		func price() int {
			return  SysParamInt("menu_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('10','EditMenu','contract EditMenu {
		data {
			Id         int
			Value      string "optional"
			Title      string "optional"
			Conditions string "optional"
		}

		func onlyConditions() bool {
			return $Conditions && !$Value && !$Title
		}

		conditions {
			RowConditions("menu", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
		}
		action {
			var pars, vals array
			if $Value {
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Title {
				pars[Len(pars)] = "title"
				vals[Len(vals)] = $Title
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if Len(vals) > 0 {
				DBUpdate("menu", $Id, Join(pars, ","), vals...)
			}			
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('11','AppendMenu','contract AppendMenu {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('12','NewPage','contract NewPage {
		data {
			Name       string
			Value      string
			Menu       string
			Conditions string
			ValidateCount int "optional"
			ApplicationId int "optional"
			ValidateMode  int "optional"
		}
		func preparePageValidateCount(count int) int {
			var min, max int
			min = Int(EcosysParam("min_page_validate_count"))
			max = Int(EcosysParam("max_page_validate_count"))
	
			if count < min {
				count = min
			} else {
				if count > max {
					count = max
				}
			}
	
			return count
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("pages").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Page %%s already exists", $Name)
			}

			$ValidateCount = preparePageValidateCount($ValidateCount)
		}
		action {
			DBInsert("pages", "name,value,menu,validate_count,conditions,app_id,validate_mode", 
				$Name, $Value, $Menu, $ValidateCount, $Conditions, $ApplicationId, $ValidateMode)
		}
		func price() int {
			return  SysParamInt("page_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('13','EditPage','contract EditPage {
		data {
			Id         int
			Value      string "optional"
			Menu      string "optional"
			Conditions string "optional"
			ValidateCount int "optional"
			ValidateMode  string "optional"
		}
		func onlyConditions() bool {
			return $Conditions && !$Value && !$Menu && !$ValidateCount 
		}
		func preparePageValidateCount(count int) int {
			var min, max int
			min = Int(EcosysParam("min_page_validate_count"))
			max = Int(EcosysParam("max_page_validate_count"))
	
			if count < min {
				count = min
			} else {
				if count > max {
					count = max
				}
			}
	
			return count
		}		
		conditions {
			RowConditions("pages", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
      $ValidateCount = preparePageValidateCount($ValidateCount)
		}
		action {
			var pars, vals array
			if $Value {
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Menu {
				pars[Len(pars)] = "menu"
				vals[Len(vals)] = $Menu
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if $ValidateCount {
				pars[Len(pars)] = "validate_count"
				vals[Len(vals)] = $ValidateCount
			}
			if $ValidateMode {
				if $ValidateMode != "1" {
					$ValidateMode = "0"
				}
				pars[Len(pars)] = "validate_mode"
				vals[Len(vals)] = $ValidateMode
			}
			if Len(vals) > 0 {
				DBUpdate("pages", $Id, Join(pars, ","), vals...)
			}
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('14','AppendPage','contract AppendPage {
		data {
			Id         int
			Value      string
		}
		conditions {
			RowConditions("pages", $Id, false)
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('15','NewLang','contract NewLang {
		data {
			Name  string
			Trans string
			AppID int
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")

			var row map
			row = DBRow("languages").Columns("id").Where("name = ? AND app_id = ?", $Name, $AppID)

			if row {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res,app_id", $Name, $Trans, $AppID)
			UpdateLang($AppID, $Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('16','EditLang','contract EditLang {
		data {
			Id    int
			Name  string
			Trans string
			AppID int
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdate("languages", $Id, "name,res,app_id", $Name, $Trans, $AppID)
			UpdateLang($AppID, $Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('17','NewSign','contract NewSign {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('18','EditSign','contract EditSign {
		data {
			Id         int
			Value      string "optional"
			Conditions string "optional"
		}

		func onlyConditions() bool {
			return $Conditions && !$Value
		}
		conditions {
			RowConditions("signatures", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
		}
		action {
			var pars, vals array
			if $Value {
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if Len(vals) > 0 {
				DBUpdate("signatures", $Id, Join(pars, ","), vals...)
			}
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('19','NewBlock','contract NewBlock {
		data {
			Name       string
			Value      string
			Conditions string
			ApplicationId int "optional"
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
			DBInsert("blocks", "name,value,conditions,app_id", $Name, $Value, $Conditions, $ApplicationId )
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('20','EditBlock','contract EditBlock {
		data {
			Id         int
			Value      string "optional"
			Conditions string "optional"
		}

		func onlyConditions() bool {
			return $Conditions && !$Value
		}

		conditions {
			RowConditions("blocks", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
		}
		action {
			var pars, vals array
			if $Value {
				pars[0] = "value"
				vals[0] = $Value
			}
			if $Conditions {
				pars[Len(pars)] = "conditions"
				vals[Len(vals)] = $Conditions
			}
			if Len(vals) > 0 {
				DBUpdate("blocks", $Id, Join(pars, ","), vals...)
			}
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('21','NewTable','contract NewTable {
		data {
			Name       string
			Columns      string
			Permissions string
			ApplicationId int "optional"
		}
		conditions {
			TableConditions($Name, $Columns, $Permissions)
		}
		action {
			CreateTable($Name, $Columns, $Permissions, $ApplicationId)
		}
		func rollback() {
			RollbackTable($Name)
		}
		func price() int {
			return  SysParamInt("table_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('22','EditTable','contract EditTable {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('23','NewColumn','contract NewColumn {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('24','EditColumn','contract EditColumn {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('25','Import','contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONDecode($Data)
		}
		func ImportList(row array, cnt string) {
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
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('26','DeactivateContract','contract DeactivateContract {
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
		func rollback() {
			Activate($Id, $ecosystem_id)
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('27','UpdateSysParam','contract UpdateSysParam {
		data {
			Name  string
			Value string
			Conditions string "optional"
		}
		action {
			DBUpdateSysParam($Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('28','NewAppParam','contract NewAppParam {
		data {
			App int
			Name string
			Value string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions, $ecosystem_id)
			if $App == 0 {
				warning "App id cannot equal 0"
			}
			var row map
			row = DBRow("app_params").Columns("id").Where("app_id = ? and name = ?", $App, $Name)
			if row {
				warning Sprintf( "App parameter %%s already exists", $Name)
			}
		}
		action {
			DBInsert("app_params", "app_id,name,value,conditions", $App, $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('29','EditAppParam','contract EditAppParam {
		data {
			Id int
			Value string
			Conditions string
		}
		func onlyConditions() bool {
			return $Conditions && !$Value
		}

		conditions {
			RowConditions("app_params", $Id, onlyConditions())
			ValidateCondition($Conditions, $ecosystem_id)
		}
		action {
			DBUpdate("app_params", $Id, "value,conditions", $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('30', 'NewDelayedContract','contract NewDelayedContract {
		data {
			Contract string
			EveryBlock int
			Conditions string
			BlockID int "optional"
			Limit int "optional"
		}
		conditions {
			ValidateCondition($Conditions, $ecosystem_id)

			if !HasPrefix($Contract, "@") {
				$Contract = "@" + Str($ecosystem_id) + $Contract
			}

			if GetContractByName($Contract) == 0 {
				error Sprintf("Unknown contract %%s", $Contract)
			}

			if $BlockID == 0 {
				$BlockID = $block + $EveryBlock
			}

			if $BlockID <= $block {
				error "The blockID must be greater than the current blockID"
			}
		}
		action {
			DBInsert("delayed_contracts", "contract,key_id,block_id,every_block,\"limit\",conditions", $Contract, $key_id, $BlockID, $EveryBlock, $Limit, $Conditions)
		}
	}','%[1]d', 'ContractConditions("MainCondition")', 1),
	('31', 'EditDelayedContract','contract EditDelayedContract {
		data {
			Id int
			Contract string
			EveryBlock int
			Conditions string
			BlockID int "optional"
			Limit int "optional"
			Deleted int "optional"
		}
		conditions {
			ConditionById("delayed_contracts", true)

			if !HasPrefix($Contract, "@") {
				$Contract = "@" + Str($ecosystem_id) + $Contract
			}

			if GetContractByName($Contract) == 0 {
				error Sprintf("Unknown contract %%s", $Contract)
			}

			if $BlockID == 0 {
				$BlockID = $block + $EveryBlock
			}

			if $BlockID <= $block {
				error "The blockID must be greater than the current blockID"
			}
		}
		action {
			DBUpdate("delayed_contracts", $Id, "contract,key_id,block_id,every_block,counter,\"limit\",deleted,conditions", $Contract, $key_id, $BlockID, $EveryBlock, 0, $Limit, $Deleted, $Conditions)
		}
	}','%[1]d', 'ContractConditions("MainCondition")', 1),
	('32', 'CallDelayedContract','contract CallDelayedContract {
		data {
			Id int
		}
		conditions {
			var rows array
			rows = DBFind("delayed_contracts").Where("id = ? and deleted = false", $Id)
			if !Len(rows) {
				error Sprintf("Delayed contract %%d does not exist", $Id)
			}
			$cur = rows[0]

			if $key_id != Int($cur["key_id"]) {
				error "Access denied"
			}

			if $block != Int($cur["block_id"]) {
				error Sprintf("Delayed contract %%d must run on block %%s, current block %%d", $Id, $cur["block_id"], $block)
			}
		}
		action {
			var limit, counter, block_id int

			limit = Int($cur["limit"])
			counter = Int($cur["counter"])+1
			block_id = $block

			if limit == 0 || limit > counter {
				block_id = block_id + Int($cur["every_block"])
			}

			DBUpdate("delayed_contracts", $Id, "counter,block_id", counter, block_id)
			CallContract($cur["contract"], nil)
		}
	}','%[1]d', 'ContractConditions("MainCondition")', 1),
	('33','UploadBinary','contract UploadBinary {
		data {
			Name  string
			Data  bytes "file"
			AppID int
			DataMimeType string "optional"
			MemberID int "optional"
		}
		conditions {
			$Id = Int(DBFind("binaries").Columns("id").Where("app_id = ? AND member_id = ? AND name = ?", $AppID, $MemberID, $Name).One("id"))
		}
		action {
			var hash string
			hash = MD5($Data)

			if $DataMimeType == "" {
				$DataMimeType = "application/octet-stream"
			}

			if $Id != 0 {
				DBUpdate("binaries", $Id, "data,hash,mime_type", $Data, hash, $DataMimeType)
			} else {
				$Id = DBInsert("binaries", "app_id,member_id,name,data,hash,mime_type", $AppID, $MemberID, $Name, $Data, hash, $DataMimeType)
			}

			$result = $Id
		}
	}', '%[1]d','ContractConditions("MainCondition")', 1),
	('34', 'NewUser','contract NewUser {
		data {
			NewPubkey string
		}
		conditions {
			$newId = PubToID($NewPubkey)
			if $newId == 0 {
				error "Wrong pubkey"
			}
			if DBFind("keys").Columns("id").WhereId($newId).One("id") != nil {
				error "User already exists"
			}

			$amount = Money(1000) * Money(1000000000000000000)
		}
		action {
			MoneyTransfer("Recipient,Amount,Comment", Str($newId), Str($amount), "New user deposit")
		}
	}','%[1]d', 'ContractConditions("NodeOwnerCondition")', 1),
	('35', 'EditEcosystemName','contract EditEcosystemName {
		data {
			EcosystemID int
			NewName string
		}
		conditions {
			var rows array
			rows = DBFind("@1_ecosystems").Where("id = ?", $EcosystemID)
			if !Len(rows) {
				error Sprintf("Ecosystem %%d does not exist", $EcosystemID)
			}
		}
		action {
			EditEcosysName($EcosystemID, $NewName)
		}
	}', '%[1]d', 'ContractConditions("MainCondition")', 1),
	('36', 'UpdateMetrics', 'contract UpdateMetrics {
		conditions {
			ContractConditions("MainCondition")
		}
		action {
			var values array
			values = DBCollectMetrics()

			var i, id int
			var v map
			while (i < Len(values)) {
				v = values[i]
				id = Int(DBFind("metrics").Columns("id").Where("time = ? AND key = ? AND metric = ?", v["time"], v["key"], v["metric"]).One("id"))
				if id != 0 {
					DBUpdate("metrics", id, "value", Int(v["value"]))
				} else {
					DBInsert("metrics", "time,key,metric,value", v["time"], v["key"], v["metric"], Int(v["value"]))
				}
				i = i + 1
			}
		}
	}','%[1]d', 'ContractConditions("MainCondition")', 1),
	('37', 'NodeOwnerCondition', 'contract NodeOwnerCondition {
		conditions {
			$full_nodes = JSONDecode(SysParamString("full_nodes"))
			var i int
			while i < Len($full_nodes) {
				$fn = $full_nodes[i]
				if $fn["key_id"] == $key_id {
					return true
				}
				i = i + 1
			}

			warning "Sorry, you do not have access to this action."
		}
	}','%[1]d', 'ContractConditions("MainCondition")', 1);
	
	INSERT INTO "1_applications" (id, name, conditions) VALUES (1, 'System', 'ContractConditions("MainCondition")');`
)
