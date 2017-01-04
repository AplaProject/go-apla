SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_append_id = TxId(AppendPage),
	type_append_menu_id = TxId(AppendMenu),
    type_new_table_id = TxId(NewTable),
	sc_conditions = "$citizen == #wallet_id#")
SetVar(sc_AddProperty = `contract AddProperty {
	data {
		Coords string "polymap"
		CitizenId string "address"
		PropertyType int "@property_types"
	}
	func conditions {
		if AddressToId($CitizenId) == 0 {
			error "invalid address"
		}
		var id int
		id = DBIntExt(Table("parlament"), "id", $citizen, "citizen_id")
		if id == 0 {
			error "access denied"
		}

	}
	func action {
		DBInsert(Table("property"), "coords,citizen_id,type", $Coords, AddressToId($CitizenId), $PropertyType)
	}
}`,
sc_EditProperty = `contract EditProperty {
	data {
		PropertyId int "hidden"
		Coords string "polymap"
		CitizenId string "address"
		PropertyType int "@property_types"
	}
	func conditions {
		if AddressToId($CitizenId) == 0 {
			error "invalid address"
		}
	}
	func action {
		DBUpdate(Table("property"), $PropertyId, "coords,citizen_id,type", $Coords, AddressToId($CitizenId),  $PropertyType)
	}
}`,
sc_PropertyAcceptOffers = `contract PropertyAcceptOffers {
	data {
		OfferId int
	}

	func conditions {
	    
		var property_id int
		property_id = DBIntExt(Table("property_offers"), "property_id", $OfferId, "id")
		
		var citizen_id int
		citizen_id = DBIntExt(Table("property"), "citizen_id", property_id, "id")
		if citizen_id!=$citizen {
		    error "incorrect citizen"
		}
	}
	func action {

		var property_id int
		property_id = DBIntExt(Table("property_offers"), "property_id", $OfferId, "id")
		
		
		var sender_citizen_id int
		sender_citizen_id = DBIntExt(Table("property_offers"), "sender_citizen_id", $OfferId, "id")

		var price int
		price = DBIntExt(Table("property_offers"), "price", $OfferId, "id")

		var sender_id int
		sender_id = DBIntExt(Table("accounts"), "id", sender_citizen_id, "citizen_id")
		var recipient_id int
		recipient_id = DBIntExt(Table("accounts"), "id", $citizen, "citizen_id")
		DBTransfer(Table("accounts"), "amount,id", sender_id, recipient_id, Money(price))

		DBUpdate(Table("property"), property_id, "citizen_id", sender_citizen_id)

	}
}`,
sc_PropertySendOffer = `contract PropertySendOffer {
	data {
		PropertyId int "hidden"
		OfferType int
		Price money
	}
	func action {
		DBInsert(Table("property_offers"), "property_id, type, price, sender_citizen_id", $PropertyId, $OfferType, $Price, $citizen)
		
		var offers int
		offers = DBInt(Table("property"), "offers", $PropertyId)
		DBUpdate(Table("property"), $PropertyId, "offers", offers+1)
	}
}`,
sc_SellProperty = `contract SellProperty {
	data {
		Id int "hidden"
		Price money
		RecipientCitizenId int "address"
	}
	func actions {
		DBUpdate(Table("citizenship_requests"), $Id, "approved", -1)
	}
}`,
sc_SetPropertyPrice = `contract SetPropertyPrice {
 data {
    ProperyId int
    Price money
}

func conditions {

}

func action {
    var count int
    count  = DBIntExt( Table("votes"), "count(id)", $VotingId, "voting_id")

    if count >= 1 {
        var name string
        name = DBStringExt( Table("voting"), "name", $VotingId, "id")
        var type string
        type = DBStringExt( Table("voting"), "type", $VotingId, "id")
        var text string
        text = DBStringExt( Table("voting"), "text", $VotingId, "id")
        if type == "contract" {
            UpdateContract(name, text, "")
        }
        if type == "param" {
            UpdateParam(name, text, "")
        }
        if type == "contract-conditions" {
            UpdateParam(name, "", text)
        }
        if type == "param-conditions" {
            UpdateParam(name, "", text)
        }
    } else {
       DBInsert(Table( "votes"), "voting_id, citizen_id, result", $VotingId, $citizen, $Result)
       
    }
  }
}
`,
sc_SetPropertyRentPrice = `contract SetPropertyRentPrice {
	data {
		Price money
	}
	func action {
		DBUpdate(Table("property"), 1, "rent_price", $Price)
	}
}`,
sc_SetPropertySellPrice = `contract SetPropertySellPrice {
	data {
		Price money
	}
	func action {
		DBUpdate(Table("property"), 1, "sell_price", $Price)
	}
}`)
TextHidden( sc_AddProperty, sc_EditProperty, sc_PropertyAcceptOffers, sc_PropertySendOffer, sc_SellProperty, sc_SetPropertyPrice, sc_SetPropertyRentPrice, sc_SetPropertySellPrice)
SetVar(`p_AddProperty #= Navigation( LiTemplate(government),Add property )
            PageTitle : Add Property
            TxForm{ Contract: AddProperty}
            PageEnd:`,
`p_EditProperty #= Title:EditProperty
Navigation(LiTemplate(government),Editing property)
PageTitle: Editing property

ValueById(1_property, #PropertyId#, "name,citizen_id,coords,type", "Name,CitizenId,Coords,PropertyType")
SetVar( CitizenId= Address(#CitizenId#))
TxForm{ Contract: EditProperty}

PageEnd:`,
`p_PropertyAcceptOffers #= Title: Best country
Navigation(LiTemplate(dashboard_default, citizen))

GetRow(offer, #state_id#_property_offers, "id", #OfferId#)


Divs(col-lg-4 data-sweet-alert)
    Divs(list-group)
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-6)
                    P(h4 mb0, Type)
                DivsEnd:
                Divs(col-xs-6)
                    P(h4 text-bold mb0, StateValue(property_prices_types, #offer_type#))
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-6)
                    P(h4 mb0, Price)
                DivsEnd:
                Divs(col-xs-6)
                    P(h4 text-bold mb0, Money(#offer_price#))
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        Divs(list-group-item)
            Divs(row row-table pv-lg)
                Divs(col-xs-12)
                    Input(OfferId, "hidden", text, text, #OfferId#)
                    TxButton{ Contract: PropertyAcceptOffers, Name: Accept, Inputs:"OfferId:OfferId" }
                DivsEnd:
            DivsEnd:
        DivsEnd:
        
        
    DivsEnd:
DivsEnd:


PageEnd:
`,
`p_PropertyDetails #= FullScreen(1)
Title: Best country
Navigation(LiTemplate(dashboard_default, citizen))
SetVar(hmap=350)

GetRow(myproperty, #state_id#_property, "id", #PropertyId#)

Divs(md-8, panel panel-default panel-body)
    Map(#myproperty_coords#)
DivsEnd:


Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Sell Price
        DivsEnd:
    DivsEnd:
    
        Input(PropertyId, "hidden", text, text, Param(PropertyId))
        
    Divs(form-group)
            InputMoney(SellPrice, "form-control input-lg ", #myproperty_sell_price#)
            
    DivsEnd:
    
    TxButton{ Contract: SetPropertySellPrice, Name: "Save", Inputs: "Price=SellPrice"}
DivsEnd:


Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Rent Price
        DivsEnd:
    DivsEnd:
    
        Input(PropertyId, "hidden", text, text, Param(PropertyId))
        
    Divs(form-group)
            InputMoney(RentPrice, "form-control input-lg ", #myproperty_rent_price#)
    DivsEnd:
    
    TxButton{ Contract: SetPropertyRentPrice, Name: "Save", Inputs: "Price=RentPrice"}
DivsEnd:


Divs(md-12, panel panel-default panel-body)
MarkDown : ## Offers
Table{
    Table: #state_id#_property_offers
    Where: property_id='#PropertyId#'
    Columns: [[ID, #id#], [price, Money(#price#)], [sender_citizen_id, #sender_citizen_id#], [Type, StateLink(property_prices_types, #type#) ], [Accept, BtnTemplate(PropertyAcceptOffers, Accept, "OfferId:#id#")] ]
}


PageEnd:
`,
`p_PropertyOffer #= Title: Property Offer
Navigation(LiTemplate(dashboard_default, Citizen))

Form(form-horizontal)
    Divs(md-12, panel panel-default panel-body)
        Divs(panel-heading)
            Divs(panel-title)
               MarkDown: Property Offer
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
           Label(Type, col-lg-4 control-label)
           Divs(col-lg-6)
                Select(OfferType, property_prices_types, input-lg m-b, #Ptype#)
           DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            Label(Price, col-lg-4 control-label)
            Divs(col-lg-6)
                InputMoney(Price, "form-control input-lg m-b", 0)
            DivsEnd:
        DivsEnd:
        
        Input(PropertyId, "hidden", 0, "text", #PropertyId#)
        
        TxButton{Contract: PropertySendOffer, Name: Send offer, "Price:Price,PropertyId:PropertyId,OfferType:OfferType"}

    DivsEnd:
FormEnd:

PageEnd:
`,
`p_PropertyOffers #= Title: Property Offers
Navigation(LiTemplate(dashboard_default, citizen))

Divs(md-12, panel panel-default panel-body)
MarkDown : ## Offers
Table{
    Table: #state_id#_property_offers
    Where: property_id='#PropertyId#'
    Columns: [[ID, #id#], [price, #price#], [sender_citizen_id, #sender_citizen_id#], [type ID, #type#] ]
}
DivsEnd:

PageEnd:
`,
`p_PropertyResults #= Title : Property results
Navigation( Citizens )

Divs(md-12, panel panel-default panel-body)
Table{
    Table: #state_id#_property
    Where: "If( #Ptype# == 1, sell_price > #PriceMin# and sell_price < #PriceMax#, rent_price > #PriceMin# and rent_price < #PriceMax#)""
    Order: id
    Columns: [[ID, #id#], [Name, property], [Coordinates, Map(#coords#)], [Citizen ID, Address(#citizen_id#)], [Send offer,BtnTemplate(PropertyOffer,Send offer,"PropertyId:#id#,Ptype:#Ptype#")], [Rent price,Money(#rent_price#)] , [Sell price,Money(#sell_price#)] ]
}
DivsEnd:

PageEnd:`,
`p_SearchProperty #= Title: Search Property
Navigation(LiTemplate(dashboard_default, Citizen))

Form(form-horizontal)
    Divs(md-12, panel panel-default panel-body)
        Divs(panel-heading)
            Divs(panel-title)
               MarkDown: Search property
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
           Label(Type, col-lg-4 control-label)
           Divs(col-lg-6)
                Select(Ptype, property_prices_types, input-lg m-b)
           DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            Label(Min price, col-lg-4 control-label)
            Divs(col-lg-6)
                InputMoney(PriceMin, "form-control input-lg m-b", 0)
            DivsEnd:
        DivsEnd:
        Divs(form-group)
            Label(Max price, col-lg-4 control-label)
            Divs(col-lg-6)
                InputMoney(PriceMax, "form-control input-lg m-b", 0)
            DivsEnd:
        DivsEnd:
        
        BtnTemplate(PropertyResults,Search,"Ptype:Val(Ptype),PriceMin:Val(PriceMin),PriceMax:Val(PriceMax)")

    DivsEnd:
FormEnd:

PageEnd:
`,

    `page_dashboard_default #= Divs(md-12, panel panel-default panel-body)
                               MarkDown : ## My property
                               Table{
                                   Table: #state_id#_property
                                   Where: citizen_id='#citizen#'
                                   Order: id
                                   Columns: [[ID, #id#], [Name, property], [Coordinates, Map(#coords#)], [Citizen ID, Address(#citizen_id#)], [Details,BtnTemplate(PropertyDetails,Details,"PropertyId:#id#")], [Rent price,Money(#rent_price#)] , [Sell price,Money(#sell_price#)] , [Offers, #offers#] ]
                               }
                               DivsEnd:
                               Divs(md-12, panel panel-default panel-body text-center)
                                   BtnTemplate(SearchProperty, Search property, '', 'btn btn-primary btn-lg')
                               DivsEnd:`,

    `page_government #=
    Divs(md-12, panel panel-default panel-body)
                MarkDown : ## Property
                Table{
                    Table: 1_property
                    Order: id
                    Columns: [[ID, #id#], [Type, StateLink(property_types, #type#)]  [Coordinates, Map(#coords#)], [Citizen ID, Address(#citizen_id#)], [Edit,BtnTemplate(EditProperty,Edit,"PropertyId:#id#")]]
                }
             BtnTemplate(AddProperty, AddProperty, '', 'btn btn-primary btn-lg') BR()
    DivsEnd:
`,
`p_SellProperty #= Title : SellProperty
Navigation( Citizens )
PageTitle : Sell property
TxForm{ Contract: SellProperty}
PageEnd:`)
TextHidden( page_dashboard_default, page_government, p_AddProperty, p_EditProperty, p_PropertyAcceptOffers, p_PropertyDetails, p_PropertyOffer, p_PropertyOffers, p_PropertyResults, p_SearchProperty, p_SellProperty)
Json(`Head: "Property",
Desc: "Property",
		Img: "/static/img/apps/property.jpg",
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
			table_name : "property",
			columns: '[["sell_price", "money", "1"],["name", "text", "0"],["type", "int64", "1"],["coords", "text", "0"],["offers", "int64", "0"],["citizen_id", "int64", "1"],["rent_price", "money", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "property_offers",
			columns: '[["type", "int64", "1"],["price", "money", "1"],["property_id", "int64", "1"],["sender_citizen_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "AddProperty",
			value: $("#sc_AddProperty").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "EditProperty",
			value: $("#sc_EditProperty").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "PropertyAcceptOffers",
			value: $("#sc_PropertyAcceptOffers").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "PropertySendOffer",
			value: $("#sc_PropertySendOffer").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SellProperty",
			value: $("#sc_SellProperty").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SetPropertyPrice",
			value: $("#sc_SetPropertyPrice").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SetPropertyRentPrice",
			value: $("#sc_SetPropertyRentPrice").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SetPropertySellPrice",
			value: $("#sc_SetPropertySellPrice").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "AddProperty",
			menu: "menu_default",
			value: $("#p_AddProperty").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "EditProperty",
			menu: "menu_default",
			value: $("#p_EditProperty").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "PropertyAcceptOffers",
			menu: "menu_default",
			value: $("#p_PropertyAcceptOffers").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "PropertyDetails",
			menu: "menu_default",
			value: $("#p_PropertyDetails").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "PropertyOffer",
			menu: "menu_default",
			value: $("#p_PropertyOffer").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "PropertyOffers",
			menu: "menu_default",
			value: $("#p_PropertyOffers").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "PropertyResults",
			menu: "menu_default",
			value: $("#p_PropertyResults").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SearchProperty",
			menu: "menu_default",
			value: $("#p_SearchProperty").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
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
			name : "SellProperty",
			menu: "menu_default",
			value: $("#p_SellProperty").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)