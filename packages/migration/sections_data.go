package migration

var sectionsDataSQL = `
INSERT INTO "1_sections" ("id","title","urlname","page","roles_access", "status", "ecosystem") VALUES
(next_id('1_sections'), 'Home', 'home', 'default_page', '[]', 2, '%[1]d'),
(next_id('1_sections'), 'Admin', 'admin', 'admin_index', '[]', 1, '%[1]d'),
(next_id('1_sections'), 'Developer', 'developer', 'developer_index', '[]', 1, '%[1]d');
`
