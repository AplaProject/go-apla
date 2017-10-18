INSERT INTO "system_states" ("id","rb_id") VALUES ('1','0');

INSERT INTO "1_contracts" ("id","value", "wallet_id", "conditions") VALUES 
('2','contract SystemFunctions {
}

func ConditionById(table string, validate bool) {
    var cond string
    cond = DBString(Table(table), `conditions`, $Id)
    if !cond {
        error Sprintf(`Item %%d has not been found`, $Id)
    }
    Eval(cond)
    if validate {
        ValidateCondition($Conditions,$state)
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
        total = Money(DBString(Table(`keys`), `amount`, $wallet))
        if $amount >= total {
            error Sprintf("Money is not enough %%v < %%v",total, $amount)
        }
    }
    action {
        DBUpdate(Table(`keys`), $wallet,`-amount`, $amount)
        DBUpdate(Table(`keys`), $recipient,`+amount`, $amount)
        DBInsert(Table(`history`), `sender_id,recipient_id,amount,comment,block_id,txhash`, 
            $wallet, $recipient, $amount, $Comment, $block, $txhash)
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
        ValidateCondition($Conditions,$state)
        $walletContract = $wallet
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
	        if IsContract(list[i], $state) {
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
        root = CompileContract($Value, $state, $walletContract, $TokenEcosystem)
        id = DBInsert(Table(`contracts`), `value,conditions, wallet_id, token_id`, 
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
        $cur = DBRow(Table(`contracts`), `id,value,conditions,active,wallet_id,token_id`, $Id)
        if Int($cur[`id`]) != $Id {
            error Sprintf(`Contract %%d does not exist`, $Id)
        }
        Eval($cur[`conditions`])
        ValidateCondition($Conditions,$state)
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
        root = CompileContract($Value, $state, Int($cur[`wallet_id`]), Int($cur[`token_id`]))
        DBUpdate(Table(`contracts`), $Id, `value,conditions`, $Value, $Conditions)
        FlushContract(root, $Id, Int($cur[`active`]) == 1)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('6','contract ActivateContract {
    data {
        Id         int
    }
    conditions {
        $cur = DBRow(Table(`contracts`), `id,conditions,active,wallet_id`, $Id)
        if Int($cur[`id`]) != $Id {
            error Sprintf(`Contract %%d does not exist`, $Id)
        }
        if Int($cur[`active`]) == 1 {
            error Sprintf(`The contract %%d has been already activated`, $Id)
        }
        Eval($cur[`conditions`])
        if $wallet != Int($cur[`wallet_id`]) {
            error Sprintf(`Wallet %%d cannot activate the contract`, $wallet)
        }
    }
    action {
        DBUpdate(Table(`contracts`), $Id, `active`, 1)
        Activate($Id, $state)
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
        id = CreateEcosystem($wallet, $Name)
    	DBInsert(Str(id) + "_pages", "name,value,menu,conditions", `default_page`, 
              SysParamString(`default_ecosystem_page`), `default_menu`, "ContractConditions(`MainCondition`)")
    	DBInsert(Str(id) + "_menu", "name,value,title,conditions", `default_menu`, 
              SysParamString(`default_ecosystem_menu`), "default", "ContractConditions(`MainCondition`)")
    	DBInsert(Str(id) + "_keys", "id,pub", $wallet, DBString("1_keys", "pub", $wallet))
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
        ValidateCondition($Conditions, $state)
        if DBIntExt(Table(`parameters`), `id`, $Name, `name`) {
            warning Sprintf( `Parameter %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(Table(`parameters`), `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('9','contract EditParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    conditions {
        EvalCondition(Table(`parameters`), $Name, `conditions`)
        ValidateCondition($Conditions, $state)
        var exist int
       	if $Name == `ecosystem_name` {
    		exist = FindEcosystem($Value)
    		if exist > 0 && exist != $state {
    			warning Sprintf(`Ecosystem %%s already exists`, $Value)
    		}
    	}
    }
    action {
        DBUpdateExt(Table(`parameters`), `name`, $Name, `value,conditions`, $Value, $Conditions )
       	if $Name == `ecosystem_name` {
            DBUpdate(`system_states`, $state, `name`, $Value)
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
        ValidateCondition($Conditions,$state)
        if DBIntExt(Table(`menu`), `id`, $Name, `name`) {
            warning Sprintf( `Menu %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(Table(`menu`), `name,value,title,conditions`, $Name, $Value, $Title, $Conditions )
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
        DBUpdate(Table(`menu`), $Id, `value,title,conditions`, $Value, $Title, $Conditions)
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
        var table string
        table = Table(`menu`)
        DBUpdate(table, $Id, `value`, DBString(table, `value`, $Id) + "\r\n" + $Value )
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
        ValidateCondition($Conditions,$state)
        if DBIntExt(Table(`pages`), `id`, $Name, `name`) {
            warning Sprintf( `Page %%s already exists`, $Name)
        }
    }
    action {
        DBInsert(Table(`pages`), `name,value,menu,conditions`, $Name, $Value, $Menu, $Conditions )
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
        DBUpdate(Table(`pages`), $Id, `value,menu,conditions`, $Value, $Menu, $Conditions)
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
        var value, table string
        table = Table(`pages`)
        value = DBString(table, `value`, $Id)
       	if Contains(value, `PageEnd:`) {
		   value = Replace(value, "PageEnd:", $Value) + "\r\nPageEnd:"
    	} else {
    		value = value + "\r\n" + $Value
    	}
        DBUpdate(table, $Id, `value`,  value )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('16','contract NewLang {
    data {
        Name  string
        Trans string
    }
    conditions {
        EvalCondition(Table(`parameters`), `changing_language`, `value`)
        var exist string
        exist = DBStringExt(Table(`languages`), `name`, $Name, `name`)
        if exist {
            error Sprintf("The language resource %%s already exists", $Name)
        }
    }
    action {
        DBInsert(Table(`languages`), `name,res`, $Name, $Trans )
        UpdateLang($Name, $Trans)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('17','contract EditLang {
    data {
        Name  string
        Trans string
    }
    conditions {
        EvalCondition(Table(`parameters`), `changing_language`, `value`)
    }
    action {
        DBUpdateExt(Table(`languages`), `name`, $Name, `res`, $Trans )
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
        ValidateCondition($Conditions,$state)
        var exist string
        exist = DBStringExt(Table(`signatures`), `name`, $Name, `name`)
        if exist {
            error Sprintf("The signature %%s already exists", $Name)
        }
    }
    action {
        DBInsert(Table(`signatures`), `name,value,conditions`, $Name, $Value, $Conditions )
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
        DBUpdate(Table(`signatures`), $Id, `value,conditions`, $Value, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('20','contract NewBlock {
    data {
    	Name       string
    	Value      string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$state)
        if DBIntExt(Table(`blocks`), `id`, $Name, `name`) {
            warning Sprintf( `Block %%s aready exists`, $Name)
        }
    }
    action {
        DBInsert(Table(`blocks`), `name,value,conditions`, $Name, $Value, $Conditions )
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
        DBUpdate(Table(`blocks`), $Id, `value,conditions`, $Value, $Conditions)
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
}', '%[1]d','ContractConditions(`MainCondition`)');