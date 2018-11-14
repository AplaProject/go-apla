package updates

var M107 = `
ALTER TABLE "1_history" 
ALTER COLUMN created_at DROP DEFAULT;
--===============================================
ALTER TABLE "1_roles" 
ALTER COLUMN date_created DROP DEFAULT,
ALTER COLUMN date_deleted DROP DEFAULT;
`
