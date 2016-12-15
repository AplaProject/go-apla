SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_append_id = TxId(AppendPage),
	type_new_column_id = TxId(NewColumn),
	sc_conditions = "$citizen == #wallet_id#",
    sc_value1 = `contract SendMoney {
                 tx {
                         RecipientAccountId int
                         Amount money
                 }

				 func front {
				 	if DBAmount(Table("accounts"),"citizen_id", $citizen ) < $Amount {
					 	error "not enough money"
					}
				 }
               	func main {
				    var sender_id int
                 	sender_id  = DBIntExt( Table("accounts"), "id", $citizen, "citizen_id")
			        DBTransfer(Table("accounts"), "amount,id", sender_id, $RecipientAccountId, $Amount)
                 }
}`,
    sc_value2 = `contract AddAccount {
                 	tx {
                 	    CitizenId string
                     }
					func front {
						if AddressToId($CitizenId)==0 {
							error "not valid citizen id"
						}
					}
                 	func main {
                        DBInsert(Table( "accounts"), "citizen_id", AddressToId($CitizenId))
                 	}
                 }`,
    sc_value3 = `contract UpdAmount {
                 	tx {
                         AccountId int
                         Amount money
                     }

                 	func main {
                         DBUpdate(Table("accounts"), $AccountId, "amount", $Amount)
                 	}
                 }`,
    `page_dashboard_default #= Divs(md-6)
                               Divs()
                               WiBalance( GetOne(amount, #state_id#_accounts, "citizen_id", #citizen#), StateValue(currency_name) )
                               DivsEnd:
                               Divs()
                               WiAccount( GetOne(id, #state_id#_accounts, "citizen_id", #citizen#) )
                               DivsEnd:
                               DivsEnd:`,

    `page_government #=
      Divs(md-12, panel panel-default panel-body)
      BtnTemplate(AddAccount, AddAccount, '', 'btn btn-primary btn-lg')
      BtnTemplate(SendMoney, SendMoney, '', 'btn btn-primary btn-lg')
      BtnTemplate(UpdAmount, UpdAmount, '', 'btn btn-primary btn-lg') BR()
      DivsEnd:
`,

     page_send_money = `Title : Best country
                        Navigation( LiTemplate(government),non-link text)
                        PageTitle : Dashboard
                        TxForm { Contract: SendMoney }
						PageEnd:`,
     page_add_account = `Title : Best country
                         Navigation( LiTemplate(government),non-link text)
                         PageTitle : Dashboard
                         TxForm { Contract: AddAccount }
						 PageEnd:`,
     page_upd_amount = `Title : Best country
                        Navigation( LiTemplate(government),non-link text)
                        PageTitle : Dashboard
                        TxForm { Contract: UpdAmount }
						PageEnd:`
)
TextHidden( sc_value1, sc_value2, sc_value3, sc_conditions )
TextHidden( page_dashboard_default, page_government, page_send_money, page_add_account, page_upd_amount )

Json(`Head: "Money",
	Desc: "Simple monetary system",
	Img: "/static/img/apps/money.jpg",
	OnSuccess: {
		script: 'template',
		page: 'government',
		parameters: {}
	},
	TX: [
		{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			name: "SendMoney",
			value: $("#sc_value1").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
		{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			name: "AddAccount",
			value: $("#sc_value2").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
		{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			name: "UpdAmount",
			value: $("#sc_value3").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
		{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #type_new_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "citizen_id",
			index: "0",
			column_type: "int64",			
			permissions: "$citizen == #wallet_id#",
			index: 1
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
			type: "AppendPage",
			typeid: #type_append_id#,
			name : "government",
			value: $("#page_government").val(),
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
			value: $("#page_send_money").val(),
			global: #global#,
			conditions: "$citizen == #wallet_id#",
		}
	},
	{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "AddAccount",
			menu: "menu_default",
			value:  $("#page_add_account").val(),
			global: #global#,
			conditions: "$citizen == #wallet_id#",
		}
	},
	{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "UpdAmount",
			menu: "menu_default",
			value: $("#page_upd_amount").val(),
			global: #global#,
			conditions: "$citizen == #wallet_id#",
		}
	}
	]
`)
