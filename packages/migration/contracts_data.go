package migration

var contractsDataSQL = `
INSERT INTO "%[1]d_contracts" ("id", "name", "value", "wallet_id","active", "conditions") VALUES 
		('1','MainCondition','contract MainCondition {
		  conditions {
			if EcosysParam("founder_account")!=$key_id
			{
			  warning "Sorry, you do not have access to this action."
			}
		  }
		}', '%[2]d', '0', 'ContractConditions("MainCondition")');
`
