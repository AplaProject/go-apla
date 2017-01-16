SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_append_page_id = TxId(AppendPage),
	type_new_menu_id = TxId(NewMenu),
	type_edit_table_id = TxId(EditTable),
	type_edit_column_id = TxId(EditColumn),
	type_append_menu_id = TxId(AppendMenu),
	type_new_contract_id = TxId(NewContract),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "$citizen == #wallet_id#")
SetVar(`sc_RF_NewIssue = contract RF_NewIssue {
 data {
    Issue string 
    Date_start_voting string "date"
    Date_stop_voting string "date"
 }
 
func conditions {
    
	    var x int
	    x = DBIntExt(Table("state_parameters"), "value", "gov_account", "name")
	    
		if x != $citizen {
			info "You do not have the right to do it"
		}


	} 

func action {
    DBInsert(Table( "rf_referendums"), "issue,date_voting_start,date_voting_finish,status,timestamp date_enter",$Issue,$Date_start_voting,$Date_stop_voting,0,$block_time)
    
  } 
}`,
`sc_RF_next_event = contract RF_next_event {
	data {

	}

	func conditions {
	    
	}
	func action {
	    
    var list array
        var vote int
        var flag int
        var war map
        var i int
        var len int
        var status int
        status=0
        flag=0
        
        list = DBGetList(Table("rf_referendums"), "id",0,10,"id desc", "status=$ and date_voting_start < now()",status)
        
        len = Len(list)
        while i < len {
            war = list[i]
            i = i + 1
            vote = DBIntWhere( Table("rf_votes"), "id", "strhash=$", war["id"]+$citizen)
            if(vote==0) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", war["id"])
                 flag=1
            }
        }
        
        if(flag==0) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", 0)
        }
    
  }
}`,
`sc_RF_NextVoting = contract RF_NextVoting {
	data {
        IdIssue int
	}

	func conditions {
	    
	}
	func action {
	    
    var list array
        var vote int
        var flag int
        var war map
        var i int
        var len int
        var status int
        status=0
        flag=0
        
        list = DBGetList(Table("rf_referendums"), "id",0,10,"id desc", "status=$ and date_voting_start < now()",status)
        
        len = Len(list)
        while i < len {
            war = list[i]
            i = i + 1
            vote = DBIntWhere( Table("rf_votes"), "id", "strhash=$", war["id"]+$citizen)
            if(vote == 0 && war["id"] != $IdIssue) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", war["id"])
                 flag=1
            }
        }
        
        if(flag==0) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", 0)
        }
    
  }
}`,
`sc_RF_Voting = contract RF_Voting {
	data {
		ReferendumId int
		RFChoice int 
	}

	func conditions {
	    

	    $sha256=$ReferendumId + $citizen

		var voted int
		voted = DBIntExt(Table("rf_votes"), "id", $sha256, "strhash")

		if voted != 0 {
			info "You already voted"
		}

		var allowed int
		allowed = DBIntWhere(Table("rf_referendums"), "id", "date_voting_start < now() and date_voting_finish > now() and id=$", $ReferendumId)

		if allowed == 0 {
			info "Voting is not available now"
		}

	}
	func action {
	    
	    
	    
	    
    
    DBInsert(Table("rf_votes"),"referendum_id,strhash,choice,timestamp time",$ReferendumId,$sha256, $RFChoice,$block_time)
    
    var counter int
    counter = DBIntExt( Table("rf_referendums"), "number_votes", $ReferendumId, "id")
    DBUpdate(Table("rf_referendums"), $ReferendumId, "number_votes", counter+1)
    
    
    var list array
        var vote int
        var flag int
        var war map
        var i int
        var len int
        var status int
        status=0
        flag=0
        
        list = DBGetList(Table("rf_referendums"), "id",0,10,"id desc", "status=$ and date_voting_start < now()",status)
        
        len = Len(list)
        while i < len {
            war = list[i]
            i = i + 1
            vote = DBIntWhere( Table("rf_votes"), "id", "strhash=$", war["id"]+$citizen)
            if(vote==0) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", war["id"])
                 flag=1
            }
        }
        
        if(flag==0) {
                 DBUpdate(Table("citizens"),  $citizen, "next_event", 0)
        }
    
  }
}`,
`sc_RF_VotingCancel = contract RF_VotingCancel {
	data {
		ReferendumId int
	}

	func conditions {

        var x int
		//x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		  x = DBIntExt(Table("state_parameters"), "value", "gov_account", "name")
		if x != $citizen {
			info "You do not have the right to stop vote"
		}


		x = DBIntWhere(Table("rf_referendums"), "id", "date_voting_start > now() and id=$ and status=2", $ReferendumId)

		if x > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate(Table("rf_referendums"),$ReferendumId,"status",2)

	}
}`,
`sc_RF_VotingResult = contract RF_VotingResult {
	data {
		ReferendumId int
	}

	func conditions {

		var x int

		//x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		x = DBIntExt(Table("state_parameters"), "value", "gov_account", "name")
		if x != $citizen {
			info "You do not have the right to stop vote"
		}

		x = DBIntExt(Table("rf_referendums"),"status",$ReferendumId,"id")
		if x == 1 {
			info "Resalt is ready"
		}

	}

	func action {
		var votes0 int
		var votes1 int
		var resalt int
		votes0 = DBIntWhere(Table("rf_votes"),"count(id)","referendum_id=$ and choice=0",$ReferendumId)
		votes1 = DBIntWhere(Table("rf_votes"),"count(id)","referendum_id=$ and choice=1",$ReferendumId)
		if (votes1 > votes0) {
			resalt = 1
		} else {
			resalt = 0
		}
	DBUpdate(Table("rf_referendums"),$ReferendumId,"result,status,number_0,number_1",resalt,1,votes0,votes1)
		
	DBInsert(Table("rf_result"),"referendum_id,choice,choice_str,value,percents",$ReferendumId,1,"Yes",votes1,100*votes1/(votes1+votes0))
	DBInsert(Table("rf_result"),"referendum_id,choice,choice_str,value,percents",$ReferendumId,0,"No",votes0,100*votes0/(votes1+votes0))

	}
}`)
TextHidden( sc_RF_NewIssue, sc_RF_next_event, sc_RF_NextVoting, sc_RF_Voting, sc_RF_VotingCancel, sc_RF_VotingResult)
SetVar(`p_RF_List #= Title : $ListVotings$

SetVar(ViewResult = BtnTemplate(RF_ViewResult, <b>$View$</b>, "ReferendumId:#id#,Issue:'#issue#',DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#',NumberVotes:#number_votes#,Back:0,Status:#status#",'btn btn-primary btn-block'))
SetVar(GetResult = BtnTemplate(RF_Result, <b>$GetResult$</b>, "ReferendumId:#id#,Issue:'#issue#'",'btn btn-primary btn-block'))
SetVar(Cancel = BtnTemplate(RF_VotingCancel, <b>$Cancel$</b>,"ReferendumId:#id#,Issue:'#issue#'",'btn btn-primary btn-block'))


Divs(md-12, panel panel-default)
    Divs(panel-heading)
        Divs(panel-title)
Table{
         Table: #state_id#_rf_referendums
         Order: #date_voting_start# DESC
         Where: #status#!=2
         Class: table-responsivee
         Adaptive: 1
      Columns: [[$Issue$, #issue#], 
      [$Start$, DateTime(#date_voting_start#, YYYY.MM.DD HH:MI)],
	  [$Finish$, DateTime(#date_voting_finish#, YYYY.MM.DD HH:MI)],
	  [$Inf$,If(#CmpTime(#date_voting_start#, Now(datetime)) == 1,#Cancel#, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1, #number_votes#, If(#status# == 1, #ViewResult#, #GetResult#)))],
	  [$Result$,If(#CmpTime(#date_voting_start#, Now(datetime)) == 1, Wait, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1, $Continues$, If(#status# == 1, If(#result#==1,$Yes$,$No$), $Finished$)))]]
     }
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_NewIssue #= Title : $NewVoting$

ValueById(#state_id#_citizens, #citizen#, "name", "FirstName")
Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
        MarkDown: <h4>$NewVoting$</h4>
          
        Form()
        Divs(form-group)
            Label($EnterIssue$)
            Input(Issue, form-control input-lg)
        DivsEnd:
        Divs(form-group)
            Label($DateStartVoting$)
            InputDate(Date_start_voting,form-control input-lg,Now(YYYY.MM.DD HH:MI))
        DivsEnd:
        Divs(form-group)
            Label($DateFinishVoting$)
            InputDate(Date_stop_voting,form-control input-lg,Now(YYYY.MM.DD HH:MI,1 days))
        DivsEnd:
        
       
        Divs()

            TxButton{Contract: RF_NewIssue,Inputs:"Issue=Issue,Date_start_voting=Date_start_voting,Date_stop_voting=Date_stop_voting", OnSuccess: "template,RF_List"}
        Div(clearfix)    
        DivsEnd:   
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_Result #= Title : $Result$
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center)
         
           MarkDown: <h4>#Issue#</h4>
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
      
        Divs(bt-block, text-center)
        TxButton{Contract: RF_VotingResult,Name: $GetResult$,Inputs:"ReferendumId=ReferendumId", OnSuccess: "template,RF_List"}
        DivsEnd:
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_StartPage #= Title : $ListVotings$


Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title text-center)

          Divs()
           MarkDown: <h1>$Welcome$</h1>
           MarkDown: <h4>$Referendapartij$</h4>
          DivsEnd: 
           MarkDown: BR() BR()
           
           
        Form()
          
           
            
             Divs()
            TxButton{Class: btn col-xs-12 btn-square btn-primary pl2, Contract: RF_next_event,Name: $StartVote$,Inputs:"Start=1", OnSuccess:"template,RF_UserVotingList"}
            DivsEnd:
            
            

          
            
        Div(clearfix)
        DivsEnd:
        FormEnd:

DivsEnd:
DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_User_Voting #= Title : $Voting$ 



Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title text-center)
             Divs(text-primary)
           MarkDown: <h4>#Issue#</h4>
           MarkDown: <br/><br/>
            DivsEnd:
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
        Input(RFChoice0, "hidden", text, text, 0)
        Input(RFChoice1, "hidden", text, text, 1)
        
          
           Divs()
            TxButton{Class: col-xs-6 pl0, Contract: RF_Voting,Name: $Yes$,Inputs:"ReferendumId=ReferendumId,RFChoice=RFChoice1", OnSuccess: "template,RF_UserList"}
            DivsEnd:

           Divs()
            TxButton{Class: col-xs-6 pl0, Contract: RF_Voting,Name: $No$,Inputs:"ReferendumId=ReferendumId,RFChoice=RFChoice0", OnSuccess: "template,RF_UserList"}
            DivsEnd:
    Div(clearfix)        
    Divs(md-12,  text-center) 
     MarkDown: <br/><br/>
      BtnTemplate(RF_UserList,<strong>$ListVotings$</strong>,"Status:0",'btn')

            
        Div(clearfix)
        DivsEnd:
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:
Div(clearfix)
DivsEnd:
PageEnd:`,
`p_RF_UserList #= Title : $FinishedVotings$


SetVar(ViewResult = BtnTemplate(RF_ViewResult, <strong>$Result$</strong>, "ReferendumId:#id#,Issue:'#issue#',DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#', NumberVotes:#number_votes#,Back:1,Status:1",'btn '))

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
    
         BtnTemplate(RF_UserVotingList,$StartVote$,"Status:1",'btn btn-square btn-primary btn-block')

   
    Table{
         Table: #state_id#_rf_referendums
         Order: #date_voting_start#  
         Where: #status#=1
         Class: table-responsive
         
      Columns: [[, <h4>#issue#</h4>],[,
        If(#CmpTime(#date_voting_start#, Now(datetime)) == -1, If(#CmpTime(#date_voting_finish#, Now(datetime)) == -1, If(#status# == 1, #ViewResult#, #Info#), #Voting#), #Info#)]]
     }
    
 DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_UserVotingList #= Title : $Voting$


SetVar(IdIssue = GetOne(next_event, #state_id#_citizens, "id", #citizen#))

SetVar(Voting1 = BtnTemplate(RF_VotingConfirm,$Yes$,"ReferendumId:#id#,Issue:'#issue#',DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#',RFChoice:1,RFChoiceT:'Yes'",'btn col-xs-6 btn-lg'))

SetVar(Voting0 = BtnTemplate(RF_VotingConfirm,$No$,"ReferendumId:#id#,Issue:'#issue#',DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#',RFChoice:0,RFChoiceT:'No'",'btn col-xs-6 btn-lg'))


Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center warning)
           Divs(text-primary)

        If(#IdIssue#==0,
        Divs(md-12,  text-center)
        MarkDown: <h4>$NoAvailablePolls$</h4>
        MarkDown: $YouVoted$<br/>
        MarkDown: <br/>
        DivsEnd:
        ,"")    
   

    Table{
         Table: #state_id#_rf_referendums
         Order: #date_voting_start# DESC
         Where: #id# = #IdIssue#
         Class: text-center text-primary
         
      Columns: [
                [,<h3>#issue#</h3><br/> #Voting1# #Voting0# ]
            ]
     }
 
      
     
    Divs(md-12,  text-center) 
     MarkDown: <br/>
      
    
        Form()
        Input(IdIssue, "hidden", text, text, #IdIssue#)
          
        TxButton{Class: text-center btn-lg, Contract:RF_NextVoting, Name: Next,Inputs:"IdIssue=IdIssue", OnSuccess: "template,RF_UserVotingList"}
            
        Div(clearfix)
        DivsEnd:
        FormEnd:


    DivsEnd:
    DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_ViewInfo #= Title : $Inf$

Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center)
            Divs(text-primary)
           MarkDown: <h4>#Issue#</h4>
           DivsEnd:
           MarkDown: <strong>$Start$:</strong> DateTime(#DateStart#, YYYY.MM.DD HH:MI)
           If(#CmpTime(#DateFinish#, Now(datetime)) == -1,
           MarkDown: $VotingFinished$
           MarkDown: <h4>$ResultSoon$</h4>)
           MarkDown: <br/>
            BtnTemplate(RF_UserList, <strong>$ListVotings$</strong>,"Status:0",'btn')
    
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_ViewResult #= Title : $Result$

Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center warning)
           Divs(text-primary)
           MarkDown: <h4>#Issue#</41>
           DivsEnd:
           MarkDown: <strong>DateTime(#DateStart#, YYYY.MM.DD HH:MI) - DateTime(#DateFinish#, YYYY.MM.DD HH:MI)</dtrong>
           
           MarkDown: <h4>$TotalVoted$: #NumberVotes#</h4>
            
            
            
    Table{
    Table: #state_id#_rf_result
    Where: referendum_id=#ReferendumId#
    Order: #choice_str# DESC
      Columns: [
    [,If(#choice#==1,$Yes$,$No$)],
    [,#value#],
    [,#percents# %]]
     }
    
    Divs(btn-lg)
    If(#Back#==1,BtnTemplate(RF_UserList, <strong>$ListVotings$</strong> ,"Status:1",'btn'), BtnTemplate(RF_List, <strong>$ListVotings$</strong>, "Status:#Status#",'btn'))
     DivsEnd:
    
        DivsEnd:
    DivsEnd:
DivsEnd:




Divs(md-6, panel panel-default panel-body)
ChartPie{Table: #state_id#_rf_result, FieldValue: percents, FieldLabel: choice_str, Colors: "5d9cec,fad732,37bc9b,f05050,23b7e5,ff902b,f05050,131e26,37bc9b,f532e5,7266ba,3a3f51,fad732,232735,3a3f51,dde6e9,e4eaec,edf1f2", Where: referendum_id = #ReferendumId#, Order: choice}
DivsEnd:


PageEnd:`,
`p_RF_VotingCancel #= Title : $Cancel$
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center)
           
           MarkDown: <h4>#Issue#</h4>
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
      
        
        TxButton{Contract: RF_VotingCancel,Name: $Cancel$,Inputs:"ReferendumId=ReferendumId", OnSuccess: "template,RF_List"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_VotingConfirm #= Title : $Voting$

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title text-center)
        Divs(text-primary)    
           MarkDown: <h4>#Issue#</h4>
        DivsEnd:
           MarkDown: <strong>$YourAnswer$</strong>
           MarkDown: <h1>If(#RFChoice#==1,$Yes$,$No$)</h1>
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
        Input(RFChoice, "hidden", text, text, #RFChoice#)
          
        Divs(md-12, btn-block)  
        TxButton{Class: btn col-xs-12 btn-primary pl2, Contract: RF_Voting,Name: $Confirm$,Inputs:"ReferendumId=ReferendumId,RFChoice=RFChoice", OnSuccess: "template,RF_UserVotingList"}
        DivsEnd:
        Divs(md-12, btn-block)
        MarkDown:BR() BR()
        BtnTemplate(RF_UserVotingList,$Cancel$,"Status:0",'btn')
        DivsEnd:
            
        Div(clearfix)
        DivsEnd:
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:
Div(clearfix)
DivsEnd:
PageEnd:`,
`p_RF_Сancel #= Title : Voting Result
Navigation( Voting Result)
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center)
           
           MarkDown: <h4>Issue</h4>
           MarkDown: <h1>#Issue#</h1>
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
      
        
        TxButton{Contract: RF_VotingResult,Name: Get Result,Inputs:"ReferendumId=ReferendumId", OnSuccess: "template,RF_List"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`)
TextHidden( p_RF_List, p_RF_NewIssue, p_RF_Result, p_RF_StartPage, p_RF_User_Voting, p_RF_UserList, p_RF_UserVotingList, p_RF_ViewInfo, p_RF_ViewResult, p_RF_VotingCancel, p_RF_VotingConfirm, p_RF_Сancel)
SetVar(`m_Referendum = [Einde Stemmen](RF_UserList)
[Beheerder](RF_List)
`,
`m_RF_UserVoting = [Stemlijst](RF_StartPage)
[Nieuwe vraag](RF_NewIssue)`)
TextHidden( m_Referendum, m_RF_UserVoting)
SetVar()
TextHidden( )
SetVar(`am_government #= [Stemlijst](RF_UserList)`,
`am_menu_default #= [Stemlijst](RF_StartPage)`)
TextHidden( am_government, am_menu_default)
Json(`Head: "",
Desc: "",
		Img: "/static/img/apps/ava.png",
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
			table_name : "rf_referendums",
			columns: '[["result", "int64", "1"],["number_1", "int64", "1"],["date_enter", "time", "1"],["number_votes", "int64", "1"],["issue", "text", "0"],["status", "int64", "1"],["number_0", "int64", "1"],["date_voting_start", "time", "1"],["date_voting_finish", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_referendums",
			general_update: "$citizen==-6000967947006608966",
			insert: "",
			new_column: "$citizen==-6000967947006608966",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "rf_result",
			columns: '[["percents", "int64", "1"],["choice_str", "text", "0"],["referendum_id", "int64", "1"],["value", "int64", "1"],["choice", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_result",
			general_update: "$citizen==-6000967947006608966",
			insert: "",
			new_column: "$citizen==-6000967947006608966",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "percents",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "choice_str",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "referendum_id",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "value",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "choice",
			permissions: "",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "rf_votes",
			columns: '[["strhash", "hash", "1"],["referendum_id", "int64", "1"],["hash", "text", "0"],["time", "time", "1"],["choice", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_votes",
			general_update: "$citizen==-6000967947006608966",
			insert: "",
			new_column: "$citizen==-6000967947006608966",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "choice",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "strhash",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "referendum_id",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "hash",
			permissions: "",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "time",
			permissions: "",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_NewIssue",
			value: $("#sc_RF_NewIssue").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_next_event",
			value: $("#sc_RF_next_event").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_NextVoting",
			value: $("#sc_RF_NextVoting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_Voting",
			value: $("#sc_RF_Voting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_VotingCancel",
			value: $("#sc_RF_VotingCancel").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_VotingResult",
			value: $("#sc_RF_VotingResult").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "Referendum",
			value: $("#m_Referendum").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "RF_UserVoting",
			value: $("#m_RF_UserVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_List",
			menu: "Referendum",
			value: $("#p_RF_List").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_NewIssue",
			menu: "Referendum",
			value: $("#p_RF_NewIssue").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_Result",
			menu: "Referendum",
			value: $("#p_RF_Result").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_StartPage",
			menu: "Referendum",
			value: $("#p_RF_StartPage").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_User_Voting",
			menu: "Referendum",
			value: $("#p_RF_User_Voting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_UserList",
			menu: "Referendum",
			value: $("#p_RF_UserList").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_UserVotingList",
			menu: "Referendum",
			value: $("#p_RF_UserVotingList").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_ViewInfo",
			menu: "Referendum",
			value: $("#p_RF_ViewInfo").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_ViewResult",
			menu: "Referendum",
			value: $("#p_RF_ViewResult").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_VotingCancel",
			menu: "Referendum",
			value: $("#p_RF_VotingCancel").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_VotingConfirm",
			menu: "Referendum",
			value: $("#p_RF_VotingConfirm").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_Сancel",
			menu: "Referendum",
			value: $("#p_RF_Сancel").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
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
		},
{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendMenu",
				typeid: #type_append_menu_id#,
				name : "menu_default",
				value: $("#am_menu_default").val(),
				global: 0
				}
		}]`
)