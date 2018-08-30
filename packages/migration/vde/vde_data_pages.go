package vde

var pagesDataSQL = `
INSERT INTO "%[1]d_pages" ("id","name","value","menu","conditions") VALUES('1', 'default_page', '', 'admin_menu', 'true'),('2','admin_index','','admin_menu','true');
`
