package translators_test

import (
	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

var _ fizz.Translator = (*translators.MsSqlServer)(nil)
var sqlsrv = translators.NewMsSqlServer()

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_SchemaMigration() {
	r := p.Require()
	ddl := `CREATE TABLE schema_migrations (
version NVARCHAR (255) NOT NULL
);
CREATE UNIQUE INDEX version_idx ON schema_migrations (version);`

	res, err := sqlsrv.CreateTable(fizz.Table{
		Name: "schema_migrations",
		Columns: []fizz.Column{
			{Name: "version", ColType: "string"},
		},
		Indexes: []fizz.Index{
			{Name: "version_idx", Columns: []string{"version"}, Unique: true},
		},
	})
	r.NoError(err)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_CreateTable() {
	r := p.Require()
	ddl := `CREATE TABLE users (
id INT PRIMARY KEY IDENTITY(1,1),
first_name NVARCHAR (255) NOT NULL,
last_name NVARCHAR (255) NOT NULL,
email NVARCHAR (20) NOT NULL,
permissions text,
age INT CONSTRAINT DF_users_age DEFAULT '40',
raw VARBINARY(MAX) NOT NULL,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "text", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("raw", "blob", {})
	}
	`, sqlsrv)

	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_CreateTable_UUID() {
	r := p.Require()
	ddl := `CREATE TABLE users (
first_name NVARCHAR (255) NOT NULL,
last_name NVARCHAR (255) NOT NULL,
email NVARCHAR (20) NOT NULL,
permissions text,
age INT CONSTRAINT DF_users_age DEFAULT '40',
uuid uniqueidentifier PRIMARY KEY,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "text", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("uuid", "uuid", {"primary": true})
	}
	`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_CreateTables_WithForeignKeys() {
	r := p.Require()
	ddl := `CREATE TABLE users (
id INT PRIMARY KEY IDENTITY(1,1),
email NVARCHAR (20) NOT NULL,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);
CREATE TABLE profiles (
id INT PRIMARY KEY IDENTITY(1,1),
user_id INT NOT NULL,
first_name NVARCHAR (255) NOT NULL,
last_name NVARCHAR (255) NOT NULL,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);
ALTER TABLE profiles ADD CONSTRAINT profiles_users_id_fk FOREIGN KEY (user_id) REFERENCES users (id);`

	res, _ := fizz.AString(`
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
	`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_CreateTables_WithCompositePrimaryKey() {
	r := p.Require()
	ddl := `CREATE TABLE user_profiles (
user_id INT NOT NULL,
profile_id INT NOT NULL,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
PRIMARY KEY([user_id], [profile_id])
);`

	res, _ := fizz.AString(`
	create_table("user_profiles") {
		t.Column("user_id", "INT")
		t.Column("profile_id", "INT")
		t.PrimaryKey("user_id", "profile_id")
	}
	`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_DropTable() {
	r := p.Require()

	ddl := `DROP TABLE users;`

	res, _ := fizz.AString(`drop_table("users")`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_RenameTable() {
	r := p.Require()

	ddl := `EXEC sp_rename 'users', 'people';`

	res, _ := fizz.AString(`rename_table("users", "people")`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_RenameTable_NotEnoughValues() {
	r := p.Require()

	_, err := sqlsrv.RenameTable([]fizz.Table{})
	r.Error(err)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_ChangeColumn() {
	r := p.Require()
	ddl := `ALTER TABLE users ALTER COLUMN mycolumn NVARCHAR (50) NOT NULL
ALTER TABLE users DROP CONSTRAINT IF EXISTS DF_users_mycolumn;
ALTER TABLE users ADD CONSTRAINT DF_users_mycolumn DEFAULT 'foo' FOR mycolumn;`

	res, _ := fizz.AString(`change_column("users", "mycolumn", "string", {"default": "foo", "size": 50})`, sqlsrv)

	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddColumn() {
	r := p.Require()
	ddl := `ALTER TABLE users ADD mycolumn NVARCHAR (50) NOT NULL CONSTRAINT DF_users_mycolumn DEFAULT 'foo';`

	res, _ := fizz.AString(`add_column("users", "mycolumn", "string", {"default": "foo", "size": 50})`, sqlsrv)

	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_DropColumn() {
	r := p.Require()
	ddl := `ALTER TABLE users DROP COLUMN mycolumn;`

	res, _ := fizz.AString(`drop_column("users", "mycolumn")`, sqlsrv)

	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_RenameColumn() {
	r := p.Require()
	ddl := `EXEC sp_rename 'users.email', 'email_address', 'COLUMN';`

	res, _ := fizz.AString(`rename_column("users", "email", "email_address")`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddIndex() {
	r := p.Require()
	ddl := `CREATE INDEX users_email_idx ON users (email);`

	res, _ := fizz.AString(`add_index("users", "email", {})`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddIndex_Unique() {
	r := p.Require()
	ddl := `CREATE UNIQUE INDEX users_email_idx ON users (email);`

	res, _ := fizz.AString(`add_index("users", "email", {"unique": true})`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddIndex_MultiColumn() {
	r := p.Require()
	ddl := `CREATE INDEX users_id_email_idx ON users (id, email);`

	res, _ := fizz.AString(`add_index("users", ["id", "email"], {})`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddIndex_CustomName() {
	r := p.Require()
	ddl := `CREATE INDEX email_index ON users (email);`

	res, _ := fizz.AString(`add_index("users", "email", {"name": "email_index"})`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_DropIndex() {
	r := p.Require()
	ddl := `DROP INDEX email_idx ON users;`

	res, _ := fizz.AString(`drop_index("users", "email_idx")`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_RenameIndex() {
	r := p.Require()

	ddl := `EXEC sp_rename 'users.email_idx', 'email_address_ix', 'INDEX';`

	res, _ := fizz.AString(`rename_index("users", "email_idx", "email_address_ix")`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_AddForeignKey() {
	r := p.Require()
	ddl := `ALTER TABLE profiles ADD CONSTRAINT profiles_users_id_fk FOREIGN KEY (user_id) REFERENCES users (id);`

	res, _ := fizz.AString(`add_foreign_key("profiles", "user_id", {"users": ["id"]}, {})`, sqlsrv)
	r.Equal(ddl, res)
}

func (p *MsSqlServerSQLSuite) Test_MsSqlServer_DropForeignKey() {
	r := p.Require()
	ddl := `ALTER TABLE profiles DROP CONSTRAINT  profiles_users_id_fk;`

	res, _ := fizz.AString(`drop_foreign_key("profiles", "profiles_users_id_fk", {})`, sqlsrv)
	r.Equal(ddl, res)
}
