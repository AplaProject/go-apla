package migration

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

INSERT INTO "1_applications" (id, name, conditions, ecosystem) VALUES (next_id('1_applications'), 'System', 'ContractConditions("MainCondition")', '1');
`
