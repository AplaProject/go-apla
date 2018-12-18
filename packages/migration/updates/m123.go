package updates

var M123 = `
ALTER TABLE "1_keys" ADD COLUMN IF NOT EXISTS "read_only" bigint NOT NULL DEFAULT '0';
`
