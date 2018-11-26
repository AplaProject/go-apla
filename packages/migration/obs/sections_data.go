package obs

var sectionsDataSQL = `
INSERT INTO "1_sections" ("id","title","urlname","page","roles_access", "delete", "ecosystem") VALUES
('1', 'Home', 'home', 'default_page', '', 0, '%[1]d');`
