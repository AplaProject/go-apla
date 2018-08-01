// Code generated by go generate; DO NOT EDIT.

package migration

var contractsDataSQL = `
INSERT INTO "%[1]d_contracts" (id, name, value, conditions, app_id, wallet_id)
VALUES
	(next_id('%[1]d_contracts'), 'MainCondition', 'contract MainCondition {
	conditions {
		if EcosysParam("founder_account")!=$key_id
		{
			warning "Sorry, you do not have access to this action."
		}
	}
}
', 'ContractConditions("MainCondition")', 1, %[2]d);
`