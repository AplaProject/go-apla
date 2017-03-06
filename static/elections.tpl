SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_menu_id = TxId(NewMenu),
	type_new_contract_id = TxId(NewContract),
	type_activate_contract_id = TxId(ActivateContract),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable),
	type_append_menu_id = TxId(AppendMenu),
	sc_conditions = "$citizen == #wallet_id#")
SetVar(`sc_SLStartVoting = contract SLStartVoting {
	data {
		SmartLaws string
		SmartLawsId int
		TabId int
	}
	func conditions {
		var x int
		x = DBIntWhere(Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		if x == 0 {
			info "You do not have the right to put to vote"
		}

		x = DBIntWhere(Table("laws_edition"), "id", "laws_id=$ and status=0", $SmartLawsId)
		if x == 0 {
			info "This law can not be put to vote"
		}

	}

	func action {
		DBUpdate(Table("laws_edition"), $TabId, "status,resalt_voting,timestamp date_voting_start", 1, 0, $block_time)
	}
}`,
`sc_GECandidateRegistration = contract GECandidateRegistration {
	data {
		CampaignName string
		FirstName string "choice"
		Description string
		CampaignId int
		PositionId int

	}

	func conditions {
		var allowed int
		allowed = DBIntWhere(Table("ge_campaigns"), "id", "date_start < now() and candidates_deadline > now() and id=$", $CampaignId)
		if allowed == 0 {
			warning "Submission of applications is not available at the moment"
		}

		allowed = DBIntWhere(Table("ge_candidates"), "id", "citizen_id=$ and id_election_campaign=$", $citizen, $CampaignId)
		if allowed != 0 {
			warning "You are already registered as a candidate"
		}

	}

	func action {
		DBInsert(Table("ge_candidates"), "candidate,citizen_id,description,id_election_campaign,timestamp application_date,position_id,campaign,result", $FirstName, $citizen, $Description, $CampaignId, $block_time, $PositionId, $CampaignName,0)

	}
}`,
`sc_GENewElectionCampaign = contract GENewElectionCampaign {
 data {
    ElectionName string "Election campaign"
    DateStart string "date"
    CandidatesDeadline "date"
    Date_start_voting string "date"
    Date_stop_voting string "date"
    PositionId int
 }

func action {
    DBInsert(Table( "ge_campaigns"), "name,date_start,candidates_deadline,date_start_voting,date_stop_voting, position_id,num_votes",$ElectionName, $DateStart,$CandidatesDeadline,$Date_start_voting,$Date_stop_voting,$PositionId,0)
    
  } 
}`,
`sc_GEVoting = contract GEVoting {
	data {
		Candidate string
		Campaign string
		ChoiceId int 
		CampaignId int 

	}

	func conditions {

	    $sha256=$CampaignId+$citizen

		var voted int
		voted = DBIntExt(Table("ge_votes"), "id", $sha256, "strhash")

		if voted != 0 {
			info "You already voted"
		}

		var allowed int
		allowed = DBIntWhere(Table("ge_campaigns"), "id", "date_start_voting < now() and date_stop_voting > now() and id=$", $CampaignId)

		if allowed == 0 {
			info "Voting is not available now"
		}

	}
	func action {
    
    
    DBInsert(Table("ge_votes"),"strhash,id_candidate,timestamp time",$sha256, $ChoiceId,$block_time)
    
    var counter int
    counter = DBIntExt( Table("ge_candidates"), "counter", $ChoiceId, "id")
    DBUpdate(Table( "ge_candidates"), $ChoiceId, "counter", counter+1)
    
    counter = DBIntExt( Table("ge_campaigns"), "num_votes", $CampaignId, "id")
    DBUpdate(Table( "ge_campaigns"), $CampaignId, "num_votes", counter+1)
    
    
  }
}`,
`sc_GEVotingResalt = contract GEVotingResalt {
 data {
    CampaignId int 
    Position string
 }
 
 func conditions {
     
    // for Assembly representative
    $numChoiceRes=4

     
    var x int
    x=DBIntWhere( Table("ge_campaigns"), "id", "date_stop_voting < now() and id=$", $CampaignId)
    if x==0 {
      info "Resalt is not available now"
    }
    
    x=DBIntExt( Table("ge_campaigns"), "status", $CampaignId, "id")
    if x==1 {
       info "Resalt is ready"
    }
    
}

func action {
    
        var list array
        var votes int
        var war map
        var i int
        var len int
        list = DBGetList("1_ge_candidates", "id,candidate,citizen_id,position_id",0,100,"counter desc", "id_election_campaign=$", $CampaignId)
        len = Len(list)
        while i < len {
            war = list[i]
            i = i + 1
            votes = DBIntWhere( Table("ge_votes"), "count(id)", "id_candidate=$",war["id"])
            DBUpdate(Table("ge_candidates"),Int(war["id"]), "counter", votes)
            if i < $numChoiceRes+1 {
                DBUpdate(Table("ge_candidates"),Int(war["id"]), "result", 1)
                
                DBInsert(Table("ge_person_position"),"position_id,name,citizen_id,timestamp date_start,position",war["position_id"],war["candidate"],war["citizen_id"],$block_time,$Position)
            
            } else {
                DBUpdate(Table("ge_candidates"),Int(war["id"]), "result", 0)
            }
        }
        
        DBUpdate(Table("ge_campaigns"),$CampaignId, "status", 1)
    
  }
}`,
`sc_LSSignature = contract LSSignature {
	data {
		SmartLawsId int
		TabId int
	}
	func conditions {
	    var x int
	    
	    x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		if x == 0 {
			info "You do not have the right to sign"
		}	
		
	    x=DBIntWhere( Table("laws_edition"), "id", "id=$ and status < 3", $TabId)
		if x != 0 {
			info "Law or an amendment to the law has not yet taken"
		}
		
		x=DBIntWhere( Table("laws_edition"), "id", "id=$ and status=4", $TabId)
		if x != 0 {
			info "The law has been signed"
		}
		
	    x=DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and resalt_voting=1", $SmartLawsId)
		if x == 0 {
			info "Law or amendments to the law have been rejected"
		}
		

	}

	func action {

    	var id_law int
    	id_law=DBIntWhere( Table("sl_list"), "id", "id_smart_contract=$", $SmartLawsId)
    	DBUpdate(Table("sl_list"), id_law, "timestamp date_last_edition", $block_time)
    	
        DBUpdate(Table("laws_edition"), $TabId, "status,timestamp date_end", 4, $block_time)
        var value string
        value=DBStringExt( Table("laws_edition"), "value",$TabId, "id")
        DBUpdate(Table("smart_contracts"), $SmartLawsId, "value", value)
    
	}
}`,
`sc_SLEdit = contract SLEdit {
	data {
		LawsValue string
		TabId int
	}
	func conditions {
		
		 var x int
	     x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		if x == 0 {
			info "You do not have the right to edit"
		}	
		x=DBIntWhere( Table("laws_edition"), "id", "id=$ and status>0", $TabId)
		if x != 0 {
			info "Change forbidden"
		}
	}

	func action {
		DBUpdate(Table("laws_edition"), $TabId, "value", $LawsValue)
	}
}`,
`sc_SLNewVoting = contract SLNewVoting {
	data {
		SmartLaws string
		SmartLawsId int
	}
	func conditions {
	    var x int
	    x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		if x == 0 {
			info "You do not have the right to do it"
		}
		
	    x=DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and status!=4", $SmartLawsId)
		if x != 0 {
			info "This law currently under consideration"
		}
		
	
		
		$value=DBStringExt( Table("smart_contracts"), "value", $SmartLawsId, "id")
	}

	func action {
	DBInsert(Table("laws_edition"), "name,laws_id,value,status,resalt_voting,resalt_for,result_against,timestamp date_start",$SmartLaws, $SmartLawsId, $value,0,0,0,0, $block_time)

	}
}`,
`sc_SLVoting = contract SLVoting {
	data {
		SmartLaws string
		SmartLawsId int 
		Vote int
		TabId int
	}

	func conditions {


		var x int
		
	    x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=1", $citizen)

		if x == 0 {
			info "You are not Assembly representative"
		}
		
        x=DBIntWhere( Table("sl_votes"), "id", "citizen_id=$ and laws_id=$", $citizen, $SmartLawsId)

		if x != 0 {
			info "You already voted"
		}

	
	    x=DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and status=1", $SmartLawsId)

		if x == 0 {
			info "Voting is not available now"
		}

	}
	func action {
    
    var citizen_name string
    citizen_name=DBStringExt( Table("ge_person_position"), "name", $citizen, "citizen_id")
    
    DBInsert(Table("sl_votes"),"citizen_id,citizen_name,laws_id,choice,id_voting,timestamp time",$citizen,citizen_name,$SmartLawsId,$Vote,$TabId, $block_time)
    
    
  }
}`,
`sc_SLVotingResalt = contract SLVotingResalt {
	data {
		SmartLawsId int
		TabId int
	}

	func conditions {

		var x int
		 x=DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and status!=1", $SmartLawsId)
		if x != 0 {
			 info "This law is not put to vote"
		}
		
		x=DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen)
		if x == 0 {
			info "You do not have the right to stop vote"
		}

		x = DBIntExt(Table("laws_edition"),"resalt_voting",$SmartLawsId,"laws_id")
		if x == 1 {
			info "Resalt is ready"
		}

	}

	func action {
		var votes0 int
		var votes1 int
		var resalt int
		votes0 = DBIntWhere(Table("sl_votes"),"count(id)","id_voting=$ and choice=0",$TabId)
		votes1 = DBIntWhere(Table("sl_votes"),"count(id)","id_voting=$ and choice=1",$TabId)
		if (votes1 > votes0) {
			resalt = 1
		} else {
			resalt = 0
		}
	DBUpdate(Table("laws_edition"),$TabId,"resalt_voting,status,timestamp date_voting_stop,result_against,resalt_for",resalt,3,$block_time,votes0,votes1)
		
	DBInsert(Table("sl_result"),"id_laws_edition,choise,value,percents",$TabId,"For",votes1,100*votes1/(votes1+votes0))
	DBInsert(Table("sl_result"),"id_laws_edition,choise,value,percents",$TabId,"Against",votes0,100*votes0/(votes1+votes0))

	}
}`)
TextHidden( sc_SLStartVoting, sc_GECandidateRegistration, sc_GENewElectionCampaign, sc_GEVoting, sc_GEVotingResalt, sc_LSSignature, sc_SLEdit, sc_SLNewVoting, sc_SLVoting, sc_SLVotingResalt)
SetVar(`p_GECampaigns #= Title: Election Campaigns
Navigation(LiTemplate(GEElections, Elections), Campaigns)
Divs(md-12, panel panel-default)
    Divs(panel-heading)
        Divs(panel-title table-responsive)

Table {
	Table: 1_ge_campaigns
	Order: id
	//Where: date_stop_voting > now()
	Columns: [
		[Position, #name#],
		[Start, Date(#date_start#, YYYY.MM.DD)],
		[Deadline
			for candidates, Date(#candidates_deadline#, YYYY.MM.DD)],
	    [Candidate Registration,If(And(#CmpTime(#date_start#, Now(datetime)) == -1, #CmpTime(#candidates_deadline#, Now(datetime)) == 1), BtnPage(GECandidateRegistration,Go,"CampaignName:'#name#',CampaignId:#id#,PositionId:#position_id#"), "Finish")],
	     [Candidate,BtnPage(GECanditatesView, View,"CampaignId:#id#,Position:'#name#'")],
		[Start Voting, DateTime(#date_start_voting#, YYYY.MM.DD HH:MI)],
		[Stop Voting, DateTime(#date_stop_voting#, YYYY.MM.DD HH:MI)],
		[Voting, If(And(#CmpTime(#date_start_voting#, Now(datetime)) == -1, #CmpTime(#date_stop_voting#, Now(datetime)) == 1), BtnPage(GEVoting, Go, "CampaignId:#id#,Position:'#name#'"), #num_votes#)],
		[Result,If(#CmpTime(#date_stop_voting#, Now(datetime)) == -1, BtnPage(GEVotingResalt, View,"CampaignId:#id#,Position:'#name#'"), "--")]
	]
}
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GECandidateRegistration #= Title : Candidate Registration
Navigation( Candidate )

ValueById(#state_id#_citizens, #citizen#, "name", "FirstName")
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading data-sweet-alert)
        Divs(panel-title)
           MarkDown: Candidate to #CampaignName#  
           MarkDown: <h1>#FirstName#</h1>
             MarkDown: Description:
          
        Form()
        Input(CampaignName, "hidden", text, text, #CampaignName#)
        Input(FirstName, "hidden", text, text, #FirstName#)
        Input(CampaignId, "hidden", text, text, #CampaignId#)
        Input(PositionId, "hidden", text, text, #PositionId#)
         Input(Description, "form-control")

          MarkDown: confirm your action by pressing the 'Send' button
          
            TxButton{Contract: GECandidateRegistration,Inputs:"CampaignName=CampaignName,FirstName=FirstName,CampaignId=CampaignId,PositionId=PositionId,Description=Description", OnSuccess: "template,GECampaigns"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_GECandidates #= Title : Candidates
Navigation( Candidates )
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
Table{
         Table: 1_ge_candidates
         Where: id_election_campaign = #CampaignId#
      Columns: [[candidate, #candidate#], [Description, #description#]]
     }
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GECanditatesView #= Title : Canditates
Navigation( Canditates )
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
        MarkDown: Candidate to #Position#

Table{
         Table: 1_ge_candidates
         Where: id_election_campaign = #CampaignId#
      Columns: [[Candidate, #candidate#], [Description, #description#]]
     }
     

DivsEnd:
DivsEnd:
DivsEnd:
PageEnd:`,
`p_GEElections #= Title: Elections
Navigation(LiTemplate(government, Government), Elections)
Divs(md-8, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
Table {
	Table: 1_ge_elective_office
	Columns: [[ID, #id#],[Election 's Type, #name#], 
[Last election, Date(#last_election#, YYYY / MM / DD)],
[Start,BtnPage(GENewCampaign,Go,"ElectionName:'#name#',PositionId:#id#")]]
			}
	        DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GENewCampaign #= Title : New Election Campaign
Navigation( New Election )
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
    MarkDown: <h4>Start Representative's Election</h4>
        Divs(panel-title)
  
    Form()
      Divs(md-10)
    Divs(md-12,help-block)
    MarkDown: Start Date  
    DivsEnd:
    InputDate(DateStart,form-control input-lg,2017/01/04 00:00)
    Divs(md-12,help-block)
    MarkDown:Deadline for candidates   
    DivsEnd:
    InputDate(CandidatesDeadline,form-control input-lg,2017/01/06 00:00)
    Divs(md-12,help-block)
    MarkDown: Start Voting   
    DivsEnd:
    InputDate(Date_start_voting,form-control input-lg,2017/01/07 07:00)
    Divs(md-12,help-block)
    MarkDown: Stop Voting  
    DivsEnd:
    InputDate(Date_stop_voting,form-control input-lg,2017/01/07 22:00)
    
    Input(ElectionName, "hidden", text, text, "Representative")
     Input(PositionId, "hidden", text, text, 1)
     Divs(md-12,help-block)
     DivsEnd:
    TxButton{Contract: GENewElectionCampaign,Inputs:"ElectionName=ElectionName,PositionId=PositionId,DateStart=DateStart,CandidatesDeadline=CandidatesDeadline,Date_start_voting=Date_start_voting,Date_stop_voting", OnSuccess: "template,GECampaigns"}
     DivsEnd:
    FormEnd:
   

        DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GEPersonsPosition #= Title : Representatives of the legislative body
Navigation( Representatives )


Table{
    Table: 1_ge_person_position
    Where: position_id=3
      Columns: [
      [Name, #name#],
      [Рosition,#position#],
      [Election date,Date(#date_start#, YYYY / MM / DD)]]
     }

Table{
    Table: 1_ge_person_position
    Where: position_id=1
      Columns: [
      [Name, #name#],
      [Рosition,#position#],
      [Election date,Date(#date_start#, YYYY / MM / DD)]]
     }

PageEnd`,
`p_GEVoteConfirmation #= Title : Vote Confirmation
Navigation( Confirmation )
PageTitle : Vote Confirmation 


Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: You vite for candidate of #Campaign#  
           MarkDown: <h1>#Candidate#</h1>
           MarkDown: confirm your choice by pressing the 'Send' button
          
        Form()
        Input(Candidate, "hidden", text, text, #Candidate#)
        Input(Campaign, "hidden", text, text, #Campaign#)
        Input(ChoiceId, "hidden", text, text, #ChoiceId#)
        Input(CampaignId, "hidden", text, text, #CampaignId#)
          
            TxButton{Contract:GEVoting,Inputs:"Candidate=Candidate,Campaign=Campaign, ChoiceId=ChoiceId, CampaignId=CampaignId", OnSuccess: "template,GECampaigns"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:
        

PageEnd:`,
`p_GEVoting #= Title : Voting
Navigation( Voting )
Divs(md-8, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
Table{
         Table: 1_ge_candidates
         Where: id_election_campaign = #CampaignId#
         Order: candidate
      Columns: [[candidate, #candidate#], [Description, #description#], [Vote,BtnPage(GEVoteConfirmation,For,"ChoiceId:#id#,CampaignId:#CampaignId#,CandidateId:#citizen_id#,Candidate:'#candidate#',Campaign:'#campaign#',Result:1")]]
     }
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GEVotingResalt #= Title : Voting Result
Navigation( Voting Result )
Divs(md-6, panel panel-default panel-body)        
        	ChartBar{Table: 1_ge_candidates, FieldValue: counter, FieldLabel: candidate, Colors: "7266ba,fad732,23b7e5", Where: id_election_campaign = #CampaignId#, Order: id DESC}
DivsEnd: 
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)

Table{
         Table: 1_ge_candidates
         Where: id_election_campaign = #CampaignId#
      Columns: [[Candidate, #candidate#], [Description, #description#], [Votes,#counter#],[Result,#result#]]
     }
     

Divs(md-12, panel panel-body text-right)
MarkDown: Click on the 'Send' button to calculation of of election results 
Form()
    Input(Position, "hidden", text, text, #Position#)
    Input(CampaignId, "hidden", text, text, #CampaignId#)
    TxButton{Contract: GEVotingResalt,Inputs:"Position=Position,CampaignId=CampaignId", OnSuccess: "template,GECampaigns"}
FormEnd:
DivsEnd:
    DivsEnd:
DivsEnd:
DivsEnd:


PageEnd:`,
`p_Legislature #= Title : Legislature
Navigation( LiTemplate(government, Government), Legislature)

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title data-sweet-alert)
 MarkDown: <h4>Speaker</h4>
Table{
    Table: 1_ge_person_position
    Where: position_id=3
      Columns: [
    [Name, #name#],
      [Election date,Date(#date_start#, YYYY.MM.DD)]]
     }
     
        DivsEnd:
    DivsEnd:
DivsEnd:
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)

 MarkDown: <h4>Representatives</h4>
Table{
    Table: 1_ge_person_position
    Where: position_id=1
      Columns: [
      [Name, #name#],
      [Election date,Date(#date_start#, YYYY.MM.DD)]]
     }
     
        DivsEnd:
    DivsEnd:
DivsEnd:


        
        PageEnd:`,
`p_LSAddVoting #= Title : New Voting of Law 
Navigation( New Voting of Law )



Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: You propose to change the law  
           MarkDown: <h1>#SmartLaws#</h1>
           MarkDown: confirm your action by pressing the 'Send' button
          
        Form()
        Input(SmartLawsId, "hidden", text, text, #SmartLawsId#)
        Input(SmartLaws, "hidden", text, text, #SmartLaws#)
          
            TxButton{Contract: SLNewVoting,Inputs:"SmartLaws=SmartLaws,SmartLawsId=SmartLawsId, ChoiceId=ChoiceId, CampaignId=CampaignId", OnSuccess: "template,SLVotingList"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_LSSignature #= Title : Signature law
Navigation( Signature )

Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading data-sweet-alert)
        Divs(panel-title)
           MarkDown: Signature the law  
           MarkDown: <h1>#SmartLaws#</h1>
           MarkDown: confirm your action by pressing the 'Send' button
          
        Form()
        Input(SmartLawsId, "hidden", text, text, #SmartLawsId#)
        Input(TabId, "hidden", text, text, #TabId#)
        
          
            TxButton{Contract: LSSignature,Inputs:"SmartLawsId=SmartLawsId,TabId=TabId", OnSuccess: "template,SLVotingList"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:



PageEnd:`,
`p_LSVoting #= Title : Laws Voting
Navigation( Laws Voting )

Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: <b>#SmartLaws#</b> 
           MarkDown: <h1>#VoteTxT#</h1>
           MarkDown: confirm your decision by pressing the 'Send' button
          
        Form()
        Input(SmartLaws, "hidden", text, text, #SmartLaws#)
        Input(SmartLawsId, "hidden", text, text, #SmartLawsId#)
        Input(Vote, "hidden", text, text, #Vote#)
         Input(TabId, "hidden", text, text, #TabId#)
          
            TxButton{Contract:  SLVoting,Inputs:"SmartLaws=SmartLaws, SmartLawsId=SmartLawsId, TabId=TabId, Vote=Vote", OnSuccess: "template,SLVotingList"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_SLEdit #= Title : Law Edit
Navigation( Law Edit )

Divs(md-12, panel panel-default panel-body)
    MarkDown : <h4>#SmartLaws#</h4>    

        Form()
        Source(LawsValue, GetOne(value, 1_laws_edition, "id", #TabId#))
        Input(TabId, "hidden", text, text, #TabId#)
          
       TxButton{Contract:SLEdit,Inputs:"LawsValue=LawsValue,TabId=TabId", OnSuccess: "template,SLVotingList"}
       FormEnd:   
DivsEnd:
PageEnd:`,
`p_SLList #= Title: Smart Laws  
Navigation( LiTemplate(Legislature, Legislature), Smart Laws )

Divs(md-8, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
Table{
   
      Table: 1_sl_list
      Order:date_adoption
      Columns: [[Law, #text_name#],
      [Adoption,DateTime(#date_adoption#, YYYY.MM.DD)],[Last edition,DateTime(#date_last_edition#, YYYY.MM.DD)]
      [Amend,BtnPage(LSAddVoting,Go,"SmartLaws:'#text_name#',SmartLawsId:#id_smart_contract#")]]
     }  
        DivsEnd:
    DivsEnd:
DivsEnd:



PageEnd:`,
`p_SLRepresentativeVoting #= Title: Result of voting of Law "#SmartLaws#" 
Navigation( Representative Voting List )

Divs(md-4, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
        
Table{
         Table: 1_sl_votes
          Where: id_voting = #TabId#
      Columns: [[Name, #citizen_name#],[Vote,If(#choice#<1,"For","Against")]]
     }   
        DivsEnd:
    DivsEnd:
DivsEnd:



Divs(md-6, panel panel-default panel-body)
ChartPie{Table: 1_sl_result, FieldValue: percents, FieldLabel: choise, Colors: "5d9cec,fad732,37bc9b,f05050,23b7e5,ff902b,f05050,131e26,37bc9b,f532e5,7266ba,3a3f51,fad732,232735,3a3f51,dde6e9,e4eaec,edf1f2", Where: id_laws_edition = #TabId#, Order: choise DESC}
DivsEnd:


PageEnd:`,
`p_SLStartVoting #= Title : Start Voting of Law 
Navigation( Start Voting )



Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Start voting the law  
           MarkDown: <h1>#SmartLaws#</h1>
           MarkDown: confirm your action by pressing the 'Send' button
          
        Form()
        Input(SmartLawsId, "hidden", text, text, #SmartLawsId#)
        Input(SmartLaws, "hidden", text, text, #SmartLaws#)
        Input(TabId, "hidden", text, text, #TabId#)
        
          
            TxButton{Contract: SLStartVoting,Inputs:"SmartLaws=SmartLaws,SmartLawsId=SmartLawsId,TabId=TabId", OnSuccess: "template,SLVotingList"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_SLView #= Title : Law View
Navigation( Law View )

Divs(md-12, panel panel-default panel-body)
MarkDown : <h4>#SmartLaws#</h4>   

    Form()
    Source(LawsValue, GetOne(value, 1_laws_edition, "id", #TabId#))

    FormEnd: 
        
DivsEnd:

PageEnd:`,
`p_SLVotingList #= Title: Laws Voting List
Navigation(LiTemplate(Legislature, Legislature), Laws Voting List )
Divs(md-12, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)

Table{
         Table: 1_laws_edition
         Order: status
      Columns: [[Law, #name#]
      [Text,If(#status#==0, BtnPage(SLEdit,Edit,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"),BtnPage(SLView,View,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"))],
      [Voting Start,If(#status#==0, BtnPage(SLStartVoting,Start,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"),DateTime(#date_voting_start#, YYYY.MM.DD HH:MI))],
      [Voting “For”,If(#status#==1, BtnPage(LSVoting,For,"SmartLaws:'#name#',SmartLawsId:#laws_id#,Vote:1,VoteTxT:'For',TabId:#id#"),#resalt_for#)],
      [Voting “Against”,If(#status#==1, BtnPage(LSVoting,Against,"SmartLaws:'#name#',SmartLawsId:#laws_id#,Vote:0,VoteTxT:'Against',TabId:#id#"), #result_against#)],
     [Voting Stop, If(#status#==1, BtnPage(SLVotingResalt,Stop,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"), DateTime(#date_voting_stop#, YYYY.MM.DD HH:MI))],
      [Voting List,If(#status#>2, BtnPage(SLRepresentativeVoting,View,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"), " ")],
      [Voting results, If(#status#==3," ",#resalt_voting#)],
      [Signature,If(And(#status#==3, #resalt_voting#==1), BtnPage(LSSignature,Signature,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"),DateTime(#date_end#, YYYY.MM.DD HH:MI))]]
     }


DivsEnd:
DivsEnd:
DivsEnd:

PageEnd:`,
`p_SLVotingResalt #= Title:  Stop Voting 
Navigation(  Stop Voting )


Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Stop voting the law  
           MarkDown: <h1>#SmartLaws#</h1>
           MarkDown: confirm your action by pressing the 'Send' button
          
        Form()
        Input(SmartLawsId, "hidden", text, text, #SmartLawsId#)
        Input(TabId, "hidden", text, text, #TabId#)
        
          
            TxButton{Contract: SLVotingResalt,Inputs:"SmartLawsId=SmartLawsId,TabId=TabId", OnSuccess: "template,SLVotingList"}
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`)
TextHidden( p_GECampaigns, p_GECandidateRegistration, p_GECandidates, p_GECanditatesView, p_GEElections, p_GENewCampaign, p_GEPersonsPosition, p_GEVoteConfirmation, p_GEVoting, p_GEVotingResalt, p_Legislature, p_LSAddVoting, p_LSSignature, p_LSVoting, p_SLEdit, p_SLList, p_SLRepresentativeVoting, p_SLStartVoting, p_SLView, p_SLVotingList, p_SLVotingResalt)
SetVar(`m_Elections = [Government dashboard](government)
[Legislature dashboard](Legislature)
[Elections](GEElections)
[Campaigns](GECampaigns)
[Start Election](GENewCampaign)
[Smart contracts](sys.contracts)
`,
 `m_Legislature = [Government dashboard](government)
 [Legislature dashboard](Legislature)
 [Smart Laws](SLList)
 [Laws Voting List](SLVotingList)`,
`menu_1 #= MenuItem(Election Campaigns, load_template, GECampaigns)`,
 `menu_2 #=
MenuItem(Legislature, load_template, Legislature)
MenuItem(Election, load_template, GEElections)`)

TextHidden( m_Elections, m_Legislature, menu_1, menu_2)
SetVar()
Json(`Head: "Elections",
Desc: "General Elections and votes in the Legislature",
		Img: "/static/img/apps/elections.jpg",
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
			table_name : "ge_campaigns",
			columns: '[["position_id", "int64", "1"],["date_stop_voting", "time", "1"],["date_start_voting", "time", "1"],["candidates_deadline", "time", "1"],["name", "text", "0"],["status", "int64", "1"],["num_votes", "int64", "0"],["date_start", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_candidates",
			columns: '[["id_election_campaign", "int64", "1"],["result", "int64", "1"],["counter", "int64", "1"],["campaign", "text", "0"],["description", "text", "0"],["position_id", "int64", "1"],["application_date", "time", "0"],["candidate", "text", "0"],["citizen_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_elective_office",
			columns: '[["name", "text", "0"],["last_election", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_person_position",
			columns: '[["position_id", "int64", "1"],["name", "text", "0"],["date_end", "time", "1"],["position", "text", "0"],["citizen_id", "int64", "1"],["date_start", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_votes",
			columns: '[["hash", "text", "0"],["time", "time", "1"],["sha256", "hash", "1"],["strhash", "hash", "1"],["userhash", "hash", "1"],["id_candidate", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "laws_edition",
			columns: '[["law_id", "int64", "1"],["status", "int64", "1"],["date_start", "time", "1"],["resalt_for", "int64", "1"],["resalt_voting", "int64", "1"],["result_against", "int64", "1"],["date_voting_stop", "time", "1"],["name", "text", "0"],["value", "text", "0"],["laws_id", "int64", "1"],["date_end", "time", "1"],["conditions", "text", "0"],["date_voting_start", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_list",
			columns: '[["text_name", "text", "0"],["date_adoption", "time", "1"],["date_last_edition", "time", "1"],["id_smart_contract", "int64", "1"],["date_submission_draft", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_result",
			columns: '[["percents", "int64", "1"],["id_laws_edition", "int64", "1"],["value", "int64", "1"],["choise", "text", "0"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_votes",
			columns: '[["laws_id", "int64", "1"],["id_voting", "int64", "1"],["citizen_id", "int64", "1"],["citizen_name", "text", "0"],["time", "time", "0"],["choice", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLStartVoting",
			value: $("#sc_SLStartVoting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLStartVoting"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GECandidateRegistration",
			value: $("#sc_GECandidateRegistration").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GECandidateRegistration"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GENewElectionCampaign",
			value: $("#sc_GENewElectionCampaign").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GENewElectionCampaign"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GEVoting",
			value: $("#sc_GEVoting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GEVoting"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GEVotingResalt",
			value: $("#sc_GEVotingResalt").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GEVotingResalt"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "LSSignature",
			value: $("#sc_LSSignature").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "LSSignature"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLEdit",
			value: $("#sc_SLEdit").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLEdit"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLNewVoting",
			value: $("#sc_SLNewVoting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLNewVoting"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLVoting",
			value: $("#sc_SLVoting").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLVoting"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLVotingResalt",
			value: $("#sc_SLVotingResalt").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLVotingResalt"
			}
	   },		   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "Elections",
			value: $("#m_Elections").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "Legislature",
			value: $("#m_Legislature").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECampaigns",
			menu: "Elections",
			value: $("#p_GECampaigns").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECandidateRegistration",
			menu: "Elections",
			value: $("#p_GECandidateRegistration").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECandidates",
			menu: "Elections",
			value: $("#p_GECandidates").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECanditatesView",
			menu: "Elections",
			value: $("#p_GECanditatesView").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEElections",
			menu: "Elections",
			value: $("#p_GEElections").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GENewCampaign",
			menu: "Elections",
			value: $("#p_GENewCampaign").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEPersonsPosition",
			menu: "Legislature",
			value: $("#p_GEPersonsPosition").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEVoteConfirmation",
			menu: "Elections",
			value: $("#p_GEVoteConfirmation").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEVoting",
			menu: "Elections",
			value: $("#p_GEVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEVotingResalt",
			menu: "Elections",
			value: $("#p_GEVotingResalt").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "Legislature",
			menu: "Legislature",
			value: $("#p_Legislature").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "LSAddVoting",
			menu: "Legislature",
			value: $("#p_LSAddVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "LSSignature",
			menu: "Legislature",
			value: $("#p_LSSignature").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "LSVoting",
			menu: "Legislature",
			value: $("#p_LSVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLEdit",
			menu: "Legislature",
			value: $("#p_SLEdit").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLList",
			menu: "Legislature",
			value: $("#p_SLList").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLRepresentativeVoting",
			menu: "Legislature",
			value: $("#p_SLRepresentativeVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLStartVoting",
			menu: "Legislature",
			value: $("#p_SLStartVoting").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLView",
			menu: "Legislature",
			value: $("#p_SLView").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLVotingList",
			menu: "Legislature",
			value: $("#p_SLVotingList").val(),
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
            	value: $("#menu_1").val(),
            	global: 0
            }
       },
       {
             Forsign: 'global,name,value',
             Data: {
             	type: "AppendMenu",
             	typeid: #type_append_menu_id#,
             	name : "government",
             	value: $("#menu_2").val(),
             	global: 0
             }
       },
    {
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "SLVotingResalt",
			menu: "Legislature",
			value: $("#p_SLVotingResalt").val(),
			global: 0,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)
