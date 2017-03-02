SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_append_page_id = TxId(AppendPage),
	type_new_menu_id = TxId(NewMenu),
	type_edit_table_id = TxId(EditTable),
	type_edit_column_id = TxId(EditColumn),
	type_append_menu_id = TxId(AppendMenu),
	type_new_lang_id = TxId(NewLang),
	type_new_contract_id = TxId(NewContract),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "ContractConditions(\"MainCondition\")")
SetVar(`sc_EditProfile = contract EditProfile {
                        	data {
                        		FirstName  string
                        		Image string "image"
                        	}
                        	func action {
                        	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
                          	  //Println("TXEditProfile new")
                        	}
                        }
`,`sc_GenCitizen = contract GenCitizen {
          	data {
          		Name      string
           		PublicKey string
          	}
          	conditions {
          	    if StateValue("gov_account") != $citizen {
          	        error "Access denied"
          	    }
          	    $idc = PubToID($PublicKey)
          	    if $idc == 0 || DBIntExt("dlt_wallets", "wallet_id", $idc, "wallet_id") == $idc {
          	        warning "Pubkey is used"
          	    }
          	}
          	action {
          		DBInsert("dlt_wallets", "wallet_id,public_key_0,address_vote", $idc, HexToBytes($PublicKey), IdToAddress($idc))
          		DBInsert(Table( "citizens"), "id,block_id,name", $idc, $block, $Name )
          	}
          }`,
`sc_TXCitizenRequest = contract TXCitizenRequest {
	data {
		StateId    int    "hidden"
		FullName   string	
	}
	conditions {
		if Balance($wallet) < Money(StateParam($StateId, "citizenship_price")) {
			error "not enough money"
		}
	}
	action {
		DBInsert(TableTx( "citizenship_requests"), "dlt_wallet_id,name,block_id", 
		    $wallet, $FullName, $block)
	}
}`,
`sc_TXEditProfile = contract TXEditProfile {
	data {
		FirstName  string
		Image string "image"
	}
	action {
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
  	  //Println("TXEditProfile new")
	}
}`,
`sc_TXNewCitizen = contract TXNewCitizen {
	data {
        RequestId int
    }
 	conditions {
		if Balance(DBInt(Table( "citizenship_requests"), "dlt_wallet_id", $RequestId )) < Money(StateParam($state, "citizenship_price")) {
			error "not enough money"
		}
	}
	action {
		var wallet int
		var towallet int
		wallet = DBInt(Table( "citizenship_requests"), "dlt_wallet_id", $RequestId )
		towallet = Int(StateValue("gov_account"))
		if towallet == 0 {
			towallet = $citizen
		}
//        DBTransfer("dlt_wallets", "amount,wallet_id", wallet, towallet, Money(StateParam($state, "citizenship_price")))
		DBInsert(Table( "citizens"), "id,block_id,name", wallet, 
		          $block, DBString(Table( "citizenship_requests"), "name", $RequestId ) )
        DBUpdate(Table( "citizenship_requests"), $RequestId, "approved", 1)
	}	
}`,
`sc_TXRejectCitizen = contract TXRejectCitizen {
   data { 
        RequestId int
   }
   action { 
	  DBUpdate(Table( "citizenship_requests"), $RequestId, "approved", -1)
   }
}`)
TextHidden( sc_GenCitizen, sc_EditProfile, sc_TXCitizenRequest, sc_TXEditProfile, sc_TXNewCitizen, sc_TXRejectCitizen)
SetVar(`p_CheckCitizens #= Title : Check citizens requests
Navigation( LiTemplate(government), Citizens)
PageTitle : Citizens requests
Table{
    Table: 1_citizenship_requests
	Order: id
	Where: approved=0
	Columns: [[ID, #id#],[Name, #name#],[Accept,BtnPage(NewCitizen,Accept,"RequestId:#id#")],[Reject,BtnPage(RejectCitizen,Reject,"RequestId:#id#")]]
}
PageEnd:
`,
`p_NewCitizen #= Title : New Citizen
Navigation( Citizens )
PageTitle : New Citizen 
TxForm{ Contract: TXNewCitizen}
PageEnd:
`,
`p_RejectCitizen #= Title : Reject Citizen
Navigation( Citizens )
PageTitle : Reject Citizen 
TxForm{ Contract: TXRejectCitizen}
PageEnd:
`,
`p_citizen_profile #= Title:Profile
Navigation(LiTemplate(Citizen),Editing profile)
PageTitle: Editing profile
ValueById(#state_id#_citizens, #citizen#, "name,avatar", "FirstName,Image")
TxForm{ Contract: TXEditProfile, OnSuccess: MenuReload()}
PageEnd:`,
`p_citizens #= Title : Citizens
Navigation( LiTemplate(government), Citizens)
PageTitle : Citizens
Table{
    Table: 1_citizens
    Columns: [[Avatar,Image(#avatar#)], [ID, #id#], [Name, #name#]]
}
PageEnd:
`)
TextHidden( p_CheckCitizens, p_NewCitizen, p_RejectCitizen, p_citizen_profile, p_citizens)
SetVar()
TextHidden( )
SetVar()
TextHidden( )
SetVar()
TextHidden( )
SetVar(`ap_government #= BtnPage(CheckCitizens, Check citizens, '', btn btn-primary btn-lg) BR() BR()`)
TextHidden( ap_government)
SetVar(`am_government #= MenuItem(Checking citizens, CheckCitizens)`)
TextHidden( am_government)
Json(`Head: "Basic",
Desc: "Basic environment ",
		Img: "/static/img/apps/ava.png",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [
		{
             		Forsign: 'global,id,value,conditions',
             		Data: {
             			typeid: #typeid#,
             			type: "EditContract",
             			global: #global#,
             			id: #sc_id#,
             			value: $("#sc_value").val(),
             			conditions: $("#sc_conditions").val()
             			}
        },
         {
        		Forsign: 'table_name,column_name,permissions,index,column_type',
        		Data: {
        			type: "NewColumn",
        			typeid: #typecolid#,
        			table_name : "#state_id#_citizens",
        			column_name: "avatar",
        			index: "0",
        			column_type: "text",
        			permissions: "ContractConditions(\"MainCondition\")",
        			index: 0
        		}
        },
        {
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "citizenship_requests",
			columns: '[["dlt_wallet_id", "int64", "1"],["public_key_0", "text", "0"],["name", "hash", "0"],["approved", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "public_key_0",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_citizenship_requests",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GenCitizen",
			value: $("#sc_GenCitizen").val(),
			conditions: $("#sc_conditions").val()
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXCitizenRequest",
			value: $("#sc_TXCitizenRequest").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXEditProfile",
			value: $("#sc_TXEditProfile").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXNewCitizen",
			value: $("#sc_TXNewCitizen").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXRejectCitizen",
			value: $("#sc_TXRejectCitizen").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CheckCitizens",
			menu: "government",
			value: $("#p_CheckCitizens").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "NewCitizen",
			menu: "menu_default",
			value: $("#p_NewCitizen").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RejectCitizen",
			menu: "menu_default",
			value: $("#p_RejectCitizen").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "citizen_profile",
			menu: "menu_default",
			value: $("#p_citizen_profile").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "citizens",
			menu: "menu_default",
			value: $("#p_citizens").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendPage",
				typeid: #type_append_page_id#,
				name : "government",
				value: $("#ap_government").val(),
				global: 0
				}
		},
{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendMenu",
				typeid: #type_append_menu_id#,
				name : "government",
				value: $("#am_government").val(),
				global: 0
				}
		}]`
)