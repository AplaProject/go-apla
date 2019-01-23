package obs

import (
	"github.com/AplaProject/go-apla/packages/consts"
)

var keysDataSQLOBS = `
INSERT INTO "1_keys" (id, pub, blocked, read_only, ecosystem) VALUES (` + consts.GuestKey + `, decode('` + consts.GuestPublic + `', 'HEX'), 0, 0, '%[1]d');`
