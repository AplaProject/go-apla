SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "$citizen == #wallet_id#")
SetVar(sc_AddAccountEuro = `contract AddAccountEuro {
	data {
		CitizenId string "address"
	}
	func conditions {
		if AddressToId($CitizenId) == 0 {
			error "not valid citizen id"
		}
	}
	func action {
		DBInsert("global_euro", "citizen_id,state", AddressToId($CitizenId), $state)
	}
}`,
sc_DisableEuroAccount = `contract DisableEuroAccount {
	data {
		AccountId  int "@global_euro.id"
	}

	func action {
		DBUpdate("global_euro", $AccountId, "disabled", "1")
	}
}`,
sc_UpdAmountEuro = `contract UpdAmountEuro {
	data {
		AccountId  int "@global_euro.id"
		Amount money
	}

	func action {
		DBUpdate("global_euro", $AccountId, "amount", $Amount)
	}
}`)
TextHidden( sc_AddAccountEuro, sc_DisableEuroAccount, sc_UpdAmountEuro)
SetVar(`p_Euro #= Title : Euro
Navigation( LiTemplate(government, Government),Euro)

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add EURO account")
        
        Divs(form-group)
            Label("Citizen ID")
            InputAddress(CitizenId, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: @AddAccountEuro, Name: Add,Inputs: "CitizenId=CitizenId",OnSuccess: "template,Euro,global:1" }
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Disable EURO account")
        
        Divs(form-group)
            Label("Account ID")
            Select(DAccountId, global_euro.id, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: @DisableEuroAccount, Name: Disable, Inputs: "AccountId=DAccountId",OnSuccess: "template,Euro,global:1" }
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Change EURO amount")
        
        Divs(form-group)
            Label("Account ID")
            Select(AccountId, global_euro.id, "form-control input-lg m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Amount")
            InputMoney(Amount, "form-control input-lg")
        DivsEnd:
        
        TxButton{ Contract: @UpdAmountEuro, Name: Change, Inputs: "AccountId=AccountId, Amount:Amount",OnSuccess: "template,Euro,global:1" }
    FormEnd:
DivsEnd:




Divs(md-12, panel panel-default panel-body)
Legend(" ", "EURO accounts")
Table{
    Table: global_euro
	Order: id
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Citizen ID, Address(#citizen_id#)],[History, If(#rb_id#>0, SysLink(rowHistory, Show, "rbId:#rb_id#,tableName:'global_euro', global:1"), "No history")],[Disabled, If(#disabled#==1, "Div(label label-danger, Yes)", "")]]
}
DivsEnd: `)
TextHidden( p_Euro)
Json(`Head: "Euro",
Desc: "Euro",
		Img: "/static/img/apps/ava.png",
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
			global: 1,
			table_name : "euro",
			columns: '[["state", "int64", "1"],["amount", "int64", "1"],["disabled", "int64", "1"],["citizen_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "AddAccountEuro",
			value: $("#sc_AddAccountEuro").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "DisableEuroAccount",
			value: $("#sc_DisableEuroAccount").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "UpdAmountEuro",
			value: $("#sc_UpdAmountEuro").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "Euro",
			menu: "menu_default",
			value: $("#p_Euro").val(),
			global: 1,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)