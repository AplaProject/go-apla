SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_append_id = TxId(AppendPage),
	type_append_menu_id = TxId(AppendMenu),
    type_new_table_id = TxId(NewTable),
	sc_conditions = "$citizen == #wallet_id#")
SetVar(sc_AddAccount = `contract AddAccount {
	data {
		CitizenId string "address"
	}
	func conditions {
		if AddressToId($CitizenId) == 0 {
			error "not valid citizen id"
		}
	}
	func action {
		DBInsert(Table("accounts"), "citizen_id", AddressToId($CitizenId))
	}
}`,
sc_DisableAccount = `contract DisableAccount {
	data {
		AccountId int
	}

	func action {
		DBUpdate(Table("accounts"), $AccountId, "disabled", "1")
	}
}`,
sc_SendMoney = `contract SendMoney {
	data {
		/*RecipientAccountId int "@1_accounts.id"
		Amount money*/
		RecipientAccountId int 
		Amount money
	}

	func conditions {
	    //Println("RecipientAccountId", $RecipientAccountId)
	    //Println("citizen", $citizen)
	    //Println("Amount", $Amount)
		if DBAmount(Table("accounts"), "citizen_id", $citizen) < $Amount {
			error "not enough money"
		}
	}
	func action {
		var sender_id int
		sender_id = DBIntExt(Table("accounts"), "id", $citizen, "citizen_id")
		DBTransfer(Table("accounts"), "amount,id", sender_id, $RecipientAccountId, $Amount)
	}
}`,
sc_UpdAmount = `contract UpdAmount {
	data {
		AccountId int "@1_accounts.id"
		Amount money
	}
	
	func conditions	{
	}

	func action {
	    //Println("AccountId", $AccountId)
		DBUpdate(Table("accounts"), $AccountId, "amount", $Amount)
	}
}`)
TextHidden( sc_AddAccount, sc_DisableAccount, sc_SendMoney, sc_UpdAmount)
SetVar(`p_CentralBank #= Title : Central bank
Navigation( LiTemplate(government, Government),Central bank)



MarkDown: ## Citizens accounts 

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add account")
        
        Divs(form-group)
            Label("Citizen ID")
            InputAddress(CitizenId, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: AddAccount, Name: Add,Inputs: "CitizenId=CitizenId",OnSuccess: "template,CentralBank,global:0" }
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Disable account")
        
        Divs(form-group)
            Label("Account ID")
            Select(DAccountId, #state_id#_accounts.id, "form-control input-lg m-b")
        DivsEnd:
        TxButton{ Contract: DisableAccount, Name: Disable, Inputs: "AccountId=DAccountId",OnSuccess: "template,CentralBank,global:0" }
       
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Change amount")
        
        Divs(form-group)
            Label("Account ID")
            Select(AccountId, #state_id#_accounts.id, "form-control input-lg m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Amount")
            InputMoney(CitizenAmount, "form-control input-lg")
        DivsEnd:
        
        TxButton{ Contract: UpdAmount, Name: Change, Inputs: "AccountId=AccountId, Amount=CitizenAmount",OnSuccess: "template,CentralBank,global:0" }
    FormEnd:
DivsEnd:


Divs(md-12, panel panel-default panel-body)
Legend(" ", "Accounts")
Table{
    Table: #state_id#_accounts
	Order: id
	Where: company_id = 0 or company_id is NULL
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Citizen ID, Address(#citizen_id#)],[History, If(#rb_id#>0, SysLink(rowHistory, Show, "rbId:#rb_id#,tableName:'#state_id#_accounts'"), "No history")],[Disabled, If(#disabled#==1, "Div(label label-danger, Yes)", "")]]
}
DivsEnd: 

Div(clearfix md)



Divs(md-12, panel panel-default panel-body)
Legend(" ", "Companies accounts")
Table{
    Table: #state_id#_accounts
	Order: id
	Where: company_id > 0
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Citizen ID, Address(#citizen_id#)],[Company ID, #company_id#],[History, If(#rb_id#>0, SysLink(rowHistory, Show, "rbId:#rb_id#,tableName:'#state_id#_company_accounts'"), "No history")],[Disabled, If(#disabled#==1, "Div(label label-danger, Yes)", "")]]
}

DivsEnd:

PageEnd:`,
    `page_dashboard_default #= Divs(md-6)
                               Divs()
                               WiBalance( GetOne(amount, #state_id#_accounts, "citizen_id", #citizen#), StateValue(currency_name) )
                               DivsEnd:
                               Divs()
                               WiAccount( GetOne(id, #state_id#_accounts, "citizen_id", #citizen#) )
                               DivsEnd:
                               DivsEnd:
                               Divs(md-12, panel panel-default panel-body text-center)
                                   BtnTemplate(SendMoney, SendMoney, '', 'btn btn-primary btn-lg')
                               DivsEnd:
                               `,
`p_SendMoney #= Title : Send money
                        Navigation( LiTemplate(dashboard_default, Citizen), Send money)
                        PageTitle : Dashboard
                        TxForm { Contract: SendMoney }
						PageEnd:`,
`menu_1 #= [CentralBank](CentralBank)`)
TextHidden( page_dashboard_default, page_government, p_CentralBank, p_SendMoney, menu_1)
Json(`Head: "Money",
	Desc: "Elements of managing the financial system",
	Img: "/static/img/apps/money.jpg",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "accounts",
			columns: '[["amount", "money", "0"],["disabled", "int64", "1"],["citizen_id", "int64", "1"],["company_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "AddAccount",
			value: $("#sc_AddAccount").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "DisableAccount",
			value: $("#sc_DisableAccount").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SendMoney",
			value: $("#sc_SendMoney").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "UpdAmount",
			value: $("#sc_UpdAmount").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CentralBank",
			menu: "menu_default",
			value: $("#p_CentralBank").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
	   		{
       			Forsign: 'global,name,value',
       			Data: {
       				type: "AppendPage",
       				typeid: #type_append_id#,
       				name : "dashboard_default",
       				value: $("#page_dashboard_default").val(),
       				global: #global#
       			}
       },
	   		{
       			Forsign: 'global,name,value',
       			Data: {
       				type: "AppendMenu",
       				typeid: #type_append_menu_id#,
       				name : "government",
       				value: $("#menu_1").val(),
       				global: #global#
       			}
       },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SendMoney",
			menu: "menu_default",
			value: $("#p_SendMoney").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)