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
            error Sprintf(`Contract %d does not exist`, $Id)
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
        root = CompileContract($Value, $state, $cur[`wallet_id`], $cur[`token_id`])
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
            error Sprintf(`Contract %d does not exist`, $Id)
        }
        if Int($cur[`active`]) == 1 {
            error Sprintf(`The contract %d has been already activated`, $Id)
        }
        Eval($cur[`conditions`])
        if $wallet != $cur[`wallet_id`] {
            error Sprintf(`Wallet %d cannot activate the contract`, $wallet)
        }
    }
    action {
        DBUpdate(Table(`contracts`), $Id, `active`, 1)
        Activate($Id, $state)
    }
}', '%[1]d','ContractConditions(`MainCondition`)');