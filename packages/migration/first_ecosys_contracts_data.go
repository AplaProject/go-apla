package migration

var firstEcosystemContractsSQL = `
INSERT INTO "1_contracts" (id, name, value, wallet_id, conditions, app_id)
VALUES ('2', 'DelApplication', 'contract DelApplication {
        data {
            ApplicationId int
            Value int "optional"
        }
    
        conditions {
            if $Value < 0 || $Value > 1 {
                error "Incorrect value"
            }
            RowConditions("applications", $ApplicationId, false)
        }
    
        action {
            DBUpdate("applications", $ApplicationId, "deleted", $Value)
        }
    }', %[1]d, 'ContractConditions("MainCondition")', 1),
('3', 'EditAppParam', 'contract EditAppParam {
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
        var pars, vals array
        if $Value {
            pars[0] = "value"
            vals[0] = $Value
        }
        if $Conditions {
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if Len(vals) > 0 {
            DBUpdate("app_params", $Id, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('4', 'EditApplication', 'contract EditApplication {
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
        var pars, vals array
        if $Conditions {
            pars[0] = "conditions"
            vals[0] = $Conditions
        }
        if Len(vals) > 0 {
            DBUpdate("applications", $ApplicationId, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('5', 'EditColumn', 'contract EditColumn {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('6', 'EditLang', 'contract EditLang {
    data {
        Id int
        Trans string
    }

    conditions {
        EvalCondition("parameters", "changing_language", "value")
        $lang = DBFind("languages").Where("id=?", $Id).Row()
    }

    action {
        EditLanguage($Id, $lang["name"], $Trans, Int($lang["app_id"]))
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('7', 'EditParameter', 'contract EditParameter {
    data {
        Id int
        Value string "optional"
        Conditions string "optional"
    }
    func onlyConditions() bool {
        return $Conditions && !$Value
    }

    conditions {
        RowConditions("parameters", $Id, onlyConditions())
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
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if Len(vals) > 0 {
            DBUpdate("parameters", $Id, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('8', 'NewTable', 'contract NewTable {
    data {
        ApplicationId int "optional"
        Name string "optional"
        Columns string "optional"
        Permissions string "optional"
        TableName string "optional"
        Id array "optional"
        Shareholding array "optional"
        Insert_con string "optional"
        Update_con string "optional"
        New_column_con string "optional"
    }
    conditions {
        if Size($Name) == 0 {
            error "Table name cannot be empty"
        }

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }
	}
    
    action {
        if Size($Name) > 0 && Size($Columns) > 0 && Size($Permissions) > 0{
            TableConditions($Name, $Columns, $Permissions)
            CreateTable($Name, $Columns, $Permissions, $ApplicationId)
        } else {
            var i,len int
            var res string
            len = Len($Id)
            
            if Size($TableName) == 0 {
                error "Table name cannot be empty"
            }
            
            while i < len {
                if i + 1 == len {
                    res = res + Sprintf("{\"name\":%%q,\"type\":%%q,\"conditions\":\"true\"}", $Id[i],$Shareholding[i])
                }
                else {
                    res = res + Sprintf("{\"name\":%%q,\"type\":%%q,\"conditions\":\"true\"},", $Id[i],$Shareholding[i])
                }
				i = i + 1
            }

            $Name = $TableName
            $Columns = Sprintf("["+"%%v"+"]", res)
            if !$Permissions {
                $Permissions = Sprintf("{\"insert\":%%q,\"update\":%%q,\"new_column\":%%q}",$Insert_con,$Update_con,$New_column_con)
            }
            TableConditions($Name, $Columns, $Permissions)
            CreateTable($Name, $Columns, $Permissions, $ApplicationId)
        }
    }
    func rollback() {
        RollbackTable($Name)
    }
    func price() int {
        return SysParamInt("table_price")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('9', 'UploadBinary', 'contract UploadBinary {
    data {
        ApplicationId int "optional"
        Name string
        Data bytes "file"
        DataMimeType string "optional"
    }

    conditions {
        $Id = Int(DBFind("binaries").Columns("id").Where("app_id = ? AND member_id = ? AND name = ?", $ApplicationId, $key_id, $Name).One("id"))
		
		if $Id == 0 {
			if $ApplicationId == 0 {
				warning "Application id cannot equal 0"
			}
		}
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
            $Id = DBInsert("binaries", "app_id,member_id,name,data,hash,mime_type", $ApplicationId, $key_id, $Name, $Data, hash, $DataMimeType)
        }

        $result = $Id
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('10', 'Export', 'contract Export {

    func EscapeSpecialSymbols(s string) string {
        s = Replace(s, ` + "`" + `\` + "`" + `, ` + "`" + `\\` + "`" + `)
        s = Replace(s, ` + "`" + `	` + "`" + `, ` + "`" + `\t` + "`" + `)
        s = Replace(s, "\n", ` + "`" + `\n` + "`" + `)
        s = Replace(s, "\r", ` + "`" + `\r` + "`" + `)
        s = Replace(s, ` + "`" + `"` + "`" + `, ` + "`" + `\"` + "`" + `)
        return s
    }

    func AssignAll(app_name string, resources string) string {
        return Sprintf(`{"name": "%%v", "data": [%%v]}`, app_name, resources)
    }

    func SerializeResource(resource map, resource_type string) string {
        var s string
        s = Sprintf(`        {
            "Type": "%%v",
            "Name": "%%v",
            "Value": "%%v",
            "Conditions": "%%v",
            "Menu": "%%v",
            "Title": "%%v",
            "Trans": "%%v",
            "Columns": "%%v"
        }`, 
            resource_type, EscapeSpecialSymbols(Str(resource["name"])), EscapeSpecialSymbols(Str(resource["value"])), EscapeSpecialSymbols(Str(resource["conditions"])),
            EscapeSpecialSymbols(Str(resource["menu"])), EscapeSpecialSymbols(Str(resource["title"])),
            EscapeSpecialSymbols(Str(resource["res"])), EscapeSpecialSymbols(Str(resource["columns"])))
        return s
    }

    func AddTypeForColumns(table_name string, table_columns string) string {
        var result string

        table_columns = Replace(table_columns, "{", "")
        table_columns = Replace(table_columns, "}", "")
        table_columns = Replace(table_columns, " ", "")

        var columns_arr array
        columns_arr = Split(table_columns, ",")

        var i int
        while (i < Len(columns_arr)){
            var s_split string
            s_split = Str(columns_arr[i])

            if Size(s_split) > 0 {
                var clm array
                clm = Split(s_split, ":")

                var s string

                if Len(clm) == 2 {
                    var col_name string
                    var col_cond string
                    var col_type string

                    col_name = Replace(Str(clm[0]), `"`, "")
                    col_cond = Str(clm[1])
                    col_type = GetColumnType(table_name, col_name)

                    s = Sprintf(`{"name":"%%v","type":"%%v","conditions":%%v}`, col_name, col_type, col_cond)
                }

                if Size(result) > 0 {
                    result = result + ","
                }
                result = result + s
            }
            i = i + 1
        }

        result = Sprintf("[%%v]", result)
        return result
    }

    func ExportTableRecords(records array, type string, entities_array array) array {
        var i int, cur_resource map
        i = 0
        while i < Len(records) {
            cur_resource = records[i]
            if type == "tables" {
                var table_name, table_columns string, table_map map
                table_map["name"] = Str(cur_resource["name"])
                table_map["columns"] = Str(cur_resource["columns"])
                table_map["columns"] = AddTypeForColumns(table_map["name"], table_map["columns"])
            }
            entities_array = Append(entities_array, SerializeResource(cur_resource, type))
            if type == "pages" {
                $menus_names = Append($menus_names, Sprintf("'%%v'", cur_resource["menu"]))
            }
            i = i + 1
        }
        return entities_array
    }

    data {}

    conditions {
        var buffer_map map
        buffer_map = DBFind("buffer_data").Columns("id,value->app_id,value->app_name").Where("member_id=$ and key=$", $key_id, "export").Row()
        if !buffer_map{
            warning "Application not found"
        }
        $ApplicationID = Int(buffer_map["value.app_id"])
        $ApplicationName = Str(buffer_map["value.app_name"])
    }

    action {
        var menus_names_arr array 
        $menus_names = menus_names_arr

        var full_result string
        var entities_array array
        var cur_resource map

        entities_array = ExportTableRecords(DBFind("pages").Limit(250).Where("app_id=?", $ApplicationID), "pages", entities_array)
        entities_array = ExportTableRecords(DBFind("contracts").Limit(250).Where("app_id=?", $ApplicationID), "contracts", entities_array)
        entities_array = ExportTableRecords(DBFind("blocks").Limit(250).Where("app_id=?", $ApplicationID), "blocks", entities_array)
        entities_array = ExportTableRecords(DBFind("languages").Limit(250).Where("app_id=?", $ApplicationID), "languages", entities_array)
        entities_array = ExportTableRecords(DBFind("app_params").Limit(250).Where("app_id=?", $ApplicationID), "params", entities_array)
        entities_array = ExportTableRecords(DBFind("tables").Limit(250).Where("app_id=?", $ApplicationID), "tables", entities_array)
        if Len($menus_names) > 0 {
            var where_for_menu string
            where_for_menu = Sprintf("name in (%%v)", Join($menus_names, ","))
            entities_array = ExportTableRecords(DBFind("menu").Limit(250).Where(where_for_menu), "menu", entities_array)
        }

        //=====================================================================================================

        full_result = AssignAll($ApplicationName, Join(entities_array, ",\r\n"))
        UploadBinary("Name,Data,ApplicationId,DataMimeType", "export", full_result, 1, "application/json")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('11', 'EditTable', 'contract EditTable {
    data {
        Name string
        InsertPerm string
        UpdatePerm string
        NewColumnPerm string
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
        $Permissions = permissions
        TableConditions($Name, "", JSONEncode($Permissions))
    }

    action {
        PermTable($Name, JSONEncode($Permissions))
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('12', 'Import_Upload', 'contract Import_Upload {
    data {
        input_file string "file"
    }

    conditions {
        $input_file = BytesToString($input_file)

        // init buffer_data, cleaning old buffer
        var initJson string
        initJson = "{}"
        $import_id = DBFind("buffer_data").Where("member_id=$ and key=$", $key_id, "import").One("id")
        if $import_id {
            $import_id = Int($import_id)
            DBUpdate("buffer_data", $import_id, "value", initJson)
        } else {
            $import_id = DBInsert("buffer_data", "member_id,key,value", $key_id, "import", initJson)
        }

        $info_id = DBFind("buffer_data").Where("member_id=$ and key=$", $key_id, "import_info").One("id")
        if $info_id {
            $info_id = Int($info_id)
            DBUpdate("buffer_data", $info_id, "value", initJson)
        } else {
            $info_id = DBInsert("buffer_data", "member_id,key,value", $key_id, "import_info", initJson)
        }
    }

    action {
        var json map
        json = JSONToMap($input_file)
        var arr_data array
        arr_data = json["data"]

        var pages_arr, blocks_arr, menu_arr, parameters_arr, languages_arr, contracts_arr, tables_arr array

        var i int
        while i<Len(arr_data){
            var tmp_object map
            tmp_object = arr_data[i]

            if tmp_object["Type"] == "pages" {
                pages_arr[Len(pages_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "blocks" {
                blocks_arr[Len(blocks_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "menu" {
                menu_arr[Len(menu_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "app_params" {
                parameters_arr[Len(parameters_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "languages" {
                languages_arr[Len(languages_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "contracts" {
                contracts_arr[Len(contracts_arr)] = Str(tmp_object["Name"])
            }
            if tmp_object["Type"] == "tables" {
                tables_arr[Len(tables_arr)] = Str(tmp_object["Name"])
            }

            i = i + 1
        }

        var info_map map
        info_map["app_name"] = json["name"]
        info_map["pages"] = Join(pages_arr, ", ")
        info_map["pages_count"] = Len(pages_arr)
        info_map["blocks"] = Join(blocks_arr, ", ")
        info_map["blocks_count"] = Len(blocks_arr)
        info_map["menu"] = Join(menu_arr, ", ")
        info_map["menu_count"] = Len(menu_arr)
        info_map["parameters"] = Join(parameters_arr, ", ")
        info_map["parameters_count"] = Len(parameters_arr)
        info_map["languages"] = Join(languages_arr, ", ")
        info_map["languages_count"] = Len(languages_arr)
        info_map["contracts"] = Join(contracts_arr, ", ")
        info_map["contracts_count"] = Len(contracts_arr)
        info_map["tables"] = Join(tables_arr, ", ")
        info_map["tables_count"] = Len(tables_arr)

        if 0 == Len(pages_arr) + Len(blocks_arr) + Len(menu_arr) + Len(parameters_arr) + Len(languages_arr) + Len(contracts_arr) + Len(tables_arr) {
            warning "Invalid or empty import file"
        }

        DBUpdate("buffer_data", $import_id, "value", $input_file)
        DBUpdate("buffer_data", $info_id, "value", info_map)

        var app_id int
        app_id = DBFind("applications").Columns("id").Where("name=$", Str(json["name"])).One("id")

        if !app_id {
            DBInsert("applications", "name,conditions", Str(json["name"]), "true")
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('13', 'NewAppParam', 'contract NewAppParam {
    data {
        ApplicationId int "optional"
        Name string
        Value string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("app_params").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Application parameter %%s already exists", $Name)
        }
    }

    action {
        DBInsert("app_params", "app_id,name,value,conditions", $ApplicationId, $Name, $Value, $Conditions)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('14', 'NewApplication', 'contract NewApplication {
    data {
        Name string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)
	
        if Size($Name) == 0 {
            warning "Application name missing"
        }

        if DBFind("applications").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Application %%s already exists", $Name)
        }
    }

    action {
        $result = DBInsert("applications", "name,conditions", $Name, $Conditions)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('15', 'NewBlock', 'contract NewBlock {
    data {
        ApplicationId int "optional"
        Name string
        Value string
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("blocks").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Block %%s already exists", $Name)
        }
    }

    action {
        DBInsert("blocks", "name,value,conditions,app_id", $Name, $Value, $Conditions, $ApplicationId)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('16', 'NewColumn', 'contract NewColumn {
    data {
        TableName string
        Name string
        Type string
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
        return SysParamInt("column_price")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('17', 'NewContract', 'contract NewContract {
    data {
        ApplicationId int "optional"
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('18', 'NewLang', 'contract NewLang {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('19', 'NewMenu', 'contract NewMenu {
    data {
        Name string
        Value string
        Title string "optional"
        Conditions string
    }

    conditions {
        ValidateCondition($Conditions,$ecosystem_id)

        if DBFind("menu").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Menu %%s already exists", $Name)
        }
    }

    action {
        DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions)
    }
    func price() int {
        return SysParamInt("menu_price")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('20', 'NewPage', 'contract NewPage {
    data {
        ApplicationId int "optional"
        Name string
        Value string
        Menu string
        Conditions string
        ValidateCount int "optional"
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

        if DBFind("pages").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Page %%s already exists", $Name)
        }

        $ValidateCount = preparePageValidateCount($ValidateCount)
    }

    action {
        DBInsert("pages", "name,value,menu,validate_count,conditions,app_id", $Name, $Value, $Menu, $ValidateCount, $Conditions, $ApplicationId)
    }
    func price() int {
        return SysParamInt("page_price")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('21', 'NewParameter', 'contract NewParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    
    conditions {
        ValidateCondition($Conditions, $ecosystem_id)
        
        if DBFind("parameters").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf("Parameter %%s already exists", $Name)
        }
    }
    
    action {
        DBInsert("parameters", "name,value,conditions", $Name, $Value, $Conditions)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('22', 'Import', 'contract Import {
    data {
        Type string
        Name string "optional"
        Value string "optional"
        Conditions string "optional"
        Menu string "optional"
        Trans string "optional"
        Columns string "optional"
        Permissions string "optional"
        Title string "optional"
    }

    conditions {
        Println(Sprintf("import: %%v, type: %%v, time: %%v", $Name, $Type, $time))
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
        var cdata, editors, creators, item map
        cdata["Value"] = $Value
        cdata["Conditions"] = $Conditions
        cdata["ApplicationId"] = $ApplicationId
        cdata["Name"] = $Name
        cdata["Title"] = $Title
        cdata["Trans"] = $Trans
        cdata["Menu"] = $Menu
        cdata["Columns"] = $Columns
        cdata["Permissions"] = $Permissions

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

        item = DBFind($Type).Where("name=?", $Name).Row()

        var contractName string
        if item {
            contractName = editors[$Type]
            cdata["Id"] = Int(item["id"])
            if $Type == "menu"{ 
                if Contains(item["value"], $Value) { 
                    // ignore repeated
                    contractName = ""
                }else{
                    cdata["Value"] = item["value"] + "\n" + $Value
                }
            }
        } else {
            contractName = creators[$Type]
        }

        if contractName != ""{
            CallContract(contractName, cdata)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('23', 'Export_NewApp', 'contract Export_NewApp {
    data {
        app_id int
    }

    conditions {
        $app_map = DBFind("applications").Columns("id,name").Where("id=$", $app_id).Row()
        if !$app_map{
            warning "Application not found"
        }
    }

    action {

        //=====================================================================================================
        //------------------------------------Menu search------------------------------------------------------
        var i int
        var pages_array array
        var menu_name_array array
		var menu_id_array array

        i = 0
        var pages_ret array
        pages_ret = DBFind("pages").Where("app_id=?", $app_id)
        while i < Len(pages_ret) {
            var page_map map
            page_map = pages_ret[i]

            pages_array[Len(pages_array)] = Sprintf("''%%v''", Str(page_map["menu"]))
            i = i + 1
        }

        if Len(pages_array) > 0 {
            var where_for_menu string
            where_for_menu = Sprintf("name in (%%v)", Join(pages_array, ","))

            i = 0
            var menu_ret array
            menu_ret = DBFind("menu").Where(where_for_menu)
            while i < Len(menu_ret) {
                var menu_map map
                menu_map = menu_ret[i]

                menu_name_array[Len(menu_name_array)] = Str(menu_map["name"])
				menu_id_array[Len(menu_id_array)] = Str(menu_map["id"])
                i = i + 1
            }
        }

        //=====================================================================================================
        //------------------------------------Creating settings------------------------------------------------
    
        var value map
        value["app_id"] = Str($app_id)
        value["app_name"] = Str($app_map["name"])
		
		if Len(menu_name_array) > 0 {
			value["menu_id"] = Str(Join(menu_id_array, ", "))
			value["menu_name"] = Str(Join(menu_name_array, ", "))
			value["count_menu"] = Str(Len(menu_name_array))
		} else {
			value["menu_id"] = "0"
			value["menu_name"] = ""
			value["count_menu"] = "0"
		}

        $buffer_id = DBFind("buffer_data").Where("member_id=$ and key=$", $key_id, "export").One("id")
        if !$buffer_id {
            DBInsert("buffer_data", "member_id,key,value", $key_id, "export", value)
        } else {
            DBUpdate("buffer_data", Int($buffer_id), "value", value)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('24', 'EditBlock', 'contract EditBlock {
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
        var pars, vals array
        if $Value {
            pars[0] = "value"
            vals[0] = $Value
        }
        if $Conditions {
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if Len(vals) > 0 {
            DBUpdate("blocks", $Id, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('25', 'EditMenu', 'contract EditMenu {
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
        var pars, vals array
        if $Value {
            pars[0] = "value"
            vals[0] = $Value
        }
        if $Title {
            pars = Append(pars, "title")
            vals = Append(vals, $Title)
        }
        if $Conditions {
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if Len(vals) > 0 {
            DBUpdate("menu", $Id, Join(pars, ","), vals...)
        }            
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('26', 'EditPage', 'contract EditPage {
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
        var pars, vals array
        if $Value {
            pars[0] = "value"
            vals[0] = $Value
        }
        if $Menu {
            pars = Append(pars, "menu")
            vals = Append(vals, $Menu)
        }
        if $Conditions {
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if $ValidateCount {
            pars = Append(pars, "validate_count")
            vals = Append(vals, $ValidateCount)
        }
        if $ValidateMode {
            if $ValidateMode != "1" {
                $ValidateMode = "0"
            }
            pars = Append(pars, "validate_mode")
            vals = Append(vals, $ValidateMode)
        }
        if Len(vals) > 0 {
            DBUpdate("pages", $Id, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('27', 'EditContract', 'contract EditContract {
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
            error Sprintf("Contract %d does not exist", $Id)
        }
        if $Value {
            ValidateEditContractNewValue($Value, $cur["value"])
        }
        if $WalletId != "" {
            $recipient = AddressToId($WalletId)
            if $recipient == 0 {
                error Sprintf("New contract owner %s is invalid", $WalletId)
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('28','MoneyTransfer','contract MoneyTransfer {
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
        var req money
		row = DBRow("keys").Columns("amount").WhereId($key_id)
        total = Money(row["amount"])
        req = $amount + Money(100000000000000000) 
        if req > total {
			error Sprintf("Money is not enough. You have got %%v but you should reserve %%v", total, req)
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('29','ActivateContract','contract ActivateContract {
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

}', %[1]d, 'ContractConditions("MainCondition")', 1),
('30','NewEcosystem','contract NewEcosystem {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('31','AppendMenu','contract AppendMenu {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('32','AppendPage','contract AppendPage {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('33','NewSign','contract NewSign {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('34','EditSign','contract EditSign {
    data {
        Id int
        Value string "optional"
        Conditions string "optional"
    }
    func onlyConditions() bool {
        return $Conditions && false
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
            pars = Append(pars, "conditions")
            vals = Append(vals, $Conditions)
        }
        if Len(vals) > 0 {
            DBUpdate("signatures", $Id, Join(pars, ","), vals...)
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('35','DeactivateContract','contract DeactivateContract {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('36','UpdateSysParam','contract UpdateSysParam {
	data {
		Name  string
		Value string
		Conditions string "optional"
	}
	action {
		DBUpdateSysParam($Name, $Value, $Conditions )
	}
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('37', 'NewDelayedContract','contract NewDelayedContract {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('38', 'EditDelayedContract','contract EditDelayedContract {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('39', 'CallDelayedContract','contract CallDelayedContract {
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

		var params map
		CallContract($cur["contract"], params)
	}
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('40', 'NewUser','contract NewUser {
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
}', %[1]d, 'ContractConditions("NodeOwnerCondition")', 1),
('41', 'EditEcosystemName','contract EditEcosystemName {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('42', 'UpdateMetrics', 'contract UpdateMetrics {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('43', 'NodeOwnerCondition', 'contract NodeOwnerCondition {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('44', 'NewBadBlock', 'contract NewBadBlock {
	data {
		ProducerNodeID int
		ConsumerNodeID int
		BlockID int
		Timestamp int
		Reason string
	}
	action {
		DBInsert("@1_bad_blocks", "producer_node_id,consumer_node_id,block_id,timestamp block_time,reason", $ProducerNodeID, $ConsumerNodeID, $BlockID, $Timestamp, $Reason)
	}
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('45', 'CheckNodesBan', 'contract CheckNodesBan {
	action {
		UpdateNodesBan($block_time)
	}
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('46', 'EditLangJoint', 'contract EditLangJoint {
    data {
        Id int
        ValueArr array
        LocaleArr array
    }

    conditions {
        var i int
        while i < Len($LocaleArr) {
            if Size($LocaleArr[i]) == 0 {
                info("Locale is empty")
            }
            if Size($ValueArr[i]) == 0 {
                info("Value is empty")
            }
            i = i + 1
        }
    }

    action {
        var i int
        var Trans map
        while i < Len($LocaleArr) {
            Trans[$LocaleArr[i]] = $ValueArr[i]
            i = i + 1
        }
        var params map
        params["Id"] = $Id 
        params["Trans"] = JSONEncode(Trans)
        CallContract("EditLang", params)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('47', 'EditSignJoint', 'contract EditSignJoint {
    data {
        Id int
        Title string
        Parameter string
        Conditions string
    }

    conditions {
        if !$Title {
            info("Title is empty")
        }
        if !$Parameter {
            info("Parameter is empty")
        }
    }

    action {
        var Value map
        Value["title"] = $Title 
        Value["params"] = $Parameter

        var params map
        params["Id"] = $Id 
        params["Value"] = JSONEncode(Value)
        params["Conditions"] = $Conditions
        CallContract("EditSign", params)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1);
`
