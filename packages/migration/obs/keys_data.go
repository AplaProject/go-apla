package obs

import (
	"github.com/AplaProject/go-apla/packages/consts"
)

var keysDataSQL = `
INSERT INTO "1_keys" (id, pub, blocked, ecosystem) 
VALUES (` + consts.GuestKey + `, decode('` + consts.GuestPublic + `', 'hex'), 1, '%[1]d');
`
