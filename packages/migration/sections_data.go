package migration

var sectionsDataSQL = `
INSERT INTO "1_sections" ("id","title","urlname","page","roles_access", "status", "ecosystem") VALUES
('1', 'Home', 'home', 'default_page', '[]', 2, '%[1]d'),
('2', 'Developer', 'admin', 'admin_index', '[]', 1, '%[1]d');
`
