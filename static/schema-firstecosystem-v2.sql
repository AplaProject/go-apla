INSERT INTO "system_states" ("rb_id") VALUES ('0');

INSERT INTO "1_contracts" ("value", "wallet_id", "conditions") VALUES 
('contract MoneyTransfer {
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
('contract NewContract {
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
			   error Sprintf(`wrong wallet %s`, $Wallet)
		    }
	    }
	    var list array
	    list = ContractsList($Value)
	    var i int
	    while i < Len(list) {
	        if IsContract(list[i], $state) {
	            warning Sprintf(`Contract %s exists`, list[i] )
	        }
	        i = i + 1
	    }
        if !$TokenEcosystem {
            $TokenEcosystem = 1
        } else {
            if !SysFuel($TokenEcosystem) {
                warning Sprintf(`Ecosystem %d is not system`, $TokenEcosystem )
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
}', '%[1]d', 'ContractConditions(`MainCondition`)'),
('contract EditContract {
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
('contract ActivateContract {
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
('contract NewEcosystem {
    data {
        Name  string
    }
    conditions {
    }
    action {
        CreateEcosystem($Name)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract NewParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    conditions {
        ValidateCondition($Conditions, $state)
    }
    action {
        DBInsert(Table(`parameters`), `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract EditParameter {
    data {
        Name string
        Value string
        Conditions string
    }
    conditions {
        EvalCondition(Table(`parameters`), $Name, `conditions`)
        ValidateCondition($Conditions, $state)
    }
    action {
        DBUpdateExt(Table(`parameters`), `name`, $Name, `value,conditions`, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
( 'contract NewMenu {
    data {
    	Name       string
    	Value      string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$state)
    }
    action {
        DBInsert(Table(`menu`), `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract EditMenu {
    data {
    	Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        Eval(DBString(Table(`menu`), $Id, `conditions`))
        ValidateCondition($Conditions,$state)
    }
    action {
        DBUpdate(Table(`menu`), $Id, `value,conditions`, $Value, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract AppendMenu {
    data {
        Id     int
    	Value      string
    }
    conditions {
        Eval(DBString(Table(`menu`), $Id, `conditions`))
    }
    action {
        var table string
        table = Table(`menu`)
        DBUpdate(table, $Id, `value`, DBString(table, $Id, `value`) + "\r\n" + $Value )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract NewPage {
    data {
    	Name       string
    	Value      string
    	Menu       string
    	Conditions string
    }
    conditions {
        ValidateCondition($Conditions,$state)
       	if HasPrefix($Name, `sys-`) || HasPrefix($Name, `app-`) {
	    	error `The name cannot start with sys- or app-`
	    }
    }
    action {
        DBInsert(Table(`pages`), `name,value,menu,conditions`, $Name, $Value, $Menu, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract EditPage {
    data {
        Id         int
    	Value      string
    	Menu      string
    	Conditions string
    }
    conditions {
        Eval(DBString(Table(`pages`), $Id, `conditions`))
        ValidateCondition($Conditions,$state)
    }
    action {
        DBUpdate(Table(`pages`), $Id, `value,menu,conditions`, $Value, $Menu, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract AppendPage {
    data {
        Id         int
    	Value      string
    }
    conditions {
        Eval(DBString(Table(`pages`), $Id, `conditions`))
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
('contract NewLang {
    data {
        Name  string
        Trans string
    }
    conditions {
        EvalCondition(Table(`parameters`), `changing_language`, `value`)
        var exist string
        exist = DBStringExt(Table(`languages`), `name`, $Name, `name`)
        if exist {
            error Sprintf("The language resource %s already exists", $Name)
        }
    }
    action {
        DBInsert(Table(`languages`), `name,res`, $Name, $Trans )
        UpdateLang($Name, $Trans)
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract EditLang {
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
('contract NewSign {
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
            error Sprintf("The signature %s already exists", $Name)
        }
    }
    action {
        DBInsert(Table(`signatures`), `name,value,conditions`, $Name, $Value, $Conditions )
    }
}', '%[1]d','ContractConditions(`MainCondition`)'),
('contract EditSign {
    data {
    	Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        Eval(DBString(Table(`signatures`), $Id, `conditions`))
        ValidateCondition($Conditions,$state)
    }
    action {
        DBUpdate(Table(`signatures`), $Id, `value,conditions`, $Value, $Conditions)
    }
}', '%[1]d','ContractConditions(`MainCondition`)');