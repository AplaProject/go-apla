SetVar(
	global = 0,
	typeid = TxId(EditContract),
	typecolid = TxId(NewColumn),
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
SetVar(`sc_AddAccount #= contract AddAccount {
	data {
		Id int
		TypeAccount string
	}
	func conditions {
	    
    CentralBankConditions()

	}
	
	func action {
	    
		DBInsert(Table("accounts"), $TypeAccount+",onhold", $Id, 0)
	}
}`,
`sc_AddCitizenAccount #= contract AddCitizenAccount {
	data {
		CitizenId string
	}
	func conditions {
	    
	    $citizen_id = AddressToId($CitizenId)
		if $citizen_id == 0 {
			warning "not valid citizen id"
		}
		
		if DBIntExt(Table("accounts"), "id",  $citizen_id, "citizen_id") {
			warning "The account has already been created"
		}
		
	}
	func action {
		
		AddAccount("Id,TypeAccount",$citizen_id, "citizen_id")
	}
}`,
`sc_addMessage #= contract addMessage {
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
		DBInsert( Table("messages"), "text, ava, username, flag, citizen_id, state_id", $Text, avatar, name, flag, $citizen, $state) 
	}		
}`,
`sc_CentralBankConditions #= contract CentralBankConditions {
	data {	}
	
	func conditions	{
	    
	    MainCondition()
	    
	}

	func action {	}
}`,
`sc_CitizenCondition #= contract CitizenCondition {
    data {   }

    conditions {
        
            if !DBInt(Table("citizens"), "id", $citizen) 
            {
                warning "Sorry, you don't have access to this action"
            }
            
             if DBIntExt(Table("citizen_del"), "status", $citizen,"citizen_id") == 1
            {
                warning "You are deprived of the rights."
            }

    }

    action {    }
}`,
`sc_CitizenDel #= contract CitizenDel {
	data {
		CitizenId int
	}

	func conditions {

        MainCondition()

		if DBIntWhere(Table("citizen_del"), "id", "citizen_id=$ and status=1", $CitizenId) > 0 {
			info "This user has deleted"
		}

	}

	func action {
	    DBInsert(Table("citizen_del"),"citizen_id,status", $CitizenId,1)

	}
}`,
`sc_DelMessage #= contract DelMessage {
    data {
        MessageId int
    }

    conditions {
        
        MainCondition()
    }

    action {
        
        DBUpdate(Table("messages"), $MessageId, "delete", 1)

    }
}`,
`sc_DisableAccount #= contract DisableAccount {
	data {
		AccountId int
	}
	
		func conditions {

	    	 CentralBankConditions()
	}

	func action {
		DBUpdate(Table("accounts"), $AccountId, "onhold", "1")
	}
}`,
`sc_EditProfile #= contract EditProfile {
                        	data {
                        		FirstName  string
                        		Image string "image"
                        	}
                        	func action {
                        	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
                          	  //Println("TXEditProfile new")
                        	}
                        }`,
`sc_GECandidateRegistration #= contract GECandidateRegistration {
	data {
		CampaignName string
		FirstName string "choice"
		Description string
		CampaignId int
		PositionId int

	}

	func conditions {
	    
	    CitizenCondition()
	    
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
`sc_GenCitizen #= contract GenCitizen {
          	data {
          		Name      string
           		PublicKey string
          	}
          	conditions {
          	    if StateVal("gov_account") != $citizen {
          	        error "Access denied"
          	    }
          	    $idc = PubToID($PublicKey)
          	    if $idc == 0 || DBIntExt("dlt_wallets", "wallet_id", $idc, "wallet_id") == $idc {
          	        warning "The key is already in use"
          	    }
          	}
          	action {
          		DBInsert("dlt_wallets", "wallet_id,public_key_0,address_vote", $idc, HexToBytes($PublicKey), IdToAddress($idc))
          		DBInsert(Table( "citizens"), "id,block_id,name", $idc, $block, $Name )
          	}
          }`,
`sc_GENewElectionCampaign #= contract GENewElectionCampaign {
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
`sc_GEVoting #= contract GEVoting {
	data {
		Candidate string
		ChoiceId int 
		CampaignId int
		Signature string "optional hidden"

	}

	func conditions {
	    
	    CitizenCondition()

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
`sc_GEVotingResult #= contract GEVotingResult {
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
    
        SmartLaw_NumResultsVoting("CampaignId",$CampaignId)
    
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
`sc_GV_NewPosition #= contract GV_NewPosition {
	data {
		position_name string
		position_type int
	}
	func conditions {

		MainCondition()

    		if DBIntExt(Table("positions_list"), "id", $position_name, "position_name") > 0 {
    			warning "The Position with the same name already exists"
    		}

    		if DBIntExt(Table("ge_elective_office"), "id", $position_name, "name") > 0 {
    			warning "The Position with the same name already exists"
    		}
    }
func action {

	if $position_type == 1 {
		DBInsert(Table("positions_list"), "position_name", $position_name)
	} else {
		DBInsert(Table("ge_elective_office"), "name", $position_name)
	}

}
}`,
`sc_GV_PositionDismiss #= contract GV_PositionDismiss {
    data {
        PositionId int
    }

    conditions {
        
        MainCondition()

    }

    action {
        
        DBUpdate(Table("positions_citizens"), $PositionId, "dismiss", "1")

    }
}`,
`sc_GV_Positions_Citizens #= contract GV_Positions_Citizens {
	data {

		Citizen_id string
		position_id int
	}
	func conditions {

		ContractConditions("MainCondition")
		if DBIntExt(Table("positions_citizens"), "id", $position_id, "position_id")
		{
		     warning "The position is already occupied."
		}

	}
	func action {

		var citizen_id int
		citizen_id = AddressToId($Citizen_id)

		var position_name string
		position_name = DBString(Table("positions_list"), "position_name", $position_id)

		var citizen_name string
		citizen_name = DBString(Table("citizens"), "name", citizen_id)

		DBInsert(Table("positions_citizens"), "citizen_id, position_id, citizen_name, position_name, timestamp date,dismiss", citizen_id, $position_id, citizen_name, position_name, $block_time,0)

	}
}`,
`sc_MoneyTransfer #= contract MoneyTransfer {
	data {
		Amount money
		SenderAccountId int
		RecipientAccountId int
	}

	func conditions {
	    
	    	if $SenderAccountId!=0 
	    	{
	    	    
	    	    if DBAmount(Table("accounts"), "id", $SenderAccountId) < $Amount {
			        warning "Not enough money"
	    	    }
	    	    
	    	}else{
	    	    
	    	    CentralBankConditions()
	    	}
	    
	}
	func action {
        if $SenderAccountId>0
        {
            var sender_amount money
            sender_amount = DBIntExt(Table("accounts"), "amount", $SenderAccountId, "id")
            sender_amount = sender_amount - $Amount
            DBUpdate(Table("accounts"), $SenderAccountId, "amount",  sender_amount)
            
        }
            var recipient_amount money
            recipient_amount = DBIntExt(Table("accounts"), "amount", $RecipientAccountId, "id")
            recipient_amount = recipient_amount + $Amount
            DBUpdate(Table("accounts"), $RecipientAccountId, "amount", recipient_amount)

	}
}`,
`sc_RechargeAccount #= contract RechargeAccount {
	data {
		AccountId int
		Amount money
	}
	
	func conditions	{
	    
	    CentralBankConditions()
	}

	func action {
	    
		MoneyTransfer("SenderAccountId,RecipientAccountId,Amount",0,$AccountId,$Amount)
	}
}`,
`sc_RF_NewIssue #= contract RF_NewIssue {
 data {
    Issue string
    Type int
    Date_start_voting string "date"
    Date_stop_voting string "date"
 }
 
func conditions {
    
	   ContractConditions("MainCondition")

	} 

func action {
    DBInsert(Table( "rf_referendums"), "issue,type,date_voting_start,date_voting_finish,status,number_votes,timestamp date_enter",$Issue,$Type,$Date_start_voting,$Date_stop_voting,0,0,$block_time)
    
  } 
}`,
`sc_RF_SaveAns #= contract RF_SaveAns {
	data {
		ReferendumId int
		Answer string 
	}

	func conditions {
	    

	    $sha256 = Sha256(Str($ReferendumId + $citizen))

		//var voted int
		//voted = DBIntExt(Table("rf_votes"), "id", $sha256, "strhash")

		//if voted != 0 {
		//	info "You already voted"
		//}

		var allowed int
		allowed = DBIntWhere(Table("rf_referendums"), "id", "date_voting_start < now() and date_voting_finish > now() and id=$", $ReferendumId)

		if allowed == 0 {
			info "Voting is not available now"
		}

	}
	func action {
    
    var id_voted int
	id_voted = DBIntWhere(Table("rf_votes"), "id", "referendum_id=$ and citizen_id=$", $ReferendumId,$citizen)
	
	if id_voted > 0
	{

        DBUpdate(Table("rf_votes"), id_voted, "answer", $Answer)
        
	}else{
	   
	   DBInsert(Table("rf_votes"),"referendum_id,strhash,answer,citizen_id,timestamp time",$ReferendumId,$sha256,$Answer, $citizen,$block_time)
        
        var counter int
        counter = DBIntExt( Table("rf_referendums"), "number_votes", $ReferendumId, "id")
        DBUpdate(Table("rf_referendums"), $ReferendumId, "number_votes", counter+1)
         
	}
    
  }
}`,
`sc_RF_Voting #= contract RF_Voting {
	data {
		ReferendumId int
		RFChoice int 
	}

	func conditions {
	    

	    $sha256 = Sha256(Str($ReferendumId + $citizen))

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
    
    var id_voted int
	id_voted = DBIntWhere(Table("rf_votes"), "id", "referendum_id=$ and citizen_id=$", $ReferendumId,$citizen)
	
	if id_voted > 0
	{

        DBUpdate(Table("rf_votes"), id_voted, "choice", $RFChoice)
        
	}else{
	   
	   DBInsert(Table("rf_votes"),"referendum_id,strhash,choice,citizen_id,timestamp time",$ReferendumId,$sha256,$RFChoice,$citizen,$block_time)
        
        var counter int
        counter = DBIntExt( Table("rf_referendums"), "number_votes", $ReferendumId, "id")
        DBUpdate(Table("rf_referendums"), $ReferendumId, "number_votes", counter+1)
         
	}
    
  }
}`,
`sc_RF_VotingCancel #= contract RF_VotingCancel {
	data {
		ReferendumId int
	}

	func conditions {

        ContractConditions("MainCondition")


		if DBIntWhere(Table("rf_referendums"), "id", "date_voting_start > now() and id=$ and status=2", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate(Table("rf_referendums"),$ReferendumId,"status",2)

	}
}`,
`sc_RF_VotingDel #= contract RF_VotingDel {
	data {
		ReferendumId int
	}

	func conditions {

        ContractConditions("MainCondition")

		if DBIntWhere(Table("rf_referendums"), "id", "date_voting_finish > now() and id=$", $ReferendumId) > 0 
		{
			info "action is not available"
		}

	}

	func action {
	    DBUpdate(Table("rf_referendums"),$ReferendumId,"status",2)

	}
}`,
`sc_RF_VotingResult #= contract RF_VotingResult {
	data {
		ReferendumId int
	}
	func conditions {

		var x int

		x = DBIntExt(Table("state_parameters"), "value", "gov_account", "name")
		if x != $citizen {
			info "You do not have the right to stop vote"
		}

		//x = DBIntExt(Table("rf_referendums"),"status",$ReferendumId,"id")
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
		
		votes0 = DBIntWhere(Table("rf_votes"),"count(id)","referendum_id=$ and choice=0",$ReferendumId)
        votes1 = DBIntWhere(Table("rf_votes"),"count(id)","referendum_id=$ and choice=1",$ReferendumId)
		
    	if(votes0==0 && votes1==0)
    	{
    	    DBUpdate(Table("rf_referendums"),$ReferendumId,"result,number_0,number_1",0,votes0,votes1)
    	
    	}else{
    		
    		if (votes1 > votes0) {
    			resalt = 1
    		} else {
    			resalt = 0
    		}
    		
    		DBUpdate(Table("rf_referendums"),$ReferendumId,"result,number_0,number_1",resalt,votes0,votes1)
    		
    		id_res0 = DBIntWhere(Table("rf_result"),"id","referendum_id=$ and choice=0",$ReferendumId)
    		
    		if id_res0 > 0 
    		{
    		    id_res1 = DBIntWhere(Table("rf_result"),"id","referendum_id=$ and choice=1",$ReferendumId)
    	    	DBUpdate(Table("rf_result"),id_res0,"value,percents",votes0,100*votes0/(votes1+votes0))
    	    	DBUpdate(Table("rf_result"),id_res1,"value,percents",votes1,100*votes1/(votes1+votes0))
    		
    		}else{
            		
            	DBInsert(Table("rf_result"),"referendum_id,choice,choice_str,value,percents",$ReferendumId,1,"Yes",votes1,100*votes1/(votes1+votes0))
            	DBInsert(Table("rf_result"),"referendum_id,choice,choice_str,value,percents",$ReferendumId,0,"No",votes0,100*votes0/(votes1+votes0))
    		}
    		
    		var finish int
		    finish = DBIntWhere(Table("rf_referendums"), "id", "date_voting_finish < now() and id=$", $ReferendumId)
		    if finish > 0
		    {
		        	DBUpdate(Table("rf_referendums"),$ReferendumId,"status",1)
		    }
    		
    	}	

	}
}`,
`sc_RF_VotingStart #= contract RF_VotingStart {
	data {
		ReferendumId int
	}

	func conditions {

        ContractConditions("MainCondition")


		if DBIntWhere(Table("rf_referendums"), "id", "date_voting_start < now() and id=$", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate(Table("rf_referendums"),$ReferendumId,"timestamp date_voting_start",$block_time)

	}
}`,
`sc_RF_VotingStop #= contract RF_VotingStop {
	data {
		ReferendumId int
	}

	func conditions {

        ContractConditions("MainCondition")

		if DBIntWhere(Table("rf_referendums"), "id", "date_voting_finish < now() and id=$", $ReferendumId) > 0 {
			info "action is not available"
		}

	}

	func action {
	
	DBUpdate(Table("rf_referendums"),$ReferendumId,"timestamp date_voting_finish,status",$block_time,0)

	}
}`,
`sc_SearchCitizen #= contract SearchCitizen {
	data {
		Name   string
	
	}
	func conditions {

	}
	func action {
	
	}
}`,
`sc_SendMoney #= contract SendMoney {
	data {

		RecipientAccountId int 
		Amount money
	}

	func conditions {
	    
	    DBInt(Table("accounts"), "citizen_id", $RecipientAccountId)
	    
	    if !DBInt(Table("accounts"), "citizen_id", $RecipientAccountId)
	    {
	        warning("The wrong account number")
	    }
	    
	    $sender_id = DBIntExt(Table("accounts"), "id", $citizen, "citizen_id")
	    if $sender_id==$RecipientAccountId
	    {
	        warning("You can not send money to your own account")
	    }

	}
	func action {

		MoneyTransfer("SenderAccountId,RecipientAccountId,Amount",$sender_id,$RecipientAccountId,$Amount)
	}
}`,
`sc_SmartLaw_NumResultsVoting #= contract SmartLaw_NumResultsVoting {
	data {
	    CampaignId int
	}
	func conditions {
	    
	}   
    func action {
        
        var position_id int
        position_id = DBInt(Table("ge_campaigns"), "position_id", $CampaignId)
        
        if(position_id == 1) {$numChoiceRes = 1}
        if(position_id == 2) {$numChoiceRes = 1}
        if(position_id == 3) {$numChoiceRes = 1}
        if(position_id == 4) {$numChoiceRes = 1}
        if(position_id > 4)  {$numChoiceRes = 1}
        
	}
}`,
`sc_TXCitizenRequest #= contract TXCitizenRequest {
	data {
		StateId    int    "hidden"
		FullName   string	
	}
	conditions {
		if Balance($wallet) < StateParam($StateId, "citizenship_price") {
			error "not enough money"
		}
	}
	action {
		DBInsert(TableTx( "citizenship_requests"), "dlt_wallet_id,name,block_id", 
		    $wallet, $FullName, $block)
	}
}`,
`sc_TXEditProfile #= contract TXEditProfile {
	data {
		FirstName  string
		Image string "image"
	}
	action {
	  DBUpdate(Table( "citizens"), $citizen, "name,avatar", $FirstName, $Image)
  	  //Println("TXEditProfile new")
	}
}`,
`sc_TXNewCitizen #= contract TXNewCitizen {
	data {
        RequestId int
    }
 	conditions {
		if Balance(DBInt(Table( "citizenship_requests"), "dlt_wallet_id", $RequestId )) < Money(StateParam($state, "citizenship_price")) {
			error "not enough money"
		}
	}
	action {
		var wallet int
		var towallet int
		wallet = DBInt(Table( "citizenship_requests"), "dlt_wallet_id", $RequestId )
		towallet = Int(StateVal("gov_account"))
		if towallet == 0 {
			towallet = $citizen
		}
//        DBTransfer("dlt_wallets", "amount,wallet_id", wallet, towallet, Money(StateParam($state, "citizenship_price")))
		DBInsert(Table( "citizens"), "id,block_id,name", wallet, 
		          $block, DBString(Table( "citizenship_requests"), "name", $RequestId ) )
        DBUpdate(Table( "citizenship_requests"), $RequestId, "approved", 1)
	}	
}`,
`sc_TXRejectCitizen #= contract TXRejectCitizen {
   data { 
        RequestId int
   }
   action { 
	  DBUpdate(Table( "citizenship_requests"), $RequestId, "approved", -1)
   }
}`)
TextHidden( sc_AddAccount, sc_AddCitizenAccount, sc_addMessage, sc_CentralBankConditions, sc_CitizenCondition, sc_CitizenDel, sc_DelMessage, sc_DisableAccount, sc_EditProfile, sc_GECandidateRegistration, sc_GenCitizen, sc_GENewElectionCampaign, sc_GEVoting, sc_GEVotingResult, sc_GV_NewPosition, sc_GV_PositionDismiss, sc_GV_Positions_Citizens, sc_MoneyTransfer, sc_RechargeAccount, sc_RF_NewIssue, sc_RF_SaveAns, sc_RF_Voting, sc_RF_VotingCancel, sc_RF_VotingDel, sc_RF_VotingResult, sc_RF_VotingStart, sc_RF_VotingStop, sc_SearchCitizen, sc_SendMoney, sc_SmartLaw_NumResultsVoting, sc_TXCitizenRequest, sc_TXEditProfile, sc_TXNewCitizen, sc_TXRejectCitizen)
SetVar(`p_citizen_profile #= Title:Profile
Navigation(LiTemplate(Citizen),Editing profile)
PageTitle: Editing profile
ValueById(#state_id#_citizens, #citizen#, "name,avatar", "FirstName,Image")
TxForm{ Contract: TXEditProfile, OnSuccess: MenuReload()}
PageEnd:`,
`p_CitizenInfo #= Title: Citizen info
Navigation(LiTemplate(government),Citizen info)

SetVar(state_name=GetOne(value,#gstate_id#_state_parameters,name='state_name'))
GetRow("user", #gstate_id#_citizens, "id", #citizenId#)

Divs(md-12, panel widget)
       Divs: half-float
        Divs: no-map h-300
        DivsEnd:
        SetVar(hmap=300)
        Map(StateVal(state_coords), StateOnTheMapCitizen)
        Divs: half-float-bottom
            Image(If(GetVar(user_avatar),#user_avatar#,"/static/img/apps/ava.png"), Image, img-thumbnail img-circle thumb-full)
        DivsEnd:
    DivsEnd:
    Divs: panel-body text-center
        Tag(h3, #user_name#, m0)
        Divs: list-comma align-center
            GetList(pos, #gstate_id#_positions_citizens, "position_name,citizen_id", "citizen_id =  #citizenId#" and dismiss = 0)
            ForList(pos)
                P(text-muted, #position_name#)
            ForListEnd:
        DivsEnd:
        Divs: list-comma align-center
            GetList(pos, #gstate_id#_ge_person_position, "position,citizen_id", "citizen_id =  #citizenId#")
            ForList(pos)
                P(text-muted, #position#)
            ForListEnd:
        DivsEnd:
    DivsEnd:
    DivsEnd:
    
Divs(md-12, panel widget data-sweet-alert)    

    Divs: panel-body text-center bg-gray-darker
        Divs: row
            Divs: col-md-6 mt-sm
                LinkPage(StateInfo, Image(GetOne(value,#gstate_id#_state_parameters,name='state_flag'), State flag, img-responsive d-inline-block align-middle w-100) Strong(d-inline-block align-middle,#state_name#), "gstate_id:#gstate_id#", profile-flag text-white h3)
            DivsEnd:
            Divs: col-md-6 mt-lg mb
                Tag(h4, Address(#user_id#) Em(clipboard fa fa-clipboard id="clipboard" aria-hidden="true" data-clipboard-action="copy" data-clipboard-text=Address(#user_id#) onClick="CopyToClipboard('#clipboard')", ), m0)
                P(text-muted m0, Citizen ID)
            DivsEnd:
        DivsEnd:

DivsEnd:
DivsEnd:

Divs(md-4, panel panel-info elastic center data-sweet-alert)
    Div(panel-heading, Div(panel-title, Send Money to #user_name#))
    Divs: panel-body
        Form()
            
            Divs(form-group)
                Label("Amount")
                InputMoney(Amount, "form-control")
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)

        Input(RecipientAccountId, "hidden", text, text, GetOne(id, #gstate_id#_accounts, "citizen_id", #user_id#))
        TxButton{ Contract: SendMoney, Name: Send, Inputs: "RecipientAccountId=RecipientAccountId, Amount=Amount", OnSuccess: "template,CitizenInfo,global:0,gstate_id:#gstate_id#,citizenId:'#citizenId#'" }
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_citizens #= Title : Citizens
Navigation( LiTemplate(government), Citizens)


SetVar(DelCitizen = BtnContract(CitizenDel, <b>$Del$</b>, Delete citezen,"CitizenId:'#id#'",'btn btn-primary',template,citizens))

GetList(del,#state_id#_citizen_del,"citizen_id,status",status=1,id)

Divs(md-8, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Search citizen")
        Divs(form-group)
            Label("Name")
            Input(Name, "form-control  m-b")
        DivsEnd:
        TxButton{ Contract: SearchCitizen, Name: Search, Inputs: "Name=Name", OnSuccess: "template,citizens,CitizenName:Val(Name),Search:1" }
    FormEnd:
DivsEnd:
Divs(md-12, panel panel-default panel-body)
GetList(account,#state_id#_accounts,"citizen_id,id",citizen_id!=0)
If(#Search#==1)
BtnPage(citizens, <b>All</b>,"Search:0",btn btn-primary)
    Table{
        Table: #state_id#_citizens
        Order: #name#
        Where: #name#='#CitizenName#'
        Columns:  [[Avatar,Image(#avatar#)], [Citizen ID, Address(#id#) Em(clipboard fa fa-clipboard id="clipboard" aria-hidden="true" data-clipboard-action="copy" data-clipboard-text=Address(#id#) onClick="CopyToClipboard('#clipboard')", )], [Account,ListVal(account,#id#,id)],[Name, If(ListVal(del,#id#,status)==1,<s>#name#</s>, LinkPage(CitizenInfo,#name#,"citizenId:'#id#',gstate_id:#state_id#",pointer) )], [$Del$,If(ListVal(del,#id#,status)==1, - ,#DelCitizen#)]]
    }
Else:

    Table{
        Table: #state_id#_citizens
        Order: #name#
        Columns: [[Avatar,Image(#avatar#)], [Citizen ID, Address(#id#) Em(clipboard fa fa-clipboard id="clipboard" aria-hidden="true" data-clipboard-action="copy" data-clipboard-text=Address(#id#) onClick="CopyToClipboard('#clipboard')", )], [Account,ListVal(account,#id#,id)],[Name, If(ListVal(del,#id#,status)==1,<s>#name#</s>, LinkPage(CitizenInfo,#name#,"citizenId:'#id#',gstate_id:#state_id#",pointer) )], [$Del$,If(ListVal(del,#id#,status)==1, - ,#DelCitizen#)]]
    }
IfEnd:
DivsEnd:
PageEnd:`,
`p_GECampaigns #= Title: Election Campaigns
Navigation( LiTemplate(dashboard_default, Dashboard), Election Campaigns)

Divs(md-12, panel panel-default panel-body data-sweet-alert)
      Legend(" ", "Elections")

Table {
	Table: #state_id#_ge_campaigns
	Order: id
	//Where: date_stop_voting > now()
	Columns: [
		[Position, #name#],
		[Start, Date(#date_start#, YYYY.MM.DD)],
		[Deadline
			for candidates, Date(#candidates_deadline#, YYYY.MM.DD)],
	    [Candidate Registration,If(And(#CmpTime(#date_start#, Now(datetime)) == -1, #CmpTime(#candidates_deadline#, Now(datetime)) == 1), BtnPage(GECandidateRegistration,Go,"CampaignName:'#name#',CampaignId:#id#,PositionId:#position_id#"),  "Finish")],
	     [Candidate,BtnPage(GECanditatesView, View,"CampaignId:#id#,Position:'#name#'")],
		[Start Voting, DateTime(#date_start_voting#, YYYY.MM.DD HH:MI)],
		[Stop Voting, DateTime(#date_stop_voting#, YYYY.MM.DD HH:MI)],
		[Voting, If(And(#CmpTime(#date_start_voting#, Now(datetime)) == -1, #CmpTime(#date_stop_voting#, Now(datetime)) == 1), BtnPage(GEVoting, Go, "CampaignId:#id#,Position:'#name#'"), #num_votes#)],
		[Result,If(#CmpTime(#date_stop_voting#, Now(datetime)) == -1, If(#status#==1,#BtnPage(GEVotingResalt, View,"CampaignId:#id#,Position:'#name#'"), BtnContract(GEVotingResult,Result,Get voting result for the election,"CampaignId:#id#")), "--")]
	]
}
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
`p_GECanditatesView #= Title : Canditates
Navigation( LiTemplate(GECampaigns, Elections), Canditates)
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
`p_gov_administration #= Title : Administration
Navigation(LiTemplate(government),Administration)

Tag(h2, Election and Assign, page-header)

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, New Election))
    Divs(panel-body)
        Table {
            Table: #state_id#_ge_elective_office
            Columns: [
                [ID, #id#],[Election's Type, #name#], 
                [Last election, Date(#last_election#, YYYY / MM / DD)],
                [Start,BtnPage(GENewCampaign,Go,"ElectionName:'#name#',PositionId:#id#")]
            ]
        }
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Assign a position to citizen))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label("Citizen ID")
                InputAddress(Citizen_id, "form-control input-lg m-b")
            DivsEnd:
            Divs(form-group)
                Label("Positions")
                Select(position_id, #state_id#_positions_list.position_name,form-control input-lg)
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: GV_Positions_Citizens, Name: Assign, OnSuccess: "template,gov_administration"}
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Add New Position))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label("Position Name")
                Input(position_name, "form-control input-lg m-b")
            DivsEnd:
            Divs(form-group)
                Label("Position Type")
                Select(position_type,type_office,form-control input-lg)
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: GV_NewPosition, Name: Add New, OnSuccess: "template,gov_administration"}
    DivsEnd:
DivsEnd:

Tag(h2, Accounts, page-header)

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Recharge Account))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label("Account ID")
                Select(AccountId, #state_id#_accounts.id, "form-control input-lg m-b")
            DivsEnd:
            Divs(form-group)
                Label("Amount")
                InputMoney(Amount, "form-control input-lg")
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: RechargeAccount, Name: Change, Inputs: "AccountId=AccountId, Amount=Amount", OnSuccess: "template,gov_administration,global:0" }
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Add citizen account))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label("Citizen ID")
                InputAddress(CitizenId, "form-control input-lg m-b")
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: AddCitizenAccount, Name: Add,Inputs: "CitizenId=CitizenId",OnSuccess: "template,gov_administration,global:0" }
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Disable account))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label("Account ID")
                Select(DAccountId, #state_id#_accounts.id, "form-control input-lg m-b")
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: DisableAccount, Name: Disable, Inputs: "AccountId=DAccountId",OnSuccess: "template,gov_administration,global:0" }
    DivsEnd:
DivsEnd:

Tag(h2, Citizens, page-header)

Divs(md-4, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Citizenship requests <span id='citizenship'></span>))
    Divs(panel-body)
        Table {
            Class: table-striped table-hover
            Table: #state_id#_citizenship_requests
            Order: id DESC
            Where: approved=0
            Columns: [
                [ID, #id#],[Name, #name#],
                [Decision, BtnContract(TXNewCitizen,Accept,Accept requests from #name#,"RequestId:#id#",'btn btn-success')],
                [ ,BtnContract(TXRejectCitizen,Reject, Reject requests from #name#,"RequestId:#id#",'btn btn-danger')]
            ]
        }
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_List #= Title : $ListVotings$
Navigation( LiTemplate(government), Pollings)

SetVar(ViewResultQues = BtnPage(RF_ViewResultQuestions, <b>$Vw$ #number_votes#</b>, "ReferendumId:#id#,Back:0,Status:#status#",btn btn-primary btn-block))
SetVar(ViewResult = BtnPage(RF_ViewResult, <b>$Vw$ #number_votes#</b>, "ReferendumId:#id#,Back:0,Status:#status#",btn btn-primary btn-block))

SetVar(Cancel = BtnContract(RF_VotingCancel, <b>$Cncl$</b>,Cancel Votin, "ReferendumId:#id#",'btn btn-primary btn-block',template,RF_List))
SetVar(Stop = BtnContract(RF_VotingStop, <b>$Stp$</b>,Stop Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,RF_List))
SetVar(Delete = BtnContract(RF_VotingDel, <b>$Del$</b>,Delete Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,RF_List))
SetVar(Start = BtnContract(RF_VotingStart, <b>$Strt$</b>,Start Voting,"ReferendumId:#id#",'btn btn-primary btn-block',template,RF_List))


Divs(md-12, panel panel-default data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title text-center)

    Table{
         Table: #state_id#_rf_referendums
         Order: #id# DESC
         Where: #status#!=2
         Class: table-responsivee
         Adaptive: 1
      Columns: [[$Iss$, #issue#],
      [Type,P(h4,If(#type#==2,Q,V))],
      [$Strt$, P(h6,DateTime(#date_voting_start#, YYYY.MM.DD HH:MI))],
	  [$Fnsh$, P(h6,DateTime(#date_voting_finish#, YYYY.MM.DD HH:MI))],
	  [$Inf$,If(#type#==2,#ViewResultQues#,If(#number_votes# == 0, #number_votes#,If(#status#==1,#ViewResult#,BtnContract(RF_VotingResult, <b>$Vw$ #number_votes#</b>, Get Result,"ReferendumId:#id#",'btn btn-primary btn-block',template,RF_ViewResult,"ReferendumId:#id#, Back:0, Status:#status#"))))],
	  [$Actn$, If(#CmpTime(#date_voting_start#, Now(datetime)) == 1,#Cancel#, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1,#Stop#, #Delete#))],
	  [$Res$,If(#CmpTime(#date_voting_start#, Now(datetime)) == 1, - , If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1, $Contin$, If(#status# == 1, If(#result#==1,$Y$,$N$), $Fnshd$)))]]
     }
     
        P(<br/>)
        BtnPage(RF_NewIssue, <b>$NewVoting$</b>,"Status:0",btn btn-oval btn-info btn-lg md5 pd5)
DivsEnd:
    DivsEnd:
DivsEnd:
PageEnd:`,
`p_RF_NewIssue #= Title : $NewVoting$
Navigation( LiTemplate(government), $NewVoting$ )

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title)
        MarkDown: <h4>$NewVoting$</h4>
          
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
        
       
        Divs(text-right)

            TxButton{ClassBtn:btn btn-primary btn-pill-right, Contract: RF_NewIssue,Name:$Save$,Inputs:"Issue=Issue,Type=Type,Date_start_voting=Date_start_voting,Date_stop_voting=Date_stop_voting", OnSuccess: "template,RF_List"}
            
            BtnPage(RF_List, $ListVotings$, "Status:1",btn btn-default btn-pill-left ml4)
        DivsEnd:     
        Div(clearfix)    
         
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_Result #= Title : $Result$
Navigation( LiTemplate(dashboard_default, Dashboard),$Result$)

SetVar(Issue = GetOne(issue, #state_id#_rf_referendums,"id",#ReferendumId#))

Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Divs(panel-heading)
        Divs(panel-title text-center)
         
           MarkDown: <h4>#Issue#</h4>
          
        Form()
        Input(ReferendumId, "hidden", text, text, #ReferendumId#)
      
        Divs(bt-block, text-center)
        TxButton{Contract: RF_VotingResult,Name: $GetResult$,Inputs:"ReferendumId=ReferendumId", OnSuccess: "template,RF_ViewResult,ReferendumId:#ReferendumId#,Back:0"}
        DivsEnd:
           
        FormEnd:   
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_UserAns #= Title : $Ans$
Navigation( LiTemplate(dashboard_default, Dashboard), $Ans$)

SetVar(Issue = GetOne(issue, #state_id#_rf_referendums, "id", #ReferendumId#))
If(#Chng#)
    SetVar(Answer = GetOne(answer, #state_id#_rf_votes, "id", #VoteId#))
Else:
    SetVar(Answer = "")
IfEnd:

Divs(md-6, panel panel-default data-sweet-alert)
    Div(panel-heading, Div(panel-title text-primary, #Issue#))
    Divs(panel-body)
        Form()
            Divs(form-group)
                Label($Ans$)
                Textarea(Answer, form-control input-lg,#Answer#)
            DivsEnd:
        FormEnd:   
    DivsEnd:
    Divs(panel-footer)
        Divs: clearfix
            Divs: pull-right
                Input(ReferendumId, "hidden", text, text, #ReferendumId#)
                BtnPage(RF_UserQuestionList, $QuestionList$, "Status:1",btn btn-default btn-pill-left ml4)
                TxButton{ClassBtn:btn btn-primary btn-pill-right, Contract: RF_SaveAns,Name:$Save$,Inputs:"ReferendumId=ReferendumId, Answer=Answer", OnSuccess: "template,RF_UserQuestionList"}
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:

PageEnd:`,
`p_RF_UserList #= Title : $ListVotings$

Navigation( LiTemplate(dashboard_default, Dashboard),$QuestionList$)

SetVar(VotingY = BtnContract(RF_Voting, $Y$, Your choice Yes,"ReferendumId:#id#,RFChoice:1",'btn btn-info',template,RF_UserList))
SetVar(VotingN = BtnContract(RF_Voting, $N$, Your choice No,"ReferendumId:#id#,RFChoice:0",'btn btn-info',template,RF_UserList))

SetVar(ViewResult = BtnPage(RF_ViewResult, <strong>$Res$</strong>, "ReferendumId:#id#,DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#', NumberVotes:#number_votes#,Back:1,Status:1",btn btn-info btn-pill-right))

SetVar(Voting = BtnPage(RF_UserVotingList, $Vote$, "ReferendumId:#id#",btn btn-primary btn-pill-right))

SetVar(ChangeVoting = BtnPage(RF_UserVotingList, $Chng$, "ReferendumId:#id#",btn btn-pill-right))

Divs(md-6, panel panel-default panel-body data-sweet-alert)

If(GetOne(id, #state_id#_rf_referendums,status!=2 and date_voting_start < now()  and type=1))

GetList(vote,#state_id#_rf_votes,"referendum_id,choice",citizen_id=#citizen#,id)

    Table{
         Table: #state_id#_rf_referendums
         Order: #date_voting_start# DESC 
         Where: #status#!=2 and #date_voting_start# < now()  and type=1
         
      Columns: [[,If(ListVal(vote,#id#,referendum_id)>0,P(h4 text-muted, <span id="tr#id#">#issue#</span>),P(h4 text-primary, #issue#)) If(#status#==1, #ViewResult#, If(ListVal(vote,#id#,referendum_id)>0," ",#VotingY# #VotingN#))],
        [,If(ListVal(vote,#id#,referendum_id)>0,If(ListVal(vote,#id#,choice)==1,P(h4 text-primary,$Y$),P(h4 text-danger,$N$)))]
        ]
     }
Else:
P(h6, No questions to polling yet.)
IfEnd:
DivsEnd:
PageEnd:`,
`p_RF_UserQuestionList #= Title : $QuestionList$
Navigation( LiTemplate(dashboard_default, Dashboard),$QuestionList$)

SetVar(ViewResult = BtnPage(RF_ViewResultQuestions, <strong>$Res$</strong>, "ReferendumId:#id#,Back:1,Status:1",btn btn-info btn-pill-right))

SetVar(Voting = BtnPage(RF_UserAns, $Ans$, "ReferendumId:#id#, Chng:0",btn btn-primary btn-pill-right))

SetVar(ChangeVoting = BtnPage(RF_UserAns, $Chng$, "ReferendumId:#id#,Chng:1,VoteId:ListVal(vote,#id#,id)",btn btn-pill-right))

GetList(vote,#state_id#_rf_votes,"referendum_id,answer,id",citizen_id=#citizen#,id)


Divs(md-6, panel panel-default panel-body data-sweet-alert)

If(GetOne(id, #state_id#_rf_referendums,status!=2 and date_voting_start < now()  and type=2))

    Table{
         Table: #state_id#_rf_referendums
         Order: #date_voting_start# DESC 
         Where: #status#!=2 and #date_voting_start# < now() and #type#=2
         
      Columns: [[,If(ListVal(vote,#id#,referendum_id)>0,P(h4 text-muted, <span id="tr#id#">#issue#</span>) P(h4 text-primary, ListVal(vote,#id#,answer)), P(h4 text-primary, #issue#)) If(#status#==1, #ViewResult#, If(#CmpTime(#date_voting_finish#, Now(datetime)) == 1,If(ListVal(vote,#id#,referendum_id)>0,#ChangeVoting#,#Voting#),))],
        ]
     }
Else:
P(h6, No questions yet.)
IfEnd:

DivsEnd:

PageEnd:`,
`p_RF_ViewResult #= Title : $Res$
Navigation( LiTemplate(dashboard_default, Dashboard),$Res$)

GetRow(vote,#state_id#_rf_referendums,"id",#ReferendumId#)


Divs(md-6, panel panel-default panel-body)
    Divs(panel-heading)
        Divs(panel-title text-center warning)
           Divs(text-primary)
          
             MarkDown: <h4>#vote_issue#</41>
           DivsEnd:
          MarkDown: <strong>DateTime(#vote_date_voting_start#, YYYY.MM.DD HH:MI) - DateTime(#vote_date_voting_finish#, YYYY.MM.DD HH:MI)</strong>
           
           MarkDown: <h4>$TotalVoted$: #vote_number_votes#</h4>
            
           Table{
    Table: #state_id#_rf_result
    Where: referendum_id=#ReferendumId#
    Order: #choice_str# DESC
      Columns: [
    [,If(#choice#==1,P(h4 text-primary,$Y$),P(h4 text-danger,$N$))],
    [,If(#choice#==1,P(h4 text-primary,#value#),P(h4 text-danger,#value#))],
    [,If(#choice#==1,P(h4 text-primary,#percents# %),P(h4 text-danger, #percents# %))]]
     } 
    
        Divs(btn-lg)
    If(#Back#==1,BtnPage(RF_UserList, <strong>$ListVotings$</strong> ,"Status:1",btn btn-pill-left btn-info), BtnPage(RF_List, <strong>$ListVotings$</strong>, "Status:0",btn btn-pill-left btn-info))
     DivsEnd:
    

        DivsEnd:
    DivsEnd:
DivsEnd:

Divs(md-6, panel panel-default)
    Divs: panel-body
        ChartPie{Table: #state_id#_rf_result, FieldValue: percents, FieldLabel: choice_str, Colors: "f05050,5d9cec,37bc9b,f05050,23b7e5,ff902b,f05050,131e26,37bc9b,f532e5,7266ba,3a3f51,fad732,232735,3a3f51,dde6e9,e4eaec,edf1f2", Where: referendum_id = #ReferendumId#, Order: choice}
    DivsEnd:
DivsEnd:


PageEnd:`,
`p_RF_ViewResultQuestions #= Title : $Res$
Navigation( LiTemplate(government),  $Res$)

SetVar(Issue = GetOne(issue, #state_id#_rf_referendums, "id", #ReferendumId#))
SetVar(Del = GetOne(id, #state_id#_citizen_del,"citizen_id", #citizen#))

Divs( md-6,panel panel-success)
    Divs: panel-body
    MarkDown: P(h4 text-primary,#Issue#)
    
    If (#Del#>0)
        MarkDown: P(h4 text-center text-danger,Uw account is opgeschort)
    Else:
    
    Divs(btn-lg)
    If(#Back#==1,BtnPage(RF_UserQuestionList, <strong>$QuestionList$</strong> ,"Status:1",btn btn-pill-left btn-info), BtnPage(RF_List, <strong>$ListVotings$</strong>, "Status:0",btn btn-pill-left btn-info))
     DivsEnd:

    Table{
         Table: #state_id#_rf_votes
         Order: #id# DESC 
         Where: #referendum_id#=#ReferendumId#
         
      Columns: [[,P(h4 text-muted, #answer#)],
        ]
     }
    IfEnd:

DivsEnd:
DivsEnd:

PageEnd:`,
`p_StateInfo #= Title: State info
Navigation(LiTemplate(government), State info)


Divs(md-4, panel panel-default elastic center)
    Divs: panel-body
        Image(GetOne("value", #gstate_id#_state_parameters, "name", "state_flag"), ALT, img-responsive)
    DivsEnd:
DivsEnd:
Divs(md-8, panel widget elastic)
    Divs: panel-body text-center
        Tag(h3, GetOne("value", #gstate_id#_state_parameters, "name", "state_name"), m0)
    DivsEnd:
    Divs: panel-body text-center bg-gray-dark
        Divs: row row-table
            Divs: col-xs-4
                Tag(h3, 01.01.2017, m0)
                P(m0 text-muted, Founded)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3, GetOne("value", #gstate_id#_state_parameters, "name", "currency_name"), m0)
                P(m0 text-muted, Currency)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3, GetOne(count(*),#gstate_id#_citizens), m0)
                P(m0 text-muted, Population)
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:


Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Recognized as the number of UN members)
    Divs: panel-body
        Ring(GetOne(id, global_states_list,gstate_id=#gstate_id#), 20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:

Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, I voted in favor of a member of UN)
    Divs: panel-body
        Ring(GetOne(num_voting, global_states_list,gstate_id=#gstate_id#),  20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:

Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Answered questions on the UN)
    Divs: panel-body
        Ring(GetOne(num_answers, global_states_list,gstate_id=#gstate_id#), 20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:


Divs(md-12)
    SetVar(hmap=400)
    Map(StateVal(state_coords))
DivsEnd:
 
PageEnd:`)
TextHidden( p_citizen_profile, p_CitizenInfo, p_citizens, p_GECampaigns, p_GECandidateRegistration, p_GECanditatesView, p_GEElections, p_GENewCampaign, p_GEVoting, p_GEVotingResalt, p_gov_administration, p_RF_List, p_RF_NewIssue, p_RF_Result, p_RF_UserAns, p_RF_UserList, p_RF_UserQuestionList, p_RF_ViewResult, p_RF_ViewResultQuestions, p_StateInfo)
SetVar()
TextHidden( )
SetVar(`pa_type_issue #= voting,question`,
`pa_type_office #= assigned,elective`)
TextHidden( pa_type_issue, pa_type_office)
SetVar()
TextHidden( )
SetVar(`ap_dashboard_default #= GetRow(my, #state_id#_citizens, "id", #citizen#)

Title : StateVal(state_name)
Navigation(Dashboard)

Divs: panel widget bg-info available_balance
    Divs: row row-table
        Divs: col-xs-3 text-center bg-info-dark pv-lg ico
            Em(glyphicons glyphicons-coins x2)
        DivsEnd:
        Divs: col-xs-9 pv-lg text
            Div(h1 m0 text-bold, Money(GetOne(amount, #state_id#_accounts, "citizen_id", #citizen#)) Span(pl, StateVal(currency_name)))
            Div(text-uppercase, Available balance)
        DivsEnd:
    DivsEnd:
DivsEnd:

Divs: panel widget bg-success available_balance
    Divs: row row-table
        Divs: col-xs-3 text-center bg-success-dark pv-lg ico
            Em(glyphicons glyphicons-credit-card x2)
        DivsEnd:
        Divs: col-xs-9 pv-lg text
            Div(h1 m0 text-bold, GetOne(id, #state_id#_accounts, "citizen_id", #citizen#))
            Div(text-uppercase, ACCOUNT NUMBER)
        DivsEnd:
    DivsEnd:
DivsEnd:

Divs(md-12, panel widget data-sweet-alert)
    Divs: half-float
        Divs: no-map h-300
        DivsEnd:
        SetVar(hmap=300)
        Map(StateVal(state_coords), StateOnTheMapCitizen)
        Divs: half-float-bottom
            Image(If(GetVar(my_avatar),#my_avatar#,"/static/img/apps/ava.png"), Image, img-thumbnail img-circle thumb-full)
        DivsEnd:
    DivsEnd:
    Divs: panel-body text-center
        Tag(h3, If(GetVar(my_name),#my_name#,Anonym), m0)
        Divs: list-comma align-center
            GetList(pos, #state_id#_positions_citizens, "position_name,citizen_id", "citizen_id =  #citizen#" and dismiss = 0)
            ForList(pos)
                P(text-muted, #position_name#)
            ForListEnd:
        DivsEnd:
        Divs: list-comma align-center
            GetList(pos, #state_id#_ge_person_position, "position,citizen_id", "citizen_id =  #citizen#")
            ForList(pos)
                P(text-muted, #position#)
            ForListEnd:
        DivsEnd:
    DivsEnd:
    Divs: panel-body text-center bg-gray-darker
        Divs: row
            Divs: col-md-6 mt-sm
                LinkPage(government, Image(StateVal(state_flag), State flag, img-responsive d-inline-block align-middle w-100) Strong(d-inline-block align-middle, StateVal(state_name)), 'id':1, profile-flag text-white h3)
                P(text-muted m0,The founder of the state LinkPage(CitizenInfo, Strong(media-box-heading text-primary, GetOne(name, #state_id#_citizens, "id=StateVal(gov_account)")), "citizenId:'StateVal(gov_account)',gstate_id:#state_id#", pointer))
            DivsEnd:
            Divs: col-md-6 mt-lg mb
                Tag(h4, Address(#my_id#) Em(clipboard fa fa-clipboard id="clipboard" aria-hidden="true" data-clipboard-action="copy" data-clipboard-text=Address(#my_id#) onClick="CopyToClipboard('#clipboard')", ), m0)
                P(text-muted m0, Citizen ID)
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:

Divs(md-4, panel panel-info elastic center data-sweet-alert)
    Div(panel-heading, Div(panel-title, Send Money))
    Divs: panel-body
        Form()
            Divs(form-group)
                Label("Account number")
                Input(AccountId, "form-control")
            DivsEnd:
            Divs(form-group)
                Label("Amount")
                InputMoney(Amount, "form-control")
            DivsEnd:
        FormEnd:
    DivsEnd:
    Divs(panel-footer)
        TxButton{ Contract: SendMoney, Name: Send, Inputs: "RecipientAccountId=AccountId, Amount=Amount", OnSuccess: "template,dashboard_default,global:0" }
    DivsEnd:
DivsEnd:

Divs(md-8, panel panel-primary elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, State Messenger))
    Divs: panel-body
         Divs: data-widget="panel-scroll" data-start="bottom" data-height="200"
            GetList(my,  #state_id#_messages, "id,username,ava,flag,text,citizen_id,stateid,delete", "delete != 1")
            SetVar(gov_account = GetOne(value, #state_id#_state_parameters, name, "gov_account"))
            ForList(my)
                Divs: list-group-item list-group-item-hover
                    Divs: media-box
                        Divs: pull-left
                            Image(#ava#, ALT, media-box-object img-circle thumb32)
                        DivsEnd:
                        Divs: media-box-body clearfix
                            Divs: pull-right
                                If (#gov_account#==#citizen#)
                                    BtnContract(DelMessage,x,Delete Message,"MessageId:#id#",'btn btn-link btn-xs',"template,dashboard_default")
                                IfEnd:
                            DivsEnd:
                            LinkPage(CitizenInfo, Strong(media-box-heading text-primary, #username#), "citizenId:'#citizen_id#',gstate_id:#state_id#", pointer)
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
                Divs(input-group-btn)
                    TxButton{ClassBtn: fa fa-paper-plane btn btn-default btn-sm bl0 radius-tl-clear radius-bl-clear,Contract: addMessage, Name:" ",Inputs:"Text=chat_message", OnSuccess: "template,dashboard_default"}
                DivsEnd:

            DivsEnd:
    DivsEnd:
DivsEnd:`,
`ap_government #= Title : Government 
Navigation(LiTemplate(dashboard_default, Citizen))

Divs(md-4, panel panel-default elastic center)
    Divs: panel-body
    SetVar(flag=StateVal(state_flag))
        If(#flag#="")
             BtnPage(sys-editStateParameters,Upload Flag,"name:'state_flag'",btn btn-primary radius-tl-clear radius-tr-clear)
        Else:
            Image(#flag#, Flag, img-responsive)
           
        IfEnd:
    DivsEnd:
DivsEnd:

Divs(md-8, panel widget elastic center)
    Divs: panel-body text-center
        Tag(h3, StateVal(state_name), m0)
        SetVar(description=StateVal(state_description))
        If(#description#)
            P(h4 text-center text-muted mb0, #description#)
        Else:
            BtnPage(sys-editStateParameters,Add Description,"name:'state_description'",btn btn-primary f0 w0 block-center mt-lg)
        IfEnd:
    DivsEnd:
    Divs: panel-body text-center bg-gray-dark f0
        Divs: row row-table
            Divs: col-xs-4
                Tag(h3, Date(GetOne(date_founded,global_states_list,gstate_id=#state_id#),DD.MM.YYYY), m0)
                P(m0 text-muted, Founded)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3,  StateVal(currency_name), m0)
                P(m0 text-muted, Currency)
            DivsEnd:
            Divs: col-xs-4
                Tag(h3, GetOne(count(*),#state_id#_citizens), m0)
                P(m0 text-muted, Population)
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:


SetVar(citizenship=GetOne(count(*), #state_id#_citizenship_requests,approved=0))
If(#citizenship#)
Divs(col-md-12, panel panel-info )
    P(h5 text-center,You have #citizenship# citizenship request. BtnPage(gov_administration,Check,"global:0",btn btn-primary btn-xs,'citizenship'))  
DivsEnd:
IfEnd:


Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Recognized as the number of UN members)
    Divs: panel-body
        Ring(GetOne(id, global_states_list,gstate_id=#state_id#), 20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:

Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, I voted in favor of a member of UN)
    Divs: panel-body
        Ring(GetOne(num_voting, global_states_list,gstate_id=#state_id#),  20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:

Divs(col-md-4, panel panel-info elastic center)
    Div(panel-heading, Answered questions on the UN)
    Divs: panel-body
        Ring(GetOne(num_answers, global_states_list,gstate_id=#state_id#), 20, 100, 3, "5d9cec", "656565", 150)
    DivsEnd:
DivsEnd:

Divs(md-12)
    SetVar(hmap=400)
    Map(StateVal(state_coords), StateMap)
DivsEnd:

Divs(md-12, panel panel-info data-sweet-alert)
    Div(panel-heading, Div(panel-title, United governments Messenger))
    Divs(panel-body data-widget=panel-scroll data-start=bottom)
         Divs: list-group
            GetList(my, global_messages, "id,username,ava,flag,text,citizen_id,stateid")
            ForList(my)
    	        Divs: list-group-item list-group-item-hover
                    Divs: media-box
                        Divs: pull-left
                            Image(#ava#, ALT, media-box-object img-circle thumb32)
                        DivsEnd:
                        Divs: media-box-body clearfix
                            Divs: flag pull-right
                        LinkPage(StateInfo,Image(#flag#, ALT, class), "gstate_id:#stateid#", pointer)
                                
                                
                            DivsEnd:
                            LinkPage(CitizenInfo, Strong(media-box-heading text-primary, #username#), "citizenId:'#citizen_id#',gstate_id:#stateid#", pointer)
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
                TxButton{ClassBtn: fa fa-paper-plane btn btn-default btn-sm bl0 radius-tl-clear radius-bl-clear,Contract: @addMessageGL, Name:" ",Inputs:"Text=chat_message", OnSuccess: "template,government"}
                DivsEnd:

            DivsEnd:
    DivsEnd:
DivsEnd:


Divs(md-6, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Appointed posts))
    Divs(panel-body)
        Table{
            Table:  #state_id#_positions_citizens
            Order: id
            Where: dismiss = 0
            Columns: [
                [Рosition, #position_name#],
                [Name,#citizen_name#],
                [Appointed date, Date(#date#, YYYY / MM / DD)],
                [Dismiss,BtnContract(GV_PositionDismiss,Dismiss,Dismiss #citizen_name#,"PositionId:#id#",'btn btn-danger')]
            ]
        }
    DivsEnd:
DivsEnd:

Divs(md-6, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Elective Posts))
    Divs(panel-body)
        Table{
            Table: #state_id#_ge_person_position
            Columns: [
                [Рosition,#position#],
                [Name, #name#],
                [Election date,Date(#date_start#, YYYY / MM / DD)]
            ]
        }
    DivsEnd:
DivsEnd:

Divs(md-6, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, States List))
    Divs(panel-body)
        Table{
            Table: global_states_list
            Order: id
            Columns: [
                [Name,#state_name#],
                [Flag, Image(#state_flag#, ALT, flag)],
                [UG,#united_governments#],
                [Votes,If(#united_governments#, BtnContract(@UG_Vote,Vote #num_votes#, Vote for #state_name#,"State_num:#id#",'btn btn-primary'),#num_votes#)]
            ]
        }
    DivsEnd:
DivsEnd:

Divs(md-6, panel panel-default elastic data-sweet-alert)
    Div(panel-heading, Div(panel-title, Polling))
    Divs(panel-body)
        Divs(table-responsive)
            GetList(vote,global_glrf_votes,"referendum_id,choice",state_id=#state_id#,id)
            Table {
                Table: global_glrf_referendums
                Order: #date_voting_start# DESC 
                Where: #status#!=2 and #date_voting_start# < now()  and type=1
                Columns: [
                    [Questions,If(ListVal(vote,#id#,referendum_id)>0,Tag(h4, <span id="tr#id#">#issue#</span>, panel-title text-muted d-inline-block mb0 wd-auto),Tag(h4, #issue#, panel-title text-primary d-inline-block mb0 wd-auto))
                    ],
                    [Action,If(ListVal(vote,#id#,referendum_id)>0,If(ListVal(vote,#id#,choice)==1,Tag(h4, $Y$, panel-title text-primary text-center mb0),Tag(h4, $N$, panel-title text-danger text-center mb0)),Div(text-center, If(ListVal(vote,#id#,referendum_id)>0," ",BtnContract(@glrf_Voting, Em(fa fa-thumbs-up text-muted), Your choice Yes,"ReferendumId:#id#,RFChoice:1",'btn btn-default btn-xs' data-toggle="tooltip" data-trigger="hover" title="$Y$")   BtnContract(@glrf_Voting, Em(fa fa-thumbs-down text-muted), Your choice No,"ReferendumId:#id#,RFChoice:0",'btn btn-default btn-xs' data-toggle="tooltip" data-trigger="hover" title="$N$"))))
                    ],
                    [Results,If(ListVal(vote,#id#,referendum_id)>0,If(#status#==1, BtnPage(glrf_ViewResult, Em(fa fa-eye) Span('', ), "ReferendumId:#id#,DateStart:'#date_voting_start#', DateFinish:'#date_voting_finish#', NumberVotes:#number_votes#,Back:1,Status:1,global:1",btn btn-info data-toggle="tooltip" data-trigger="hover" title="$Res$")), Tag(button, Em(fa fa-eye-slash) Span('', ), btn btn-info data-toggle="tooltip" data-trigger="hover" title="$Res$ not available" disabled))
                    ]
                ]
            }
        DivsEnd:
    DivsEnd:
     If(GetOne(admin, global_states_list, gstate_id=#state_id#))
    Divs(panel-footer)
        Divs: clearfix
            Divs: pull-right
                BtnPage(glrf_List, Polling ,global:1)
            DivsEnd:
        DivsEnd:
    DivsEnd:
    IfEnd:
DivsEnd:`)
TextHidden( ap_dashboard_default, ap_government)
SetVar(`am_government #= MenuItem(Administration,  gov_administration)
MenuItem(Citizens,  citizens)
MenuItem(Pollings List,  RF_List)
MenuItem(New Polling,  RF_NewIssue)`,
`am_menu_default #= MenuItem(ListVotings, RF_UserList)
MenuItem(QuestionList, RF_UserQuestionList)
MenuItem(Elections, GECampaigns)`)
TextHidden( am_government, am_menu_default)
SetVar(`l_lang #= {" ResultSoon":"{\"en\": \" Result will be soon\", \"nl\": \" Result will be soon\", \"ru\": \"Результат будет скоро\"}","Actn":"{\"en\": \"Actions\", \"nl\": \"Acties\", \"ru\": \"Действия\"}","Actual":"{\"en\": \"Actual\", \"nl\": \"Actueel\", \"ru\": \"Текущие\"}","Ans":"{\"en\": \"Answer\", \"nl\": \"Antwoord\", \"ru\": \"Ответить\"}","Cancel":"{\"en\": \"Cancel\", \"nl\": \"Annuleer\", \"ru\": \"Отменить\"}","Chng":"{\"en\": \"Change\", \"nl\": \"Wijzigen\", \"ru\": \"Изменить\"}","Confirm":"{\"en\": \"Confirm\", \"nl\": \"Bevestig\", \"ru\": \"Подтвердить\"}","Contin":"{\"en\": \"Continues\", \"nl\": \"Doorgaan\", \"ru\": \"Продолжаются\"}","Continues":"{\"en\": \"Continues\", \"nl\": \"Doorgaan\", \"ru\": \"Продолжаются\"}","DateFinishVoting":"{\"en\": \"Date Finish Voting\", \"nl\": \"Eind datum stem vraag\", \"ru\": \"Дата завершения голосования\"}","DateStartVoting":"{\"en\": \"Date Start Voting\", \"nl\": \"Begin datum stem vraag\", \"ru\": \"Дата начала голосования\"}","Del":"{\"en\": \"Delete\", \"nl\": \"Verwijdering\", \"ru\": \"Удалить\"}","EnterIssue":"{\"en\": \"Enter Issue\", \"nl\": \"Onderwerp\", \"ru\": \"Введите вопрос\"}","Finish":"{\"en\": \"Finish\", \"nl\": \"Einde\", \"ru\": \"Окончание\"}","Finished":"{\"en\": \"Finished\", \"nl\": \"Einde\", \"ru\": \"Завершено\"}","FinishedVotings":"{\"en\": \"Finished Votings\", \"nl\": \"Einde Stemmen\", \"ru\": \"Завершенные\"}","Fnsh":"{\"en\": \"Finish\", \"nl\": \"Einde\", \"ru\": \"Окончание\"}","Gender":"{\"en\": \"Gender\", \"ru\": \"Пол\"}","GetResult":"{\"en\": \"Get Result\", \"nl\": \"Haa resultaat op\", \"ru\": \"Получить результат\"}","GovernmentDashboard":"{\"en\": \"Government dashboard\", \"nl\": \"Land overzicht\", \"ru\": \"Панель государства\"}","Inf":"{\"en\": \"Info\", \"nl\": \"Info\", \"ru\": \"Информация\"}","Info":"{\"en\": \"Info\", \"nl\": \"Info\", \"ru\": \"Информация\"}","Iss":"{\"en\": \"Issue\", \"nl\": \"Onderwerp\", \"ru\": \"Вопрос\"}","Issue":"{\"en\": \"Issue\", \"nl\": \"Onderwerp\", \"ru\": \"Вопрос\"}","ListVotings":"{\"en\": \"List of Polling\", \"nl\": \"Stemlijst\", \"ru\": \"Опросы\"}","N":"{\"en\": \"No\", \"nl\": \"Nee\", \"ru\": \"Нет\"}","NewVoting":"{\"en\": \"New Polling\", \"nl\": \"Nieuwe vraag\", \"ru\": \"Новый опрос\"}","Next":"{\"en\": \"Next\", \"nl\": \"Naast\", \"ru\": \"Дальше\"}","No":"{\"en\": \"No\", \"nl\": \"Nee\", \"ru\": \"Нет\"}","NoAvailablePolls":"{\"en\": \"No Available Polls\", \"nl\": \"Geen beschikbare vragen\", \"ru\": \"Нет доступных голосований\"}","QuestionList":"{\"en\": \"Questions List\", \"nl\": \"Lijst van vragen\", \"ru\": \"Список вопросов\"}","Referendapartij":"{\"en\": \"Referendapartij\", \"nl\": \"stemNLwijzer.nl - directe democratie\", \"ru\": \"Referendapartij\"}","Res":"{\"en\": \"Result\", \"nl\": \"Resultaat\", \"ru\": \"Результат\"}","Result":"{\"en\": \"Result\", \"nl\": \"Resultaat\", \"ru\": \"Результат\"}","ResultSoon":"{\"en\": \" Result will be soon\", \"nl\": \" Result will be soon\", \"ru\": \"Результат будет скоро\"}","Save":"{\"en\": \"Save\", \"nl\": \"Bewaren\", \"ru\": \"Сохранить\"}","Start":"{\"en\": \"Start\", \"nl\": \"Sart\", \"ru\": \"Начало\"}","StartVote":"{\"en\": \"Start Vote\", \"nl\": \"Begin stemmen\", \"ru\": \"Начать голосование\"}","Stp":"{\"en\": \"Stop\", \"nl\": \"Stop\", \"ru\": \"Стоп\"}","Strt":"{\"en\": \"Start\", \"nl\": \"Sart\", \"ru\": \"Начало\"}","TotalVoted":"{\"en\": \"Total voted\", \"nl\": \"Aantal stemmen\", \"ru\": \"Всего проголосовало\"}","TypeIssue":"{\"en\": \"Type\", \"nl\": \"Type\", \"ru\": \"Тип опроса\"}","View":"{\"en\": \"View\", \"nl\": \"Uitzicht\", \"ru\": \"Смотреть\"}","Vote":"{\"en\": \"Vote\", \"nl\": \"Stemmen\", \"ru\": \"Голосовать\"}","Voting":"{\"en\": \"Voting\", \"nl\": \"Stemmen\", \"ru\": \"Голосование\"}","VotingFinished":"{\"en\": \" Voting finished\", \"nl\": \"Ende stemmen\", \"ru\": \"Голосование закончено\"}","Vw":"{\"en\": \"View\", \"nl\": \"Uitzicht\", \"ru\": \"Смотреть\"}","Welcome":"{\"en\": \"Welcome\", \"nl\": \"Welkom\", \"ru\": \"Добро пожаловать\"}","Y":"{\"en\": \"Yes\", \"nl\": \"Ja\", \"ru\": \"Да\"}","Yes":"{\"en\": \"Yes\", \"nl\": \"Ja\", \"ru\": \"Да\"}","YouVoted":"{\"en\": \"You voted for all available issues\", \"nl\": \"U stemt op alle beschikbare onderwerpen\", \"ru\": \"Вы проголосовали за все доступные вопросы\"}","YourAnswer":"{\"en\": \"Your Answer\", \"nl\": \"Uw antwoord\", \"ru\": \"Ваш ответ\"}","dateformat":"{\"en\": \"YYYY-MM-DD\", \"ru\": \"DD.MM.YYYY\"}","female":"{\"en\": \"Female\", \"ru\": \"Женский\"}","male":"{\"en\": \"Male\", \"ru\": \"Мужской\"}","qes1":"{\"en\": \"the first question \", \"nl\": \"De eerste vraag\"}","ques1":"{\"en\": \"the first question \", \"nl\": \"De eerste vraag\"}","timeformat":"{\"en\": \"YYYY-MM-DD HH:MI:SS\", \"ru\": \"DD.MM.YYYY HH:MI:SS\"}"}`)
TextHidden(l_lang)
Json(`Head: "Basic Apps",
Desc: "Election and Assign, Polling, Messenger, Simple Money System",
		Img: "/static/img/apps/money.png",
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
        			table_name : "#state_id#_citizens",
        			column_name: "avatar",
        			index: "0",
        			column_type: "text",
        			permissions: "ContractConditions(\"CitizenCondition\")"
        		}
        },
		{
        		Forsign: 'table_name,column_name,permissions,index,column_type',
        		Data: {
        			type: "NewColumn",
        			typeid: #typecolid#,
        			table_name : "#state_id#_citizens",
        			column_name: "name",
        			index: "0",
        			column_type: "hash",
        			permissions: "ContractConditions(\"CitizenCondition\")"
        		}
        },
		{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "accounts",
			columns: '[["amount", "money", "0"],["onhold", "int64", "1"],["agency_id", "int64", "1"],["citizen_id", "int64", "1"],["company_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_accounts",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"AddAccount\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "amount",
			permissions: "ContractAccess(\"MoneyTransfer\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "onhold",
			permissions: "ContractAccess(\"DisableAccount\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "agency_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_accounts",
			column_name: "company_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "citizen_del",
			columns: '[["status", "int64", "1"],["citizen_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_citizen_del",
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
			table_name : "#state_id#_citizen_del",
			column_name: "status",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizen_del",
			column_name: "citizen_id",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "citizenship_requests",
			columns: '[["public_key_0", "text", "0"],["dlt_wallet_id", "int64", "1"],["name", "hash", "0"],["approved", "int64", "1"],["block_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_citizenship_requests",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "true",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "approved",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "block_id",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "public_key_0",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "dlt_wallet_id",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_citizenship_requests",
			column_name: "name",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_campaigns",
			columns: '[["candidates_deadline", "time", "1"],["name", "text", "0"],["status", "int64", "1"],["num_votes", "int64", "0"],["date_start", "time", "1"],["position_id", "int64", "1"],["date_stop_voting", "time", "1"],["date_start_voting", "time", "1"]]',
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
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_start",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "position_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_stop_voting",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "date_start_voting",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "candidates_deadline",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_campaigns",
			column_name: "name",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_candidates",
			columns: '[["candidate", "text", "0"],["description", "text", "0"],["position_id", "int64", "1"],["counter", "int64", "1"],["campaign", "text", "0"],["citizen_id", "int64", "1"],["application_date", "time", "0"],["id_election_campaign", "int64", "1"],["result", "int64", "1"]]',
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
			column_name: "candidate",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "position_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "application_date",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "id_election_campaign",
			permissions: "false",
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
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_candidates",
			column_name: "description",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_elective_office",
			columns: '[["name", "hash", "1"],["last_election", "time", "1"]]',
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
			column_name: "last_election",
			permissions: "ContractConditions(\"MainCondition\")",
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
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "ge_person_position",
			columns: '[["position", "text", "0"],["citizen_id", "int64", "1"],["date_start", "time", "1"],["position_id", "int64", "1"],["name", "text", "0"],["date_end", "time", "1"]]',
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
			column_name: "date_start",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "position_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "name",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "date_end",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "position",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_person_position",
			column_name: "citizen_id",
			permissions: "false",
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
			column_name: "time",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_votes",
			column_name: "strhash",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_ge_votes",
			column_name: "id_candidate",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "messages",
			columns: '[["flag", "text", "0"],["text", "text", "0"],["delete", "int64", "1"],["stateid", "int64", "1"],["username", "hash", "1"],["ava", "text", "0"],["time", "time", "1"],["state_id", "int64", "1"],["citizen_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_messages",
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
			table_name : "#state_id#_messages",
			column_name: "flag",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "stateid",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "state_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "username",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "ava",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "text",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "time",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_messages",
			column_name: "delete",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "positions_citizens",
			columns: '[["citizen_name", "hash", "1"],["position_name", "hash", "1"],["date", "time", "0"],["dismiss", "int64", "1"],["citizen_id", "int64", "1"],["position_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_positions_citizens",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"GV_Positions_Citizens\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "date",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "dismiss",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "position_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "citizen_name",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_citizens",
			column_name: "position_name",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "positions_list",
			columns: '[["position_name", "hash", "1"],["position_type", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_positions_list",
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
			table_name : "#state_id#_positions_list",
			column_name: "position_name",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_positions_list",
			column_name: "position_type",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "rf_referendums",
			columns: '[["issue", "text", "0"],["result", "int64", "1"],["status", "int64", "1"],["number_0", "int64", "1"],["number_1", "int64", "1"],["number_votes", "int64", "1"],["date_voting_finish", "time", "1"],["type", "int64", "1"],["date_enter", "time", "1"],["date_voting_start", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_referendums",
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
			table_name : "#state_id#_rf_referendums",
			column_name: "issue",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "number_1",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "date_enter",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "number_votes",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "date_voting_start",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "date_voting_finish",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "type",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "result",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "status",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_referendums",
			column_name: "number_0",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "rf_result",
			columns: '[["value", "int64", "1"],["choice", "int64", "1"],["percents", "int64", "1"],["choice_str", "text", "0"],["referendum_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_result",
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
			table_name : "#state_id#_rf_result",
			column_name: "value",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "choice",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "percents",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "choice_str",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_result",
			column_name: "referendum_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "rf_votes",
			columns: '[["choice", "int64", "1"],["strhash", "hash", "1"],["citizen_id", "int64", "1"],["referendum_id", "int64", "1"],["time", "time", "1"],["answer", "text", "0"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_rf_votes",
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
			table_name : "#state_id#_rf_votes",
			column_name: "citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "referendum_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "time",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "answer",
			permissions: "ContractConditions(\"CitizenCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "choice",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_rf_votes",
			column_name: "strhash",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "CentralBankConditions",
			value: $("#sc_CentralBankConditions").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "CentralBankConditions"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SmartLaw_NumResultsVoting",
			value: $("#sc_SmartLaw_NumResultsVoting").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SmartLaw_NumResultsVoting"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "CitizenCondition",
			value: $("#sc_CitizenCondition").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "CitizenCondition"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "MoneyTransfer",
			value: $("#sc_MoneyTransfer").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "MoneyTransfer"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "AddAccount",
			value: $("#sc_AddAccount").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "AddAccount"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "AddCitizenAccount",
			value: $("#sc_AddCitizenAccount").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "AddCitizenAccount"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
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
			global: 0,
			id: "addMessage"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "CitizenDel",
			value: $("#sc_CitizenDel").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "CitizenDel"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "DelMessage",
			value: $("#sc_DelMessage").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "DelMessage"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "DisableAccount",
			value: $("#sc_DisableAccount").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "DisableAccount"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "EditProfile",
			value: $("#sc_EditProfile").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "EditProfile"
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
			name: "GenCitizen",
			value: $("#sc_GenCitizen").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GenCitizen"
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
			name: "GV_NewPosition",
			value: $("#sc_GV_NewPosition").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GV_NewPosition"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GV_PositionDismiss",
			value: $("#sc_GV_PositionDismiss").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GV_PositionDismiss"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "GV_Positions_Citizens",
			value: $("#sc_GV_Positions_Citizens").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "GV_Positions_Citizens"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RechargeAccount",
			value: $("#sc_RechargeAccount").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RechargeAccount"
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
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_NewIssue"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_SaveAns",
			value: $("#sc_RF_SaveAns").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_SaveAns"
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
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_Voting"
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
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_VotingCancel"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_VotingDel",
			value: $("#sc_RF_VotingDel").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_VotingDel"
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
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_VotingResult"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_VotingStart",
			value: $("#sc_RF_VotingStart").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_VotingStart"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "RF_VotingStop",
			value: $("#sc_RF_VotingStop").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "RF_VotingStop"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SearchCitizen",
			value: $("#sc_SearchCitizen").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SearchCitizen"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SendMoney",
			value: $("#sc_SendMoney").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "SendMoney"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXCitizenRequest",
			value: $("#sc_TXCitizenRequest").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "TXCitizenRequest"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXEditProfile",
			value: $("#sc_TXEditProfile").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "TXEditProfile"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXNewCitizen",
			value: $("#sc_TXNewCitizen").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "TXNewCitizen"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "TXRejectCitizen",
			value: $("#sc_TXRejectCitizen").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: 0,
			id: "TXRejectCitizen"
			}
	   },
{
		Forsign: 'name,value,conditions',
		Data: {
			type: "NewStateParameters",
			typeid: #type_new_state_params_id#,
			name : "type_issue",
			value: $("#pa_type_issue").val(),
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'name,value,conditions',
		Data: {
			type: "NewStateParameters",
			typeid: #type_new_state_params_id#,
			name : "type_office",
			value: $("#pa_type_office").val(),
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'name,value,conditions',
		Data: {
			type: "NewStateParameters",
			typeid: #type_new_state_params_id#,
			name : "state_description",
			value: $("#pa_state_description").val(),
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "citizen_profile",
			menu: "menu_default",
			value: $("#p_citizen_profile").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "citizens",
			menu: "government",
			value: $("#p_citizens").val(),
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
			menu: "government",
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
			name : "gov_administration",
			menu: "government",
			value: $("#p_gov_administration").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_List",
			menu: "government",
			value: $("#p_RF_List").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_NewIssue",
			menu: "government",
			value: $("#p_RF_NewIssue").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_Result",
			menu: "government",
			value: $("#p_RF_Result").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_UserAns",
			menu: "menu_default",
			value: $("#p_RF_UserAns").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_UserList",
			menu: "menu_default",
			value: $("#p_RF_UserList").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_UserQuestionList",
			menu: "menu_default",
			value: $("#p_RF_UserQuestionList").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_ViewResult",
			menu: "menu_default",
			value: $("#p_RF_ViewResult").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "RF_ViewResultQuestions",
			menu: "menu_default",
			value: $("#p_RF_ViewResultQuestions").val(),
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
			menu: "government",
			value: $("#p_StateInfo").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
				Forsign: 'name,trans',
				Data: {
					type: "NewLang",
					typeid: #type_new_lang_id#,
					name : "",
					trans: $("#l_lang").val(),
					}
				},
{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendPage",
				typeid: #type_append_page_id#,
				name : "dashboard_default",
				value: $("#ap_dashboard_default").val(),
				global: 0
				}
		},
{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendPage",
				typeid: #type_append_page_id#,
				name : "government",
				value: $("#ap_government").val(),
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
