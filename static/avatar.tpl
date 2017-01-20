GetRow(sc, #state_id#_smart_contracts, name, TXEditProfile )
SetVar(
	global = 0,
	typeid = TxId(EditContract),
	typecolid = TxId(NewColumn),
	sc_value = `contract TXEditProfile {
	data {
		FirstName  string
		Image string "image"
	}
	func action {
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
  	  //Println("TXEditProfile new")
	}
}`
)
TextHidden( sc_value, sc_conditions )
Json(`Head: "Avatar",
	Desc: "Adding an image",
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
			permissions: "$citizen == #wallet_id#",
			index: 0
		}
		}
	]
`)
