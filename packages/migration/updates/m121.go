package updates

var M121 = `
insert into "1_parameters" (id, name, value, conditions, ecosystem) values
 (next_id('1_parameters'), 'error_page', '@1error_page', 'ContractConditions("@1DeveloperCondition")', 1);
 `
