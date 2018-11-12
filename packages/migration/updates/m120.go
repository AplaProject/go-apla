package updates

var M120 = `
ALTER TABLE "1_history" ADD COLUMN "created_new" bigint NOT NULL DEFAULT '0';
UPDATE "1_history" SET created_new = EXTRACT(EPOCH FROM created_at::timestamp with time zone);
ALTER TABLE "1_history" DROP COLUMN "created_at";
ALTER TABLE "1_history" RENAME COLUMN "created_new" TO "created_at";
`
