INSERT INTO "system_states" ("id","rb_id") VALUES ('1','0');

INSERT INTO "1_contracts" ("id","value", "wallet_id", "conditions") VALUES 
('2','contract SystemFunctions {
}

func DBFind(table string).Columns(columns string).Where(where string, params ...)
     .WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
    return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
}

func DBString(table, column string, id int) string {
    var ret array
    var result string
    
    ret = DBFind(table).Columns(column).WhereId(id)
    if Len(ret) > 0 {
        var vmap map
        vmap = ret[0]
        result = vmap[column]
    }
    return result
}

func ConditionById(table string, validate bool) {
    var cond string
    cond = DBString(table, `conditions`, $Id)
    if !cond {
        error Sprintf(`Item %%d has not been found`, $Id)
    }
    Eval(cond)
    if validate {
        ValidateCondition($Conditions,$ecosystem_id)
    }
}

', '%[1]d','ContractConditions(`MainCondition`)'),
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
        total = Money(DBString(`keys`, `amount`, $key_id))
        if $amount >= total {
            error Sprintf("Money is not enough %%v < %%v",total, $amount)
        }
    }
    action {
        DBUpdate(`keys`, $key_id,`-amount`, $amount)
        DBUpdate(`keys`, $recipient,`+amount`, $amount)
        DBInsert(`history`, `sender_id,recipient_id,amount,comment,block_id,txhash`, 
            $key_id, $recipient, $amount, $Comment, $block, $txhash)
    }
}', '%[1]d', 'ContractConditions(`MainCondition`)'),
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
			   error Sprintf(`wrong wallet %%s`, $Wallet)
		    }
	    }
	    var list array
	    list = ContractsList($Value)
	    var i int
	    while i < Len(list) {
	        if IsContract(list[i], $ecosystem_id) {
	            warning Sprintf(`Contract %%s exists`, list[i] )
	        }
	        i = i + 1
	    }
        if !$TokenEcosystem {
            $TokenEcosystem = 1
        } else {
            if !SysFuel($TokenEcosystem) {
                warning Sprintf(`Ecosystem %%d is not system`, $TokenEcosystem )
            }
        }
    }
    action {
        var root, id int
        root = CompileContract($Value, $ecosystem_id, $walletContract, $TokenEcosystem)
        id = DBInsert(`contracts`, `value,conditions, wallet_id, token_id`, 
               $Value, $Conditions, $walletContract, $TokenEcosystem)
        FlushContract(root, id, false)
    }
    func price() int {
        return  SysParamInt(`contract_price`)
    }
}', '%[1]d', 'ContractConditions(`MainCondition`)'),
('5','contract EditContract {
    data {
        Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        $cur = DBRow(`contracts`, `id,value,conditions,active,wallet_id,token_id`, $Id)
        if Int($cur[`id`]) != $Id {
            error Sprintf(`Contract %%d does not exist`, $Id)
        }
        Eval($cur[`conditions`])
        ValidateCondition($Conditions,$ecosystem_id)
	    var list, curlist array
	    list = ContractsList($Value)
	    curlist = ContractsList($cur[`value`])
	    if Len(list) != Len(curlist) {
	        error `Contracts cannot be removed or inserted`
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
	            error `Contracts names cannot be changed`
	        }
	        i = i + 1
	    }
    }
    action {
        var root int
        root = CompileContract($Value, $ecosystem_id, Int($cur[`wallet_id`]), Int($cur[`token_id`]))
        DBUpdate(`contracts`, $Id, `value,conditions`, $Value, $Conditions)
        FlushContract(root, $Id, Int($cur[`active`]) == 1)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('6','contract ActivateContract {
    data {
        Id         int
    }
    conditions {
        $cur = DBRow(`contracts`, `id,conditions,active,wallet_id`, $Id)
        if Int($cur[`id`]) != $Id {
            error Sprintf(`Contract %%d does not exist`, $Id)
        }
        if Int($cur[`active`]) == 1 {
            error Sprintf(`The contract %%d has been already activated`, $Id)
        }
        Eval($cur[`conditions`])
        if $key_id != Int($cur[`wallet_id`]) {
            error Sprintf(`Wallet %%d cannot activate the contract`, $key_id)
        }
    }
    action {
        DBUpdate(`contracts`, $Id, `active`, 1)
        Activate($Id, $ecosystem_id)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('7','contract NewEcosystem {
    data {
        Name  string "optional"
    }
    conditions {
        if $Name && FindEcosystem($Name) {
            error Sprintf(`Ecosystem %%s is already existed`, $Name)
        }
    }
    action {
        var id int
        id = CreateEcosystem($key_id, $Name)
    	DBInsert(Str(id) + "_pages", "name,value,menu,conditions", `default_page`, 
              SysParamString(`default_ecosystem_page`), `default_menu`, "ContractConditions(`MainCondition`)")
    	DBInsert(Str(id) + "_menu", "name,value,title,conditions", `default_menu`, 
              SysParamString(`default_ecosystem_menu`), "default", "ContractConditions(`MainCondition`)")
    	DBInsert(Str(id) + "_keys", "id,pub", $key_id, DBString("1_keys", "pub", $key_id))
        $result = id
    }
    func price() int {
        return  SysParamInt(`ecosystem_price`)
    }
    func rollback() {
        RollbackEcosystem()
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('8','contract NewParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    conditions {
        ValidateCondition($Conditions, $ecosystem_id)
        if DBIntExt(`parameters`, `id`, $Name, `name`) {
            warning Sprintf( `Parameter %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(`parameters`, `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('9','contract EditParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    conditions {
        EvalCondition(`parameters`, $Name, `conditions`)
        ValidateCondition($Conditions, $ecosystem_id)
        var exist int
       	if $Name == `ecosystem_name` {
    		exist = FindEcosystem($Value)
    		if exist > 0 && exist != $ecosystem_id {
    			warning Sprintf(`Ecosystem %%s already exists`, $Value)
    		}
    	}
    }
    action {
        DBUpdateExt(`parameters`, `name`, $Name, `value,conditions`, $Value, $Conditions )
       	if $Name == `ecosystem_name` {
            DBUpdate(`system_states`, $ecosystem_id, `name`, $Value)
        }
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('10', 'contract NewMenu {
    data {
    	Name       string
    	Value      string
    	Title      string "optional"
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$ecosystem_id)
        if DBIntExt(`menu`, `id`, $Name, `name`) {
            warning Sprintf( `Menu %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(`menu`, `name,value,title,conditions`, $Name, $Value, $Title, $Conditions )
    }
    func price() int {
        return  SysParamInt(`menu_price`)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('11','contract EditMenu {
    data {
    	Id         int
    	Value      string
        Title      string "optional"
    	Conditions string
    }
    conditions {
        ConditionById(`menu`, true)
    }
    action {
        DBUpdate(`menu`, $Id, `value,title,conditions`, $Value, $Title, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('12','contract AppendMenu {
    data {
        Id     int
    	Value      string
    }
    conditions {
        ConditionById(`menu`, false)
    }
    action {
        DBUpdate(`menu`, $Id, `value`, DBString(`menu`, `value`, $Id) + "\r\n" + $Value )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('13','contract NewPage {
    data {
    	Name       string
    	Value      string
    	Menu       string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$ecosystem_id)
        if DBIntExt(`pages`, `id`, $Name, `name`) {
            warning Sprintf( `Page %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(`pages`, `name,value,menu,conditions`, $Name, $Value, $Menu, $Conditions )
    }
    func price() int {
        return  SysParamInt(`page_price`)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('14','contract EditPage {
    data {
        Id         int
    	Value      string
    	Menu      string
    	Conditions string
    }
    conditions {
        ConditionById(`pages`, true)
    }
    action {
        DBUpdate(`pages`, $Id, `value,menu,conditions`, $Value, $Menu, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('15','contract AppendPage {
    data {
        Id         int
    	Value      string
    }
    conditions {
        ConditionById(`pages`, false)
    }
    action {
        var value string
        value = DBString(`pages`, `value`, $Id)
       	if Contains(value, `PageEnd:`) {
		   value = Replace(value, "PageEnd:", $Value) + "\r\nPageEnd:"
    	} else {
    		value = value + "\r\n" + $Value
    	}
        DBUpdate(`pages`, $Id, `value`,  value )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('16','contract NewLang {
    data {
        Name  string
        Trans string
    }
    conditions {
        EvalCondition(`parameters`, `changing_language`, `value`)
        var exist string
        exist = DBStringExt(`languages`, `name`, $Name, `name`)
        if exist {
            error Sprintf("The language resource %%s already exists", $Name)
        }
    }
    action {
        DBInsert(`languages`, `name,res`, $Name, $Trans )
        UpdateLang($Name, $Trans)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('17','contract EditLang {
    data {
        Name  string
        Trans string
    }
    conditions {
        EvalCondition(`parameters`, `changing_language`, `value`)
    }
    action {
        DBUpdateExt(`languages`, `name`, $Name, `res`, $Trans )
        UpdateLang($Name, $Trans)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('18','contract NewSign {
    data {
    	Name       string
    	Value      string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$ecosystem_id)
        var exist string
        exist = DBStringExt(`signatures`, `name`, $Name, `name`)
        if exist {
            error Sprintf("The signature %%s already exists", $Name)
        }
    }
    action {
        DBInsert(`signatures`, `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('19','contract EditSign {
    data {
    	Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        ConditionById(`signatures`, true)
    }
    action {
        DBUpdate(`signatures`, $Id, `value,conditions`, $Value, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('20','contract NewBlock {
    data {
    	Name       string
    	Value      string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$ecosystem_id)
        if DBIntExt(`blocks`, `id`, $Name, `name`) {
            warning Sprintf( `Block %%s aready exists`, $Name)
        }
    }
    action {
        DBInsert(`blocks`, `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('21','contract EditBlock {
    data {
        Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        ConditionById(`blocks`, true)
    }
    action {
        DBUpdate(`blocks`, $Id, `value,conditions`, $Value, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
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
        return  SysParamInt(`table_price`)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('23','contract EditTable {
    data {
    	Name       string
    	Permissions string
    }
    conditions {
        TableConditions($Name, ``, $Permissions)
    }
    action {
        PermTable($Name, $Permissions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('24','contract NewColumn {
    data {
    	TableName   string
	    Name        string
	    Type        string
	    Permissions string
	    Index       string "optional"
    }
    conditions {
        ColumnCondition($TableName, $Name, $Type, $Permissions, $Index)
    }
    action {
        CreateColumn($TableName, $Name, $Type, $Permissions, $Index)
    }
    func rollback() {
        RollbackColumn($TableName, $Name)
    }
    func price() int {
        return  SysParamInt(`column_price`)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('25','contract EditColumn {
    data {
    	TableName   string
	    Name        string
	    Permissions string
    }
    conditions {
        ColumnCondition($TableName, $Name, ``, $Permissions, ``)
    }
    action {
        PermColumn($TableName, $Name, $Permissions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('26','func ImportList(row array, cnt string) {
    if !row {
        return
    }
    var i int
    while i < Len(row) {
        var idata map
        idata = row[i]
        CallContract(cnt, idata)
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
        tblname = idata[`Table`]
        columns = Join(idata[`Columns`], `,`)
        list = idata[`Data`] 
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
        i = i + 1
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
        ImportList($list["pages"], "NewPage")
        ImportList($list["blocks"], "NewBlock")
        ImportList($list["menus"], "NewMenu")
        ImportList($list["parameters"], "NewParameter")
        ImportList($list["languages"], "NewLang")
        ImportList($list["contracts"], "NewContract")
        ImportList($list["tables"], "NewTable")
        ImportData($list["data"])
    }
}', '%[1]d','ContractConditions(`MainCondition`)');