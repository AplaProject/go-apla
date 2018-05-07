package migration

var firstEcosystemContractsSQL = `
INSERT INTO "1_contracts" (id, name, value, wallet_id, conditions, app_id)
VALUES ('2', 'DelApplication', 'contract DelApplication {
    data {
        ApplicationId int
        Value int "optional"
    }

    conditions {
        RowConditions("applications", $ApplicationId, false)
    }

    action {
        if $Value == 1 {
            DBUpdate("applications", $ApplicationId, "deleted", 1)
        }
        else {
            DBUpdate("applications", $ApplicationId, "deleted", 0)
        } 
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
			pars[Len(pars)] = "conditions"
			vals[Len(vals)] = $Conditions
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
			pars[Len(pars)] = "conditions"
			vals[Len(vals)] = $Conditions
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
        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }
	}
    
    action {
        if Size($Name) > 0 && Size($Columns) > 0 && Size($Permissions) > 0{
            CreateTable($Name, $Columns, $Permissions, $ApplicationId)
        } else {
            var i,len int
            var res string
            len = Len($Id)
			
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

    func ReplaceValue(s string) string {
		s = Replace(s, ` + "`" + `\` + "`" + `, ` + "`" + `\\` + "`" + `)
        s = Replace(s, ` + "`" + `	` + "`" + `, ` + "`" + `\t` + "`" + `)
        s = Replace(s, "\n", ` + "`" + `\n` + "`" + `)
        s = Replace(s, "\r", ` + "`" + `\r` + "`" + `)
        s = Replace(s, ` + "`" + `"` + "`" + `, ` + "`" + `\"` + "`" + `)
        return s
    }

    func AssignAll(app_name string, all_blocks string, all_contracts string, all_data string, all_languages string, all_menus string, all_pages string, all_parameters string, all_tables string) string {

        var res_str string
        res_str = res_str + all_blocks

        if  Size(res_str)>0 && Size(all_contracts)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_contracts

        if  Size(res_str)>0 && Size(all_data)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_data

        if  Size(res_str)>0 && Size(all_languages)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_languages

        if  Size(res_str)>0 && Size(all_menus)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_menus

        if  Size(res_str)>0 && Size(all_pages)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_pages

        if  Size(res_str)>0 && Size(all_parameters)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_parameters

        if  Size(res_str)>0 && Size(all_tables)>0  {
            res_str = res_str + ","
        }
        res_str = res_str + all_tables

        res_str = Sprintf(` + "`" + `{
    "name": "%%v",
    "data": [%%v
    ]
}` + "`" + `, app_name, res_str)

        return res_str
    }

    func AddPage(page_name string, page_value string, page_conditions string, page_menu string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "pages",
            "Name": "%%v",
            "Value": "%%v",
            "Conditions": "%%v",
            "Menu": "%%v"
        }` + "`" + `, page_name, page_value, page_conditions, page_menu)
        return s
    }

    func AddMenu(menu_name string, menu_value string, menu_title string, menu_conditions string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "menu",
            "Name": "%%v",
            "Value": "%%v",
            "Title": "%%v",
            "Conditions": "%%v"
        }` + "`" + `, menu_name, menu_value, menu_title, menu_conditions)
        return s
    }

    func AddContract(contract_name string, contract_value string, contract_conditions string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "contracts",
            "Name": "%%v",
            "Value": "%%v",
            "Conditions": "%%v"
        }` + "`" + `, contract_name, contract_value, contract_conditions)
        return s
    }

    func AddBlock(block_name string, block_value string, block_conditions string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "blocks",
            "Name": "%%v",
            "Value": "%%v",
            "Conditions": "%%v"
        }` + "`" + `, block_name, block_value, block_conditions)
        return s
    }

    func AddLanguage(language_name string, language_conditions string, language_trans string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "languages",
            "Name": "%%v",
            "Conditions": "%%v",
            "Trans": "%%v"
        }` + "`" + `, language_name, language_conditions, language_trans)
        return s
    }

    func AddParameter(parameter_name string, parameter_value string, parameter_conditions string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "app_params",
            "Name": "%%v",
            "Value": "%%v",
            "Conditions": "%%v"
        }` + "`" + `, parameter_name, parameter_value, parameter_conditions)
        return s
    }

    func AddTable(table_name string, table_columns string, table_permissions string) string {
        var s string
        s = Sprintf(` + "`" + `        {
            "Type": "tables",
            "Name": "%%v",
            "Columns": "%%v",
            "Permissions": "%%v"
        }` + "`" + `, table_name, table_columns, table_permissions)
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

					col_name = Replace(Str(clm[0]), ` + "`" + `"` + "`" + `, "")
					col_cond = Str(clm[1])
					col_type = GetColumnType(table_name, col_name)

					s = Sprintf(` + "`" + `{"name":"%%v","type":"%%v","conditions":%%v}` + "`" + `, col_name, col_type, col_cond)
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
        //warning $ApplicationID

        var full_result string
        var i int

        var all_blocks string
        var all_contracts string
        var all_data string
        var all_languages string
        var all_menus string
        var all_pages string
        var all_parameters string
        var all_tables string

        //=====================================================================================================
        //------------------------------------Export pages-----------------------------------------------------
        var string_for_menu string

        i = 0
        var pages_array array
        pages_array = DBFind("pages").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(pages_array) {
            var page_map map
            page_map = pages_array[i]

            var page_name string
            var page_value string
            var page_conditions string
            var page_menu string

            page_name = ReplaceValue(Str(page_map["name"]))
            page_value = ReplaceValue(Str(page_map["value"]))
            page_conditions = ReplaceValue(Str(page_map["conditions"]))
            page_menu = ReplaceValue(Str(page_map["menu"]))

            if Size(all_pages) > 0 {
                all_pages = all_pages + ",\r\n"
            } else {
                all_pages = all_pages + "\r\n"
            }

            if Size(string_for_menu) > 0 {
                string_for_menu = string_for_menu + ","
            }
            string_for_menu = string_for_menu + Sprintf("''%%v''", page_menu)           

            all_pages = all_pages + AddPage(page_name, page_value, page_conditions, page_menu)
            i = i + 1
        }

        //=====================================================================================================
        //------------------------------------Export menus-----------------------------------------------------
        if Size(string_for_menu) > 0 {

            var where_for_menu string
            where_for_menu = Sprintf("name in (%%v)", string_for_menu)
            //warning where_for_menu 

            i = 0
            var menus_array array
            menus_array = DBFind("menu").Limit(250).Where(where_for_menu)
            while i < Len(menus_array) {
                var menu_map map
                menu_map = menus_array[i]

                var menu_name string
                var menu_value string
                var menu_title string
                var menu_conditions string

                menu_name = ReplaceValue(Str(menu_map["name"]))
                menu_value = ReplaceValue(Str(menu_map["value"]))
                menu_title = ReplaceValue(Str(menu_map["title"]))
                menu_conditions = ReplaceValue(Str(menu_map["conditions"]))

                if Size(all_menus) > 0 {
                    all_menus = all_menus + ",\r\n"
                } else {
                    all_menus = all_menus + "\r\n"
                }

                all_menus = all_menus + AddMenu(menu_name, menu_value, menu_title, menu_conditions)
                i = i + 1
            }

        }

        //=====================================================================================================
        //------------------------------------Export contracts-------------------------------------------------

        i = 0
        var contracts_array array
        contracts_array = DBFind("contracts").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(contracts_array) {
            var contract_map map
            contract_map = contracts_array[i]

            var contract_name string
            var contract_value string
            var contract_conditions string

            contract_name = ReplaceValue(Str(contract_map["name"]))
            contract_value = ReplaceValue(Str(contract_map["value"]))
            contract_conditions = ReplaceValue(Str(contract_map["conditions"]))

            if Size(all_contracts) > 0 {
                all_contracts = all_contracts + ",\r\n"
            } else {
                all_contracts = all_contracts + "\r\n"
            }

            all_contracts = all_contracts + AddContract(contract_name, contract_value, contract_conditions)
            i = i + 1
        }

        //=====================================================================================================
        //------------------------------------Export blocks----------------------------------------------------

        i = 0
        var blocks_array array
        blocks_array = DBFind("blocks").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(blocks_array) {
            var block_map map
            block_map = blocks_array[i]

            var block_name string
            var block_value string
            var block_conditions string

            block_name = ReplaceValue(Str(block_map["name"]))
            block_value = ReplaceValue(Str(block_map["value"]))
            block_conditions = ReplaceValue(Str(block_map["conditions"]))

            if Size(all_blocks) > 0 {
                all_blocks = all_blocks + ",\r\n"
            } else {
                all_blocks = all_blocks + "\r\n"
            }

            all_blocks = all_blocks + AddBlock(block_name, block_value, block_conditions)
            i = i + 1
        }

        //=====================================================================================================
        //------------------------------------Export languages-------------------------------------------------

        i = 0
        var languages_array array
        languages_array = DBFind("languages").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(languages_array) {
            var language_map map
            language_map = languages_array[i]

            var language_name string
            var language_conditions string
            var language_trans string

            language_name = ReplaceValue(Str(language_map["name"]))
            language_conditions = ReplaceValue(Str(language_map["conditions"]))
            language_trans = ReplaceValue(Str(language_map["res"]))

            if Size(all_languages) > 0 {
                all_languages = all_languages + ",\r\n"
            } else {
                all_languages = all_languages + "\r\n"
            }

            all_languages = all_languages + AddLanguage(language_name, language_conditions, language_trans)
            i = i + 1
        }

        //=====================================================================================================
        //------------------------------------Export params----------------------------------------------------

        i = 0
        var parameters_array array
        parameters_array = DBFind("app_params").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(parameters_array) {
            var parameter_map map
            parameter_map = parameters_array[i]

            var parameter_name string
            var parameter_value string
            var parameter_conditions string

            parameter_name = ReplaceValue(Str(parameter_map["name"]))
            parameter_value = ReplaceValue(Str(parameter_map["value"]))
            parameter_conditions = ReplaceValue(Str(parameter_map["conditions"]))

            if Size(all_parameters) > 0 {
                all_parameters = all_parameters + ",\r\n"
            } else {
                all_parameters = all_parameters + "\r\n"
            }

            all_parameters = all_parameters + AddParameter(parameter_name, parameter_value, parameter_conditions)
            i = i + 1
        }

        //=====================================================================================================
        //------------------------------------Export tables----------------------------------------------------

        i = 0
        var tables_array array
        tables_array = DBFind("tables").Limit(250).Where("app_id=?", $ApplicationID)
        while i < Len(tables_array) {
            var table_map map
            table_map = tables_array[i]

            var table_name string
            var table_columns string
            var table_permissions string

            table_name = Str(table_map["name"])
            table_columns = Str(table_map["columns"])
			table_permissions = Str(table_map["permissions"])

			table_columns = AddTypeForColumns(table_name, table_columns)
            
			table_name = ReplaceValue(table_name)
			table_columns = ReplaceValue(table_columns)
			table_permissions = ReplaceValue(table_permissions)

            if Size(all_tables) > 0 {
                all_tables = all_tables + ",\r\n"
            } else {
                all_tables = all_tables + "\r\n"
            }

            all_tables = all_tables + AddTable(table_name, table_columns, table_permissions)
            i = i + 1
        }

        //=====================================================================================================

        full_result = AssignAll($ApplicationName, all_blocks, all_contracts, all_data, all_languages, all_menus, all_pages, all_parameters, all_tables)
        UploadBinary("Name,Data,ApplicationId,DataMimeType", "export", full_result, 1, "application/json")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('11', 'EditTable', 'contract EditTable {
	data {
		Name string
		Permissions string "optional"
        Insert_con string "optional"
    	Update_con string "optional"
    	New_column_con string "optional"
	}
	
	conditions {
        if !$Permissions {
            var permissions string
            permissions = Sprintf("{\"insert\":%%q,\"update\":%%q,\"new_column\":%%q}",$Insert_con,$Update_con,$New_column_con)
            $Permissions = permissions
        }
		TableConditions($Name, "", $Permissions)
	}
	
	action {
		PermTable($Name, $Permissions )
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
        var list array
        list = ContractsList($Value)

        if Len(list) == 0 {
            error "must be the name"
        }

        var i int
        while i < Len(list) {
            if IsObject(list[i], $ecosystem_id) {
                warning Sprintf("Contract or function %%s exists", list[i])
            }
            i = i + 1
        }

        $contract_name = list[0]
        if !$TokenEcosystem {
            $TokenEcosystem = 1
        } else {
            if !SysFuel($TokenEcosystem) {
                warning Sprintf("Ecosystem %%d is not system", $TokenEcosystem)
            }
        }
    }

    action {
        var root, id int
        root = CompileContract($Value, $ecosystem_id, $walletContract, $TokenEcosystem)
        id = DBInsert("contracts", "name,value,conditions, wallet_id, token_id,app_id", $contract_name, $Value, $Conditions, $walletContract, $TokenEcosystem, $ApplicationId)
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
			pars[Len(pars)] = "conditions"
			vals[Len(vals)] = $Conditions
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('26', 'EditPage', 'contract EditPage {
	data {
		Id int
		Value string "optional"
		Menu string "optional"
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
        MoneyTransfer("Recipient,Amount,Comment", Str($newId), Str($amount), "New user deposit")
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
	}
	action {
		DBInsert("bad_blocks", "producer_node_id,consumer_node_id,block_id,timestamp block_time", $ProducerNodeID, $ConsumerNodeID, $BlockID, $Timestamp)
	}
}', %[1]d, 'ContractConditions("NodeOwnerCondition")', 1),
('45', 'CheckNodesBan', 'contract CheckNodesBan {
	action {
		UpdateNodesBan($block_time)
	}
}', %[1]d, 'ContractConditions("MainCondition")', 1);
`
