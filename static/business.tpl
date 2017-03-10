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
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "ContractConditions(\"MainCondition\")")
SetVar(`sc_AddAccount = contract AddAccount {
	data {
		Id int
		TypeAccount string
	}
	func conditions {
	    
	    
	    if ($TypeAccount == "citizen_id" && $Id != $citizen)
	    {
	        warning "You have no right to this action"
	    }

	    
	    if $TypeAccount == "company_id"
	    {
	        CompanyConditions("CompanyId",$Id)
	    }
	}
	
	func action {
	    
		DBInsert(Table("accounts"), $TypeAccount+",onhold", $Id, 0)
	}
}`,
`sc_AddCitizenAccount = contract AddCitizenAccount {
	data {
		CitizenId string
	}
	func conditions {
	    
	    $citizen_id = AddressToId($CitizenId)
		if $citizen_id == 0 {
			warning "not valid citizen id"
		}
	}
	func action {
		
		AddAccount("Id,TypeAccount",$citizen_id, "citizen_id")
	}
}`,
`sc_BuyItem = contract BuyItem {
	data {
		ItemId int
		Price money
		ItemCode int
		SellCompanyId int
		BuyCompanyId int
	}

	func conditions {
	     
	     //SL_CompanyCertificate("ItemCode,BuyCompanyId",$ItemCode,$BuyCompanyId)
	     CompanyConditions("CompanyId",$BuyCompanyId)
	     
	}
	func action {
        
        var sender_id int
    	var recipient_id int
    	
    	recipient_id = DBIntExt(Table("accounts"), "id",$SellCompanyId, "company_id")
    	sender_id = DBIntExt(Table("accounts"), "id", $BuyCompanyId, "company_id")
        
        $sales_tax=0
        SL_SalesTax("SenderAccountId,Price",sender_id,$Price)
        MoneyTransfer("SenderAccountId,RecipientAccountId,Amount",sender_id,recipient_id,$Price)
        
        var itemname string
	    itemname=DBStringExt( Table("items"), "name", $ItemId, "id")
        DBInsert(Table("buysell"), "buyer_company_id,item_id,name_item,price,seller_company_id,sales_tax,timestamp date", $BuyCompanyId, $ItemId, itemname, $Price, $SellCompanyId, $sales_tax, $block_time)
        
         BuyItemCompanyContracts("Price,SellCompanyId",$Price, $SellCompanyId)
	}
}`,
`sc_BuyItemCompanyContracts = contract BuyItemCompanyContracts {
	data {
		Price money
		SellCompanyId int
	}

	func conditions {
	    
	}
	func action {
	    
        
        var list array
        var contr map
        var par map
        var i int
        var len int
        
        par["Price"] = $Price
        
        list = DBGetList(Table("jobs"), "smart_contract_name",0,1000,"id","company_id=$",$SellCompanyId)
        
        len = Len(list)
        while i < len {
            contr = list[i]
            i = i + 1
            CallContract(contr["smart_contract_name"], par)
        }
	}
}`,
`sc_BZ_AcceptApplication = contract BZ_AcceptApplication {
	data {
	    
		CompanyId int
		VacancyApplId int
		CitizenId int
	}
	func conditions {
	    
        CompanyConditions("CompanyId",$CompanyId)
        
		 var x int
		 x = DBInt(Table( "job_vacancy_application"), "status", $VacancyApplId )
	    if x == 1 {
			warning "You have already signed a contract with the employee"
		}

	}
	func action {
	    
	   DBUpdate(Table("job_vacancy_application"), $VacancyApplId, "status", 1)
	   
	    var company_name string
	    var occupation_name string
	    var citizen_name string
	    var contract_id int
	    var contract_name string
	    var vacancy_id int
	    var occupation_id int
	    
	     
        company_name = DBStringExt(Table("companies"), "name", $CompanyId, "id")
	    citizen_name =  DBStringExt(Table("citizens"), "name", $CitizenId, "id")
	    contract_id = DBStringExt(Table("job_vacancy_application"), "smart_contract_id", $VacancyApplId, "id")
	    contract_name = DBStringExt(Table("job_vacancy_application"), "smart_contract_name", $VacancyApplId, "id")
	    vacancy_id = DBStringExt(Table("job_vacancy_application"), "vacancy_id", $VacancyApplId, "id")
	    occupation_id = DBStringExt(Table("jobs_vacancies"), "occupation_id", vacancy_id, "id")
	    occupation_name = DBStringExt(Table("jobs_vacancies"), "occupation_name", vacancy_id, "id")
	    
		DBInsert(Table("jobs"), "citizen_id, citizen_name, company_id, company_name, occupation_id,	occupation_name, smart_contract_id, smart_contract_name, status, vacancy_id, timestamp date_start", $CitizenId, citizen_name, $CompanyId, company_name, occupation_id, occupation_name, contract_id, contract_name, 1, vacancy_id, $block_time)
	   
	}
}`,
`sc_BZ_CertificateConfirm = contract BZ_CertificateConfirm {
	data {
	    
		CertificateId int
		ItemId int
		
	}
	func conditions {
	   
	    var tl_conclusion int
		tl_conclusion = DBIntExt(Table("bz_certificates"), "tl_conclusion", $CertificateId, "id")
		if tl_conclusion==0
		{
		    warning "No Test Lab conclusion"
		}
	    
	}
	func action {
	    
	   DBUpdate(Table( "bz_certificates"), $CertificateId, "certificate", 1)
	   DBUpdate(Table( "items"), $ItemId, "certificate", 1)
	   
	}
}`,
`sc_BZ_JobApplication = contract BZ_JobApplication {
	data {
		OccupationId int
	}
	func conditions {
	    
	}
	func action {
	    var occupation_name string
	    var citizen_name string
	    occupation_name = DBStringExt(Table("occupations"), "name", $OccupationId, "id")
	    citizen_name =  DBStringExt(Table("citizens"), "name", $citizen, "id")
	    
		DBInsert(Table("jobs_application"), "citizen_id,citizen_name,occupation_id,occupation_name,status,timestamp date", $citizen, citizen_name, $OccupationId, occupation_name, 1,$block_time)
		
	}
}`,
`sc_BZ_JobRecruitment = contract BZ_JobRecruitment {
	data {
		VacancyId int
		CompanyId int
		CitizenId int
		ContractName string
		ContractId int
		OccupationId int
	}
	func conditions {
	    
	}
	func action {
	    
	    var company_name string
	    var occupation_name string
	    var citizen_name string
	    
	    occupation_name = DBStringExt(Table("occupations"), "name", $OccupationId, "id")
        company_name = DBStringExt(Table("companies"), "name", $CompanyId, "id")
	    citizen_name =  DBStringExt(Table("citizens"), "name", $citizen, "id")
	    
		DBInsert(Table("jobs"), "citizen_id, citizen_name, company_id, company_name, occupation_id,	occupation_name, smart_contract_id, smart_contract_name, status, vacancy_id, timestamp date_start", $CitizenId, citizen_name, $CompanyId, company_name, $OccupationId, occupation_name, $ContractId, $ContractName, 0, $VacancyId, $block_time)
		
	}
}`,
`sc_BZ_JobVacancy = contract BZ_JobVacancy {
	data {
		OccupationId int
		CompanyId int
		SCName string
	}
	func conditions {
	    var x int
	    
	    x = DBIntWhere(Table("jobs_vacancies"), "id", "company_id=$ and occupation_id=$ and status=1", $CompanyId, $OccupationId)
	    if x != 0 {
			warning "Vacancy is already open"
		}
	    
	    $smart_contract_id = DBIntExt(Table("smart_contracts"), "id", $SCName, "name")
	    if $smart_contract_id == 0 {
			warning "The contract with the same name does not exist"
		}
	    
	     CompanyConditions("CompanyId",$CompanyId) 
	}
	func action {
	    
	    var company_name string
	    company_name = DBStringExt(Table("companies"), "name", $CompanyId, "id")
	    var occupation_name string
	    occupation_name = DBStringExt(Table("occupations"), "name", $OccupationId, "id")
	    
		DBInsert(Table("jobs_vacancies"), "company_id,company_name,occupation_id,occupation_name,smart_contract_name,smart_contract_id,status,timestamp date", $CompanyId, company_name, $OccupationId,occupation_name, $SCName,$smart_contract_id,1, $block_time)
		
	}
}`,
`sc_BZ_RejectApplication = contract BZ_RejectApplication {
	data {
	    
		CompanyId int
		VacancyApplId int
		
	}
	func conditions {
	    
	    CompanyConditions("CompanyId",$CompanyId)
		
		var x int
		x = DBInt(Table( "job_vacancy_application"), "status", $VacancyApplId )
	    if x == 1 {
			warning "You have already signed a contract with the employee"
		}

	}
	func action {
	    
	   DBUpdate(Table("job_vacancy_application"), $VacancyApplId, "status", 2)
	   
	}
}`,
`sc_BZ_TestLabConfirm = contract BZ_TestLabConfirm {
	data {
	    
		CertificateId int
		
	}
	func conditions {
	    
        	MainCondition()

	}
	func action {
	    
	   DBUpdate(Table( "bz_certificates"), $CertificateId, "tl_conclusion", 1)
	   
	}
}`,
`sc_BZ_VacancyAgree = contract BZ_VacancyAgree {
	data {
		VacancyId int
		CompanyId int
		CitizenId int
		ContractName string
		ContractId int
		OccupationId int
	}
	func conditions {
	    
	    var x int
	    x = DBIntWhere(Table("job_vacancy_application"), "id", "citizen_id=$ and vacancy_id=$", $CitizenId, $VacancyId)
	    if x != 0 {
			warning "You have already applied"
		}
	    
	}
	func action {
	    
	    var occupation_name string
	    var citizen_name string
	    
	    occupation_name = DBStringExt(Table("occupations"), "name", $OccupationId, "id")
	    citizen_name =  DBStringExt(Table("citizens"), "name", $citizen, "id")
	    
		DBInsert(Table("job_vacancy_application"), "citizen_id,citizen_name,company_id,occupation_name, smart_contract_id,	smart_contract_name,status, vacancy_id, timestamp date", $CitizenId, citizen_name, $CompanyId, occupation_name, $ContractId, $ContractName, 0, $VacancyId, $block_time)
	}
}`,
`sc_CentralBankConditions = contract CentralBankConditions {
	data {	}
	
	func conditions	{
	    
	    MainCondition()
	    
	}

	func action {	}
}`,
`sc_CompanyConditions = contract CompanyConditions {
	data {
	    
	    CompanyId int
	}
	
	func conditions	{
	  
	     if $citizen != DBInt(Table("companies"), "owner_citizen_id", $CompanyId)
	    {
	        warning "You have no right to this action"
	    }
	    
	}

	func action {	}
}`,
`sc_DisableAccount = contract DisableAccount {
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
`sc_JobContract_Test = contract JobContract_Test {
	data {
		Price money
	}
	func conditions {
		    
	    /////////////////////////////////////////
		
		$company_id = #
		$citizen_id = #########
		
		/////////////////////////////////////////
		
	    $percentage_of_sales = 1 
	    
		/////////////////////////////////////////
	    
	    
	    if DBString(Table("citizens"), "public_key_0", $citizen_id)=="" {
			warning "Check JobContract_Test"
		}
		
		if DBString(Table("companies"), "name", $company_id)=="" {
			warning "Check JobContract_Test"
		}
	    
	}
	func action {
	    
	    var income int
	    var citizen_account_id int
	    var company_account_id int
		
		company_account_id = DBIntExt(Table("accounts"), "id", $company_id, "company_id")
        citizen_account_id = DBIntExt(Table("accounts"), "id", $citizen_id, "citizen_id")
        income = $Price * $percentage_of_sales/100

	    MoneyTransfer("SenderAccountId,RecipientAccountId,Amount",company_account_id,citizen_account_id,income)
        
	}
}`,
`sc_MoneyTransfer = contract MoneyTransfer {
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
`sc_NewCompany = contract NewCompany {
	data {
		Name string
	}
	func conditions {
	    if DBIntWhere(Table("companies"), "id", "name=$", $Name) > 0
	    {
	    	    warning "The company with the same name already exists in the registry"
	    }
	}
	func action {
	    
		DBInsert(Table("companies"), "name, owner_citizen_id,timestamp opened_time", $Name, $citizen, $block_time)
		
		var companyid int
	    companyid=DBIntWhere(Table("companies"), "id", "owner_citizen_id=$ and name=$",  $citizen, $Name)
	    
		AddAccount("Id,TypeAccount", companyid, "company_id")
		
	}
}`,
`sc_NewGovernmentAgency = contract NewGovernmentAgency {
	data {
		Name string
	}
	func conditions {
	    
	    MainCondition()
	   
	}
	func action {
		DBInsert(Table("government_agencies"), "name, timestamp opened_time", $Name, $block_time)
		
		var agencyid int
	    agencyid=DBIntWhere(Table("government_agencies"), "id", "name=$ and id>0", $Name)
	    
	    AddAccount("Id,TypeAccount", agencyid, "agency_id")
		
	}
}`,
`sc_NewItem = contract NewItem {
		data {
		ItemName string
		CompanyId int
		ItemPrice money
		ItemCode int
	}
	func conditions {
	    
	    
	    	if DBIntWhere(Table("items"), "id", "name=$", $ItemName) > 0
	    	{
	    	     warning "The item with the same name already exists in the registry"
	    	}
	    	
	        CompanyConditions("CompanyId",$CompanyId)
	    
	}
	func action {
	        
	   SL_ItemCertificate("ItemCode", $ItemCode)

	    DBInsert(Table("items"), "name, company_id, timestamp added_time, price, code_item, certificate", $ItemName, $CompanyId, $block_time, $ItemPrice,$ItemCode,$certificate)
	    	
	    if $certificate==0 
	    {
	    	var itemid int
	    	itemid=DBIntWhere(Table("items"), "id", "company_id=$ and code_item=$", $CompanyId, $ItemCode)
	    	
	    	DBInsert(Table("bz_certificates"), "item_name, item_id, company_id, timestamp added_time, item_code, tl_conclusion, certificate", $ItemName, itemid, $CompanyId, $block_time, $ItemCode,0,0)
	    }   
	    
		
	}
}`,
`sc_RechargeAccount = contract RechargeAccount {
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
`sc_SendMoney = contract SendMoney {
	data {

		RecipientAccountId int 
		Amount money
	}

	func conditions {
	    
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
`sc_SL_ItemCertificate = contract SL_ItemCertificate {
	data {
	    ItemCode int
	}

	func conditions {
	    
	}
	func action {
	    
	    
	    if $ItemCode < 1000 && $ItemCode > 899
	    {
	        $certificate=0
	        
	    }else{
	        
	        $certificate=2
	    }
	    

	}
}`,
`sc_SL_SalesTax = contract SL_SalesTax {
	data {
		Price money
		SenderAccountId int
	}
	func conditions {
	    
	}
	func action {
	    
	    var tax_rate int
	    var tax_agency_id int
	    var tax_account int
	    
	    tax_agency_id = 1	    
	    tax_rate = 5
	    $sales_tax = $Price*tax_rate/100

	    tax_account=DBIntExt(Table("accounts"), "id",tax_agency_id, "agency_id")
	    
	    MoneyTransfer("SenderAccountId,RecipientAccountId,Amount",$SenderAccountId,tax_account,$sales_tax)
	
	}
}`)
TextHidden( sc_AddAccount, sc_AddCitizenAccount, sc_BuyItem, sc_BuyItemCompanyContracts, sc_BZ_AcceptApplication, sc_BZ_CertificateConfirm, sc_BZ_JobApplication, sc_BZ_JobRecruitment, sc_BZ_JobVacancy, sc_BZ_RejectApplication, sc_BZ_TestLabConfirm, sc_BZ_VacancyAgree, sc_CentralBankConditions, sc_CompanyConditions, sc_DisableAccount, sc_JobContract_Test, sc_MoneyTransfer, sc_NewCompany, sc_NewGovernmentAgency, sc_NewItem, sc_RechargeAccount, sc_SendMoney, sc_SL_ItemCertificate, sc_SL_SalesTax)
SetVar(`p_AgencyInfo #= Title : Agency info
Navigation( LiTemplate(government, Government), AgencyInfo)

Divs(md-6 )
Divs(panel widget bg-danger)
    Divs(row row-table)
        Divs(col-xs-4 text-center bg-danger-dark pv-lg)
             MarkDown: Em(icon-globe fa-3x)
        DivsEnd:
        Divs(col-xs-8 pv-lg)
            Divs(h1 m0 text-bold)
                MarkDown: GetOne(name, #state_id#_government_agencies, "id", #AgencyId#)
            DivsEnd:
            
        DivsEnd:
    DivsEnd:
DivsEnd:


    SetVar(Account = GetOne(id, #state_id#_accounts, "agency_id", #AgencyId#))
   
        WiAccount( #Account# )
    
   
   
        WiBalance(Money(GetOne(amount, #state_id#_accounts, "agency_id", #AgencyId#)), StateVal(currency_name))
    DivsEnd:
 
If (#AgencyId#==1) 
Divs(md-12, panel panel-default panel-body)
Legend(" ", "Sales Tax")
Table {
    Class: table-striped table-hover
    Table: #state_id#_buysell
	Order: date DESC
	Where: sales_tax > 0
	Columns: [
	[Date, Date(#date#, DD.MM.YYYY)],
	[Seller, #seller_company_id#],
	[Item ID,#item_id#],
	[Item Name,#name_item#],
	[Buyer, #buyer_company_id#],
	[Price, Money(#price#)],
	[Tax Rate, 5%]
	[Tax, Money(#sales_tax#)]
	]
}
DivsEnd:
IfEnd:
PageEnd:`,
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
Divs(md-8, panel panel-default panel-body)
Legend(" ", "My companies")
Table {
    Class: table-striped table-hover
    Table: #state_id#_companies
    Where: owner_citizen_id=#citizen#
	Order: id DESC
	Columns: [[ID, #id#],[Name, #name#],
	[Registration date, Date(#opened_time#, DD.MM.YYYY)], 
	[Info, BtnPage(CompanyDetails,Company Page,"CompanyId:#id#",btn btn-info btn-pill-right)] ]
}

DivsEnd:
Div(clearfix)


Div(clearfix)
Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add Job application")
        
        Divs(form-group)
            Label("Occupation")
            Select(OccupationId, #state_id#_occupations.name)
        DivsEnd:
        
        TxButton{ Contract: BZ_JobApplication, Name: Add,Inputs: "OccupationId=OccupationId",OnSuccess: "template,business" }
    FormEnd:
DivsEnd:

Divs(md-8, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "My Job Application")
        Table {
    Class: table-striped table-hover
    Table: #state_id#_jobs_application
	Order: id DESC
	Where: status!=0 and citizen_id=#citizen#
	Columns: [[Id, #id#],
	[Occupation, #occupation_name#],
	[Date, Date(#date#, DD.MM.YYYY)],
	[Vacancies, BtnPage(BZ_Vacancies,View,"OccupationID:#occupation_id#",btn btn-info btn-pill-right)]]
    }
DivsEnd:
PageEnd:`,
`p_BZ_Certificate #= Title : Certificate
Navigation( LiTemplate(dashboard_default, Citizen), Business)

Divs(md-12, panel panel-default panel-body data-sweet-alert)

SetVar(BtnTestLab = BtnContract(BZ_TestLabConfirm, Confirm, Confirm Test Lab Conclusion for #item_name#,"CertificateId:#id#,ItemId:#item_id#",'btn btn-primary btn-sm'))

SetVar(BtnCertificate = BtnContract(BZ_CertificateConfirm, Certify,Confirm Certificate for #item_name#,"CertificateId:#id#,ItemId:#item_id#",'btn btn-primary btn-sm'))

Legend(" ", "Certificate")
If (#CitizenId# = #citizen#)
BtnPage(CompanyDetails, Back to Company, "CompanyId:#CompanyId#",btn btn-info btn-pill-left ml4)
IfEnd:
Table {
    Class: table-striped table-hover
    Table: #state_id#_bz_certificates
	Order: id DESC
	Where: company_id=#CompanyId#
	Columns: [[Item Name, #item_name#],
		[Code, #item_code#],
    	[Registration date, Date(#added_time#, DD.MM.YYYY)],
    	[Test Lab Conclusion, If(#tl_conclusion# == 0, #BtnTestLab#,"Confirmed")],
    	[Certificate, If(And(#tl_conclusion# == 1, #certificate# == 0), #BtnCertificate#, If(#certificate# == 1, "Yes", "No"))]]
}

DivsEnd:
PageEnd:`,
`p_BZ_ContractView #= Navigation( Contract View )

DivsEnd:
    If(#Agree#==1)
    Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Employee consent")
        MarkDown: You confirm that you agree to the terms of the contract and agree to comply with them 
        
        Input(VacancyId, "hidden", text, text, #VacancyId#)
        Input(CompanyId, "hidden", text, text, #CompanyId#)
        Input(CitizenId, "hidden", text, text, #citizen#)
        Input(ContractName, "hidden", text, text, #ContractName#) 
		Input(ContractId, "hidden", text, text, #ContractId#) 
		Input(OccupationId, "hidden", text, text, #OccupationId#) 

        TxButton{ Contract: BZ_VacancyAgree, Name: Agree,Inputs: "VacancyId=VacancyId, CompanyId=CompanyId, CitizenId=CitizenId, ContractName=ContractName, ContractId=ContractId, OccupationId=OccupationId",OnSuccess: "template,business"}
    FormEnd:
    DivsEnd:
    IfEnd:

Divs(md-12, panel panel-default panel-body)
MarkDown : <h4>Smart Contract #ContractName#</h4>
MarkDown : Company Id: #CompanyId#
BtnPage(#Back#, Back, "CompanyId:#CompanyId#,CitizenId:'#CitizenId#'",btn btn-info btn-pill-left ml4)
     MarkDown: <br/>

    Form()
    Source(LawsValue, GetOne(value, #state_id#_smart_contracts, "id", #ContractId#))

    FormEnd: 
        



PageEnd:`,
`p_BZ_Vacancies #= Title : Job Vacancy
Navigation( LiTemplate(dashboard_default, Citizen), Vacancy)


Divs(md-8, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Job Vacancy")
    
    If(#OccupationID#>0) 
    BtnPage(BZ_Vacancies,View All,"OccupationID:0,Back:'#Back#'",btn btn-info)
    
        Table {
        Class: table-striped table-hover
        Table: #state_id#_jobs_vacancies
    	Order: id DESC
    	Where: status=1 and occupation_id=#OccupationID#
    	Columns: [[Id, #id#],
    	[Occupation, #occupation_name#],
    	[Company, BtnPage(CompanyDetails,#company_name#,"CompanyId:#company_id#",btn btn-info btn-pill-right)],
    	[Smart Contract, BtnPage(BZ_ContractView,View & Agree,"Agree:1,VacancyId:#id#, ContractId:#smart_contract_id#,CompanyId:#company_id#,OccupationId:#occupation_id#,ContractName:'#smart_contract_name#',Back:'BZ_Vacancies'",btn btn-info btn-pill-right)]
    	[Date Add, Date(#date#, DD.MM.YYYY)]]
        }
        
    Else:
    
        Table {
        Class: table-striped table-hover
        Table: #state_id#_jobs_vacancies
    	Order: id DESC
    	Where: status=1
    	Columns: [[Id, #id#],
    	[Occupation, #occupation_name#],
    	[Company, BtnPage(CompanyDetails,#company_name#,"CompanyId:#company_id#",btn btn-info btn-pill-right)],
    	[Smart Contract, BtnPage(BZ_ContractView,View & Agree,"Agree:1,VacancyId:#id#, ContractId:#smart_contract_id#,CompanyId:#company_id#,OccupationId:#occupation_id#,ContractName:'#smart_contract_name#',Back:'BZ_Vacancies'",btn btn-info btn-pill-right)]
    	[Date Add, Date(#date#, DD.MM.YYYY)]]
        }
        
    IfEnd:    
        
DivsEnd:
PageEnd:`,
`p_CentralBank #= Title : Central bank
Navigation( LiTemplate(government, Government),Central bank)



MarkDown: ## Accounts 

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add citizen account")
        
        Divs(form-group)
            Label("Citizen ID")
            InputAddress(CitizenId, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: AddCitizenAccount, Name: Add,Inputs: "CitizenId=CitizenId",OnSuccess: "template,CentralBank,global:0" }
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Disable account")
        
        Divs(form-group)
            Label("Account ID")
            Select(DAccountId, #state_id#_accounts.id, "form-control input-lg m-b")
        DivsEnd:
        TxButton{ Contract: DisableAccount, Name: Disable, Inputs: "AccountId=DAccountId",OnSuccess: "template,CentralBank,global:0" }
       
    FormEnd:
DivsEnd:

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Recharge Account")
        
        Divs(form-group)
            Label("Account ID")
            Select(AccountId, #state_id#_accounts.id, "form-control input-lg m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Amount")
            InputMoney(Amount, "form-control input-lg")
        DivsEnd:
        
        TxButton{ Contract: RechargeAccount, Name: Change, Inputs: "AccountId=AccountId, Amount=Amount", OnSuccess: "template,CentralBank,global:0" }
    FormEnd:
DivsEnd:

Div(clearfix md)

Divs(md-6, panel panel-default panel-body)
Legend(" ", "Companies accounts")
Table {
    Class: table-striped table-hover
    Table: #state_id#_accounts
	Order: id
	Where: company_id > 0 and onhold=0
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Company ID, BtnPage(CompanyDetails,#company_id#,"CompanyId:#company_id#",btn btn-info btn-pill-right)],[History, If(#rb_id#>0, LinkPage(sys-rowHistory, Show, "rbId:#rb_id#,tableName:'#state_id#_company_accounts'"), "No history")]]
}
DivsEnd:

Divs(md-6, panel panel-default panel-body)
Legend(" ", "Citizens accounts")
Table {
    Class: table-striped table-hover
    Table: #state_id#_accounts
	Order: id
	Where: citizen_id != 0 and onhold=0
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Citizen ID, BtnPage(CitizenPage,Address(#citizen_id#),"CitizenId:'#citizen_id#'",btn btn-info btn-pill-right)],[History, If(#rb_id#>0, LinkPage(sys-rowHistory, Show, "rbId:#rb_id#,tableName:'#state_id#_accounts'"), "No history")]]
}
DivsEnd: 

Divs(md-6, panel panel-default panel-body)
Legend(" ", "Goverment Agencies accounts")
Table {
    Class: table-striped table-hover
    Table: #state_id#_accounts
	Order: id
	Where: agency_id > 0 and onhold=0
	Columns: [[ID, #id#],[Amount, Money(#amount#)],[Agency ID, BtnPage(AgencyInfo,#agency_id#,"AgencyId:#agency_id#",btn btn-info btn-pill-right)],[History, If(#rb_id#>0, LinkPage(sys-rowHistory, Show, "rbId:#rb_id#,tableName:'#state_id#_accounts'"), "No history")]]
}
DivsEnd:



DivsEnd:

PageEnd:`,
`p_CitizenPage #= Title : Citizen page
Navigation( Dashboard )


Divs(md-6)
    GetRow(ci, #state_id#_citizens, "id", #CitizenId#)
    WiCitizen( #ci_name#, #ci_id#, #ci_avatar#, StateValue(state_flag) )
DivsEnd:
Divs(md-6)
    WiAccount(GetOne(id, #state_id#_accounts, "citizen_id", #CitizenId#))
    WiBalance(GetOne(amount, #state_id#_accounts, "citizen_id", #CitizenId#), StateVal(currency_name))
DivsEnd:

Div(clearfix)   
    
Divs(md-6, panel panel-default panel-body)
    
			Legend(" ", "Companies")
	Table {

		Table: #state_id#_companies
		Where: owner_citizen_id=#CitizenId#
		Order: id
		Columns: [
		[ID, #id#],
		[Name, #name#],
		[Registration date, Date(#opened_time#, DD.MM.YYYY)],
		[Details, BtnPage(CompanyDetails,Details,"CompanyId:#id#",btn btn-info btn-pill-right)]
		]
	}
    
DivsEnd:
       


Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Jobs on contract")
        Table {
    Class: table-striped table-hover
    Table: #state_id#_jobs
	Order: id DESC
	Where: status=1 and citizen_id=#CitizenId#
	Columns: [[Occupation, #occupation_name#],
	[Conmany, BtnPage(CompanyDetails,#company_name#,"CompanyId:'#company_id#'",btn btn-info btn-pill-right)],
	[Date, Date(#date_start#, DD.MM.YYYY)],
	[Smart Contract, BtnPage(BZ_ContractView,View Contract,"Agree:0,Back:'CitizenPage',CitizenId:'#citizen_id#',CompanyId:'#company_id#', ContractId:#smart_contract_id#,CompanyId:#company_id#,ContractName:'#smart_contract_name#'",btn btn-info btn-pill-right)]

	]
}
    
DivsEnd:
PageEnd:`,
`p_CompanyDetails #= Title : Company
Navigation( LiTemplate(dashboard_default, Citizen), Business)

SetVar(CompanyName = GetOne(name, #state_id#_companies, "id", #CompanyId#))
SetVar(Account = GetOne(id, #state_id#_accounts, "company_id", #CompanyId#))
SetVar(OwnerId = GetOne(owner_citizen_id, #state_id#_companies, "id", #CompanyId#))




Divs(md-4 )
    Divs(panel widget bg-danger)
        Divs(row row-table)
            Divs(col-xs-4 text-center bg-danger-dark pv-lg)
                 MarkDown: Em(fa fa-support fa-3x)
            DivsEnd: 
            Divs(col-xs-8 pv)
                Divs(h1 text-bold)
                    MarkDown: #CompanyName#
                DivsEnd:
                
            DivsEnd:
        DivsEnd:
    DivsEnd:
DivsEnd:    
    Divs(md-4 )
        WiAccount( #Account# )
    DivsEnd:
    Divs(md-4 )
        WiBalance(Money(GetOne(amount, #state_id#_accounts, "company_id", #CompanyId#)), StateVal(currency_name))
    DivsEnd:

Div(clearfix)
   
Divs(md-8, panel panel-default panel-body)

Legend(" ", "Company items")
Table {
    Class: table-striped table-hover
    Table: #state_id#_items
	Order: id DESC
	Where: company_id=#CompanyId#
	Columns: [[Name, #name#],
	[Date, Date(#added_time#, DD.MM.YYYY)],
	[Code, #code_item#],
	[Price, Money(#price#)],
	[Certificate,If(#certificate#==2,"Not required",If(#certificate#==1,"<b>Yes"</b>, BtnPage(BZ_Certificate,<b>No</b>,"CompanyId:#CompanyId#,CitizenId:#citizen#",btn btn-info btn-pill-right)))]]
}

DivsEnd:

If(#OwnerId#==#citizen#)

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add item")
        
        Divs(form-group)
            Label("Item Name")
            Input(ItemName, "form-control m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Price")
            InputMoney(ItemPrice, "form-control m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Code")
            Input(ItemCode, "form-control m-b")
        DivsEnd:
        
        
        Input(CompanyId, "hidden", text, text, #CompanyId#)
        TxButton{ Contract: NewItem, Name: Add,Inputs: "ItemName=ItemName, ItemPrice=ItemPrice, CompanyId=CompanyId, ItemCode=ItemCode", OnSuccess: "template,CompanyDetails,CompanyId:#CompanyId#" }
    FormEnd:
DivsEnd:

IfEnd:

Div(clearfix)


Divs(md-8, panel panel-default panel-body)
Legend(" ", "Company purchase")

If(#OwnerId#==#citizen#)
Divs(text-left)
BtnPage(shops, Buy,"CompanyId:#CompanyId#",btn btn-info btn-pill-right ml4)
DivsEnd:
IfEnd:

Table {
    Class: table-striped table-hover
    Table: #state_id#_buysell
	Order: date DESC
	Where: buyer_company_id=#CompanyId#
	Columns: [[Name, #name_item#],
	[Date, Date(#date#, DD.MM.YYYY)],
	[Price, Money(#price#)]]
}

DivsEnd:



Div(clearfix)
 Divs(md-8, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Employees on contract")
        Table {
    Class: table-striped table-hover
    Table: #state_id#_jobs
	Order: id DESC
	Where: status=1 and company_id=#CompanyId#
	Columns: [[Employee, BtnPage(CitizenPage,#citizen_name#,"CitizenId:'#citizen_id#'",btn btn-info btn-pill-right)],
	[Occupation, #occupation_name#],
	[Date, Date(#date_start#, DD.MM.YYYY)],
	[Smart Contract, BtnPage(BZ_ContractView,View,"Agree:0,Back:'CompanyDetails',CitizenId:#citizen_id#, ContractId:#smart_contract_id#,CompanyId:#company_id#,ContractName:'#smart_contract_name#'",btn btn-info btn-pill-right)]

	]
}
    
DivsEnd:


Divs(md-8, panel panel-default panel-body data-sweet-alert)

        Legend(" ", "Job Application")
        Table {
    Class: table-striped table-hover
    Table: #state_id#_job_vacancy_application
	Order: id DESC
	Where: status=0 and company_id=#CompanyId#
	Columns: [[Occupation, #occupation_name#],
	[Applicant, BtnPage(CitizenPage,#citizen_name#,"CitizenId:'#citizen_id#'",btn btn-info btn-pill-right)],
	[Date, Date(#date#, DD.MM.YYYY)],
	[Decision,BtnContract(BZ_AcceptApplication,Accept,Sign contract with#citizen_name#,"CompanyId:#CompanyId#, VacancyApplId:#id#,CitizenId:'#citizen_id#'",'btn btn-success btn-pill-left')],
	[ ,BtnContract(BZ_RejectApplication,Reject, Reject Application,"CompanyId:#CompanyId#,VacancyApplId:#id#,CitizenId:'#citizen_id#'",'btn btn-danger btn-pill-right')]
	]
}
DivsEnd:

Divs(md-8, panel panel-default panel-body data-sweet-alert)

  Legend(" ", "Job Vacancy")
 Table {
        Class: table-striped table-hover
        Table: #state_id#_jobs_vacancies
    	Order: id DESC
    	Where: status=1 and company_id=#CompanyId#
    	Columns: [[Id, #id#],
    	[Occupation, #occupation_name#],
    	[Smart Contract, BtnPage(BZ_ContractView,View,"Agree:0,VacancyId:#id#, ContractId:#smart_contract_id#,CompanyId:#company_id#,OccupationId:#occupation_id#,ContractName:'#smart_contract_name#',Back:'CompanyDetails'",btn btn-info btn-pill-right)]
    	[Date Add, Date(#date#, DD.MM.YYYY)]]
        }
DivsEnd:

Div(clearfix)

If(#OwnerId#==#citizen#)
Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add Job Vacancy")
        
        Divs(form-group)
            Label("Occupation")
            Select(OccupationId, #state_id#_occupations.name)
        DivsEnd:
        Divs(form-group)
            Label("Smart Contract")
            Input(SCName, "form-control m-b",text,text,"JobContract_Test")
        DivsEnd:
         Input(CompanyId, "hidden", text, text, #CompanyId#)
         Input(CompanyName, "hidden", text, text, #CompanyName#)
         
        TxButton{ Contract: BZ_JobVacancy, Name: Add,Inputs: "OccupationId=OccupationId,CompanyId=CompanyId,SCName=SCName",OnSuccess: "template,CompanyDetails,CompanyId:#CompanyId#" }
    FormEnd:
DivsEnd:
IfEnd:
PageEnd:`,
`p_shops #= Title : Shops
Navigation( LiTemplate(dashboard_default, Citizen), Shops)

SetVar(CompanyName = GetOne(name, #state_id#_companies, "id", #CompanyId#))


Divs(md-6, panel panel-default panel-body data-sweet-alert)
Legend(" ", "Buyer: <b>#CompanyName#</b>")
BtnPage(CompanyDetails, Back to Company page, "CompanyId:#CompanyId#",btn btn-info btn-pill-left ml4)
Table {
    Class: table-striped table-hover
    Table: #state_id#_items
	Order: id DESC
	Where: #certificate# > 0 and company_id!=#CompanyId#
	Columns: [[Goods name, #name#],
	[Price, Money(#price#)],
	[Buy, BtnContract(BuyItem, Buy, "Buy #name#<br/> Price - Money(#price#)","ItemId:#id#,Price:#price#,ItemCode:#code_item#,SellCompanyId:#company_id#,BuyCompanyId:#CompanyId#")]
	
	]
}

DivsEnd:`)
TextHidden( p_AgencyInfo, p_business, p_BZ_Certificate, p_BZ_ContractView, p_BZ_Vacancies, p_CentralBank, p_CitizenPage, p_CompanyDetails, p_shops)
SetVar()
TextHidden( )
SetVar()
TextHidden( )
SetVar(`d_Export0_accounts #= contract Export0_accounts {
func action {
	var tblname, fields string
	tblname = Table("accounts")
	fields = "onhold,agency_id,citizen_id,company_id,amount"
	DBInsert(tblname, fields, "0","1","0","0","0")
	}
}`,
`d_Export0_government_agencies #= contract Export0_government_agencies {
func action {
	var tblname, fields string
	tblname = Table("government_agencies")
	fields = "name,opened_time"
	DBInsert(tblname, fields, "Tax Agency","2017-03-03T14:24:42Z")
	}
}`,
`d_Export0_occupations #= contract Export0_occupations {
func action {
	var tblname, fields string
	tblname = Table("occupations")
	fields = "code,name"
	DBInsert(tblname, fields, "1","Analyst Programmer")
	DBInsert(tblname, fields, "2","Developer Programmer")
	DBInsert(tblname, fields, "3","Electronics Engineer")
	DBInsert(tblname, fields, "4","Accountant (General)")
	DBInsert(tblname, fields, "5","Construction Project Manager")
	DBInsert(tblname, fields, "6","Computer Network Engineer")
	DBInsert(tblname, fields, "7","Electrician (General)")
	}
}`)
TextHidden( d_Export0_accounts, d_Export0_government_agencies, d_Export0_occupations)
SetVar(`ap_government #= 
Div(clearfix)

Divs(md-4, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Add Government Agency")
        
        Divs(form-group)
            Label("Government Agency Name")
            Input(Name, "form-control input-lg m-b")
        DivsEnd:
        
        TxButton{ Contract: NewGovernmentAgency, Name: Add,Inputs: "Name=Name",OnSuccess: "template,government" }
    FormEnd:
DivsEnd:

     Divs(md-8, panel panel-default panel-body)
    
        Legend(" ", "Government Agencies")
Table {

    Table: #state_id#_government_agencies
	Order: id
	Columns: [
	[ID, #id#],
	[Name, #name#],
	[Registration date, Date(#opened_time#, DD.MM.YYYY)],
	[Details, BtnPage(AgencyInfo,Info,"AgencyId:#id#",btn btn-info btn-pill-right)]
	]
}
    
       DivsEnd:`,
`ap_dashboard_default #= Divs(md-6)
     Divs()
     WiBalance(GetOne(amount, #state_id#_accounts, "citizen_id", #citizen#), StateVal(currency_name) )
     DivsEnd:
     Divs()
     WiAccount( GetOne(id, #state_id#_accounts, "citizen_id", #citizen#) )
     DivsEnd:
     DivsEnd:
     
    
 Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Send Money")
        
        Divs(form-group)
            Label("Account ID")
            Select(AccountId, #state_id#_accounts.id, "form-control  m-b")
        DivsEnd:
        
        Divs(form-group)
            Label("Amount")
            InputMoney(Amount, "form-control")
        DivsEnd:
        
        TxButton{ Contract: SendMoney, Name: Send, Inputs: "RecipientAccountId=AccountId, Amount=Amount", OnSuccess: "template,dashboard_default,global:0" }
    FormEnd:
DivsEnd:
    
     Divs(md-6, panel panel-default panel-body)
    
        Legend(" ", "Companies")
Table {

    Table: #state_id#_companies
    Where: owner_citizen_id=#citizen#
	Order: id
	Columns: [
	[ID, #id#],
	[Name, #name#],
	[Registration date, Date(#opened_time#, DD.MM.YYYY)],
	[Info,  BtnPage(CompanyDetails,Company page,"CompanyId:#id#",btn btn-info btn-pill-right)]
	]
}
    
       DivsEnd:

DivsEnd:

DivsEnd:
 Divs(md-6, panel panel-default panel-body data-sweet-alert)
    Form()
        Legend(" ", "Jobs on contract")
        Table {
    Class: table-striped table-hover
    Table: #state_id#_jobs
	Order: id DESC
	Where: status=1 and citizen_id=#citizen#
	Columns: [[Occupation, #occupation_name#],
	[Conmany, BtnPage(CompanyDetails,#company_name#,"CompanyId:'#company_id#'",btn btn-info btn-pill-right)],
	[Date, Date(#date_start#, DD.MM.YYYY)],
	[Smart Contract, BtnPage(BZ_ContractView,View Contract,"Agree:0,Back:'CitizenPage',CitizenId:'#citizen_id#',CompanyId:'#company_id#', ContractId:#smart_contract_id#,CompanyId:#company_id#,ContractName:'#smart_contract_name#'",btn btn-info btn-pill-right)]

	]
}
DivsEnd:`)
TextHidden( ap_government, ap_dashboard_default)
SetVar(`am_menu_default #= MenuItem(Business, business)
MenuBack(Government dashboard,government)`,
`am_government #= MenuItem(CentralBank, CentralBank)`)
TextHidden( am_menu_default, am_government)
Json(`Head: "Money & Busines",
Desc: "Money & Busines",
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
			table_name : "accounts",
			columns: '[["onhold", "int64", "1"],["agency_id", "int64", "1"],["citizen_id", "int64", "1"],["company_id", "int64", "1"],["amount", "money", "0"]]',
			permissions: "ContractConditions(\"MainCondition\")"
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
			table_name : "buysell",
			columns: '[["seller_company_id", "int64", "1"],["date", "time", "1"],["price", "money", "1"],["item_id", "int64", "1"],["name_item", "text", "0"],["sales_tax", "int64", "1"],["buyer_company_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_buysell",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"BuyItem\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "buyer_company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "seller_company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "date",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "price",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "item_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "name_item",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_buysell",
			column_name: "sales_tax",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "bz_certificates",
			columns: '[["certificate", "int64", "1"],["tl_conclusion", "int64", "1"],["item_id", "int64", "1"],["item_code", "int64", "1"],["company_id", "int64", "1"],["company_name", "text", "0"],["testing_laboratory", "text", "0"],["item_name", "text", "0"],["added_time", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_bz_certificates",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"NewItem\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "testing_laboratory",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "item_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "added_time",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "company_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "tl_conclusion",
			permissions: "ContractAccess(\"BZ_TestLabConfirm\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "item_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "item_code",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_bz_certificates",
			column_name: "certificate",
			permissions: "ContractAccess(\"BZ_CertificateConfirm\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "companies",
			columns: '[["name", "hash", "1"],["opened_time", "time", "1"],["owner_citizen_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_companies",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"NewCompany\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_companies",
			column_name: "owner_citizen_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_companies",
			column_name: "name",
			permissions: "ContractConditions(\"CompanyConditions\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_companies",
			column_name: "opened_time",
			permissions: "false",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "government_agencies",
			columns: '[["name", "hash", "1"],["opened_time", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },

{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_government_agencies",
			column_name: "name",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_government_agencies",
			column_name: "opened_time",
			permissions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "items",
			columns: '[["added_time", "time", "1"],["company_id", "int64", "1"],["certificate", "int64", "1"],["name", "hash", "1"],["price", "money", "1"],["code_item", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_items",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"NewItem\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "name",
			permissions: "ContractConditions(\"CompanyConditions\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "price",
			permissions: "ContractConditions(\"CompanyConditions\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "code_item",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "added_time",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "company_id",
			permissions: "false",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_items",
			column_name: "certificate",
			permissions: "ContractAccess(\"BZ_CertificateConfirm\")",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "job_vacancy_application",
			columns: '[["date", "time", "1"],["company_id", "int64", "1"],["citizen_name", "text", "0"],["occupation_id", "int64", "1"],["occupation_name", "hash", "1"],["status", "int64", "1"],["citizen_id", "int64", "1"],["vacancy_id", "int64", "1"],["smart_contract_id", "int64", "1"],["smart_contract_name", "hash", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_job_vacancy_application",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"BZ_VacancyAgree\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "date",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "status",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "citizen_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "occupation_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "smart_contract_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "vacancy_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "occupation_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_job_vacancy_application",
			column_name: "smart_contract_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "jobs",
			columns: '[["vacancy_id", "int64", "1"],["company_name", "text", "0"],["smart_contract_name", "hash", "1"],["date_company_agreement", "time", "1"],["date_citizen_agreement", "time", "1"],["status", "int64", "1"],["date_finish", "time", "1"],["occupation_id", "int64", "1"],["occupation_name", "hash", "1"],["smart_contract_id", "int64", "1"],["citizen_id", "int64", "1"],["company_id", "int64", "1"],["date_start", "time", "1"],["citizen_name", "text", "0"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_jobs",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"BZ_AcceptApplication\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "vacancy_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "date_finish",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "smart_contract_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "date_start",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "occupation_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "smart_contract_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "status",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "company_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "occupation_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "date_citizen_agreement",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "citizen_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs",
			column_name: "date_company_agreement",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "jobs_application",
			columns: '[["citizen_name", "text", "0"],["occupation_id", "int64", "1"],["occupation_name", "hash", "1"],["date", "time", "1"],["status", "int64", "1"],["citizen_id", "int64", "1"],["date_finish", "time", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_jobs_application",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"BZ_JobApplication\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "occupation_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "date",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "status",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "citizen_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "date_finish",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "citizen_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_application",
			column_name: "occupation_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "jobs_vacancies",
			columns: '[["date_finish", "time", "1"],["company_name", "hash", "1"],["occupation_id", "int64", "1"],["occupation_name", "hash", "1"],["smart_contract_id", "int64", "1"],["smart_contract_name", "hash", "1"],["date", "time", "1"],["status", "int64", "1"],["company_id", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_jobs_vacancies",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractAccess(\"BZ_JobVacancy\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "date",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "occupation_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "occupation_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "smart_contract_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "status",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "company_id",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "date_finish",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "company_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_jobs_vacancies",
			column_name: "smart_contract_name",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 0,
			table_name : "occupations",
			columns: '[["name", "hash", "1"],["code", "int64", "1"]]',
			permissions: "ContractConditions(\"MainCondition\")"
			}
	   },

{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_occupations",
			column_name: "code",
			permissions: "true",
			}
	   },
{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "#state_id#_occupations",
			column_name: "name",
			permissions: "true",
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "CompanyConditions",
			value: $("#sc_CompanyConditions").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SL_ItemCertificate",
			value: $("#sc_SL_ItemCertificate").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "SL_SalesTax",
			value: $("#sc_SL_SalesTax").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
	{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BuyItemCompanyContracts",
			value: $("#sc_BuyItemCompanyContracts").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },   
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "Export0_accounts",
			value: $("#d_Export0_accounts").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "Export0_government_agencies",
			value: $("#d_Export0_government_agencies").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "Export0_occupations",
			value: $("#d_Export0_occupations").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BuyItem",
			value: $("#sc_BuyItem").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },

{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_AcceptApplication",
			value: $("#sc_BZ_AcceptApplication").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_CertificateConfirm",
			value: $("#sc_BZ_CertificateConfirm").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_JobApplication",
			value: $("#sc_BZ_JobApplication").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_JobRecruitment",
			value: $("#sc_BZ_JobRecruitment").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_JobVacancy",
			value: $("#sc_BZ_JobVacancy").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_RejectApplication",
			value: $("#sc_BZ_RejectApplication").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_TestLabConfirm",
			value: $("#sc_BZ_TestLabConfirm").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "BZ_VacancyAgree",
			value: $("#sc_BZ_VacancyAgree").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "JobContract_Test",
			value: $("#sc_JobContract_Test").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 0,
			name: "NewGovernmentAgency",
			value: $("#sc_NewGovernmentAgency").val(),
			conditions: "ContractConditions(\"MainCondition\")"
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
			conditions: "ContractConditions(\"MainCondition\")"
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
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_occupations",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "#state_id#_government_agencies",
			general_update: "ContractConditions(\"MainCondition\")",
			insert: "ContractConditions(\"MainCondition\")",
			new_column: "ContractConditions(\"MainCondition\")",
			}
	   },

{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "AgencyInfo",
			menu: "government",
			value: $("#p_AgencyInfo").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "BZ_Certificate",
			menu: "menu_default",
			value: $("#p_BZ_Certificate").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "BZ_ContractView",
			menu: "menu_default",
			value: $("#p_BZ_ContractView").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "BZ_Vacancies",
			menu: "menu_default",
			value: $("#p_BZ_Vacancies").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CentralBank",
			menu: "government",
			value: $("#p_CentralBank").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "CitizenPage",
			menu: "menu_default",
			value: $("#p_CitizenPage").val(),
			global: 0,
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
			conditions: "ContractConditions(\"MainCondition\")",
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
