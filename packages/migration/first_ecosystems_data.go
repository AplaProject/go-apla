package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/types"
)

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

INSERT INTO "1_applications" (id, name, conditions, ecosystem) VALUES (next_id('1_applications'), 'System', 'ContractConditions("MainCondition")', '1');
`

var firstEcosystemData = []Row{
	{
		Registry:   &types.Registry{Name: "ecosystem", Ecosystem: &types.Ecosystem{Name: "1"}, Type: types.RegistryTypePrimary},
		PrimaryKey: "1",
		//Data:       model.Ecosystem{ID: 1, Name: "platform ecosystem"},
	},
}
