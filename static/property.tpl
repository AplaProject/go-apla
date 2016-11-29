SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_append_id = TxId(AppendPage),
	type_new_table_id = TxId(NewTable),
	sc_conditions = "$citizen == #wallet_id#",
    sc_value1 = `contract AddProperty {
                 	tx {
                         Coords string "map"
                 	CitizenId string
                 	Name string

                     }
                    func front {
                 	    if AddressToId($CitizenId) == 0 {
                            error "invalid address"
                       }

                 	}
                 	func main {
                 		DBInsert(Table( "property"), "coords,citizen_id,name", $Coords, AddressToId($CitizenId), $Name)
                 	}
                 }`,
    sc_value2 = `contract EditProperty {
                 	tx {
                 		PropertyId  int
                 	        Coords string "map"
                 	        CitizenId string
                 	        Name string
                 	}
                 	func front {
                               if AddressToId($CitizenId) == 0 {
                                                error "invalid address"
                                }
                    }
                 	func main {
                 	  DBUpdate(Table( "property"), $PropertyId, "coords,citizen_id,name", $Coords, AddressToId($CitizenId), $Name)
                 	}
                 }`,

    page_add_property = `Navigation( LiTemplate(government),Add property )
            PageTitle : Add Property
            TxForm{ Contract: AddProperty}
            PageEnd:`,

    page_edit_property = `Title:EditProperty
                          Navigation(LiTemplate(government),Editing property)
                          PageTitle: Editing property
                          ValueById(#state_id#_property, #PropertyId#, "name,citizen_id,coords", "Name,CitizenId,Coords")
                          TxForm{ Contract: EditProperty}
                          PageEnd:`,

    `page_dashboard_default #= MarkDown : ## My property
           Table{
               Table: #state_id#_property
               Where: citizen_id='#!citizen#'
               Order: id
               Columns: [[ID, #!id#], [Name, #!name#], [Coordinates, #!coords#], [Citizen ID, #!citizen_id#]]
           }`,

    `page_government #=
            MarkDown : ## Property
            Table{
                Table: #state_id#_property
                Order: id
                Columns: [[ID, #!id#], [Name, #!name#], [Coordinates, #!coords#], [Citizen ID, #!citizen_id#], [Edit,BtnTemplate(EditProperty,Edit,"PropertyId:#!id#")]]
            }
             BtnTemplate(AddProperty, AddProperty, '', 'btn btn-primary btn-lg')
            `

)
TextHidden( sc_value1, sc_value2, sc_conditions )
TextHidden( page_add_property, page_edit_property, page_dashboard_default, page_government )
Json(`Head: "Property",
	Desc: "Property",
	Img: "/static/img/apps/property.jpg",
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
			name: "AddProperty",
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
			name: "EditProperty",
			value: $("#sc_value2").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
		{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: #global#,
			table_name : "property",
			columns: '["citizen_id","coords","name"]',
			permissions: "$citizen == #wallet_id#"
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
			Forsign: 'global,name,value,menu,conditions',
			Data: {
				type: "NewPage",
				typeid: #type_new_page_id#,
				name : "EditProperty",
				value: $("#page_edit_property").val(),
				menu: "menu_default",
				global: #global#,
				conditions: "$citizen == #wallet_id#",
			}
		},
		{
			Forsign: 'global,name,value,menu,conditions',
			Data: {
				type: "NewPage",
				typeid: #type_new_page_id#,
				name : "AddProperty",
				value: $("#page_add_property").val(),
				menu: "menu_default",									   
				global: #global#,
				conditions: "$citizen == #wallet_id#",
			}
		}
	]
`)
