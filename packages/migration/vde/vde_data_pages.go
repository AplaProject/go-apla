package vde

var pagesDataSQL = `
INSERT INTO "1_pages" ("id","name","value","menu","conditions","ecosystem") VALUES(next_id('1_pages'), 'default_page', '', 'admin_menu', 'true','%[1]d'),(next_id('1_pages'),'admin_index','','admin_menu','true','%[1]d');
`
