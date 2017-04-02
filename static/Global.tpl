SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	typecolid = TxId(NewColumn),
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
SetVar(`sc_addMessageGL #= contract addMessageGL {
	data {
        Text string
    }
 	func conditions {
		
 	    if !DBIntExt( "global_states_list", "united_governments", $state, "gstate_id")
 	    {
 	        warning "Sorry, you don't have access to this action. Only for members of the United Governments."
 	    } 	    
	}
	func action {
	    
		var avatar string
		avatar = DBStringExt( Table("citizens"), "avatar", $citizen, "id")
		var name string
		name = DBStringExt( Table("citizens"), "name", $citizen, "id")
		var flag string
		flag = DBStringExt( Table("state_parameters"), "value", "state_flag", "name")
		var statename string
		statename = DBStringExt( Table("state_parameters"), "value", "state_name", "name")
		DBInsert( "global_messages", "text, ava, username, flag, citizen_id, stateid,statename", $Text, avatar, name, flag, $citizen, $state, statename) 
		
	}	
}`,
`sc_GlobalCondition #= contract GlobalCondition {
	data {}
	conditions {
		if $state != DBIntExt("global_states_list", "gstate_id", 1, "admin") {
			warning "Sorry, you don't have access to this action."
		}
	}
	action {}
}`,
`sc_glrf_NewIssue #= contract glrf_NewIssue {
 data {
    Issue string
    Type int
    Date_start_voting string "date"
    Date_stop_voting string "date"
 }
 
func conditions {
    
	   GlobalCondition()

	} 

func action {
    DBInsert("global_glrf_referendums", "issue,type,date_voting_start,date_voting_finish,status,number_votes,timestamp date_enter",$Issue,$Type,$Date_start_voting,$Date_stop_voting,0,0,$block_time)
    
  } 
}`,
`sc_glrf_SaveAns #= contract glrf_SaveAns {
	data {
		ReferendumId int
		Answer string 
	}

	func conditions {
	    

	    $sha256 = Sha256(Str($ReferendumId + $citizen))

		//var voted int
		//voted = DBIntExt(Table("glrf_votes"), "id", $sha256, "strhash")

		//if voted != 0 {
		//	info "You already voted"
		//}

		var allowed int
		allowed = DBIntWhere("global_glrf_referendums", "id", "date_voting_start < now() and date_voting_finish > now() and id=$", $ReferendumId)

		if allowed == 0 {
			info "Voting is not available now"
		}

	}
	func action {
    
    var id_voted int
	id_voted = DBIntWhere("global_glrf_votes", "id", "referendum_id=$ and citizen_id=$", $ReferendumId,$citizen)
	
	if id_voted > 0
	{

        DBUpdate("global_glrf_votes", id_voted, "answer", $Answer)
        
	}else{
	   
	   DBInsert("global_glrf_votes","referendum_id,strhash,answer,citizen_id,timestamp time",$ReferendumId,$sha256,$Answer, $citizen,$block_time)
        
        var counter int
        counter = DBIntExt( "global_glrf_referendums", "number_votes", $ReferendumId, "id")
        DBUpdate("global_glrf_referendums", $ReferendumId, "number_votes", counter+1)
         
	}
    
  }
}`,
`sc_glrf_Voting #= contract glrf_Voting {
	data {
		ReferendumId int
		RFChoice int
	}

	func conditions {

		ContractConditions("MainCondition")

		$sha256 = Sha256(Str($ReferendumId + $state))

		var voted int
		voted = DBIntExt("global_glrf_votes", "id", $sha256, "strhash")

		if voted != 0 {
			info "You already voted"
		}

		var allowed int
		allowed = DBIntWhere("global_glrf_referendums", "id", "date_voting_start < now() and date_voting_finish > now() and id=$", $ReferendumId)

		if allowed == 0 {
			info "Voting is not available now"
		}

	}
	func action {

		var id_voted int
		id_voted = DBIntWhere("global_glrf_votes", "id", "referendum_id=$ and state_id=$", $ReferendumId, $state)

		if id_voted > 0 {

			DBUpdate("global_glrf_votes", id_voted, "choice", $RFChoice)

		} else {

			DBInsert("global_glrf_votes", "referendum_id,strhash,choice,state_id,timestamp time", $ReferendumId, $sha256, $RFChoice,$state, $block_time)

			var counter, id_table int
			counter = DBIntExt("global_glrf_referendums", "number_votes", $ReferendumId, "id")
			DBUpdate("global_glrf_referendums", $ReferendumId, "number_votes", counter + 1)
			
			counter = DBIntExt("global_states_list", "num_answers", $state, "gstate_id")
			id_table = DBIntExt("global_states_list", "id", $state, "gstate_id")
			DBUpdate("global_states_list", id_table, "num_answers", counter+1)

		}


	}
}`,
`sc_glrf_VotingCancel #= contract glrf_VotingCancel {
	data {
		ReferendumId int
	}

	func conditions {

        GlobalCondition()


		if DBIntWhere("global_glrf_referendums", "id", "date_voting_start > now() and id=$ and status=2", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate("global_glrf_referendums",$ReferendumId,"status",2)

	}
}`,
`sc_glrf_VotingDel #= contract glrf_VotingDel {
	data {
		ReferendumId int
	}

	func conditions {

    GlobalCondition()

		if DBIntWhere("global_glrf_referendums", "id", "date_voting_finish > now() and id=$", $ReferendumId) > 0 
		{
			info "action is not available"
		}

	}

	func action {
	    DBUpdate("global_glrf_referendums",$ReferendumId,"status",2)

	}
}`,
`sc_glrf_VotingResult #= contract glrf_VotingResult {
	data {
		ReferendumId int
	}
	func conditions {
	    
	     GlobalCondition()

		var x int

		x = DBIntExt(Table("state_parameters"), "value", "gov_account", "name")
		if x != $citizen {
			info "You do not have the right to stop vote"
		}

		//x = DBIntExt(Table("glrf_referendums"),"status",$ReferendumId,"id")
		//if x == 1 {
		//	info "Resalt is ready"
		//}

	}

	func action {
		var votes0 int
		var votes1 int
		var resalt int
		var id_res0 int
		var id_res1 int
		
		votes0 = DBIntWhere("global_glrf_votes","count(id)","referendum_id=$ and choice=0",$ReferendumId)
        votes1 = DBIntWhere("global_glrf_votes","count(id)","referendum_id=$ and choice=1",$ReferendumId)
		
    	if(votes0==0 && votes1==0)
    	{
    	    DBUpdate("global_glrf_referendums",$ReferendumId,"result,number_0,number_1",0,votes0,votes1)
    	
    	}else{
    		
    		if (votes1 > votes0) {
    			resalt = 1
    		} else {
    			resalt = 0
    		}
    		
    		DBUpdate("global_glrf_referendums",$ReferendumId,"result,number_0,number_1",resalt,votes0,votes1)
    		
    		id_res0 = DBIntWhere("global_glrf_result","id","referendum_id=$ and choice=0",$ReferendumId)
    		
    		if id_res0 > 0 
    		{
    		    id_res1 = DBIntWhere("global_glrf_result","id","referendum_id=$ and choice=1",$ReferendumId)
    	    	DBUpdate("global_glrf_result",id_res0,"value,percents",votes0,100*votes0/(votes1+votes0))
    	    	DBUpdate("global_glrf_result",id_res1,"value,percents",votes1,100*votes1/(votes1+votes0))
    		
    		}else{
            		
            	DBInsert("global_glrf_result","referendum_id,choice,choice_str,value,percents",$ReferendumId,1,"Yes",votes1,100*votes1/(votes1+votes0))
            	DBInsert("global_glrf_result","referendum_id,choice,choice_str,value,percents",$ReferendumId,0,"No",votes0,100*votes0/(votes1+votes0))
    		}
    		
    		var finish int
		    finish = DBIntWhere("global_glrf_referendums", "id", "date_voting_finish < now() and id=$", $ReferendumId)
		    if finish > 0
		    {
		        	DBUpdate("global_glrf_referendums",$ReferendumId,"status",1)
		    }
    		
    	}	

	}
}`,
`sc_glrf_VotingStart #= contract glrf_VotingStart {
	data {
		ReferendumId int
	}

	func conditions {

         GlobalCondition()


		if DBIntWhere("global_glrf_referendums", "id", "date_voting_start < now() and id=$", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate("global_glrf_referendums",$ReferendumId,"timestamp date_voting_start",$block_time)

	}
}`,
`sc_glrf_VotingStop #= contract glrf_VotingStop {
	data {
		ReferendumId int
	}

	func conditions {

         GlobalCondition()

		if DBIntWhere("global_glrf_referendums", "id", "date_voting_finish < now() and id=$", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate("global_glrf_referendums",$ReferendumId,"timestamp date_voting_finish,status",$block_time,0)

	}
}`,
`sc_UG_Vote #= contract UG_Vote {
	data {
		State_num int
	}

	conditions {

		ContractConditions("MainCondition")

		$State_id = DBIntExt("global_states_list", "gstate_id", $State_num, "id")
		$sha256 = Sha256(Str($state)+"-"+Str($State_id))

		var voted int
		voted = DBIntExt("global_ug_votes", "id", $sha256, "strhash")

		if voted != 0 {
			info "You already voted"
		}
		
		 if $State_id==$state
        {
             warning "You can not vote for your state."
        }
	}

	action {

		var counter_vote, counter_voting, id_table_voting int
		counter_vote = DBIntExt("global_states_list", "num_votes", $State_id, "gstate_id")
		counter_voting = DBIntExt("global_states_list", "num_voting", $state, "gstate_id")
		id_table_voting = DBIntExt("global_states_list", "id", $state, "gstate_id")


		DBUpdate("global_states_list", $State_num, "num_votes", counter_vote + 1)
		DBUpdate("global_states_list", id_table_voting, "num_voting", counter_voting + 1)

		DBInsert("global_ug_votes", "strhash,timestamp time", $sha256, $state)

		if counter_vote == 4 {
			DBUpdate("global_states_list", $State_num, "united_governments", 1)
		}


	}
}`)
TextHidden(sc_addMessageGL, sc_GlobalCondition, sc_glrf_NewIssue, sc_glrf_SaveAns, sc_glrf_Voting, sc_glrf_VotingCancel, sc_glrf_VotingDel, sc_glrf_VotingResult, sc_glrf_VotingStart, sc_glrf_VotingStop, sc_UG_Vote)
SetVar(`p_glrf_List #= Title : $ListVotings$
Navigation( LiTemplate(government,Government), Citizens)

SetVar(ViewResultQues = BtnPage(glrf_ViewResultQuestions, <b>$Vw$ #number_votes#</b>, "ReferendumId:#id#,Back:0,Status:#status#,global:1",btn btn-primary btn-block))
SetVar(ViewResult = BtnPage(glrf_ViewResult, <b>$Vw$ #number_votes#</b>, "ReferendumId:#id#,Back:0,Status:#status#,global:1",btn btn-primary btn-block))

SetVar(Cancel = BtnContract(@glrf_VotingCancel, <b>$Cncl$</b>,Cancel Votin, "ReferendumId:#id#",'btn btn-primary btn-block',template,glrf_List,global:1))
SetVar(Stop = BtnContract(@glrf_VotingStop, <b>$Stp$</b>,Stop Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,glrf_List,global:1))
SetVar(Delete = BtnContract(@glrf_VotingDel, <b>$Del$</b>,Delete Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,glrf_List,global:1))
SetVar(Start = BtnContract(@glrf_VotingStart, <b>$Strt$</b>,Start Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,glrf_List,global:1))

Divs(md-12, panel panel-default data-sweet-alert)
    Divs(panel-body)
        Divs(table-responsive)
            Table{
                Table: global_glrf_referendums
                Class: table-striped table-bordered table-hover data-role="table"
                Order: #id# DESC
                Where: #status#!=2
                Adaptive: 1
                Columns: [
                    [$Iss$, #issue#],
                    [Type,P(h4,If(#type#==2,Q,V))],
                    [$Strt$, P(h6,DateTime(#date_voting_start#, YYYY.MM.DD HH:MI))],
                    [$Fnsh$, P(h6,DateTime(#date_voting_finish#, YYYY.MM.DD HH:MI))],
                    [$Inf$,If(#type#==2,#ViewResultQues#,If(#number_votes# == 0, #number_votes#,If(#status#==1,#ViewResult#,BtnContract(@glrf_VotingResult, <b>$Vw$ #number_votes#</b>, Get Result,"ReferendumId:#id#",'btn btn-primary btn-block',template,glrf_ViewResult,"ReferendumId:#id#, Back:0, Status:#status#,global:1"))))],
                    [$Actn$, If(#CmpTime(#date_voting_start#, Now(datetime)) == 1,#Cancel#, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1,#Stop#, #Delete#))],
                    [$Res$,If(#CmpTime(#date_voting_start#, Now(datetime)) == 1, - , If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1, $Contin$, If(#status# == 1, If(#result#==1,$Y$,$N$), $Fnshd$)))]
                ]
            }
        DivsEnd:
    DivsEnd:
    Divs(panel-footer text-center)
        BtnPage(glrf_NewIssue, <b>$NewVoting$</b>,"Status:0,global:1",btn btn-oval btn-info btn-lg md5 pd5)
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_glrf_NewIssue #= Title : $NewVoting$
Navigation( LiTemplate(government,Government), $NewVoting$ )

Divs(md-6, panel panel-default data-sweet-alert)
    Div(panel-heading, Div(panel-title, $NewVoting$))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label($EnterIssue$)
                Textarea(Issue, form-control input-lg)
            DivsEnd:
            Divs(form-group)
                Label($TypeIssue$)
                Select(Type,type_issue,form-control input-lg)
            DivsEnd:
            Divs(form-group)
                Label($DateStartVoting$)
                InputDate(Date_start_voting,form-control input-lg,Now(YYYY.MM.DD HH:MI))
            DivsEnd:
            Divs(form-group)
                Label($DateFinishVoting$)
                InputDate(Date_stop_voting,form-control input-lg,Now(YYYY.MM.DD HH:MI,60 days))
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        Divs: clearfix
            Divs: pull-right
                BtnPage(glrf_List, $ListVotings$, "Status:1,global:1",btn btn-default btn-pill-left pull-left ml4)
                TxButton{ClassBtn:btn btn-primary btn-pill-right, Contract: @glrf_NewIssue,Name:$Save$,Inputs:"Issue=Issue,Type=Type,Date_start_voting=Date_start_voting,Date_stop_voting=Date_stop_voting", OnSuccess: "template,glrf_List,global:1"}
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_glrf_UserAns #= Title : $Ans$
Navigation( LiTemplate(dashboard_default, Dashboard), $Ans$)

SetVar(Issue = GetOne(issue, global_glrf_referendums, "id", #ReferendumId#))
If(#Chng#)
    SetVar(Answer = GetOne(answer, global_glrf_votes, "id", #VoteId#))
Else:
    SetVar(Answer = "")
IfEnd:

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
        MarkDown: P(h4 text-primary,#Issue#)
          
        Form()
        Divs(form-group)
            Label($Ans$)
            Textarea(Answer, form-control input-lg,#Answer#)
        DivsEnd:
        
       
        Divs(text-right)
        
            Input(ReferendumId, "hidden", text, text, #ReferendumId#)

            TxButton{ClassBtn:btn btn-primary btn-pill-right, Contract: glrf_SaveAns,Name:$Save$,Inputs:"ReferendumId=ReferendumId, Answer=Answer", OnSuccess: "template,glrf_UserQuestionList"}
            
            BtnPage(glrf_UserQuestionList, $QuestionList$, "Status:1",btn btn-default btn-pill-left ml4)
        DivsEnd:     
        Div(clearfix)    
         
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_glrf_UserList #= Title : $ListVotings$
Navigation( LiTemplate(dashboard_default, Dashboard),$QuestionList$)

Divs(md-6, panel panel-default panel-body data-sweet-alert)

SetVar(VotingY = BtnContract(@glrf_Voting, $Y$, Your choice Yes,"ReferendumId:#id#,RFChoice:1",'btn btn-info',template,glrf_UserList))
SetVar(VotingN = BtnContract(@glrf_Voting, $N$, Your choice No,"ReferendumId:#id#,RFChoice:0",'btn btn-info',template,glrf_UserList))

SetVar(ViewResult = BtnPage(glrf_ViewResult, <strong>$Res$</strong>, "ReferendumId:#id#,DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#', NumberVotes:#number_votes#,Back:1,Status:1,global:1",btn btn-info btn-pill-right))

SetVar(Voting = BtnPage(glrf_UserVotingList, $Vote$, "ReferendumId:#id#",btn btn-primary btn-pill-right))

SetVar(ChangeVoting = BtnPage(glrf_UserVotingList, $Chng$, "ReferendumId:#id#",btn btn-pill-right))

GetList(vote,global_glrf_votes,"referendum_id,choice",state_id=#state_id#,id)



    Table{
         Table: global_glrf_referendums
         Order: #date_voting_start# DESC 
         Where: #status#!=2 and #date_voting_start# < now()  and type=1
         
      Columns: [[,If(ListVal(vote,#id#,referendum_id)>0,P(h4 text-muted, <span id="tr#id#">#issue#</span>),P(h4 text-primary, #issue#)) If(#status#==1, #ViewResult#, If(ListVal(vote,#id#,referendum_id)>0," ",#VotingY# #VotingN#))],
        [,If(ListVal(vote,#id#,referendum_id)>0,If(ListVal(vote,#id#,choice)==1,P(h4 text-primary,$Y$),P(h4 text-danger,$N$)))]
        ]
     }
DivsEnd:

PageEnd:`,
`p_glrf_UserQuestionList #= Title : $QuestionList$
Navigation( LiTemplate(dashboard_default, Dashboard),$QuestionList$)

SetVar(ViewResult = BtnPage(glrf_ViewResultQuestions, <strong>$Res$</strong>, "ReferendumId:#id#,Back:1,Status:1",btn btn-info btn-pill-right))

SetVar(Voting = BtnPage(glrf_UserAns, $Ans$, "ReferendumId:#id#, Chng:0",btn btn-primary btn-pill-right))

SetVar(ChangeVoting = BtnPage(glrf_UserAns, $Chng$, "ReferendumId:#id#,Chng:1,VoteId:ListVal(vote,#id#,id)",btn btn-pill-right))

GetList(vote,global_glrf_votes,"referendum_id,answer,id",state_id=#state_id#,id)


Divs( md-6,panel panel-success)
    Divs: panel-body


    Table{
         Table: global_glrf_referendums
         Order: #date_voting_start# DESC 
         Where: #status#!=2 and #date_voting_start# < now() and #type#=2
         
      Columns: [[,If(ListVal(vote,#id#,referendum_id)>0,P(h4 text-muted, <span id="tr#id#">#issue#</span>) P(h4 text-primary, ListVal(vote,#id#,answer)), P(h4 text-primary, #issue#)) If(#status#==1, #ViewResult#, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1,If(ListVal(vote,#id#,referendum_id)>0,#ChangeVoting#,#Voting#),))],
        ]
     }

DivsEnd:
DivsEnd:

PageEnd:`,
`p_glrf_ViewResult #= Title : $Res$
Navigation(LiTemplate(government),$Res$)

GetRow(vote,global_glrf_referendums,"id",#ReferendumId#)

Divs(md-6, panel panel-info elastic center)
    Div(panel-heading, Div(panel-title, #vote_issue#))
    Divs(panel-body f0)
        Divs(panel-title text-center)
            MarkDown: <strong>DateTime(#vote_date_voting_start#, YYYY.MM.DD HH:MI) - DateTime(#vote_date_voting_finish#, YYYY.MM.DD HH:MI)</strong>
            MarkDown: <h4>$TotalVoted$: #vote_number_votes#</h4>
        DivsEnd:
    DivsEnd:
    Divs(panel-body)
        Divs(panel-title text-center)
            Divs(table-responsive)
                Table{
                    Table: global_glrf_result
                    Class: table-striped table-bordered table-hover data-role="table"
                    Where: referendum_id=#ReferendumId#
                    Order: #choice_str# DESC
                    Columns: [
                        [
                            Ответ,
                            If(#choice#==1,P(h4 text-primary text-bold m0,$Y$),P(h4 text-danger text-bold m0,$N$))
                        ],
                        [
                            Голоса,
                            If(#choice#==1,P(h4 text-primary m0,#value#),P(h4 text-danger m0,#value#))
                        ],
                        [
                            Процент,
                            If(#choice#==1,P(h4 text-primary m0,#percents# %),P(h4 text-danger m0, #percents# %))
                        ]
                    ]
                }
            DivsEnd:
        DivsEnd:
    DivsEnd:
    Divs(panel-footer text-center)
            If(#Back#==1,BtnPage(government, <strong>$ListVotings$</strong> ,"Status:1",btn btn-oval btn-info f0,'list'), BtnPage(glrf_List, <strong>$ListVotings$</strong>, "Status:0,global:1",btn btn-oval btn-info f0,'list'))
        DivsEnd:
    DivsEnd:
DivsEnd:

Divs(md-6, panel panel-default elastic center)
    Divs: panel-body canvas-responsive
        ChartPie{Table: global_glrf_result, FieldValue: percents, FieldLabel: choice_str, Colors: "f05050,5d9cec,37bc9b,f05050,23b7e5,ff902b,f05050,131e26,37bc9b,f532e5,7266ba,3a3f51,fad732,232735,3a3f51,dde6e9,e4eaec,edf1f2", Where: referendum_id = #ReferendumId#, Order: choice}
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_glrf_ViewResultQuestions #= Title : $Res$
Navigation( LiTemplate(government),  $Res$)

SetVar(Issue = GetOne(issue, global_glrf_referendums, "id", #ReferendumId#))
SetVar(Del = GetOne(id, global_citizen_del,"citizen_id", #citizen#))

Divs( md-6,panel panel-success)
    Divs: panel-body
    MarkDown: P(h4 text-primary,#Issue#)
    
    If (#Del#>0)
        MarkDown: P(h4 text-center text-danger,Uw account is opgeschort)
    Else:
    
    Divs(btn-lg)
    If(#Back#==1,BtnPage(glrf_UserQuestionList, <strong>$QuestionList$</strong> ,"Status:1",btn btn-pill-left btn-info), BtnPage(glrf_List, <strong>$ListVotings$</strong>, "Status:0",btn btn-pill-left btn-info))
     DivsEnd:

    Table{
         Table: global_glrf_votes
         Order: #id# DESC 
         Where: #referendum_id#=#ReferendumId#
         
      Columns: [[,P(h4 text-muted, #answer#)],
        ]
     }
    IfEnd:

DivsEnd:
DivsEnd:

PageEnd:`)
TextHidden( p_glrf_List, p_glrf_NewIssue, p_glrf_UserAns, p_glrf_UserList, p_glrf_UserQuestionList, p_glrf_ViewResult, p_glrf_ViewResultQuestions)
SetVar(`m_Global #= MenuItem(List Votings,  glrf_List, global:1)
MenuItem(New Voting,  glrf_NewIssue, global:1)`,
`m_Goverment #= MenuItem(Citizen dashboard, dashboard_default)
MenuItem(Government dashboard, government)`)
TextHidden( m_Global, m_Goverment)
SetVar()
TextHidden( )
SetVar()
TextHidden( )
Json(`Head: "Global",
Desc: "",
		Img: "/static/img/apps/ava.png",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #typecolid#,
			table_name : "global_states_list",
			column_name: "num_votes",
			index: "1",
			column_type: "int64",
			permissions: "ContractConditions(\"MainCondition\")"

			}
	   },
{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #typecolid#,
			table_name : "global_states_list",
			column_name: "num_answers",
			index: "1",
			column_type: "int64",
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #typecolid#,
			table_name : "global_states_list",
			column_name: "united_governments",
			index: "1",
			column_type: "int64",
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #typecolid#,
			table_name : "global_states_list",
			column_name: "num_voting",
			index: "1",
			column_type: "int64",
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,column_name,permissions,index,column_type',
		Data: {
			type: "NewColumn",
			typeid: #typecolid#,
			table_name : "global_states_list",
			column_name: "state_flag",
			index: "0",
			column_type: "text",
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "GlobalCondition",
			value: $("#sc_GlobalCondition").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "GlobalCondition"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "glrf_referendums",
			columns: '[["date_voting_start", "time", "1"],["date_voting_finish", "time", "1"],["type", "int64", "1"],["result", "int64", "1"],["status", "int64", "1"],["number_1", "int64", "1"],["date_enter", "time", "1"],["number_votes", "int64", "1"],["issue", "text", "0"],["number_0", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "global_glrf_referendums",
			general_update: "ContractConditions(\"GlobalCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "result",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "number_0",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "number_votes",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "date_voting_start",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "issue",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "status",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "number_1",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "date_enter",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "date_voting_finish",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_referendums",
			column_name: "type",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "glrf_result",
			columns: '[["choice_str", "text", "0"],["referendum_id", "int64", "1"],["value", "int64", "1"],["choice", "int64", "1"],["percents", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "global_glrf_result",
			general_update: "ContractConditions(\"GlobalCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_result",
			column_name: "choice",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_result",
			column_name: "percents",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_result",
			column_name: "choice_str",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_result",
			column_name: "referendum_id",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_result",
			column_name: "value",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "glrf_votes",
			columns: '[["time", "time", "1"],["answer", "text", "0"],["choice", "int64", "1"],["strhash", "hash", "1"],["state_id", "int64", "1"],["referendum_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "global_glrf_votes",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"CitizenCondition\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "choice",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "strhash",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "state_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "referendum_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "time",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_glrf_votes",
			column_name: "answer",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "messages",
			columns: '[["stateid", "int64", "1"],["state_id", "int64", "1"],["username", "hash", "1"],["statename", "hash", "1"],["citizen_id", "int64", "1"],["ava", "text", "0"],["flag", "text", "0"],["text", "text", "0"],["time", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "global_messages",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "stateid",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "state_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "username",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "statename",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "ava",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "flag",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "text",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_messages",
			column_name: "time",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "ug_votes",
			columns: '[["time", "time", "0"],["strhash", "hash", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "global_ug_votes",
			general_update: "ContractConditions(\"GlobalCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_ug_votes",
			column_name: "time",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "global_ug_votes",
			column_name: "strhash",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "addMessageGL",
			value: $("#sc_addMessageGL").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "addMessageGL"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_NewIssue",
			value: $("#sc_glrf_NewIssue").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_NewIssue"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_SaveAns",
			value: $("#sc_glrf_SaveAns").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_SaveAns"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_Voting",
			value: $("#sc_glrf_Voting").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_Voting"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_VotingCancel",
			value: $("#sc_glrf_VotingCancel").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_VotingCancel"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_VotingDel",
			value: $("#sc_glrf_VotingDel").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_VotingDel"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_VotingResult",
			value: $("#sc_glrf_VotingResult").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_VotingResult"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_VotingStart",
			value: $("#sc_glrf_VotingStart").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_VotingStart"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "glrf_VotingStop",
			value: $("#sc_glrf_VotingStop").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "glrf_VotingStop"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "UG_Vote",
			value: $("#sc_UG_Vote").val(),
			conditions: "ContractConditions(\"GlobalCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 1,
			id: "UG_Vote"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "Global",
			value: $("#m_Global").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "Goverment",
			value: $("#m_Goverment").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_List",
			menu: "Global",
			value: $("#p_glrf_List").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_NewIssue",
			menu: "Global",
			value: $("#p_glrf_NewIssue").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_UserAns",
			menu: "menu_default",
			value: $("#p_glrf_UserAns").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_UserList",
			menu: "Global",
			value: $("#p_glrf_UserList").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_UserQuestionList",
			menu: "menu_default",
			value: $("#p_glrf_UserQuestionList").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_ViewResult",
			menu: "Goverment",
			value: $("#p_glrf_ViewResult").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "glrf_ViewResultQuestions",
			menu: "menu_default",
			value: $("#p_glrf_ViewResultQuestions").val(),
			global: 1,
			conditions: "ContractConditions(\"GlobalCondition\")",
			}
	   }
]`
)