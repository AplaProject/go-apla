package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/types"
)

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

INSERT INTO "1_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator","roles_access") VALUES
	('3','', 'Apla Consensus asbl', '0', '3', NOW(), '{}', '{"rids": "1"}'),
	('4','', 'Candidate for validators', '0', '3', NOW(), '{}', '{"rids": "1"}'),
	('5','', 'Validator', '0', '3', NOW(), '{}', '{"rids": "1"}'),
	('6','', 'Investor with voting rights', '0', '3', NOW(), '{}', '{"rids": "1"}'),
	('7','', 'Delegate', '0', '3', NOW(), '{}', '{"rids": "1"}');
`

var firstEcosystemData = []Row{
	{
		Registry:   &types.Registry{Name: "ecosystem", Ecosystem: &types.Ecosystem{Name: "1"}, Type: types.RegistryTypePrimary},
		PrimaryKey: "1",
		//Data:       model.Ecosystem{ID: 1, Name: "platform ecosystem"},
	},
}
