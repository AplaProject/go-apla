package translators

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/fizz"
)

type Postgres struct {
}

func NewPostgres() *Postgres {
	return &Postgres{}
}

func (Postgres) Name() string {
	return "postgres"
}

func (p *Postgres) CreateTable(t fizz.Table) (string, error) {
	sql := []string{}
	cols := []string{}
	var s string
	for _, c := range t.Columns {
		if c.Primary {
			switch c.ColType {
			case "string", "uuid": // make sure that we don't fall into default
			case "integer", "INT", "int":
				c.ColType = "SERIAL"
			case "bigint", "BIGINT":
				c.ColType = "BIGSERIAL"
			default:
				return "", fmt.Errorf("can not use %s as a primary key", c.ColType)
			}
		}
		cols = append(cols, p.buildAddColumn(c))
		if c.Primary {
			cols = append(cols, fmt.Sprintf(`PRIMARY KEY("%s")`, c.Name))
		}
	}

	for _, fk := range t.ForeignKeys {
		cols = append(cols, p.buildForeignKey(t, fk, true))
	}

	primaryKeys := t.PrimaryKeys()
	if len(primaryKeys) > 1 {
		pks := make([]string, len(primaryKeys))
		for i, pk := range primaryKeys {
			pks[i] = fmt.Sprintf("\"%s\"", pk)
		}
		cols = append(cols, fmt.Sprintf("PRIMARY KEY(%s)", strings.Join(pks, ", ")))
	}

	s = fmt.Sprintf("CREATE TABLE \"%s\" (\n%s\n);", t.Name, strings.Join(cols, ",\n"))
	sql = append(sql, s)

	for _, i := range t.Indexes {
		s, err := p.AddIndex(fizz.Table{
			Name:    t.Name,
			Indexes: []fizz.Index{i},
		})
		if err != nil {
			return "", err
		}
		sql = append(sql, s)
	}

	return strings.Join(sql, "\n"), nil
}

func (p *Postgres) DropTable(t fizz.Table) (string, error) {
	return fmt.Sprintf("DROP TABLE \"%s\";", t.Name), nil
}

func (p *Postgres) RenameTable(t []fizz.Table) (string, error) {
	if len(t) < 2 {
		return "", fmt.Errorf("not enough table names supplied")
	}
	return fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\";", t[0].Name, t[1].Name), nil
}

func (p *Postgres) ChangeColumn(t fizz.Table) (string, error) {
	if len(t.Columns) == 0 {
		return "", fmt.Errorf("not enough columns supplied")
	}
	c := t.Columns[0]
	s := fmt.Sprintf("ALTER TABLE \"%s\" ALTER COLUMN %s;", t.Name, p.buildChangeColumn(c))
	return s, nil
}

func (p *Postgres) AddColumn(t fizz.Table) (string, error) {
	if len(t.Columns) == 0 {
		return "", fmt.Errorf("not enough columns supplied")
	}
	c := t.Columns[0]
	s := fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s;", t.Name, p.buildAddColumn(c))
	return s, nil
}

func (p *Postgres) DropColumn(t fizz.Table) (string, error) {
	if len(t.Columns) == 0 {
		return "", fmt.Errorf("not enough columns supplied")
	}
	c := t.Columns[0]
	return fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\";", t.Name, c.Name), nil
}

func (p *Postgres) RenameColumn(t fizz.Table) (string, error) {
	if len(t.Columns) < 2 {
		return "", fmt.Errorf("not enough columns supplied")
	}
	oc := t.Columns[0]
	nc := t.Columns[1]
	s := fmt.Sprintf("ALTER TABLE \"%s\" RENAME COLUMN \"%s\" TO \"%s\";", t.Name, oc.Name, nc.Name)
	return s, nil
}

func (p *Postgres) AddIndex(t fizz.Table) (string, error) {
	if len(t.Indexes) == 0 {
		return "", fmt.Errorf("not enough indexes supplied")
	}
	i := t.Indexes[0]
	s := fmt.Sprintf("CREATE INDEX \"%s\" ON \"%s\" (%s);", i.Name, t.Name, strings.Join(i.Columns, ", "))
	if i.Unique {
		s = strings.Replace(s, "CREATE", "CREATE UNIQUE", 1)
	}
	return s, nil
}

func (p *Postgres) DropIndex(t fizz.Table) (string, error) {
	if len(t.Indexes) == 0 {
		return "", fmt.Errorf("not enough indexes supplied")
	}
	i := t.Indexes[0]
	return fmt.Sprintf("DROP INDEX \"%s\";", i.Name), nil
}

func (p *Postgres) RenameIndex(t fizz.Table) (string, error) {
	ix := t.Indexes
	if len(ix) < 2 {
		return "", fmt.Errorf("not enough indexes supplied")
	}
	oi := ix[0]
	ni := ix[1]
	return fmt.Sprintf("ALTER INDEX \"%s\" RENAME TO \"%s\";", oi.Name, ni.Name), nil
}

func (p *Postgres) AddForeignKey(t fizz.Table) (string, error) {
	if len(t.ForeignKeys) == 0 {
		return "", fmt.Errorf("not enough foreign keys supplied")
	}

	return p.buildForeignKey(t, t.ForeignKeys[0], false), nil
}

func (p *Postgres) DropForeignKey(t fizz.Table) (string, error) {
	if len(t.ForeignKeys) == 0 {
		return "", fmt.Errorf("not enough foreign keys supplied")
	}

	fk := t.ForeignKeys[0]

	var ifExists string
	if v, ok := fk.Options["if_exists"]; ok && v.(bool) {
		ifExists = "IF EXISTS "
	}

	s := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s\"%s\";", p.escapeIdentifier(t.Name), ifExists, fk.Name)
	return s, nil
}

func (p *Postgres) buildAddColumn(c fizz.Column) string {
	s := fmt.Sprintf("\"%s\" %s", c.Name, p.colType(c))

	if c.Options["null"] == nil || c.Primary {
		s = fmt.Sprintf("%s NOT NULL", s)
	}
	if c.Options["default"] != nil {
		s = fmt.Sprintf("%s DEFAULT '%v'", s, c.Options["default"])
	}
	if c.Options["default_raw"] != nil {
		s = fmt.Sprintf("%s DEFAULT %s", s, c.Options["default_raw"])
	}

	return s
}

func (p *Postgres) buildChangeColumn(c fizz.Column) string {
	s := fmt.Sprintf("\"%s\" TYPE %s", c.Name, p.colType(c))

	var sets []string
	if c.Options["null"] == nil {
		sets = append(sets, fmt.Sprintf("ALTER COLUMN \"%s\" SET NOT NULL", c.Name))
	} else {
		sets = append(sets, fmt.Sprintf("ALTER COLUMN \"%s\" DROP NOT NULL", c.Name))
	}
	if c.Options["default"] != nil {
		sets = append(sets, fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT '%v'", c.Name, c.Options["default"]))
	}
	if c.Options["default_raw"] != nil {
		sets = append(sets, fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT %s", c.Name, c.Options["default_raw"]))
	}
	if len(sets) > 0 {
		s += ", " + strings.Join(sets, ", ")
	}

	return s
}

func (p *Postgres) colType(c fizz.Column) string {
	switch c.ColType {
	case "string":
		s := "255"
		if c.Options["size"] != nil {
			s = fmt.Sprintf("%d", c.Options["size"])
		}
		return fmt.Sprintf("VARCHAR (%s)", s)
	case "uuid":
		return "UUID"
	case "time", "datetime":
		return "timestamptz"
	case "blob", "[]byte":
		return "bytea"
	case "float", "decimal":
		if c.Options["precision"] != nil {
			precision := c.Options["precision"]
			if c.Options["scale"] != nil {
				scale := c.Options["scale"]
				return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
			}
			return fmt.Sprintf("DECIMAL(%d)", precision)
		}

		return "DECIMAL"
	case "[]string":
		return "varchar[]"
	case "[]float":
		return "decimal[]"
	case "[]int":
		return "integer[]"
	case "json":
		return "jsonb"
	default:
		return c.ColType
	}
}

func (p *Postgres) buildForeignKey(t fizz.Table, fk fizz.ForeignKey, onCreate bool) string {
	rcols := []string{}
	for _, c := range fk.References.Columns {
		rcols = append(rcols, fmt.Sprintf("\"%s\"", c))
	}
	refs := fmt.Sprintf("%s (%s)", p.escapeIdentifier(fk.References.Table), strings.Join(rcols, ", "))
	s := fmt.Sprintf("FOREIGN KEY (\"%s\") REFERENCES %s", fk.Column, refs)

	if onUpdate, ok := fk.Options["on_update"]; ok {
		s += fmt.Sprintf(" ON UPDATE %s", onUpdate)
	}

	if onDelete, ok := fk.Options["on_delete"]; ok {
		s += fmt.Sprintf(" ON DELETE %s", onDelete)
	}

	if !onCreate {
		s = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT \"%s\" %s;", p.escapeIdentifier(t.Name), fk.Name, s)
	}

	return s
}

func (Postgres) escapeIdentifier(s string) string {
	if !strings.ContainsRune(s, '.') {
		return fmt.Sprintf("\"%s\"", s)
	}
	parts := strings.Split(s, ".")
	for _, p := range parts {
		p = fmt.Sprintf("\"%s\"", p)
	}
	return strings.Join(parts, ".")
}
