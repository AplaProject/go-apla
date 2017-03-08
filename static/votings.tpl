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
SetVar(`sc_GECandidateRegistration = contract GECandidateRegistration {
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
		DBInsert(Table("ge_candidates"), "candidate,citizen_id,description,id_election_campaign,timestamp application_date,position_id,campaign,result,counter", $FirstName, $citizen, $Description, $CampaignId, $block_time, $PositionId, $CampaignName,0,0)

	}
}`,
`sc_SLaw_NumResultsVoting = contract SLaw_NumResultsVoting {
	data {
	    CampaignId int
	}
	func conditions {
	}   
    func action {
        var position_id int
         
        position_id = DBInt(Table("ge_campaigns"), "position_id", $CampaignId)
        if(position_id == 1) 
        {
            $numChoiceRes = 3
        }
        if(position_id == 2  || position_id == 3) 
        {
            $numChoiceRes = 1
        }
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
 
 func conditions {
     
    MainCondition()
    
}

func action {
    DBInsert(Table( "ge_campaigns"), "name,date_start,candidates_deadline,date_start_voting,date_stop_voting, position_id,num_votes",$ElectionName, $DateStart,$CandidatesDeadline,$Date_start_voting,$Date_stop_voting,$PositionId,0)
    
  } 
}`,
`sc_GEVoting = contract GEVoting {
	data {
		Candidate string
		ChoiceId int 
		CampaignId int
		Signature string "optional hidden"

	}

	func conditions {

	    $sha256 = Sha256(Str($CampaignId + $citizen))

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

    counter = DBIntExt( Table("ge_campaigns"), "num_votes", $CampaignId, "id")
    DBUpdate(Table( "ge_campaigns"), $CampaignId, "num_votes", counter+1)
    
    
  }
}`,
`sc_GEVotingResult = contract GEVotingResult {
 data {
    CampaignId int 
    
 }
 
 func conditions {
     
    MainCondition()
     
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
    
        SLaw_NumResultsVoting("CampaignId",$CampaignId)
    
        var list array
        var votes int
        var war map
        var i int
        var len int
        var position string
        
        position = DBString(Table("ge_campaigns"), "name", $CampaignId)
        
        list = DBGetList(Table("ge_candidates"), "id,candidate,citizen_id,position_id",0,100,"counter desc", "id_election_campaign=$", $CampaignId)
        len = Len(list)
        while i < len {
            war = list[i]
            i = i + 1
            votes = DBIntWhere( Table("ge_votes"), "count(id)", "id_candidate=$",war["id"])
            DBUpdate(Table("ge_candidates"),Int(war["id"]), "counter", votes)
            if i < $numChoiceRes+1 {
                DBUpdate(Table("ge_candidates"),Int(war["id"]), "result", 1)
                
                DBInsert(Table("ge_person_position"),"position_id,name,citizen_id,timestamp date_start,position",war["position_id"],war["candidate"],war["citizen_id"],$block_time, position)
            
            } else {
                DBUpdate(Table("ge_candidates"),Int(war["id"]), "result", 0)
            }
        }
        
        DBUpdate(Table("ge_campaigns"),$CampaignId, "status", 1)
    
  }
}`,
`sc_LegislatureConditions = contract LegislatureConditions {
    data {    }

    conditions {
        
		if DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=1", $citizen) == 0 {
			info "You do not have the right to do it"
		}

    }

    action {    }
}`,
`sc_LSSignature = contract LSSignature {
	data {
		SmartLawsId int
		Id_voting int
		Signature string "optional hidden"
	}
	func conditions {

		if DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen) == 0 {
			info "You do not have the right to sign"
		}	
		
		if DBIntWhere( Table("laws_edition"), "id", "id=$ and status < 3", $Id_voting) != 0 {
			info "Law or an amendment to the law has not yet taken"
		}
		
		if DBIntWhere( Table("laws_edition"), "id", "id=$ and status=4", $Id_voting) != 0 {
			info "The law has been signed"
		}
		
		if DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and resalt_voting=1", $SmartLawsId) == 0 {
			info "Law or amendments to the law have been rejected"
		}

	}

	func action {

    	 var id_law_list int
    	id_law_list=DBIntWhere( Table("sl_list"), "id", "id_smart_contract=$", $SmartLawsId)
    	DBUpdate(Table("sl_list"), id_law_list, "timestamp date_last_edition", $block_time)
        DBUpdate(Table("laws_edition"), $Id_voting, "status,timestamp date_end", 4, $block_time)
        var value string
		var name_smart_contract string
        value=DBStringExt( Table("laws_edition"), "value",$Id_voting, "id")
		name_smart_contract=DBStringExt( Table("sl_list"), "name_smart_contract",$SmartLawsId, "id_smart_contract")
        UpdateContract(name_smart_contract, value, "true")
    
	}
}`,
`sc_SLAddLaw = contract SLAddLaw {
    data {
        SLawName string
        SLawTxtName string
    }

    conditions {

	    if DBIntExt(Table("sl_list"), "id",$SLawName ,"name_smart_contract") != 0 {
			warning "Vacancy is already open"
		}
	    
	    $smart_contract_id = DBIntExt(Table("smart_contracts"), "id", $SLawName, "name")
	    if $smart_contract_id == 0 {
			warning "The law with the same name does not exist"
		}

    }

    action {
        
        DBInsert(Table("sl_list"),"id_smart_contract,name_smart_contract,text_name,timestamp date_submission_draft", $smart_contract_id,$SLawName, $SLawTxtName,$block_time)
    }
}`,
`sc_SLEdit = contract SLEdit {
	data {
		LawsValue string
		TabId int
	}
	func conditions {
		
        LegislatureConditions()	
		if DBIntWhere( Table("laws_edition"), "id", "id=$ and status>0", $TabId) != 0 {
			info "Change forbidden"
		}
	}

	func action {
		DBUpdate(Table("laws_edition"), $TabId, "value", $LawsValue)
	}
}`,
`sc_SLNewVoting = contract SLNewVoting {
	data {
		SmartLawsListID int
		SmartLawsId int
	}
	func conditions {
	    
	    LegislatureConditions()

		if DBIntWhere( Table("laws_edition"), "id", "laws_id=$ and status!=4", $SmartLawsId) != 0 {
			info "This law currently under consideration"
		}
		
		$value=DBStringExt( Table("smart_contracts"), "value", $SmartLawsId, "id")
	}

	func action {
	var name string
	name = DBString(Table("sl_list"), "text_name", $SmartLawsListID)
	
	DBInsert(Table("laws_edition"), "name,laws_id,value,status,resalt_voting,resalt_for,result_against,timestamp date_start",name, $SmartLawsId, $value,0,0,0,0, $block_time)

	}
}`,
`sc_SLStartVoting = contract SLStartVoting {
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
		DBUpdate(Table("laws_edition"), $TabId, "status,timestamp date_voting_start", 1, $block_time)
	}
}`,
`sc_SLVoting = contract SLVoting {
	data {
		SmartLawsId int 
		Vote int
		Id_voting int
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
    
    DBInsert(Table("sl_votes"),"citizen_id,citizen_name,laws_id,choice,id_voting,timestamp time",$citizen,citizen_name,$SmartLawsId,$Vote,$Id_voting, $block_time)
    
    
  }
}`,
`sc_SLVotingResult = contract SLVotingResult {
	data {
		SmartLawsId int
		TabId int
	}

	func conditions {
	    
		if DBIntWhere( Table("ge_person_position"), "id", "citizen_id=$ and position_id=3", $citizen) == 0 {
			info "You do not have the right to stop vote"
		}

		if DBIntExt(Table("laws_edition"),"resalt_voting",$TabId,"id") == 1  {
			info "Resalt is ready"
		}

	}

	func action {
		var votes0 int
		var votes1 int
		
		
		votes0 = DBIntWhere(Table("sl_votes"),"count(id)","id_voting=$ and choice=0",$TabId)
		votes1 = DBIntWhere(Table("sl_votes"),"count(id)","id_voting=$ and choice=1",$TabId)
		if (votes1 > votes0) {
		    
			DBUpdate(Table("laws_edition"),$TabId,"resalt_voting,status,timestamp date_voting_stop,result_against,resalt_for",1,3,$block_time,votes0,votes1)
			
		} else {
			
			DBUpdate(Table("laws_edition"),$TabId,"resalt_voting,status,timestamp date_voting_stop,result_against,resalt_for",0,4,$block_time,votes0,votes1)
		}
	

		
	DBInsert(Table("sl_result"),"id_laws_edition,choise,value,percents",$TabId,"For",votes1,100*votes1/(votes1+votes0))
	DBInsert(Table("sl_result"),"id_laws_edition,choise,value,percents",$TabId,"Against",votes0,100*votes0/(votes1+votes0))

	}
}`)
TextHidden( sc_GECandidateRegistration, sc_GENewElectionCampaign, sc_GEVoting, sc_GEVotingResult, sc_LegislatureConditions, sc_LSSignature, sc_SLAddLaw, sc_SLaw_NumResultsVoting, sc_SLEdit, sc_SLNewVoting, sc_SLStartVoting, sc_SLVoting, sc_SLVotingResult)
SetVar(`p_GECampaigns #= Title: Election Campaigns
Navigation(LiTemplate(GEElections, Elections), Campaigns)
Divs(md-12, panel panel-default)
    Divs(panel-heading)
        Divs(panel-title table-responsive)

Table {
	Table: #state_id#_ge_campaigns
	Order: id
	//Where: date_stop_voting > now()
	Columns: [
		[Position, #name#],
		[Start, P(h6,Date(#date_start#, YYYY.MM.DD))],
		[Deadline
			for candidates, P(h6,Date(#candidates_deadline#, YYYY.MM.DD))],
	    [Candidate Registration,If(And(#CmpTime(#date_start#, Now(datetime)) == -1, #CmpTime(#candidates_deadline#, Now(datetime)) == 1), BtnPage(GECandidateRegistration,Go,"CampaignName:'#name#',CampaignId:#id#,PositionId:#position_id#"), "Finish")],
	     [Candidate,BtnPage(GECanditatesView, View,"CampaignId:#id#,Position:'#name#'")],
		[Start Voting, P(h6,DateTime(#date_start_voting#, YYYY.MM.DD HH:MI))],
		[Stop Voting, P(h6,DateTime(#date_stop_voting#, YYYY.MM.DD HH:MI))],
		[Voting, If(And(#CmpTime(#date_start_voting#, Now(datetime)) == -1, #CmpTime(#date_stop_voting#, Now(datetime)) == 1), BtnPage(GEVoting, Go, "CampaignId:#id#,Position:'#name#'"), #num_votes#)],
		[Result,If(#CmpTime(#date_stop_voting#, Now(datetime)) == -1, If(#status#==1,#BtnPage(GEVotingResalt, View,"CampaignId:#id#,Position:'#name#'"), BtnContract(GEVotingResult,Result,Get voting result for the election,"CampaignId:#id#")), "--")]
	]
}
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GECandidateRegistration #= Title : Candidate Registration
Navigation( Candidate )

ValueById(#state_id#_citizens, #citizen#, "name", "FirstName")
Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Candidate to #CampaignName#  
           MarkDown: <h1>#FirstName#</h1>
          
        Form()
        Divs(form-group)
            Label(Description)
            Textarea(Description, form-control input-lg)
        DivsEnd:
        
        Input(CampaignName, "hidden", text, text, #CampaignName#)
        Input(FirstName, "hidden", text, text, #FirstName#)
        Input(CampaignId, "hidden", text, text, #CampaignId#)
        Input(PositionId, "hidden", text, text, #PositionId#)
          
            TxButton{Contract: GECandidateRegistration,Name:Registration,Inputs:"CampaignName=CampaignName,FirstName=FirstName,CampaignId=CampaignId,PositionId=PositionId,Description=Description", OnSuccess: "template,GECampaigns"}
           
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
         Table: #state_id#_ge_candidates
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
         Table: #state_id#_ge_candidates
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
	Table: #state_id#_ge_elective_office
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
    MarkDown: <h4>Start #ElectionName#'s Election</h4>
        Divs(panel-title)
  
    Form()
      Divs(md-10)
    Divs(md-12,help-block)
    
    Divs(form-group)
        Label(Start Date)
        InputDate(DateStart,form-control input-lg,Now(YYYY.MM.DD 00:00))
    DivsEnd:
    Divs(form-group)
        Label(Deadline for candidates)
        InputDate(CandidatesDeadline,form-control input-lg,Now(YYYY.MM.DD 00:00,1 day))
    DivsEnd:
    Divs(form-group)
        Label(Start Voting)
        InputDate(Date_start_voting,form-control input-lg,Now(YYYY.MM.DD 08:00,2 day))
    DivsEnd:
    
    Divs(form-group)
        Label(Stop Voting)
        InputDate(Date_stop_voting,form-control input-lg,Now(YYYY.MM.DD 22:00,2 day))
    DivsEnd:    

    Input(ElectionName, "hidden", text, text, #ElectionName#)
    Input(PositionId, "hidden", text, text, #PositionId#)
    Divs(md-12,help-block)
    DivsEnd:
    TxButton{ClassBtn:btn btn-primary btn-pill-right, Contract: GENewElectionCampaign,Name:Start,Inputs:"ElectionName=ElectionName,PositionId=PositionId,DateStart=DateStart,CandidatesDeadline=CandidatesDeadline,Date_start_voting=Date_start_voting,Date_stop_voting=Date_stop_voting", OnSuccess: "template,GECampaigns"}
     DivsEnd:
    FormEnd:
   

        DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GEPersonsPosition #= Title : Representatives of the legislative body
Navigation( Representatives )


Table{
    Table: #state_id#_ge_person_position
    Where: position_id=3
      Columns: [
      [Name, #name#],
      [Рosition,#position#],
      [Election date,Date(#date_start#, YYYY / MM / DD)]]
     }

Table{
    Table: #state_id#_ge_person_position
    Where: position_id=1
      Columns: [
      [Name, #name#],
      [Рosition,#position#],
      [Election date,Date(#date_start#, YYYY / MM / DD)]]
     }

PageEnd`,
`p_GEVoting #= Title : Voting
Navigation( Voting )
Divs(md-8, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
Table{
         Table: #state_id#_ge_candidates
         Where: id_election_campaign = #CampaignId#
         Order: candidate
      Columns: [[candidate, #candidate#], [Description, #description#], [Vote,BtnContract(GEVoting,For,You vote for candidate to #campaign#<br/>  #candidate#,"ChoiceId:#id#,CampaignId:#CampaignId#,Candidate:'#candidate#'",'btn btn-primary',template,GECampaigns)]]
     }
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_GEVotingResalt #= Title : Voting Result
Navigation( Voting Result )
Divs(md-6, panel panel-default panel-body)        
        	ChartBar{Table: #state_id#_ge_candidates, FieldValue: counter, FieldLabel: candidate, Colors: "7266ba,fad732,23b7e5", Where: id_election_campaign = #CampaignId#, Order: counter DESC}
DivsEnd: 
Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
        MarkDown: P(h4,#Position#)

Table{
         Table: #state_id#_ge_candidates
         Where: id_election_campaign = #CampaignId#
         Order: #counter# DESC
      Columns: [[Candidate, #candidate#], [Description, #description#], [Votes,#counter#],[Result,#result#]]
     }
     


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
    Table: #state_id#_ge_person_position
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
    Table: #state_id#_ge_person_position
    Where: position_id=1
      Columns: [
      [Name, #name#],
      [Election date,Date(#date_start#, YYYY.MM.DD)]]
     }
     
        DivsEnd:
    DivsEnd:
DivsEnd:


        
        PageEnd:`,
`p_SLEdit #= Title : Law Edit
Navigation( Law Edit )

Divs(md-12, panel panel-default panel-body)
    MarkDown : <h4>#SmartLaws#</h4>    

        Form()
        Source(LawsValue, GetOne(value, #state_id#_laws_edition, "id", #TabId#))
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
   
      Table: #state_id#_sl_list
      Order:date_adoption
      Columns: [[Law, #text_name#],
      [Adoption,DateTime(#date_adoption#, YYYY.MM.DD)],[Last edition,DateTime(#date_last_edition#, YYYY.MM.DD)],
      [Amend,BtnContract(SLNewVoting,Go,Amend Law #text_name#,"SmartLawsListID:#id#, SmartLawsId:#id_smart_contract#",'btn btn-primary',template,SLVotingList)]
      ]

     }  
        DivsEnd:
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add Law")
        
        Divs(form-group)
            Label("System Name")
            Input(SLawName, "form-control m-b",text,text,"SLaw_NumResultsVoting")
        DivsEnd:
        Divs(form-group)
            Label("Text Name")
            Input(SLawTxtName, "form-control m-b")
        DivsEnd:
         
        TxButton{ Contract: SLAddLaw, Name: Add,OnSuccess: "template,SLList"}
    FormEnd:
DivsEnd:


PageEnd:`,
`p_SLRepresentativeVoting #= Title: Result of voting of Law "#SmartLaws#" 
Navigation( Representative Voting List )

Divs(md-4, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title)
        
Table{
         Table: #state_id#_sl_votes
          Where: id_voting = #TabId#
      Columns: [[Name, #citizen_name#],[Vote,If(#choice#<1,"For","Against")]]
     }   
        DivsEnd:
    DivsEnd:
DivsEnd:



Divs(md-6, panel panel-default panel-body)
ChartPie{Table: #state_id#_sl_result, FieldValue: percents, FieldLabel: choise, Colors: "5d9cec,fad732,37bc9b,f05050,23b7e5,ff902b,f05050,131e26,37bc9b,f532e5,7266ba,3a3f51,fad732,232735,3a3f51,dde6e9,e4eaec,edf1f2", Where: id_laws_edition = #TabId#, Order: choise DESC}
DivsEnd:


PageEnd:`,
`p_SLView #= Title : Law View
Navigation( Law View )

Divs(md-12, panel panel-default panel-body)
MarkDown : <h4>#SmartLaws#</h4>   

    Form()
    Source(LawsValue, GetOne(value, #state_id#_laws_edition, "id", #TabId#))

    FormEnd: 
        
DivsEnd:

PageEnd:`,
`p_SLVotingList #= Title: Laws Voting List
Navigation(LiTemplate(Legislature, Legislature), Laws Voting List )
Divs(md-12, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)

Table{
         Table: #state_id#_laws_edition
         Order: date_voting_stop DESC
      Columns: [[Law, #name#]
      [Text,If(#status#==0, BtnPage(SLEdit,Edit,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"),BtnPage(SLView,View,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"))],
      [Voting Start,If(#status#==0, BtnContract(SLStartVoting,Start,Start voting #name#,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"),P(h6,DateTime(#date_voting_start#, YYYY.MM.DD HH:MI)))],
      [Voting “For”,If(#status#==1, BtnContract(SLVoting, For, Voting “For” #name#,"SmartLawsId:#laws_id#,Vote:1,Id_voting:#id#"), #resalt_for#)],
      [Voting “Against”,If(#status#==1, BtnContract(SLVoting,Against,Voting “Against” #name#,"SmartLawsId:#laws_id#,Vote:0,Id_voting:#id#"), #result_against#)],
     [Voting Stop, If(#status#==1, BtnContract(SLVotingResult, Stop, Stop voting #name#, "SmartLawsId:#laws_id#,TabId:#id#"), P(h6,DateTime(#date_voting_stop#, YYYY.MM.DD HH:MI)))],
      [Voting List,If(#status#>2, BtnPage(SLRepresentativeVoting,View,"SmartLaws:'#name#',SmartLawsId:#laws_id#,TabId:#id#"), " ")],
      [Voting results, If(#status#>2,If(#resalt_voting#==1,"For","Against")," ")],
      [Signature,If(And(#status#==3, #resalt_voting#==1), BtnContract(LSSignature,Signature,Signature #name#,"SmartLawsId:#laws_id#,Id_voting:#id#"),P(h6,DateTime(#date_end#, YYYY.MM.DD HH:MI)))]]
     }


DivsEnd:
DivsEnd:
DivsEnd:

PageEnd:`)
TextHidden( p_GECampaigns, p_GECandidateRegistration, p_GECandidates, p_GECanditatesView, p_GEElections, p_GENewCampaign, p_GEPersonsPosition, p_GEVoting, p_GEVotingResalt, p_Legislature, p_SLEdit, p_SLList, p_SLRepresentativeVoting, p_SLView, p_SLVotingList)
SetVar(`m_Legislature = MenuItem(Citizen dashboard, dashboard_default)
MenuItem(Legislature dashboard, Legislature)
MenuItem(Smart Laws, SLList)
MenuItem(Laws Voting List, SLVotingList)
MenuItem(Election, GEElections)`)
TextHidden( m_Legislature)
SetVar()
TextHidden( )
SetVar(`d_Export0_ge_elective_office #= contract Export0_ge_elective_office {
func action {
	var tblname, fields string
	tblname = Table("ge_elective_office")
	fields = "last_election,name"
	DBInsert(tblname, fields, "2015-12-01T00:00:00Z","Representative")
	DBInsert(tblname, fields, "2016-12-01T00:00:00Z","President")
	DBInsert(tblname, fields, "2016-12-01T00:00:00Z","Speaker")
	}
}`)
TextHidden( d_Export0_ge_elective_office)
SetVar(`am_menu_default #= MenuItem(Elections, GECampaigns)`,
`am_government #= MenuItem(Legislature, Legislature)`)
TextHidden( am_menu_default, am_government)
Json(`Head: "Legislature and Viting",
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
			table_name : "ge_campaigns",
			columns: '[["date_stop_voting", "time", "1"],["date_start_voting", "time", "1"],["candidates_deadline", "time", "1"],["name", "text", "0"],["status", "int64", "1"],["num_votes", "int64", "0"],["date_start", "time", "1"],["position_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_ge_campaigns",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"GENewElectionCampaign\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_start",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "position_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_stop_voting",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_start_voting",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "candidates_deadline",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "status",
			permissions: "ContractAccess(\"GEVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "num_votes",
			permissions: "ContractAccess(\"GEVoting\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_candidates",
			columns: '[["counter", "int64", "1"],["campaign", "text", "0"],["citizen_id", "int64", "1"],["id_election_campaign", "int64", "1"],["result", "int64", "1"],["candidate", "text", "0"],["description", "text", "0"],["position_id", "int64", "1"],["application_date", "time", "0"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_ge_candidates",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"GECandidateRegistration\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "counter",
			permissions: "ContractAccess(\"GEVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "campaign",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "description",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "position_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "application_date",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "id_election_campaign",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "result",
			permissions: "ContractAccess(\"GEVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "candidate",
			permissions: "true",
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
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_ge_elective_office",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_elective_office",
			column_name: "name",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_elective_office",
			column_name: "last_election",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_person_position",
			columns: '[["name", "text", "0"],["date_end", "time", "1"],["position", "text", "0"],["citizen_id", "int64", "1"],["date_start", "time", "1"],["position_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_ge_person_position",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"GEVotingResult\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "date_end",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "position",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "date_start",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "position_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "name",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_votes",
			columns: '[["time", "time", "1"],["strhash", "hash", "1"],["id_candidate", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_ge_votes",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"GEVoting\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_votes",
			column_name: "strhash",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_votes",
			column_name: "id_candidate",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_votes",
			column_name: "time",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "laws_edition",
			columns: '[["conditions", "text", "0"],["date_start", "time", "1"],["resalt_voting", "int64", "1"],["result_against", "int64", "1"],["date_voting_start", "time", "1"],["name", "text", "0"],["law_id", "int64", "1"],["laws_id", "int64", "1"],["date_end", "time", "1"],["resalt_for", "int64", "1"],["date_voting_stop", "time", "1"],["value", "text", "0"],["status", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_laws_edition",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"SLNewVoting\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "result_against",
			permissions: "ContractAccess(\"SLVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "date_voting_stop",
			permissions: "ContractAccess(\"SLVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "conditions",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "date_start",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "resalt_voting",
			permissions: "ContractAccess(\"SLVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "status",
			permissions: "ContractAccess(\"SLVotingResult\", \"SLStartVoting\",\"LSSignature\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "laws_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "date_end",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "resalt_for",
			permissions: "ContractAccess(\"SLVotingResult\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "date_voting_start",
			permissions: "ContractAccess(\"SLStartVoting\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "value",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_laws_edition",
			column_name: "law_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_list",
			columns: '[["name_smart_contract", "hash", "1"],["date_submission_draft", "time", "1"],["text_name", "text", "0"],["date_adoption", "time", "1"],["date_last_edition", "time", "1"],["id_smart_contract", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_sl_list",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"SLAddLaw\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "id_smart_contract",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "name_smart_contract",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "date_submission_draft",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "text_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "date_adoption",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_list",
			column_name: "date_last_edition",
			permissions: "ContractAccess(\"LSSignature\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_result",
			columns: '[["id_laws_edition", "int64", "1"],["value", "int64", "1"],["choise", "text", "0"],["percents", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_sl_result",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"SLVotingResult\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_result",
			column_name: "value",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_result",
			column_name: "choise",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_result",
			column_name: "percents",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_result",
			column_name: "id_laws_edition",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "sl_votes",
			columns: '[["id_voting", "int64", "1"],["citizen_id", "int64", "1"],["citizen_name", "text", "0"],["time", "time", "0"],["choice", "int64", "1"],["laws_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_sl_votes",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"SLVoting\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "id_voting",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "citizen_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "time",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "choice",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_sl_votes",
			column_name: "laws_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SLaw_NumResultsVoting",
			value: $("#sc_SLaw_NumResultsVoting").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLaw_NumResultsVoting"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "LegislatureConditions",
			value: $("#sc_LegislatureConditions").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "LegislatureConditions"
			}
	   },


{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "Export0_ge_elective_office",
			value: $("#d_Export0_ge_elective_office").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "Export0_ge_elective_office"
			}
	   },
{
				Forsign: '',
				Data: {
					type: "Contract",
					global: 0,
					name: "Export0_ge_elective_office"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			name: "GEVotingResult",
			value: $("#sc_GEVotingResult").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GEVotingResult"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			name: "SLAddLaw",
			value: $("#sc_SLAddLaw").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLAddLaw"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
			name: "SLStartVoting",
			value: $("#sc_SLStartVoting").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
			name: "SLVoting",
			value: $("#sc_SLVoting").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
			name: "SLVotingResult",
			value: $("#sc_SLVotingResult").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SLVotingResult"
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
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECampaigns",
			menu: "menu_default",
			value: $("#p_GECampaigns").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECandidateRegistration",
			menu: "menu_default",
			value: $("#p_GECandidateRegistration").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GECanditatesView",
			menu: "menu_default",
			value: $("#p_GECanditatesView").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEElections",
			menu: "Legislature",
			value: $("#p_GEElections").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GENewCampaign",
			menu: "Legislature",
			value: $("#p_GENewCampaign").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEVoting",
			menu: "menu_default",
			value: $("#p_GEVoting").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "GEVotingResalt",
			menu: "menu_default",
			value: $("#p_GEVotingResalt").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
		}]`
)
