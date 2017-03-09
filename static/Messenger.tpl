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
	type_activate_contract_id = TxId(ActivateContract),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable))
SetVar(`sc_addMessage = contract addMessage {
	data {
        Text string
    }
 	func conditions {
 	    
	}
	func action {
	    
		var avatar string
		avatar = DBStringExt( Table("citizens"), "avatar", $citizen, "id")
		var name string
		name = DBStringExt( Table("citizens"), "name", $citizen, "id")
		var flag string
		flag = DBStringExt( Table("state_parameters"), "value", "state_flag", "name")
		DBInsert( "global_messages", "text, ava, username, flag, citizen_id, state_id", $Text, avatar, name, flag, $citizen, $state) 
	}	
}`)
TextHidden( sc_addMessage)
SetVar(`p_CitizenInfo #= Title: Citizen info
Navigation(LiTemplate(Messenger, Messenger), Citizen info)


GetRow("user", #stateId#_citizens, "id", #citizenId#)

Divs(md-12, panel widget)
    Divs: half-float
        SetVar(hmap=300)
        Map(StateVal(state_coords))
        Divs: half-float-bottom
            Image(#user_avatar#, Image, img-thumbnail img-circle thumb-full)
        DivsEnd:
    DivsEnd:
    Divs: panel-body text-center
        Tag(h3, #user_name#, m0)
        P(text-muted, Head of state)
        P(class, Proin metus justo, commodo in ultrices at, lobortis sit amet dui. Fusce dolor purus, adipiscing a tempus at, gravida vel purus.)
    DivsEnd:
    Divs: panel-body text-center bg-gray-darker
        Divs: row row-table
            Divs: col-xs-12
                LinkPage(StateInfo, Image(StateVal(state_flag), Image, w50 img-responsive d-inline-block align-middle) Strong(d-inline-block align-middle, USA), 'id':1, text-white h3, "stateId:#stateId#")
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:`,
`p_Messenger #= Title : Messenger
Navigation( LiTemplate(dashboard_default, Dashboard), Messenger)


Divs(md-12, panel panel-info data-sweet-alert)
    Div(panel-heading, Div(panel-title, Messenger name))
    Divs(panel-body data-widget=panel-scroll data-start=bottom)
         Divs: list-group
GetList(my, global_messages, "id,username,ava,flag,text,citizen_id,stateid")
ForList(my)
	        Divs: list-group-item list-group-item-hover pointer
                Divs: media-box
                    Divs: pull-left
                        Image(#ava#, ALT, media-box-object img-circle thumb32)
                    DivsEnd:
                    Divs: media-box-body clearfix
                        Divs: flag pull-right
                            Image(#flag#, ALT, class)
                        DivsEnd:
                        LinkPage(CitizenInfo, Strong(media-box-heading text-primary, #username#), "citizenId:'#citizen_id#',stateId:1", class)
                        P(small, #text#)
                    DivsEnd:
                DivsEnd:
            DivsEnd:
ForListEnd:
         DivsEnd:
    DivsEnd:
    Divs(panel-footer)
            Divs(input-group)
                Input(chat_message,form-control input-sm,Write a message...,text)
                Divs( input-group-btn)
                TxButton{ClassBtn: fa fa-paper-plane btn btn-default btn-sm,Contract: @addMessage, Name:" ",Inputs:"Text=chat_message", OnSuccess: "template,Messenger"}
                DivsEnd:

            DivsEnd:
    DivsEnd:
DivsEnd:`,
`p_StateInfo #= Title: State info
Navigation(LiTemplate(Messenger, Messenger), State info)


Divs(md-4, panel panel-default elastic center)
    Divs: panel-body
        Image(GetOne("value", #id#_state_parameters, "name", "state_flag"), ALT, img-responsive)
    DivsEnd:
DivsEnd:
Divs(md-8, panel widget elastic)
    Divs: panel-body text-center
        Tag(h3, GetOne("value", #id#_state_parameters, "name", "state_name"), m0)
    DivsEnd:
    Divs: panel-body text-center bg-gray-dark
        Divs: row row-table
            Divs: col-xs-4
                Tag(h3, 01.01.2017, m0)
                P(m0 text-muted, Founded)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3, GetOne("value", #id#_state_parameters, "name", "currency_name"), m0)
                P(m0 text-muted, Currency)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3, 500, m0)
                P(m0 text-muted, Population)
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:


Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Recognized as the number of UN members)
    Divs: panel-body
        Ring(24, 20, 100, 3, "23b7e5", "656565", 150, 20)
    DivsEnd:
DivsEnd:
Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, I voted in favor of a member of UN)
    Divs: panel-body
        Ring(126,  20, 100, 3, "7266ba", "656565", 150, 20)
    DivsEnd:
DivsEnd:
Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Answered questions on the UN)
    Divs: panel-body
        Ring(111, 20, 100, 3, "27c24c", "656565", 150, 20)
    DivsEnd:
DivsEnd:


Divs(md-12)
    SetVar(hmap=400)
    Map(StateVal(state_coords))
DivsEnd:

`,`menu_1 #= MenuItem(Messenger, load_template, Messenger)`)
TextHidden( p_CitizenInfo, p_Messenger, p_StateInfo, menu_1)
SetVar()
TextHidden( )
SetVar()
TextHidden( )
SetVar()
TextHidden( )
Json(`Head: "Messenger",
Desc: "Messenger",
		Img: "/static/img/apps/messenger.png",
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
			global: 1,
			table_name : "messages",
			columns: '[["stateid", "int64", "1"],["state_id", "int64", "1"],["username", "hash", "1"],["citizen_id", "int64", "1"],["ava", "text", "0"],["flag", "text", "0"],["text", "text", "0"],["time", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "stateid",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "addMessage",
			value: $("#sc_addMessage").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "addMessage"
			}
	   },	   
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CitizenInfo",
			menu: "government",
			value: $("#p_CitizenInfo").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			name : "Messenger",
			menu: "menu_default",
			value: $("#p_Messenger").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "StateInfo",
			menu: "menu_default",
			value: $("#p_StateInfo").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   }]`
)
