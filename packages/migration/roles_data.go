package migration

var rolesDataSQL = `INSERT INTO "%[1]d_members" ("id", "member_name") VALUES('%[2]d', 'founder');`
