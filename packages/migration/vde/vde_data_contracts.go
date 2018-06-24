package vde

var contractsDataSQL = `INSERT INTO "%[1]d_contracts" ("id", "name", "value", "conditions") VALUES 
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
			ApplicationId int
			Value string
			Conditions string
			Wallet string "optional"
			TokenEcosystem int "optional"
		}
	
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
	
			if $ApplicationId == 0 {
				warning "Application id cannot equal 0"
			}
	
			$walletContract = $key_id
			if $Wallet {
				$walletContract = AddressToId($Wallet)
				if $walletContract == 0 {
					error Sprintf("wrong wallet %%s", $Wallet)
				}
			}
	
			$contract_name = ContractName($Value)
	
			if !$contract_name {
				error "must be the name"
			}
	
			if !$TokenEcosystem {
				$TokenEcosystem = 1
			} else {
				if !SysFuel($TokenEcosystem) {
					warning Sprintf("Ecosystem %%d is not system", $TokenEcosystem)
				}
			}
		}
	
		action {
			$result = CreateContract($contract_name, $Value, $Conditions, $walletContract, $TokenEcosystem, $ApplicationId)
		}
		func rollback() {
			RollbackNewContract($Value)
		}
		func price() int {
			return SysParamInt("contract_price")
		}
	}', 'ContractConditions("MainCondition")'),
	  ('3','EditContract','contract EditContract {
		data {
			Id int
			Value string "optional"
			Conditions string "optional"
			WalletId string "optional"
		}
		func onlyConditions() bool {
			return $Conditions && !$Value && !$WalletId
		}
	
		conditions {
			RowConditions("contracts", $Id, onlyConditions())
			if $Conditions {
				ValidateCondition($Conditions, $ecosystem_id)
			}
			$cur = DBFind("contracts").Columns("id,value,conditions,active,wallet_id,token_id").WhereId($Id).Row()
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if $Value {
				ValidateEditContractNewValue($Value, $cur["value"])
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
			UpdateContract($Id, $Value, $Conditions, $WalletId, $recipient, $cur["active"], $cur["token_id"])
		}
		func rollback() {
			RollbackEditContract()
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
	  ('18','NewLang', 'contract NewLang {
		data {
			ApplicationId int "optional"
			Name string
			Trans string "optional"
			Value array "optional"
			IdLanguage array "optional"
		}
	
		conditions {
			if $ApplicationId == 0 {
				warning "Application id cannot equal 0"
			}
	
			if DBFind("languages").Columns("id").Where("name = ?", $Name).One("id") {
				warning Sprintf( "Language resource %%s already exists", $Name)
			}
		
			var j int
			while j < Len($IdLanguage) {
				if $IdLanguage[j] == "" {
					info("Locale empty")
				}
				if $Value[j] == "" {
					info("Value empty")
				}
				j = j + 1
			}
			EvalCondition("parameters", "changing_language", "value")
		}
	
		action {
			var i,len,lenshar int
			var res,langarr string
			len = Len($IdLanguage)
			lenshar = Len($Value)	
			while i < len {
				if i + 1 == len {
					res = res + Sprintf("%%q: %%q",$IdLanguage[i],$Value[i])
				} else {
					res = res + Sprintf("%%q: %%q,",$IdLanguage[i],$Value[i])
				}
				i = i + 1
			}
			if len > 0 {
				langarr = Sprintf("{"+"%%v"+"}", res)
				$Trans = langarr
			}
			$result = CreateLanguage($Name, $Trans, $ApplicationId)
		}
	}', 'ContractConditions("MainCondition")'),
	('19','EditLang','contract EditLang {
		data {
			Id int
			Name string "optional"
			ApplicationId int "optional"
			Trans string "optional"
			Value array "optional"
			IdLanguage array "optional"
		}
		
		conditions {
			var j int
			while j < Len($IdLanguage) {
				if ($IdLanguage[j] == ""){
					info("Locale empty")
				}
				if ($Value[j] == ""){
					info("Value empty")
				}
				j = j + 1
			}
			EvalCondition("parameters", "changing_language", "value")
		}
		
		action {
			var i,len int
			var res,langarr string
			len = Len($IdLanguage)
			while i < len {
				if (i + 1 == len){
					res = res + Sprintf("%%q: %%q", $IdLanguage[i],$Value[i])
				}
				else {
					res = res + Sprintf("%%q: %%q, ", $IdLanguage[i],$Value[i])
				}
				i = i + 1
			}
	
			$row = DBFind("languages").Columns("name,app_id").WhereId($Id).Row()
			if !$row{
				warning "Language not found"
			}
	
			if $ApplicationId == 0 {
				$ApplicationId = Int($row["app_id"])
			}
			if $Name == "" {
				$Name = $row["name"]
			}
	
			if (len > 0){
				langarr = Sprintf("{"+"%%v"+"}", res)
				$Trans = langarr
				
			}
			EditLanguage($Id, $Name, $Trans, $ApplicationId)
		}
	}', 'ContractConditions("MainCondition")'),
	('20','Import','contract Import {
		data {
			Data string
		}
		func ReplaceValue(s string) string {
			s = Replace(s, "#IMPORT_ECOSYSTEM_ID#", "#ecosystem_id#")
			s = Replace(s, "#IMPORT_KEY_ID#", "#key_id#")
			s = Replace(s, "#IMPORT_ISMOBILE#", "#isMobile#")
			s = Replace(s, "#IMPORT_ROLE_ID#", "#role_id#")
			s = Replace(s, "#IMPORT_ECOSYSTEM_NAME#", "#ecosystem_name#")
			s = Replace(s, "#IMPORT_APP_ID#", "#app_id#")
			return s
		}
	
		conditions {
			$Data = ReplaceValue($Data)
	
			$ApplicationId = 0
			var app_map map
			app_map = DBFind("buffer_data").Columns("value->app_name").Where("key=''import_info'' and member_id=$", $key_id).Row()
			if app_map{
				var app_id int
				app_id = DBFind("applications").Columns("id").Where("name=$", Str(app_map["value.app_name"])).One("id")
				if app_id {
					$ApplicationId = Int(app_id)
				}
			}
		}
	
		action {
			var editors, creators map
			editors["pages"] = "EditPage"
			editors["blocks"] = "EditBlock"
			editors["menu"] = "EditMenu"
			editors["app_params"] = "EditAppParam"
			editors["languages"] = "EditLang"
			editors["contracts"] = "EditContract"
			editors["tables"] = "" // nothing
	
			creators["pages"] = "NewPage"
			creators["blocks"] = "NewBlock"
			creators["menu"] = "NewMenu"
			creators["app_params"] = "NewAppParam"
			creators["languages"] = "NewLang"
			creators["contracts"] = "NewContract"
			creators["tables"] = "NewTable"
	
			var dataImport array
			dataImport = JSONDecode($Data)
			var i int
			while i<Len(dataImport){
				var item, cdata map
				cdata = dataImport[i]
				if cdata {
					cdata["ApplicationId"] = $ApplicationId
					$Type = cdata["Type"]
					$Name = cdata["Name"]
	
					// Println(Sprintf("import %%v: %%v", $Type, cdata["Name"]))
	
					item = DBFind($Type).Where("name=?", $Name).Row()
					var contractName string
					if item {
						contractName = editors[$Type]
						cdata["Id"] = Int(item["id"])
						if $Type == "menu"{
							var menu menuItem string
							menu = Replace(item["value"], " ", "")
							menu = Replace(menu, "\n", "")
							menu = Replace(menu, "\r", "")
							menuItem = Replace(cdata["Value"], " ", "")
							menuItem = Replace(menuItem, "\n", "")
							menuItem = Replace(menuItem, "\r", "")
							if Contains(menu, menuItem) {
								// ignore repeated
								contractName = ""
							}else{
								cdata["Value"] = item["value"] + "\n" + cdata["Value"]
							}
						}
					} else {
						contractName = creators[$Type]
					}
	
					if contractName != ""{
						CallContract(contractName, cdata)
					}
				}
				i=i+1
			}
			// Println(Sprintf("> time: %%v", $time))
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
	('23', 'UploadBinary', 'contract UploadBinary {
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
	}', 'ContractConditions("MainCondition")'),
	('24', 'NewUser','contract NewUser {
		data {
			NewPubkey string
		}
		conditions {
			Println($NewPubkey)
			$newId = PubToID($NewPubkey)
			if $newId == 0 {
				error "Wrong pubkey"
			}
			if DBFind("keys").Columns("id").WhereId($newId).One("id") != nil {
				error "User already exists"
			}
		}
		action {
			DBInsert("keys", "id", $newId)
			SetPubKey($newId, StringToBytes($NewPubkey))
		}
	}', 'ContractConditions("MainCondition")'),
	('25', 'NewVDE', 'contract NewVDE {
		data {
			VDEName string
			DBUser string
			DBPassword string
			VDEAPIPort int
		}
	
		conditions {
		}
	
		action {
			CreateVDE($VDEName, $DBUser, $DBPassword, $VDEAPIPort)
		}
	}', 'ContractConditions("MainCondition")'),
	('26', 'ListVDE', 'contract ListVDE {
		data {}
	
		conditions {}
	
		action {
			return GetVDEList()
		}
	}', 'ContractConditions("MainCondition")'),
	('27', 'RunVDE', 'contract RunVDE {
		data {
			VDEName string
		}
	
		conditions {
		}
	
		action {
			StartVDE($VDEName)
		}
	}', 'ContractConditions("MainCondition")'),
	('28', 'StopVDE', 'contract StopVDE {
		data {
			VDEName string
		}
	
		conditions {
		}
	
		action {
			StopVDEProcess($VDEName)
		}
	}', 'ContractConditions("MainCondition")'),
	('29', 'RemoveVDE', 'contract RemoveVDE {
		data {
			VDEName string
		}
		conditions {}
		action{
			DeleteVDE($VDEName)
		}
	}', 'ContractConditions("MainCondition")');`
