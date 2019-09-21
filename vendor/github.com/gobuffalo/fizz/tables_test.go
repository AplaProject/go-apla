package fizz_test

import (
	"testing"

	"github.com/gobuffalo/fizz"
	"github.com/stretchr/testify/require"
)

func Test_Table_Stringer(t *testing.T) {
	r := require.New(t)

	expected :=
		`create_table("users") {
	t.Column("name", "string")
	t.Column("alive", "boolean", {null: true})
	t.Column("birth_date", "timestamp", {null: true})
	t.Column("bio", "text", {null: true})
	t.Column("price", "numeric", {default: "1.00", null: true})
	t.Column("email", "string", {default: "foo@example.com", size: 50})
	t.Timestamps()
}`

	table := fizz.NewTable("users", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("alive", "boolean", fizz.Options{
		"null": true,
	}))
	r.NoError(table.Column("birth_date", "timestamp", fizz.Options{
		"null": true,
	}))
	r.NoError(table.Column("bio", "text", fizz.Options{
		"null": true,
	}))
	r.NoError(table.Column("price", "numeric", fizz.Options{
		"null":    true,
		"default": "1.00",
	}))
	r.NoError(table.Column("email", "string", fizz.Options{
		"size":    50,
		"default": "foo@example.com",
	}))

	r.Equal(expected, table.String())
}

func Test_Table_StringerOpts(t *testing.T) {
	r := require.New(t)

	// Timestamps
	expected :=
		`create_table("users") {
	t.Column("name", "string")
	t.Timestamps()
}`

	table := fizz.NewTable("users", map[string]interface{}{
		"timestamps": true,
	})
	r.NoError(table.Column("name", "string", nil))

	r.Equal(expected, table.String())

	// Random option
	expected =
		`create_table("users", {myopt: "test"}) {
	t.Column("name", "string")
	t.Timestamps()
}`

	table = fizz.NewTable("users", map[string]interface{}{
		"myopt": "test",
	})
	r.NoError(table.Column("name", "string", nil))

	r.Equal(expected, table.String())
}

func Test_Table_StringerAutoDisableTimestamps(t *testing.T) {
	r := require.New(t)

	// Custom type timestamps
	expected :=
		`create_table("users") {
	t.Column("name", "string")
	t.Column("created_at", "int")
	t.Column("updated_at", "int")
}`

	table := fizz.NewTable("users", map[string]interface{}{
		"timestamps": true,
	})
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("created_at", "int", nil))
	r.NoError(table.Column("updated_at", "int", nil))

	r.Equal(expected, table.String())

	// only one timestamp override
	expected =
		`create_table("users") {
	t.Column("name", "string")
	t.Column("created_at", "int")
	t.Column("updated_at", "timestamp")
}`

	table = fizz.NewTable("users", map[string]interface{}{
		"timestamps": true,
	})
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("created_at", "int", nil))

	r.Equal(expected, table.String())

	// timestamp columns provided but same as default
	expected =
		`create_table("users") {
	t.Column("name", "string")
	t.Timestamps()
}`

	table = fizz.NewTable("users", map[string]interface{}{
		"timestamps": true,
	})
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("created_at", "timestamp", nil))
	r.NoError(table.Column("updated_at", "timestamp", nil))

	r.Equal(expected, table.String())
}

func Test_Table_StringerIndex(t *testing.T) {
	r := require.New(t)

	// Single column index
	expected :=
		`create_table("users") {
	t.Column("name", "string")
	t.Column("email", "string")
	t.Timestamps()
	t.Index("email", {name: "users_email_idx", unique: true})
}`

	table := fizz.NewTable("users", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("email", "string", nil))
	r.NoError(table.Index("email", fizz.Options{
		"unique": true,
	}))

	r.Equal(expected, table.String())

	// Multiple-column index
	expected =
		`create_table("users") {
	t.Column("name", "string")
	t.Column("email", "string")
	t.Timestamps()
	t.Index(["name", "email"], {name: "users_name_email_idx"})
}`

	table = fizz.NewTable("users", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("email", "string", nil))
	r.NoError(table.Index([]string{"name", "email"}, nil))

	r.Equal(expected, table.String())
}

func Test_Table_StringerForeignKey(t *testing.T) {
	r := require.New(t)

	// Single column
	expected :=
		`create_table("users_color") {
	t.Column("name", "string")
	t.Column("user_id", "int")
	t.Timestamps()
	t.ForeignKey("user_id", {"users": ["id"]}, {on_delete: "cascade"})
}`

	table := fizz.NewTable("users_color", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("user_id", "int", nil))
	r.NoError(table.ForeignKey("user_id",
		map[string]interface{}{
			"users": []interface{}{"id"},
		},
		fizz.Options{
			"on_delete": "cascade",
		}))

	r.Equal(expected, table.String())

	// Multiple columns
	expected =
		`create_table("users_color") {
	t.Column("name", "string")
	t.Column("user_id", "int")
	t.Timestamps()
	t.ForeignKey("user_id", {"users": ["id", "id2"]}, {on_delete: "cascade", on_update: "restrict"})
}`

	table = fizz.NewTable("users_color", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("user_id", "int", nil))
	r.NoError(table.ForeignKey("user_id",
		map[string]interface{}{
			"users": []interface{}{"id", "id2"},
		},
		fizz.Options{
			"on_delete": "cascade",
			"on_update": "restrict",
		}))

	r.Equal(expected, table.String())
}

func Test_Table_UnFizz(t *testing.T) {
	r := require.New(t)
	table := fizz.NewTable("users", nil)
	r.Equal(`drop_table("users")`, table.UnFizz())
}

func Test_Table_HasColumn(t *testing.T) {
	r := require.New(t)
	table := fizz.NewTable("users", nil)
	table.Column("firstname", "string", nil)
	table.Column("lastname", "string", nil)
	r.True(table.HasColumns("firstname", "lastname"))
	r.False(table.HasColumns("age"))
}

func Test_Table_ColumnNames(t *testing.T) {
	r := require.New(t)
	table := fizz.NewTable("users", nil)
	table.Column("firstname", "string", nil)
	table.Column("lastname", "string", nil)
	r.Equal([]string{"firstname", "lastname"}, table.ColumnNames())
}

func Test_Table_DuplicateColumn(t *testing.T) {
	r := require.New(t)
	table := fizz.NewTable("users", map[string]interface{}{})
	r.NoError(table.Column("name", "string", fizz.Options{}))
	r.Error(table.Column("name", "string", fizz.Options{}))
	r.Error(table.Column("name", "string", fizz.Options{
		"null": true,
	}))
}

func Test_Table_AddEmptyIndex(t *testing.T) {
	r := require.New(t)
	// Empty index
	table := fizz.NewTable("users", nil)
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("email", "string", nil))
	r.Error(table.Index([]string{}, nil))
}

func Test_Table_AddPrimaryKey(t *testing.T) {
	r := require.New(t)

	// Add single primary key
	expected :=
		`create_table("users") {
	t.Column("id", "int", {primary: true})
	t.Column("name", "string")
	t.Column("email", "string")
	t.Timestamps()
}`
	table := fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", fizz.Options{"primary": true}))
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("email", "string", nil))
	r.Equal(expected, table.String())

	table = fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", nil))
	r.NoError(table.Column("name", "string", nil))
	r.NoError(table.Column("email", "string", nil))
	r.NoError(table.PrimaryKey("id"))
	r.Equal(expected, table.String())

	// Add composite primary key
	expected =
		`create_table("user_privileges") {
	t.Column("user_id", "int")
	t.Column("privilege_id", "int")
	t.Timestamps()
	t.PrimaryKey("user_id", "privilege_id")
}`
	table = fizz.NewTable("user_privileges", nil)
	r.NoError(table.Column("user_id", "int", nil))
	r.NoError(table.Column("privilege_id", "int", nil))
	r.NoError(table.PrimaryKey("user_id", "privilege_id"))
	r.Equal(expected, table.String())
}

func Test_Table_AddPrimaryKey_Errors(t *testing.T) {
	r := require.New(t)

	// Primary key on unknown column
	table := fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", nil))
	r.Error(table.PrimaryKey("id2"))

	// Duplicate primary key
	table = fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", nil))
	r.NoError(table.PrimaryKey("id"))
	r.Error(table.PrimaryKey("id"))

	// Duplicate primary key
	table = fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", fizz.Options{"primary": true}))
	r.Error(table.PrimaryKey("id"))

	// Duplicate inline primary key
	table = fizz.NewTable("users", nil)
	r.NoError(table.Column("id", "int", fizz.Options{"primary": true}))
	r.Error(table.Column("id2", "int", fizz.Options{"primary": true}))
}
