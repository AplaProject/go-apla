SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_append_id = TxId(AppendPage),
	type_append_menu_id = TxId(AppendMenu),
    type_new_table_id = TxId(NewTable),
	sc_conditions = "$citizen == #wallet_id#")
SetVar(sc_AddCompanyAccount = `contract AddCompanyAccount {
	data {
		CompanyId int
	}
	func conditions {

	}
	func action {
		DBInsert(Table("accounts"), "citizen_id, company_id", $citizen, $CompanyId)
	}
}`,
sc_BuyItem = `contract BuyItem {
	data {
		ItemId int
		Price money
	}

	func conditions {
	}
	func action {

		var company_id int
		company_id = DBIntExt(Table("items"), "company_id", $ItemId, "id")

		var recipient_id int
		recipient_id = DBIntExt(Table("accounts"), "id", company_id, "company_id")
		
		var sender_id int
		sender_id = DBIntWhere( Table("accounts"), "id", "citizen_id=$ and (disabled is NULL or disabled=0)", $citizen)
		
		//Println("sender_id", sender_id)
		//Println("recipient_id", recipient_id)
		//Println("Price", $Price)
		DBTransfer(Table("accounts"), "amount,id", sender_id, recipient_id, $Price)


	}
}`,
sc_NewCompany = `contract NewCompany {
	data {
		Name string
	}
	func conditions {
	}
	func action {
		DBInsert(Table("companies"), "name, owner_citizen_id,timestamp opened_time", $Name, $citizen, $block_time)
	}
}`,
sc_NewItem = `contract NewItem {
	data {
		ItemName string
		CompanyId int
		ItemPrice money
	}
	func conditions {
	}
	func action {
		DBInsert(Table("items"), "name, company_id, timestamp added_time, price", $ItemName, $CompanyId, $block_time, $ItemPrice)
	}
}`)
TextHidden( sc_AddCompanyAccount, sc_BuyItem, sc_NewCompany, sc_NewItem)
SetVar(`p_BuyItem #= Title: Buy good
Navigation(LiTemplate(dashboard_default, Citizen))

GetRow(item, #state_id#_items, "id", #ItemId#)


Divs(col-lg-4 data-sweet-alert)
    Divs(list-group)
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-6)
                    P(h4 mb0, Name)
                DivsEnd:
                Divs(col-xs-6)
                    P(h4 text-bold mb0, #item_name#)
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-6)
                    P(h4 mb0, Price)
                DivsEnd:
                Divs(col-xs-6)
                    P(h4 text-bold mb0, Money(#item_price#))
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-12)
                    Input(ItemId, "hidden", text, text, #ItemId#)
                    Input(Price, "hidden", text, text, #item_price#)
                    TxButton{ Contract: BuyItem, Name: Accept, Inputs:"ItemId:ItemId,Price:Price" }
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        
    DivsEnd:
DivsEnd:


PageEnd:
`,
`p_CompanyDetails #= Title : Company details
Navigation( LiTemplate(dashboard_default, Citizen), Business)

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add item")
        
        Divs(form-group)
            Label("Item Name")
            Input(ItemName, "form-control input-lg m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Price")
            InputMoney(ItemPrice, "form-control input-lg m-b")
        DivsEnd:
        
        Input(CompanyId, "hidden", text, text, #CompanyId#)
        TxButton{ Contract: NewItem, Name: Add,Inputs: "ItemName=ItemName, ItemPrice=ItemPrice, CompanyId=CompanyId", OnSuccess: "template,CompanyDetails,CompanyId:#CompanyId#" }
    FormEnd:
DivsEnd:


Divs(md-6)
    Divs()
        WiBalance( Money(GetOne(amount, #state_id#_accounts, "company_id", #CompanyId#)), StateValue(currency_name) )
    DivsEnd:
    Divs()
        WiAccount( GetOne(id, #state_id#_accounts, "company_id", #CompanyId#) )
    DivsEnd:
    
    Divs()
        TxButton{ Contract: AddCompanyAccount, Name: Open account, Inputs: "CompanyId=CompanyId", OnSuccess: "template,CompanyDetails,CompanyId:#CompanyId#" }
     DivsEnd:

DivsEnd:


Divs(md-12, panel panel-default panel-body)
Legend(" ", "My items")
Table {
    Class: table-striped table-hover
    Table: #state_id#_items
	Order: id DESC
	Where: company_id=#CompanyId#
	Columns: [[ID, #id#],[Name, #name#],[Registration date, Date(#added_time#, DD.MM.YYYY)], [Price, Money(#price#)] ]
}

DivsEnd: `,
`p_business #= Title : Business
Navigation( LiTemplate(dashboard_default, Citizen), Business)

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add company")
        
        Divs(form-group)
            Label("Company Name")
            Input(Name, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: NewCompany, Name: Add,Inputs: "Name=Name",OnSuccess: "template,business" }
    FormEnd:
DivsEnd:


Divs(md-12, panel panel-default panel-body)
Legend(" ", "My companies")
Table {
    Class: table-striped table-hover
    Table: #state_id#_companies
    Where: owner_citizen_id=#citizen#
	Order: id
	Columns: [[ID, #id#],[Name, #name#],[Registration date, Date(#opened_time#, DD.MM.YYYY)], [Details, BtnTemplate(CompanyDetails,Details,"CompanyId:#id#")] ]
}

DivsEnd: `,
`p_shops #= Title : Shops
Navigation( LiTemplate(dashboard_default, Citizen), Shops)

Divs(md-12, panel panel-default panel-body)
Legend(" ", "Goods")
Table {
    Class: table-striped table-hover
    Table: #state_id#_items
	Order: id DESC
	Columns: [[Name, #name#],[Price, Money(#price#)], [Buy, BtnTemplate(BuyItem, Buy,"ItemId:#id#")] ]
}

DivsEnd:
`, `menu_1 #= MenuItem(Business, load_template, business)
MenuItem(Shops, load_template, shops)`)
TextHidden( p_BuyItem, p_CompanyDetails, p_business, p_shops, menu_1)
Json(`Head: "Business",
Desc: "Company register, buying and sales tools",
		Img: "/static/img/apps/business.png",
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
			table_name : "companies",
			columns: '[["name", "hash", "1"],["opened_time", "time", "1"],["owner_citizen_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "items",
			columns: '[["name", "text", "0"],["price", "money", "1"],["added_time", "time", "1"],["company_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "AddCompanyAccount",
			value: $("#sc_AddCompanyAccount").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BuyItem",
			value: $("#sc_BuyItem").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "NewCompany",
			value: $("#sc_NewCompany").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "NewItem",
			value: $("#sc_NewItem").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "BuyItem",
			menu: "menu_default",
			value: $("#p_BuyItem").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CompanyDetails",
			menu: "menu_default",
			value: $("#p_CompanyDetails").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "business",
			menu: "menu_default",
			value: $("#p_business").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },

	   	   		{
              			Forsign: 'global,name,value',
              			Data: {
              				type: "AppendMenu",
              				typeid: #type_append_menu_id#,
              				name : "menu_default",
              				value: $("#menu_1").val(),
              				global: #global#
              			}
              },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "shops",
			menu: "menu_default",
			value: $("#p_shops").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)