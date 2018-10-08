package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/migration/vde"
)

var membersDataSQL = `
	INSERT INTO "1_members" ("id", "member_name", "ecosystem") VALUES('%[2]d', 'founder', '%[1]d'),
	('` + vde.GuestKey + `', 'guest', '%[1]d');

`
