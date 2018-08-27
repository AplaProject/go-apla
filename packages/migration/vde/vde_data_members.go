package vde

var membersDataSQL = `
INSERT INTO "1_members" ("id", "member_name", "ecosystem") VALUES('%[2]d', 'founder', '%[1]d'),
('` + GuestKey + `', 'guest', '%[1]d');`
