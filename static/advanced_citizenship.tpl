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
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar,PlaceOfBirth,DateOfBirth,Gender,DateOfIssue,DateOfExpiry", $NickName, $Image, $PlaceOfBirth, $DateOfBirth, $Gender, $DateOfIssue, $DateOfExpiry)
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
        			column_name: "PlaceOfBirth",
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
                 			column_name: "DateOfBirth",
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
                         			column_name: "Gender",
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
                                 			column_name: "DateOfIssue",
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
                                         			column_name: "DateOfExpiry",
                                         			permissions: "$citizen == #wallet_id#",
                                         			index: 0
                                         		}
                                         		}
	]
`)
