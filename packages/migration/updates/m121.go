package updates

var M121 = `
insert into "1_parameters" (name, value, conditions, ecosystem) values
 ('error_page', '1@error_page', 'ContractConditions("@1DeveloperCondition")', 1);
 `
