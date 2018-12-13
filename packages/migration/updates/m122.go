package updates

var M122 = `
ALTER TABLE "1_ecosystems" ADD COLUMN "emission_amount" jsonb;
ALTER TABLE "1_ecosystems" ADD COLUMN "token_title" varchar(255);
ALTER TABLE "1_ecosystems" ADD COLUMN "type_emission" bigint NOT NULL DEFAULT '0';
ALTER TABLE "1_ecosystems" ADD COLUMN "type_withdraw" bigint NOT NULL DEFAULT '0';

UPDATE "1_tables" SET permissions = '
{	"insert": "ContractAccess(\"@1NewEcosystem\")",
	"update": "ContractAccess(\"@1EditEcosystemName\",\"@1VotingDecisionCheck\",\"@1EcManageInfo\",\"@1TeCreate\",\"@1TeChange\",\"@1TeBurn\")",
	"new_column": "ContractConditions(\"@1AdminCondition\")"
}', columns = '
{	"name": "ContractAccess(\"@1EditEcosystemName\")",
	"info": "ContractAccess(\"@1EcManageInfo\")",
	"is_valued": "ContractAccess(\"@1VotingDecisionCheck\")",
	"emission_amount": "ContractAccess(\"@1TeCreate\",\"@1TeBurn\")",
	"token_title": "ContractAccess(\"@1TeCreate\")",
	"type_emission": "ContractAccess(\"@1TeCreate\",\"@1TeChange\")",
	"type_withdraw": "ContractAccess(\"@1TeCreate\",\"@1TeChange\")"
}' WHERE name='ecosystems';

UPDATE "1_tables" SET permissions = '
{
	"insert": "true",
	"update": "ContractAccess(\"@1TokensTransfer\",\"@1TokensLockoutMember\",\"@1MultiwalletCreate\",\"@1TeCreate\",\"@1TeBurn\")",
	"new_column": "ContractConditions(\"@1AdminCondition\")"
}',	columns = '
{	"pub": "false",
    "amount": "ContractAccess(\"@1TokensTransfer\",\"@1TeCreate\",\"@1TeBurn\")",
    "maxpay": "ContractConditions(\"@1AdminCondition\")",
    "deleted": "ContractConditions(\"@1AdminCondition\")",
    "blocked": "ContractAccess(\"@1TokensLockoutMember\")"
}' WHERE name='keys';

UPDATE "1_tables" SET permissions = '
{
	"insert": "ContractAccess(\"@1TokensTransfer\",\"@1TeCreate\",\"@1TeBurn\")",
	"update": "ContractConditions(\"@1AdminCondition\")",
	"new_column": "ContractConditions(\"@1AdminCondition\")"
}' WHERE name='history';
`
