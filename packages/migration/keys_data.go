package migration

import "github.com/AplaProject/go-apla/packages/consts"

var keysDataSQL = `INSERT INTO "1_keys" (id, pub, blocked, ecosystem) VALUES (` + consts.GuestKey + `, '', 1, '%[1]d');`
