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
    func rollback() {
        RollbackTable($Name)
    }
    func price() int {
        return SysParamInt("table_price")
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('9', 'UploadBinary', 'contract UploadBinary {
    data {
        ApplicationId int
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
    data {}

    func escapeSpecials(s string) string {
        s = Replace(s, ` + "`" + `\` + "`" + `, ` + "`" + `\\` + "`" + `)
        s = Replace(s, ` + "`" + `	` + "`" + `, ` + "`" + `\t` + "`" + `)
        s = Replace(s, "\n", ` + "`" + `\n` + "`" + `)
        s = Replace(s, "\r", ` + "`" + `\r` + "`" + `)
        s = Replace(s, ` + "`" + `"` + "`" + `, ` + "`" + `\"` + "`" + `)
        if s == "0"{
            s = ""
        }
        return s
    }

    func AssignAll(app_name string, resources string) string {
        return Sprintf(` + "`" + `{
            "name": "%%v",
            "data": [
                %%v
            ]
        }` + "`" + `, app_name, resources)
    }

    func serializeItem(item map, type string) string {
        var s string
        s = Sprintf(
            ` + "`" + `{
                "Type": "%%v",
                "Name": "%%v",
                "Value": "%%v",
                "Conditions": "%%v",
                "Menu": "%%v",
                "Title": "%%v",
                "Trans": "%%v",
                "Columns": "%%v",
                "Permissions": "%%v"
            }` + "`" + `, type, escapeSpecials(Str(item["name"])), escapeSpecials(Str(item["value"])), escapeSpecials(Str(item["conditions"])), escapeSpecials(Str(item["menu"])), escapeSpecials(Str(item["title"])), escapeSpecials(Str(item["res"])), escapeSpecials(Str(item["columns"])), escapeSpecials(Str(item["permissions"]))
        )
        return s
    }

    func getTypeForColumns(table_name string, columnsJSON string) string {
        var colsMap map, result columns array
        colsMap = JSONDecode(columnsJSON)
        columns = GetMapKeys(colsMap)
        var i int
        while i < Len(columns){
            if Size(columns[i]) > 0 {
                var col map
                col["name"] = columns[i]
                col["conditions"] = colsMap[col["name"]]
                col["type"] = GetColumnType(table_name, col["name"])
                result = Append(result, col)
            }
            i = i + 1
        }
        return JSONEncode(result)
    }

    func exportTable(type string, result array) array {
        var items array, limit offset int
        limit = 250
        while true{
            var rows array, where string
            if type == "menu" {
                if Len($menus_names) > 0 {
                    where = Sprintf("name in (%%v)", Join($menus_names, ","))
                }
            }else{
                where = Sprintf("app_id=%%v", $ApplicationID)
            }
            if where {
                rows = DBFind(type).Limit(limit).Offset(offset).Where(where)
            }
            if Len(rows) > 0{
                var i int
                while i<Len(rows){
                    items = Append(items, rows[i])
                    i=i+1
                }
            }else{
                break
            }
            offset = offset+limit
        }
        var i int, item map
        while i < Len(items) {
            item = items[i]
            if type == "tables" {
                var table map
                table["name"] = item["name"]
                table["permissions"] = item["permissions"]
                table["conditions"] = item["conditions"]
                table["columns"] = getTypeForColumns(item["name"], item["columns"])
                item = table
            }
            result = Append(result, serializeItem(item, type))
            if type == "pages" {
                $menus_names = Append($menus_names, Sprintf("''%%v''", item["menu"]))
            }
            i = i + 1
        }
        return result
    }

    conditions {
        var buffer_map map
        buffer_map = DBFind("buffer_data").Columns("id,value->app_id,value->app_name").Where("member_id=$ and key=$", $key_id, "export").Row()
        if !buffer_map{
            warning "Application not found"
        }
        $ApplicationID = Int(buffer_map["value.app_id"])
        $ApplicationName = Str(buffer_map["value.app_name"])

        var menus_names array
        $menus_names = menus_names
    }

    action {
        var exportJSON string, items array
        items = exportTable("pages", items)
        items = exportTable("contracts", items)
        items = exportTable("blocks", items)
        items = exportTable("languages", items)
        items = exportTable("app_params", items)
        items = exportTable("tables", items)
        items = exportTable("menu", items)

        exportJSON = AssignAll($ApplicationName, Join(items, ",\r\n"))
        UploadBinary("Name,Data,ApplicationId,DataMimeType", "export", exportJSON, 1, "application/json")
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
('12', 'ImportUpload', 'contract ImportUpload {
    data {
        input_file string "file"
    }
    func ReplaceValue(s string) string {
        s = Replace(s, "#ecosystem_id#", "#IMPORT_ECOSYSTEM_ID#")
        s = Replace(s, "#key_id#", "#IMPORT_KEY_ID#")
        s = Replace(s, "#isMobile#", "#IMPORT_ISMOBILE#")
        s = Replace(s, "#role_id#", "#IMPORT_ROLE_ID#")
        s = Replace(s, "#ecosystem_name#", "#IMPORT_ECOSYSTEM_NAME#")
        s = Replace(s, "#app_id#", "#IMPORT_APP_ID#")
        return s
    }

    conditions {
        $input_file = BytesToString($input_file)
        $input_file = ReplaceValue($input_file)
        $limit = 5 // data piece size of import

        // init buffer_data, cleaning old buffer
        var initJson map
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
        var input map
        input = JSONDecode($input_file)
        var arr_data array
        arr_data = input["data"]

        var pages_arr, blocks_arr, menu_arr, parameters_arr, languages_arr, contracts_arr, tables_arr array

        // import info
        var i int
        while i<Len(arr_data){
            var tmp_object map
            tmp_object = arr_data[i]

            if tmp_object["Type"] == "pages" {
                pages_arr = Append(pages_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "blocks" {
                blocks_arr = Append(blocks_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "menu" {
                menu_arr = Append(menu_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "app_params" {
                parameters_arr = Append(parameters_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "languages" {
                languages_arr = Append(languages_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "contracts" {
                contracts_arr = Append(contracts_arr, Str(tmp_object["Name"]))
            }
            if tmp_object["Type"] == "tables" {
                tables_arr = Append(tables_arr, Str(tmp_object["Name"]))
            }

            i = i + 1
        }

        var info_map map
        info_map["app_name"] = input["name"]
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

        // import data
        // the contracts is imported in one piece, the rest is cut under the $limit, a crutch to bypass the error when you import dependent contracts in different pieces
        i=0
        var sliced contracts array, arr_data_len int
        arr_data_len = Len(arr_data)
        while i <arr_data_len{
            var part array, l int, tmp map
            while l < $limit && (i+l < arr_data_len) {
                tmp = arr_data[i+l]
                if tmp["Type"] == "contracts" {
                    contracts = Append(contracts, tmp)
                }else{
                    part = Append(part, tmp)
                }
                l=l+1
            }
            var batch map
            batch["Data"] = JSONEncode(part)
            sliced = Append(sliced, batch)
            i=i+$limit
        }
        if Len(contracts) > 0{
            var batch map
            batch["Data"] = JSONEncode(contracts)
            sliced = Append(sliced, batch)
        }
        input["data"] = sliced

        // storing
        DBUpdate("buffer_data", $import_id, "value", input)
        DBUpdate("buffer_data", $info_id, "value", info_map)

        var app_id int
        app_id = DBFind("applications").Columns("id").Where("name=$", Str(input["name"])).One("id")

        if !app_id {
            DBInsert("applications", "name,conditions", Str(input["name"]), "true")
        }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('13', 'NewAppParam', 'contract NewAppParam {
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('18', 'NewLang', 'contract NewLang {
    data {
        ApplicationId int
        Name string
        Trans string
    }

    conditions {
        if $ApplicationId == 0 {
            warning "Application id cannot equal 0"
        }

        if DBFind("languages").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf( "Language resource %%s already exists", $Name)
        }

        EvalCondition("parameters", "changing_language", "value")
    }

    action {
        CreateLanguage($Name, $Trans, $ApplicationId)
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

        if DBFind("pages").Columns("id").Where("name = ?", $Name).One("id") {
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
        DBInsert("pages", "name,value,menu,validate_count,validate_mode,conditions,app_id", $Name, $Value, $Menu, $ValidateCount, $ValidateMode, $Conditions, $ApplicationId)
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('23', 'ExportNewApp', 'contract ExportNewApp {
    data {
        ApplicationId int
    }

    conditions {
        $app_map = DBFind("applications").Columns("id,name").Where("id=$", $ApplicationId).Row()
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
        pages_ret = DBFind("pages").Where("app_id=?", $ApplicationId)
        while i < Len(pages_ret) {
            var page_map map
            page_map = pages_ret[i]

            pages_array = Append(pages_array, Sprintf("''%%v''", Str(page_map["menu"])))
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

                menu_name_array = Append(menu_name_array, Str(menu_map["name"]))
                menu_id_array = Append(menu_id_array, Str(menu_map["id"]))
                i = i + 1
            }
        }

        //=====================================================================================================
        //------------------------------------Creating settings------------------------------------------------

        var value map
        value["app_id"] = Str($ApplicationId)
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
        Name string
        Value string
        Conditions string
    }
    conditions {
        ValidateCondition($Conditions, $ecosystem_id)

        if DBFind("signatures").Columns("id").Where("name = ?", $Name).One("id") {
            warning Sprintf("The signature %%s already exists", $Name)
        }
    }
    action {
        DBInsert("signatures", "name,value,conditions", $Name, $Value, $Conditions)  
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
        Name string
        Value string
        Conditions string "optional"
    }

    conditions {
        if GetContractByName($Name){
            var params map
            params["Value"] = $Value
            CallContract($Name, params)
        } else {
            warning "System parameter not found"
        }
    }

    action {
        DBUpdateSysParam($Name, $Value, $Conditions)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
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
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('48', 'NewLangJoint', 'contract NewLangJoint {
    data {
        ApplicationId int
        Name string
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
        params["ApplicationId"] = $ApplicationId 
        params["Name"] = $Name
        params["Trans"] = JSONEncode(Trans)
        CallContract("NewLang", params)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('49', 'NewSignJoint', 'contract NewSignJoint {
        data {
            Name string
            Title string
            ParamArr array
            ValueArr array
            Conditions string
        }
    
        conditions {
            var i int
            while i < Len($ParamArr) {
                if Size($ParamArr[i]) == 0 {
                    info("Parameter is empty")
                }
                if Size($ValueArr[i]) == 0 {
                    info("Value is empty")
                }
                i = i + 1
            }
        }
    
        action {
            var par_arr array
    
            var i int
            while i < Len($ParamArr) {
                var par_map map
                par_map["name"] = $ParamArr[i]
                par_map["text"] = $ValueArr[i]
                par_arr = Append(par_arr, JSONEncode(par_map))
                i = i + 1
            }
    
            var params map
            params["Name"] = $Name 
            params["Value"] = Sprintf(` + "`" + `{"title":"%%v","params":[%%v]}` + "`" + `, $Title, Join(par_arr, ","))
            params["Conditions"] = $Conditions
            CallContract("NewSign", params)
        }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('50', 'NewTableJoint', 'contract NewTableJoint {
    data {
        ApplicationId int
        Name string
        ColumnsArr array
        TypesArr array
        InsertPerm string
        UpdatePerm string
        NewColumnPerm string
    }

    conditions {
        var i int
        while i < Len($ColumnsArr) {
            if Size($ColumnsArr[i]) == 0 {
                info("Columns is empty")
            }
            if Size($TypesArr[i]) == 0 {
                info("Type is empty")
            }
            i = i + 1
        }
    }

    action {
        var i int
        var col_arr array
        while i < Len($ColumnsArr) {
            var col_map map
            col_map["name"] = $ColumnsArr[i]
            col_map["type"] = $TypesArr[i]
            col_map["conditions"] = "true"
            col_arr[i] = JSONEncode(col_map)
            i = i + 1
        }

        var Permissions map
        Permissions["insert"] = $InsertPerm 
        Permissions["update"] = $UpdatePerm
        Permissions["new_column"] = $NewColumnPerm

        var params map
        params["ApplicationId"] = $ApplicationId 
        params["Name"] = $Name
        params["Columns"] = JSONEncode(col_arr)
        params["Permissions"] = JSONEncode(Permissions)
        CallContract("NewTable", params)
    }
}', %[1]d, 'ContractConditions("MainCondition")', 1),
('51', 'blockchain_url', 'contract blockchain_url {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if !(HasPrefix($Value, "http://") || HasPrefix($Value, "https://")) {
        warning "URL ivalid (not found protocol)"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('52', 'block_reward', 'contract block_reward {
      data {
          Value string
      }
  
      conditions {
          if Size($Value) == 0 {
              warning "Value was not received"
          }
          if Int($Value) < 3 || Int($Value) > 9999 {
              warning "Value must be between 3 and 9999"
          }
      }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('53', 'column_price', 'contract column_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('54', 'commission_size', 'contract commission_size {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('55', 'commission_wallet', 'contract commission_wallet {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('56', 'contract_price', 'contract contract_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('57', 'default_ecosystem_contract', 'contract default_ecosystem_contract {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('58', 'default_ecosystem_menu', 'contract default_ecosystem_menu {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
    }
  }', %[1]d, 'ContractConditions("MainCondition")', 2),
('59', 'default_ecosystem_page', 'contract default_ecosystem_page {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('60', 'ecosystem_price', 'contract ecosystem_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('61', 'extend_cost_activate', 'contract extend_cost_activate {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('62', 'extend_cost_address_to_id', 'contract extend_cost_address_to_id {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('63', 'extend_cost_column_condition', 'contract extend_cost_column_condition {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('64', 'extend_cost_compile_contract', 'contract extend_cost_compile_contract {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('65', 'extend_cost_contains', 'contract extend_cost_contains {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('66', 'extend_cost_contracts_list', 'contract extend_cost_contracts_list {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('67', 'extend_cost_create_column', 'contract extend_cost_create_column {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('68', 'extend_cost_create_ecosystem', 'contract extend_cost_create_ecosystem {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('69', 'extend_cost_create_table', 'contract extend_cost_create_table {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('70', 'extend_cost_deactivate', 'contract extend_cost_deactivate {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('71', 'extend_cost_ecosys_param', 'contract extend_cost_ecosys_param {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('72', 'extend_cost_eval_condition', 'contract extend_cost_eval_condition {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('73', 'extend_cost_eval', 'contract extend_cost_eval {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('74', 'extend_cost_flush_contract', 'contract extend_cost_flush_contract {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('75', 'extend_cost_has_prefix', 'contract extend_cost_has_prefix {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('76', 'extend_cost_id_to_address', 'contract extend_cost_id_to_address {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('77', 'extend_cost_is_object', 'contract extend_cost_is_object {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('78', 'extend_cost_join', 'contract extend_cost_join {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('79', 'extend_cost_json_to_map', 'contract extend_cost_json_to_map {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('80', 'extend_cost_len', 'contract extend_cost_len {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('81', 'extend_cost_new_state', 'contract extend_cost_new_state {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('82', 'extend_cost_perm_column', 'contract extend_cost_perm_column {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('83', 'extend_cost_perm_table', 'contract extend_cost_perm_table {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('84', 'extend_cost_pub_to_id', 'contract extend_cost_pub_to_id {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('85', 'extend_cost_replace', 'contract extend_cost_replace {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('86', 'extend_cost_sha256', 'contract extend_cost_sha256 {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('87', 'extend_cost_size', 'contract extend_cost_size {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('88', 'extend_cost_substr', 'contract extend_cost_substr {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('89', 'extend_cost_sys_fuel', 'contract extend_cost_sys_fuel {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('90', 'extend_cost_sys_param_int', 'contract extend_cost_sys_param_int {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('91', 'extend_cost_sys_param_string', 'contract extend_cost_sys_param_string {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('92', 'extend_cost_table_conditions', 'contract extend_cost_table_conditions {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('93', 'extend_cost_update_lang', 'contract extend_cost_update_lang {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('94', 'extend_cost_validate_condition', 'contract extend_cost_validate_condition {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('95', 'fuel_rate', 'contract fuel_rate {
    data {
      Value string
    }
  
    conditions {
      $Value = TrimSpace($Value)
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      // [["x1","number"]]
      if !(HasPrefix($Value, "[") && "]" == Substr($Value, Size($Value)-1, 1)){
        warning "Invalid value"
      }
      var rates newRate array
      rates = JSONDecode($Value)
      if Len(rates) > 1{
        warning "Invalid size array"
      }
      newRate = rates[0]
      if Len(newRate) != 2{
        warning "Invalid size new rate array"
      }
      if newRate[0] != 1 {
        warning "Invalid ecosystem number"
      }
      if Int(newRate[1]) <= 0 {
        warning "Invalid fuel value"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('96', 'full_nodes', 'contract full_nodes {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
  
      var full_nodes_arr array
      full_nodes_arr = JSONDecode($Value)
  
      var len_arr int
      len_arr = Len(full_nodes_arr)
  
      if len_arr == 0 {
          warning "Wrong array structure"
      }
  
      var i int
      while(i < len_arr){
          var node_map map 
          node_map = full_nodes_arr[i]
  
          var public_key string
          var tcp_address string
          var api_address string
          var key_id string
  
          public_key = node_map["public_key"]
          tcp_address = node_map["tcp_address"]
          api_address = node_map["api_address"]
          key_id = node_map["key_id"]
  
          if Size(public_key) == 0 {
              warning "Public key was not received"
          }
          if Size(tcp_address) == 0 {
              warning "TCP address was not received"
          }
          if Size(api_address) == 0 {
              warning "API address was not received"
          }
          if Size(key_id) == 0 {
              warning "Key ID was not received"
          }
  
          i = i + 1
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('97', 'gap_between_blocks', 'contract gap_between_blocks {
      data {
          Value string
      }
  
      conditions {
          if Size($Value) == 0 {
              warning "Value was not received"
          }
          if Int($Value) <= 0 || Int($Value) >= 86400 {
              warning "Value must be between 1 and 86399"
          }
      }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('98', 'max_block_generation_time', 'contract max_block_generation_time {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('99', 'max_block_size', 'contract max_block_size {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('100', 'max_block_user_tx', 'contract max_block_user_tx {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('101', 'max_columns', 'contract max_columns {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('102', 'max_fuel_block', 'contract max_fuel_block {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('103', 'max_fuel_tx', 'contract max_fuel_tx {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('104', 'max_indexes', 'contract max_indexes {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('105', 'max_tx_count', 'contract max_tx_count {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('106', 'max_tx_size', 'contract max_tx_size {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('107', 'menu_price', 'contract menu_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('108', 'new_version_url', 'contract new_version_url {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('109', 'number_of_nodes', 'contract number_of_nodes {
      data {
          Value string
      }
  
      conditions {
          if Size($Value) == 0 {
              warning "Value was not received"
          }
          if Int($Value) < 1 || Int($Value) > 999 {
              warning "Value must be between 1 and 999"
          }
      }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('110', 'page_price', 'contract page_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('111', 'rb_blocks_1', 'contract rb_blocks_1 {
      data {
          Value string
      }
  
      conditions {
          if Size($Value) == 0 {
              warning "Value was not received"
          }
          if Int($Value) < 1 || Int($Value) > 999 {
              warning "Value must be between 1 and 999"
          }
      }
}', %[1]d, 'ContractConditions("MainCondition")', 2),
('112', 'table_price', 'contract table_price {
    data {
      Value string
    }
  
    conditions {
      if Size($Value) == 0 {
        warning "Value was not received"
      }
      if Int($Value) <= 0 {
        warning "Value must be greater than zero"
      }
    }
}', %[1]d, 'ContractConditions("MainCondition")', 2);
`
