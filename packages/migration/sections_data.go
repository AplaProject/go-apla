package migration

var sectionsDataSQL = `
INSERT INTO "%[1]d_sections" ("id","title","urlname","page","roles_access", "status") VALUES
('1', 'Home', 'home', 'default_page', '[]', 2),
('2', 'Developer', 'admin', 'admin_index', '[]', 1);
`
