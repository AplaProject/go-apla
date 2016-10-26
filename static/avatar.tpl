GetRow(sc, #state_id#_smart_contracts, name, TXEditProfile )
SetVar(
	global = 0,
	typeid = TxId(EditContract),
	typecolid = TxId(NewColumn),
	sc_value = `contract TXEditProfile {
	tx {
		FirstName  string
		Image string "image"
	}
	func main {
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
  	  Println("TXEditProfile new")
	}
}`
)
TextHidden( sc_value, sc_conditions )
Json(`Head: "Adding avatar column",
	Desc: "This application adds avatar column into citizens table.	Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
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
			column_name: "avatar",
			permissions: "$citizen == #wallet_id#",
			index: 0
		}
		}
	]
`)
