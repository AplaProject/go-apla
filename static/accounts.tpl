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
            func main {
                   var cur_amount money
                   var sender_id int
                   var cur_amount_sender money

                   cur_amount = Money(DBString(Table("accounts"), "amount", $RecipientAccountId ))
                   DBUpdate(Table( "accounts"), $RecipientAccountId, "amount", cur_amount + $Amount)

                   sender_id  = DBIntExt( Table("accounts"), "id", $citizen, "citizen_id")
                   cur_amount_sender  = Money(DBString(Table("accounts"), "amount", sender_id))
                   DBUpdate(Table( "accounts"), sender_id, "amount", cur_amount_sender - $Amount)
            }
}`,
    sc_value2 = `contract AddAccount {
                 	tx {
                 	    Citizen string
  }
                 	func main {
     DBInsert(Table( "accounts"), "citizen_id", $Citizen)
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
}`
    page_dashboard_default = `
             Table{
                 Table: #state_id#_accounts
  Where: citizen_id='#citizen#'
  Columns: [[amount, #amount#]]
             }

             Table{
                 Table: #state_id#_accounts
                 Order: id
                 Columns: [[ID, #id#], [Amount, #amount#], [Send money,BtnTemplate(SendMoney,Send,"RecipientAccountId:#id#")]]
             }

             PageEnd:`

    page_government = `TemplateNav(AddAccount, AddAccount) BR()
     TemplateNav(SendMoney, SendMoney) BR()
     TemplateNav(UpdAmount, UpdAmount) BR()

     MarkDown : ## Citizens
     Table{
         Table: #state_id#_citizens
         Order: id
         Columns: [[Avatar,Image(#avatar#)], [ID, Address(#id#)], [Name, #name#]]
     }
     PageEnd:`
     page_send_money = `Title : Best country
                        Navigation( LiTemplate(government),non-link text)
                        PageTitle : Dashboard
                        TxForm { Contract: SendMoney }
                        PageEnd:`
     page_add_account = `Title : Best country
                         Navigation( LiTemplate(government),non-link text)
                         PageTitle : Dashboard
                         TxForm { Contract: AddAccount }
                         PageEnd:`
     page_upd_amount = `Title : Best country
                        Navigation( LiTemplate(government),non-link text)
                        PageTitle : Dashboard
                        TxForm { Contract: UpdAmount }
                        PageEnd:`

)
TextHidden( sc_value1, sc_value2, sc_value3, sc_conditions )
Json(`Head: "Adding account column",
	Desc: "This application adds citizen_id column into account table.",
	OnSuccess: {
		script: 'template',
		page: 'government',
		parameters: {}
	},
	TX: [
		{
		Forsign: 'global,id,value,conditions',
		Data: {
			type: "AddContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			value: $("#sc_value1").val(),
			conditions: $("#sc_conditions1").val()
			}
	   },
		{
		Forsign: 'global,id,value,conditions',
		Data: {
			type: "AddContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			value: $("#sc_value2").val(),
			conditions: $("#sc_conditions2").val()
			}
	   },
		{
		Forsign: 'global,id,value,conditions',
		Data: {
			type: "AddContract",
			typeid: #type_new_contract_id#,
			global: #global#,
			value: $("#sc_value3").val(),
			conditions: $("#sc_conditions3").val()
			}
	   },
        	   {
        		Forsign: 'table_name,column_name,permissions,index',
        		Data: {
        			type: "NewColumn",
        			typeid: #type_new_column_id#,
        			table_name : "#state_id#_accounts",
        			column_name: "citizen_id",
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
               			value: "#page_dashboard_default#",
               			global: #global#
               		}
               },
           {
           		Forsign: 'global,name,value',
           		Data: {
           			type: "AppendPage",
           			typeid: #type_append_id#,
           			name : "goventment",
           			value: "#page_dashboard_goventment#",
           			global: #global#
           		}
           },
                   {
                   		Forsign: 'global,name,value,conditions',
                   		Data: {
                   			type: "NewPage",
                   			typeid: #type_new_page_id#,
                   			name : "SendMoney",
                   			value: "#page_send_money#",
                   			global: #global#,
                    		conditions: "$citizen == #wallet_id#",
                   		}
                   },
                           {
                           		Forsign: 'global,name,value,conditions',
                           		Data: {
                           			type: "NewPage",
                           			typeid: #type_new_page_id#,
                           			name : "AddAccount",
                           			value: "#page_add_account#",
                           			global: #global#,
                            		conditions: "$citizen == #wallet_id#",
                           		}
                           },
                                   {
                                   		Forsign: 'global,name,value,conditions',
                                   		Data: {
                                   			type: "NewPage",
                                   			typeid: #type_new_page_id#,
                                   			name : "UpdAmount",
                                   			value: "#page_upd_amount#",
                                   			global: #global#,
                                    		conditions: "$citizen == #wallet_id#",
                                   		}
                                   }
	]
`)
