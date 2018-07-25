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
			id = DBInsert("contracts", {name: $contract_name, value: $Value,
				conditions: $Conditions, wallet_id: $walletContract, token_id: $TokenEcosystem,
				 app_id: $ApplicationId})
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
			var pars map

			if $Value {
				root = CompileContract($Value, $ecosystem_id, 0, 0)
				pars["value"] = $Value
			}
			if $Conditions {
				pars["conditions"] = $Conditions
			}
			if pars {
				DBUpdate("contracts", $Id, pars)
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
			  $result = DBInsert("parameters", {name: $Name, value: $Value, conditions: $Conditions})
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
			  DBUpdate("parameters", $Id, {"value": $Value,"conditions": $Conditions})
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
			DBInsert("menu", {name: $Name,value: $Value,title: $Title, conditions: $Conditions})
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
		  var pars map
		  if $Value {
			  pars["value"] = $Value
		  }
		  if $Title {
			  pars["title"] = $Title
		  }
		  if $Conditions {
			  pars["conditions"] = $Conditions
		  }
		  if pars {
			  DBUpdate("menu", $Id, pars)
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
			var val string
			val = row["value"] + "\r\n" + $Value
			DBUpdate("menu", $Id, {"value": val})
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
			DBInsert("pages", {name:$Name,value:$Value,menu:$Menu,validate_count:$ValidateCount,
				conditions:$Conditions,app_id:$ApplicationId,validate_mode:$ValidateMode})
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
			var pars map
			if $Value {
				pars["value"] = $Value
			}
			if $Menu {
				pars["menu"] = $Menu
			}
			if $Conditions {
				pars["conditions"] = $Conditions
			}
			if $ValidateCount {
				pars["validate_count"] = $ValidateCount
			}
			if $ValidateMode {
				if $ValidateMode != "1" {
					$ValidateMode = "0"
				}
				pars["validate_mode"] = $ValidateMode
			}
			if pars {
				DBUpdate("pages", $Id, pars)
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
			  var val string
			  val = row["value"] + "\r\n" + $Value
			  DBUpdate("pages", $Id, {"value": val})
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
			DBInsert("blocks", {name:$Name,value:$Value,conditions:$Conditions,app_id: $ApplicationId })
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
		  var pars map
		  if $Value {
			  pars["value"] = $Value
		  }
		  if $Conditions {
			  pars["conditions"] = $Conditions
		  }
		  if pars {
			  DBUpdate("blocks", $Id, pars)
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
				var list acol array
				var tblname string
				idata = row[i]
				i = i + 1
				tblname = idata["Table"]
				list = idata["Data"] 
				if !list {
					continue
				}
				var j int
				acol = idata["Columns"]
				while j < Len(list) {
					var pars map
					var ilist array
					ilist = list[j]
					var k int
					while k < Len(acol) {
						pars[acol[k]] = ilist[k]
						k = k + 1
					}
					DBInsert(tblname, pars)
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
			$result = DBInsert("cron", {owner: $key_id,cron:$Cron,contract: $Contract,
				counter:$Limit, till: $Till,conditions: $Conditions})
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
			DBUpdate("cron", $Id, {"cron": $Cron,"contract": $Contract,
			    "counter":$Limit, "till": $Till, "conditions":$Conditions})
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
				DBUpdate("binaries", $Id, {data: $Data,hash: hash,mime_type: $DataMimeType"})
			} else {
				$Id = DBInsert("binaries", {app_id: $AppID, member_id: $MemberID, name: $Name,
					data: $Data,hash: hash, mime_type: $DataMimeType })
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
			DBInsert("keys", {"id": $newId})
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
