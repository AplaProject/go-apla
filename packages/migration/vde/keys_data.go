package vde

import "github.com/GenesisKernel/go-genesis/packages/consts"

var keysDataSQL = `
INSERT INTO "1_keys" (id, pub, blocked, ecosystem) 
VALUES (` + consts.GuestKey + `, '', 1, '%[1]d');
`
