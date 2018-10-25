package updates

var M1O7 = `
ALTER TABLE "1_history" 
ADD COLUMN created bigint;

UPDATE "1_history" 
SET created = extract(epoch from created_at) * 1000;

ALTER TABLE "1_history" 
ALTER COLUMN created_at TYPE bigint USING (created), 
ALTER COLUMN created_at DROP DEFAULT;

UPDATE "1_history" 
SET created_at = created;

ALTER TABLE "1_history"
DROP COLUMN created;
--===============================================

ALTER TABLE "1_roles" 
ADD COLUMN created bigint,
ADD COLUMN deleted_dt bigint;

UPDATE "1_roles" 
SET created = extract(epoch from date_created) * 1000,
deleted_dt = extract(epoch from date_deleted) * 1000;

ALTER TABLE "1_roles" 
ALTER COLUMN date_deleted TYPE bigint USING (deleted_dt),
ALTER COLUMN date_created TYPE bigint USING (created);

ALTER TABLE "1_roles"
DROP COLUMN deleted_dt,
DROP COLUMN created;
`
