package translators_test

import (
	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

var _ fizz.Translator = (*translators.Postgres)(nil)
var pgt = translators.NewPostgres()

func (p *PostgreSQLSuite) Test_Postgres_CreateTable() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"id" SERIAL NOT NULL,
PRIMARY KEY("id"),
"first_name" VARCHAR (255) NOT NULL,
"last_name" VARCHAR (255) NOT NULL,
"email" VARCHAR (20) NOT NULL,
"permissions" jsonb,
"age" integer DEFAULT '40',
"raw" bytea NOT NULL,
"company_id" UUID NOT NULL DEFAULT uuid_generate_v1(),
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("id", "integer", {"primary": true})
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("raw", "blob", {})
		t.Column("company_id", "uuid", {"default_raw": "uuid_generate_v1()"})
	}
	`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_CreateTable_UUID() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"first_name" VARCHAR (255) NOT NULL,
"last_name" VARCHAR (255) NOT NULL,
"email" VARCHAR (20) NOT NULL,
"permissions" jsonb,
"age" integer DEFAULT '40',
"integer" integer NOT NULL,
"float" DECIMAL NOT NULL,
"bytes" bytea NOT NULL,
"strings" varchar[] NOT NULL,
"floats" decimal[] NOT NULL,
"ints" integer[] NOT NULL,
"jason" jsonb NOT NULL,
"mydecimal" DECIMAL NOT NULL,
"mydecimal2" DECIMAL(5,2) NOT NULL,
"uuid" UUID NOT NULL,
PRIMARY KEY("uuid"),
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("integer", "integer", {})
		t.Column("float", "float", {})
		t.Column("bytes", "[]byte", {})
		t.Column("strings", "[]string", {})
		t.Column("floats", "[]float", {})
		t.Column("ints", "[]int", {})
		t.Column("jason", "json", {})
		t.Column("mydecimal", "decimal", {})
		t.Column("mydecimal2", "decimal", {"precision": 5, "scale": 2})
		t.Column("uuid", "uuid", {"primary": true})
	}
	`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_CreateTable_UUID_With_Default() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"uuid" UUID NOT NULL DEFAULT uuid_generate_v4(),
PRIMARY KEY("uuid")
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("uuid", "uuid", {"primary": true, "default_raw": "uuid_generate_v4()"})
		t.DisableTimestamps()
	}
	`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_CreateTable_Cant_Set_PK_To_Nullable() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"uuid" UUID NOT NULL,
PRIMARY KEY("uuid")
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("uuid", "uuid", {"primary": true, "null": true})
		t.DisableTimestamps()
	}
	`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_CreateTables_WithForeignKeys() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"id" SERIAL NOT NULL,
PRIMARY KEY("id"),
"email" VARCHAR (20) NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL
);
CREATE TABLE "profiles" (
"id" SERIAL NOT NULL,
PRIMARY KEY("id"),
"user_id" INT NOT NULL,
"first_name" VARCHAR (255) NOT NULL,
"last_name" VARCHAR (255) NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL,
FOREIGN KEY ("user_id") REFERENCES "users" ("id")
);`

	res, err := fizz.AString(`
	create_table("users") {
		t.Column("id", "INT", {"primary": true})
		t.Column("email", "string", {"size":20})
	}
	create_table("profiles") {
		t.Column("id", "INT", {"primary": true})
		t.Column("user_id", "INT", {})
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.ForeignKey("user_id", {"users": ["id"]}, {})
	}
	`, pgt)
	r.NoError(err)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_CreateTables_WithCompositePrimaryKey() {
	r := p.Require()
	ddl := `CREATE TABLE "user_profiles" (
"user_id" INT NOT NULL,
"profile_id" INT NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL,
PRIMARY KEY("user_id", "profile_id")
);`

	res, _ := fizz.AString(`
	create_table("user_profiles") {
		t.Column("user_id", "INT")
		t.Column("profile_id", "INT")
		t.PrimaryKey("user_id", "profile_id")
	}
	`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_DropTable() {
	r := p.Require()

	ddl := `DROP TABLE "users";`

	res, _ := fizz.AString(`drop_table("users")`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_RenameTable() {
	r := p.Require()

	ddl := `ALTER TABLE "users" RENAME TO "people";`

	res, _ := fizz.AString(`rename_table("users", "people")`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_RenameTable_NotEnoughValues() {
	r := p.Require()

	_, err := pgt.RenameTable([]fizz.Table{})
	r.Error(err)
}

func (p *PostgreSQLSuite) Test_Postgres_ChangeColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "mytable" ALTER COLUMN "mycolumn" TYPE VARCHAR (50), ALTER COLUMN "mycolumn" SET NOT NULL, ALTER COLUMN "mycolumn" SET DEFAULT 'foo';`

	res, _ := fizz.AString(`change_column("mytable", "mycolumn", "string", {"default": "foo", "size": 50})`, pgt)

	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "mytable" ADD COLUMN "mycolumn" VARCHAR (50) NOT NULL DEFAULT 'foo';`

	res, _ := fizz.AString(`add_column("mytable", "mycolumn", "string", {"default": "foo", "size": 50})`, pgt)

	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_DropColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "table_name" DROP COLUMN "column_name";`

	res, _ := fizz.AString(`drop_column("table_name", "column_name")`, pgt)

	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_RenameColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "table_name" RENAME COLUMN "old_column" TO "new_column";`

	res, _ := fizz.AString(`rename_column("table_name", "old_column", "new_column")`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddIndex() {
	r := p.Require()
	ddl := `CREATE INDEX "table_name_column_name_idx" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {})`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddIndex_Unique() {
	r := p.Require()
	ddl := `CREATE UNIQUE INDEX "table_name_column_name_idx" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {"unique": true})`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddIndex_MultiColumn() {
	r := p.Require()
	ddl := `CREATE INDEX "table_name_col1_col2_col3_idx" ON "table_name" (col1, col2, col3);`

	res, _ := fizz.AString(`add_index("table_name", ["col1", "col2", "col3"], {})`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddIndex_CustomName() {
	r := p.Require()
	ddl := `CREATE INDEX "custom_name" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {"name": "custom_name"})`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_DropIndex() {
	r := p.Require()
	ddl := `DROP INDEX "my_idx";`

	res, _ := fizz.AString(`drop_index("users", "my_idx")`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_RenameIndex() {
	r := p.Require()

	ddl := `ALTER INDEX "old_ix" RENAME TO "new_ix";`

	res, _ := fizz.AString(`rename_index("table", "old_ix", "new_ix")`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_AddForeignKey() {
	r := p.Require()

	ddl := `ALTER TABLE "profiles" ADD CONSTRAINT "profiles_users_id_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id");`

	res, _ := fizz.AString(`add_foreign_key("profiles", "user_id", {"users": ["id"]}, {})`, pgt)
	r.Equal(ddl, res)
}

func (p *PostgreSQLSuite) Test_Postgres_DropForeignKey() {
	r := p.Require()

	ddl := `ALTER TABLE "profiles" DROP CONSTRAINT "profiles_users_id_fk";`

	res, _ := fizz.AString(`drop_foreign_key("profiles", "profiles_users_id_fk", {})`, pgt)
	r.Equal(ddl, res)
}
