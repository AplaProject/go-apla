// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package vde

var contractsDataSQL = `
INSERT INTO "1_contracts" (id, name, value, conditions, app_id, ecosystem)
VALUES
	(next_id('1_contracts'), 'BindWallet', 'contract BindWallet {
	data {
		Id  int
	}
	conditions {
		$cur = DBRow("contracts").Columns("id,conditions,wallet_id").WhereId($Id)
		if !$cur {
			error Sprintf("Contract %%d does not exist", $Id)
		}
		
		Eval($cur["conditions"])
		if $key_id != Int($cur["wallet_id"]) {
			error Sprintf("Wallet %%d cannot activate the contract", $key_id)
		}
	}
	action {
		BndWallet($Id, $ecosystem_id)
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'CallDelayedContract', 'contract CallDelayedContract {
	data {
		Id int
	}
	conditions {
		var rows array
		rows = DBFind("delayed_contracts").Where({id: $Id, deleted: 0} )

		if !Len(rows) {
			error Sprintf("Delayed contract %%d does not exist", $Id)
		}
		$cur = rows[0]
		$limit = Int($cur["limit"])
		$counter = Int($cur["counter"])

		if $key_id != Int($cur["key_id"]) {
			error "Access denied"
		}

		if $block < Int($cur["block_id"]) {
			error Sprintf("Delayed contract %%d must run on block %%s, current block %%d", $Id, $cur["block_id"], $block)
		}

		if $limit > 0 && $counter >= $limit {
			error Sprintf("Delayed contract %%d is limited by number of launches", $Id)
		}
	}
	action {
		$counter = $counter + 1

		var block_id int
		block_id = $block
		if $limit == 0 || $limit > $counter {
			block_id = block_id + Int($cur["every_block"])
		}

		DBUpdate("delayed_contracts", $Id, {"counter": $counter, "block_id": block_id})

		var params map
		CallContract($cur["contract"], params)
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'CheckNodesBan', 'contract CheckNodesBan {
	action {
		UpdateNodesBan($block_time)
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditAppParam', 'contract EditAppParam {
    data {
        Id int
        Value string "optional"
        Conditions string "optional"
    }
    func onlyConditions() bool {
        return $Conditions && !$Value
    }

    conditions {
        RowConditions("app_params", $Id, onlyConditions())
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
            DBUpdate("app_params", $Id, pars)
        }
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditApplication', 'contract EditApplication {
    data {
        ApplicationId int
        Conditions string "optional"
    }
    func onlyConditions() bool {
        return $Conditions && false
    }

    conditions {
        RowConditions("applications", $ApplicationId, onlyConditions())
        if $Conditions {
            ValidateCondition($Conditions, $ecosystem_id)
        }
    }

    action {
        var pars map
        if $Conditions {
            pars["conditions"] = $Conditions
        }
        if pars {
            DBUpdate("applications", $ApplicationId, pars)
        }
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditBlock', 'contract EditBlock {
    data {
        Id int
        Value string "optional"
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
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditColumn', 'contract EditColumn {
    data {
        TableName string
        Name string
        Permissions string
    }

    conditions {
        ColumnCondition($TableName, $Name, "", $Permissions)
    }

    action {
        PermColumn($TableName, $Name, $Permissions)
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditContract', 'contract EditContract {
    data {
        Id int
        Value string "optional"
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
        $cur = DBFind("contracts").Columns("id,value,conditions,wallet_id,token_id").WhereId($Id).Row()
        if !$cur {
            error Sprintf("Contract %%d does not exist", $Id)
        }
        if $Value {
            ValidateEditContractNewValue($Value, $cur["value"])
        }
   
        $recipient = Int($cur["wallet_id"])
    }

    action {
        UpdateContract($Id, $Value, $Conditions, $recipient, $cur["token_id"])
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditCron', 'contract EditCron {
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
	}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditLang', 'contract EditLang {
    data {
        Id int
        Trans string
    }

    conditions {
        EvalCondition("parameters", "changing_language", "value")
        $lang = DBFind("languages").Where({id: $Id}).Row()
    }

    action {
        EditLanguage($Id, $lang["name"], $Trans)
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditMenu', 'contract EditMenu {
    data {
        Id int
        Value string "optional"
        Title string "optional"
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
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditPage', 'contract EditPage {
    data {
        Id int
        Value string "optional"
        Menu string "optional"
        Conditions string "optional"
        ValidateCount int "optional"
        ValidateMode string "optional"
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
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'EditTable', 'contract EditTable {
    data {
        Name string
        InsertPerm string
        UpdatePerm string
        NewColumnPerm string
        ReadPerm string "optional"
    }

    conditions {
        if !$InsertPerm {
            info("Insert condition is empty")
        }
        if !$UpdatePerm {
            info("Update condition is empty")
        }
        if !$NewColumnPerm {
            info("New column condition is empty")
        }

        var permissions map
        permissions["insert"] = $InsertPerm
        permissions["update"] = $UpdatePerm
        permissions["new_column"] = $NewColumnPerm
        if $ReadPerm {
            permissions["read"] = $ReadPerm
        }
        $Permissions = permissions
        TableConditions($Name, "", JSONEncode($Permissions))
    }

    action {
        PermTable($Name, JSONEncode($Permissions))
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'Import', 'contract Import {
    data {
        Data string
    }
    func decodeGlobals(s string) string {
        var globs array i lenGlobs int r map prefix from to string
        prefix = "import_"
        globs = ["guest_key", "ecosystem_id", "key_id", "isMobile", "role_id", "ecosystem_name", "app_id", "info_value_pages_count", "info_value_pages", "info_value_blocks_count", "info_value_blocks","info_value_menu_count", "info_value_menu", "info_value_parameters_count", "info_value_parameters","info_value_languages_count", "info_value_languages", "info_value_contracts_count", "info_value_contracts", "info_value_tables_count", "info_value_tables", "DataName", "DataCount", "DataInfo", "info_value_app_name", "import_id"]
        lenGlobs = Len(globs)
        while i < lenGlobs{
            r = globs[i]
            from = ToUpper("#" + prefix + r + "#")
            to = "#" + r + "#"
            s = Replace(s, from, to)
            i = i + 1
        }
        return s
    }

    conditions {
        $Data = decodeGlobals($Data)
        $ApplicationId = 0
        var app_map map
        app_map = DBFind("@1buffer_data").Columns("value->app_name").Where({key: "import_info", member_id: $key_id, ecosystem: $ecosystem_id}).Row()

        if app_map{
            var app_id int ival string
            ival = Str(app_map["value.app_name"])
            app_id = DBFind("@1applications").Columns("id").Where({name: ival, ecosystem: $ecosystem_id}).One("id")
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
        while i < Len(dataImport){
            var item cdata map type name string
            cdata = dataImport[i]
            if cdata {
                cdata["ApplicationId"] = $ApplicationId
                type = cdata["Type"]
                name = cdata["Name"]
                // Println(Sprintf("import %%v: %%v", type, cdata["Name"]))

                var tbl string
                tbl = "@1" + Str(type)
                item = DBFind(tbl).Where({name: name, ecosystem: $ecosystem_id}).Row()
                var contractName string
                if item {
                    contractName = editors[type]
                    cdata["Id"] = Int(item["id"])
                    if type == "contracts" {
                        if item["conditions"] == "false"{
                            // ignore updating impossibled
                            contractName = ""
                        }
                    } elif type == "menu"{
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
                    contractName = creators[type]
                }

                if contractName != ""{
                    CallContract(contractName, cdata)
                }
            }
            i = i + 1
        }
        // Println(Sprintf("> time: %%v", $time))
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'ImportUpload', 'contract ImportUpload {
    data {
        input_file file
    }
    func encodeGlobals(s string) string {
        var globs array i lenGlobs int r map prefix from to string
        prefix = "import_"
        globs = ["guest_key", "ecosystem_id", "key_id", "isMobile", "role_id", "ecosystem_name", "app_id", "info_value_pages_count", "info_value_pages", "info_value_blocks_count", "info_value_blocks","info_value_menu_count", "info_value_menu", "info_value_parameters_count", "info_value_parameters","info_value_languages_count", "info_value_languages", "info_value_contracts_count", "info_value_contracts", "info_value_tables_count", "info_value_tables", "DataName", "DataCount", "DataInfo", "info_value_app_name", "import_id"]
        lenGlobs = Len(globs)
        while i < lenGlobs{
            r = globs[i]
            from = "#" + r + "#"
            to = ToUpper("#" + prefix + r + "#")
            s = Replace(s, from, to)
            i = i + 1
        }
        return s
    }

    conditions {
        $input_file = BytesToString($input_file["Body"])
        $input_file = encodeGlobals($input_file)
        $limit = 10 // data piece size of import
    }

    action {
        // init buffer_data, cleaning old buffer
        var initJson map
        $import_id = DBFind("@1buffer_data").Where({member_id:$key_id, key: "import", ecosystem: $ecosystem_id}).One("id")
        if $import_id {
            $import_id = Int($import_id)
            DBUpdate("@1buffer_data", $import_id, {"value": initJson})
        } else {
            $import_id = DBInsert("@1buffer_data", {"member_id":$key_id,"key": "import", "value": initJson,"ecosystem": $ecosystem_id})
        }

        $info_id = DBFind("@1buffer_data").Where({member_id:$key_id, key: "import_info", ecosystem: $ecosystem_id}).One("id")
        if $info_id {
            $info_id = Int($info_id)
            DBUpdate("@1buffer_data", $info_id, {"value": initJson})
        } else {
            $info_id = DBInsert("@1buffer_data", {"member_id":$key_id,"key": "import_info", "value": initJson,"ecosystem": $ecosystem_id})
        }
        
        var input map arrData array
        input = JSONDecode($input_file)
        arrData = input["data"]

        var pages_arr blocks_arr menu_arr parameters_arr languages_arr contracts_arr tables_arr array

        // import info
        var i lenArrData int item map
        lenArrData = Len(arrData)
        while i < lenArrData{
            item = arrData[i]

            if item["Type"] == "pages" {
                pages_arr = Append(pages_arr, item["Name"])
            }elif item["Type"] == "blocks" {
                blocks_arr = Append(blocks_arr, item["Name"])
            }elif item["Type"] == "menu" {
                menu_arr = Append(menu_arr, item["Name"])
            }elif item["Type"] == "app_params" {
                parameters_arr = Append(parameters_arr, item["Name"])
            }elif item["Type"] == "languages" {
                languages_arr = Append(languages_arr, item["Name"])
            }elif item["Type"] == "contracts" {
                contracts_arr = Append(contracts_arr, item["Name"])
            }elif item["Type"] == "tables" {
                tables_arr = Append(tables_arr, item["Name"])
            }

            i = i + 1
        }

        var inf map
        inf["app_name"] = input["name"]
        inf["pages"] = Join(pages_arr, ", ")
        inf["pages_count"] = Len(pages_arr)
        inf["blocks"] = Join(blocks_arr, ", ")
        inf["blocks_count"] = Len(blocks_arr)
        inf["menu"] = Join(menu_arr, ", ")
        inf["menu_count"] = Len(menu_arr)
        inf["parameters"] = Join(parameters_arr, ", ")
        inf["parameters_count"] = Len(parameters_arr)
        inf["languages"] = Join(languages_arr, ", ")
        inf["languages_count"] = Len(languages_arr)
        inf["contracts"] = Join(contracts_arr, ", ")
        inf["contracts_count"] = Len(contracts_arr)
        inf["tables"] = Join(tables_arr, ", ")
        inf["tables_count"] = Len(tables_arr)

        if 0 == inf["pages_count"] + inf["blocks_count"] + inf["menu_count"] + inf["parameters_count"] + inf["languages_count"] + inf["contracts_count"] + inf["tables_count"] {
            warning "Invalid or empty import file"
        }

        // import data
        // the contracts is imported in one piece, the rest is cut under the $limit, a crutch to bypass the error when you import dependent contracts in different pieces
        var sliced contracts array
        i = 0
        while i <lenArrData{
            var items array l int item map
            while l < $limit && (i + l < lenArrData) {
                item = arrData[i + l]
                if item["Type"] == "contracts" {
                    contracts = Append(contracts, item)
                }else{
                    items = Append(items, item)
                }
                l = l + 1
            }
            var batch map
            batch["Data"] = JSONEncode(items)
            sliced = Append(sliced, batch)
            i = i + $limit
        }
        if Len(contracts) > 0{
            var batch map
            batch["Data"] = JSONEncode(contracts)
            sliced = Append(sliced, batch)
        }
        input["data"] = sliced

        // storing
        DBUpdate("@1buffer_data", $import_id, {"value": input})
        DBUpdate("@1buffer_data", $info_id, {"value": inf})

        var name string
        name = Str(input["name"])
        var cndns string
        cndns = Str(input["conditions"])

        if !DBFind("@1applications").Columns("id").Where({name:name, ecosystem: $ecosystem_id}).One("id") {
            DBInsert("@1applications", {"name": name, "conditions": cndns,"ecosystem": $ecosystem_id})
        }
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'ListVDE', 'contract ListVDE {
		data {}
	
		conditions {}
	
		action {
			$result = GetVDEList()
		}
	}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'MainCondition', 'contract MainCondition {
		conditions {
		  if EcosysParam("founder_account")!=$key_id
		  {
			warning "Sorry, you do not have access to this action."
		  }
		}
	  }
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewAppParam', 'contract NewAppParam {
    data {
        ApplicationId int
        Name string
        Value string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("app_params").Columns("id").Where({"name":$Name}).One("id") {
            warning Sprintf( "Application parameter %%s already exists", $Name)
        }
    }

    action {
        DBInsert("app_params", {app_id: $ApplicationId, name: $Name, value: $Value,
              conditions: $Conditions})
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewApplication', 'contract NewApplication {
    data {
        Name string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if Size($Name) == 0 {
            warning "Application name missing"
        }

        if DBFind("applications").Columns("id").Where({name:$Name}).One("id") {
            warning Sprintf( "Application %%s already exists", $Name)
        }
    }

    action {
        $result = DBInsert("applications", {name: $Name,conditions: $Conditions})
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewBadBlock', 'contract NewBadBlock {
	data {
		ProducerNodeID int
		ConsumerNodeID int
		BlockID int
		Timestamp int
		Reason string
	}
	action {
        DBInsert("@1bad_blocks", {producer_node_id: $ProducerNodeID,consumer_node_id: $ConsumerNodeID,
            block_id: $BlockID, "timestamp block_time": $Timestamp, reason: $Reason})
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewBlock', 'contract NewBlock {
    data {
        ApplicationId int
        Name string
        Value string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("blocks").Columns("id").Where({name:$Name}).One("id") {
            warning Sprintf( "Block %%s already exists", $Name)
        }
    }

    action {
        DBInsert("blocks", {name: $Name, value: $Value, conditions: $Conditions,
              app_id: $ApplicationId})
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewContract', 'contract NewContract {
    data {
        ApplicationId int
        Value string
        Conditions string
        TokenEcosystem int "optional"
    }

    conditions {
        ValidateCondition($Conditions,$ecosystem_id)

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
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
        $result = CreateContract($contract_name, $Value, $Conditions, $TokenEcosystem, $ApplicationId)
    }
    func price() int {
        return SysParamInt("contract_price")
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewCron', 'contract NewCron {
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
	}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewEcosystem', 'contract NewEcosystem {
	data {
		Name  string
	}
	action {
		$result = CreateEcosystem($key_id, $Name)
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewLang', 'contract NewLang {
    data {
        ApplicationId int
        Name string
        Trans string
    }

    conditions {
        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("languages").Columns("id").Where({name: $Name}).One("id") {
            warning Sprintf( "Language resource %%s already exists", $Name)
        }

        EvalCondition("parameters", "changing_language", "value")
    }

    action {
        CreateLanguage($Name, $Trans)
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewMenu', 'contract NewMenu {
    data {
        Name string
        Value string
        Title string "optional"
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions,$ecosystem_id)

        if DBFind("menu").Columns("id").Where({name: $Name}).One("id") {
            warning Sprintf( "Menu %%s already exists", $Name)
        }
    }

    action {
        DBInsert("menu", {name:$Name,value: $Value, title: $Title, conditions: $Conditions})
    }
    func price() int {
        return SysParamInt("menu_price")
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewPage', 'contract NewPage {
    data {
        ApplicationId int
        Name string
        Value string
        Menu string
        Conditions string
        ValidateCount int "optional"
        ValidateMode string "optional"
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

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("pages").Columns("id").Where({name: $Name}).One("id") {
            warning Sprintf( "Page %%s already exists", $Name)
        }

        $ValidateCount = preparePageValidateCount($ValidateCount)

        if $ValidateMode {
            if $ValidateMode != "1" {
                $ValidateMode = "0"
            }
        }
    }

    action {
        DBInsert("pages", {name: $Name,value: $Value, menu: $Menu,
             validate_count:$ValidateCount,validate_mode: $ValidateMode,
             conditions: $Conditions,app_id: $ApplicationId})
    }
    func price() int {
        return SysParamInt("page_price")
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewTable', 'contract NewTable {
    data {
        ApplicationId int
        Name string
        Columns string
        Permissions string
    }
    conditions {
        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }
        TableConditions($Name, $Columns, $Permissions)
    }
    
    action {
        CreateTable($Name, $Columns, $Permissions, $ApplicationId)
    }
    func price() int {
        return SysParamInt("table_price")
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewUser', 'contract NewUser {
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
        NewMoney($newId, Str($amount), "New user deposit")
        SetPubKey($newId, StringToBytes($NewPubkey))
	}
}
', 'ContractConditions("NodeOwnerCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NewVDE', 'contract NewVDE {
		data {
			VDEName string
			DBUser string
			DBPassword string
			VDEAPIPort int
		}
	
		conditions {
            if Size($VDEName) == 0 {
                warning "VDEName was not received"
            }
            if Contains($VDEName, " ") {
                error "VDEName can not contain spaces"
            }
            if Size($DBUser) == 0 {
                warning "DBUser was not received"
            }
            if Size($DBPassword) == 0 {
                warning "DBPassword was not received"
            }
            if $VDEAPIPort <= 0  {
                warning "VDE API PORT not received"
            }
            
		}
	
		action {
            $VDEName = ToLower($VDEName)
            $DBUser = ToLower($DBUser)
            CreateVDE($VDEName, $DBUser, $DBPassword, $VDEAPIPort)
            $result = "VDE " + $VDEName + " created"
		}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'NodeOwnerCondition', 'contract NodeOwnerCondition {
	conditions {
        $raw_full_nodes = SysParamString("full_nodes")
        if Size($raw_full_nodes) == 0 {
            ContractConditions("MainCondition")
        } else {
            $full_nodes = JSONDecode($raw_full_nodes)
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
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'RemoveVDE', 'contract RemoveVDE {
	data {
			VDEName string
	}
	conditions {}
	action{
        $VDEName = ToLower($VDEName)
        DeleteVDE($VDEName)
        $result = "VDE " + $VDEName + " removed"
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'RunVDE', 'contract RunVDE {
		data {
			VDEName string
		}
	
		conditions {
		}
	
		action {
            $VDEName = ToLower($VDEName)
            StartVDE($VDEName)
            $result = "VDE " + $VDEName + " running"
		}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'StopVDE', 'contract StopVDE {
		data {
			VDEName string
		}
	
		conditions {
		}
	
		action {
            $VDEName = ToLower($VDEName)
            StopVDEProcess($VDEName)
            $result = "VDE " + $VDEName + " stopped"
		}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'UnbindWallet', 'contract UnbindWallet {
	data {
		Id         int
	}
	conditions {
		$cur = DBRow("contracts").Columns("id,conditions,wallet_id").WhereId($Id)
		if !$cur {
			error Sprintf("Contract %%d does not exist", $Id)
		}
		
		Eval($cur["conditions"])
		if $key_id != Int($cur["wallet_id"]) {
			error Sprintf("Wallet %%d cannot deactivate the contract", $key_id)
		}
	}
	action {
		UnbndWallet($Id, $ecosystem_id)
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'UpdateMetrics', 'contract UpdateMetrics {
	conditions {
		ContractConditions("MainCondition")
	}
	action {
		var values array
		values = DBCollectMetrics()

		var i, id int
		var v map
		while (i < Len(values)) {
            var inmap map

            v = values[i]
            inmap["time"] = v["time"]
            inmap["key"] = v["key"]
            inmap["metric"] = v["metric"]
            
            id = Int(DBFind("metrics").Columns("id").Where(inmap).One("id"))
            var ival int
			if id != 0 {
                ival = Int(v["value"])
				DBUpdate("metrics", id, {"value": ival})
			} else {
                inmap["value"] = Int(v["value"])
				DBInsert("metrics", inmap )
			}
			i = i + 1
		}
	}
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'UpdateSysParam', 'contract UpdateSysParam {
     data {
        Name string
        Value string
        Conditions string "optional"
     }
     conditions {
         if !GetContractByName($Name){
            warning "System parameter not found"
         }
     }
     action {
        var params map
        params["Value"] = $Value
        CallContract($Name, params)
        
        DBUpdateSysParam($Name, $Value, $Conditions)
     }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'UploadBinary', 'contract UploadBinary {
    data {
        ApplicationId int
        Name string
        Data bytes
        DataMimeType string "optional"
    }

    conditions {
        $Id = Int(DBFind("@1binaries").Columns("id").Where({app_id: $ApplicationId,
            member_id: $key_id, name: $Name, ecosystem: $ecosystem_id}).One("id"))

        if $Id == 0 {
            if $ApplicationId == 0 {
                warning LangRes("@1aid_cannot_zero", "en")
            }
        }
    }
    action {
        var hash  string
        hash = Hash($Data)

        if $DataMimeType == "" {
            $DataMimeType = "application/octet-stream"
        }

        if $Id != 0 {
            DBUpdate("@1binaries", $Id, {"data":$Data, "hash":hash, "mime_type":$DataMimeType})
        } else {
            $Id = DBInsert("@1binaries", {"app_id":$ApplicationId, "member_id":$key_id,
               "name":$Name, "data":$Data, "hash":hash, "mime_type":$DataMimeType, "ecosystem":$ecosystem_id})
        }

        $result = $Id
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d'),
	(next_id('1_contracts'), 'UploadFile', 'contract UploadFile {
    data {
        ApplicationId int
        Data file
        Name string "optional"
    }

    conditions {
        if $Name == "" {
            $Name = $Data["Name"]
        }
        $Body = $Data["Body"]
        $DataMimeType = $Data["MimeType"] 
    }
    
    action {
        $Id = @1UploadBinary("ApplicationId,Name,Data,DataMimeType", $ApplicationId, $Name, $Body, $DataMimeType)
        $result = $Id
    }
}
', 'ContractConditions("MainCondition")', '1', '%[1]d');
`
