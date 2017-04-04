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
	type_new_sign_id = TxId(NewSign),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable))
SetVar()
TextHidden( )
SetVar(`p_GameDescription #= Title : Description
Navigation(LiTemplate(dashboard_default, Citizen), Game)

Tag(h2, State settings, page-header)
UList(class, ul)
    Li("The state is created through the entry of a unique name and currency code.")
    Li("Basic Apps shall be installed: Election and Assign, Polling, Messenger, Simple Money System, etc. (this may take some time).")
    Li("A flag and a short description of the state (the Upload Flag and Add Description buttons) must be entered on the &quot;Government&quot; page. In the table of the state parameters its territory may be specified (Admon Tools> Tables> State parameters> state_coords). The flag, description and territory can be changed in the government parameters table.")
UListEnd:

Tag(h2, Administration, page-header)
Tag(h4, Citizenship)
UList(class, ul)
    Li("Notifications of new requests for citizenship are displayed on the &quot;Government&quot; page. The applications can be accepted or rejected on the &quot;Administration&quot; page on the 'Citizenship requests' panel.")
    Li("The list of citizens is displayed on the &quot;Citizens&quot; page. In the table it is possible to deprive of citizenship.")
UListEnd:

Tag(h4, Polling of citizens)
UList(class, ul)
    Li("On the &quot;New Polling&quot; page it is possible to ask citizens questions for")
    UList(class, ul)
        Li("voting (Yes / No answers) - type 'Voting'")
        Li("polling (with detailed answers) - type ‘Question’.")
    UListEnd:
    Li("The voting process is displayed on the &quot;Pollings list&quot; page:")
    UList(class, ul)
        Li("voting terms")
        Li("number of voters")
        Li("the interim results may be viewed")
        Li("the voting may be stopped and")
        Li("deleted")
    UListEnd:
    Li("Voting results will be available to citizens after the end of voting or in case of forced stop.")
UListEnd:

Tag(h4, Assigned positions)
UList(class, ul)
    Li("A list of assigned and selected positions may be set on the &quot;Administration&quot; page (the &quot;Add New Position&quot; panel).")
    Li("In the &quot;Assign a position to citizen&quot; any citizen may be appointed to a position.")
    Li("The list of officials is displayed in the 'Appointed posts' panel on the 'Government' page. The assignment may be canceled on this panel.")
UListEnd:

Tag(h4, General voting)
UList(class, ul)
    Li("The elected positions are displayed in the 'New Election' panel, where an elective campaign may be run.")
    Li("The dates of the campaign events shall be set on the &quot;New Election Campaign&quot; page: Start of campaign, Deadline for candidates, Start and end of voting. (The dates do not change, so you should be less dismissive of the timing of the election campaign).")
    Li("The data in the campaign are reflected on the &quot;Election Campaigns&quot; page, where citizens may register as candidates, vote and view the voting results. (The number of persons elected to elective office is indicated in the contract SmartLaw_NumResultsVoting.)")
    Li("The election winners are displayed in the &quot;Elective Posts&quot; panel on the &quot;Government&quot; page.")
UListEnd:

Tag(h4, Finance)
UList(class, ul)
    Li("An account for a new citizen may be opened on the &quot;Administration&quot; page in the ‘Accounts’")
    UList(class, ul)
        Li("(if it was not automatically opened during registration)")
        Li("block a citizen's account")
        Li("replenish the citizen's account with a certain amount of money.")
    UListEnd:
    Li("The numbers of the citizens' accounts are reflected on their pages and in the table on the &quot;Citizens&quot; page.")
    Li("Citizens can transfer money to each other.")
UListEnd:

Tag(h2, United Government, page-header)
UList(class, ul)
    Li("In order to demonstrate its position the Government may answer a number of questions reflected in the 'Polling' panel on the &quot;Government&quot; page.")
    Li("The government may vote for other governments in the 'States List' panel. The Governments that have won more than 5 votes shall be admitted to the United Government.")
    Li("Members of United Government can participate in the discussion in a special messenger on the &quot;Government&quot; page.")
UListEnd:

Tag(h2, Citizens, page-header)
P(ml-xl, "Citizens may:")
UList(class, ul)
    Li("Edit their profile (Welcome menu > click on the avatar, or click on the avatar of the citizen page).")
    Li("Transfer money to each other")
    Li("Discuss problems in State Messenger")
    Li("Propose themselves as the candidates for elective positions")
    Li("Participate in elections")
    Li("Answer polling questions")
UListEnd:
PageEnd:`,
`pc_GameDescription #= ContractConditions("MainCondition")`)
TextHidden( p_GameDescription, pc_GameDescription)
SetVar()
TextHidden( )
SetVar()
TextHidden( )
SetVar()
TextHidden( )
Json(`Head: "description",
Desc: "",
		Img: "/static/img/apps/ava.png",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GameDescription",
			menu: "Goverment",
			value: $("#p_GameDescription").val(),
			global: 1,
			conditions: $("#pc_GameDescription").val(),
			}
	   }]`
)