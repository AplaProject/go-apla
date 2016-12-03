GetRow(sc, #state_id#_smart_contracts, name, TXEditProfile )
SetVar(
	global = 0,
	typeid = TxId(EditContract),
	typecolid = TxId(NewColumn),
	sc_value = `contract TXEditProfile {
	tx {
		NickName  string
		Image string "image"
		PlaceOfBirth  string "map"
		DateOfBirth  string "date"
		Gender  string
		DateOfIssue  string "date"
		DateOfExpiry  string "date"
	}
	func main {
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar,place_of_birth,date_of_birth,gender,date_of_issue,date_of_expiry", $NickName, $Image, $PlaceOfBirth, $DateOfBirth, $Gender, $DateOfIssue, $DateOfExpiry)
	}
}`
)
TextHidden( sc_value, sc_conditions )
Json(`Head: "Advanced citizenship",
	Desc: "Adding a fields",
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
        		Forsign: 'table_name,column_name,permissions,index',
        		Data: {
        			type: "NewColumn",
        			typeid: #typecolid#,
        			table_name : "#state_id#_citizens",
        			column_name: "place_of_birth",
        			permissions: "$citizen == #wallet_id#",
        			index: 0
        		}
        		},
                 	   {
                 		Forsign: 'table_name,column_name,permissions,index',
                 		Data: {
                 			type: "NewColumn",
                 			typeid: #typecolid#,
                 			table_name : "#state_id#_citizens",
                 			column_name: "date_of_birth",
                 			permissions: "$citizen == #wallet_id#",
                 			index: 0
                 		}
                 		},
                         	   {
                         		Forsign: 'table_name,column_name,permissions,index',
                         		Data: {
                         			type: "NewColumn",
                         			typeid: #typecolid#,
                         			table_name : "#state_id#_citizens",
                         			column_name: "gender",
                         			permissions: "$citizen == #wallet_id#",
                         			index: 0
                         		}
                         		},
                                 	   {
                                 		Forsign: 'table_name,column_name,permissions,index',
                                 		Data: {
                                 			type: "NewColumn",
                                 			typeid: #typecolid#,
                                 			table_name : "#state_id#_citizens",
                                 			column_name: "date_of_issue",
                                 			permissions: "$citizen == #wallet_id#",
                                 			index: 0
                                 		}
                                 		},
                                         	   {
                                         		Forsign: 'table_name,column_name,permissions,index',
                                         		Data: {
                                         			type: "NewColumn",
                                         			typeid: #typecolid#,
                                         			table_name : "#state_id#_citizens",
                                         			column_name: "date_of_expiry",
                                         			permissions: "$citizen == #wallet_id#",
                                         			index: 0
                                         		}
                                         		}
	]
`)
