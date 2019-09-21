package fizz

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Index is the index definition for fizz.
type Index struct {
	Name    string
	Columns []string
	Unique  bool
	Options Options
}

func (i Index) String() string {
	var opts map[string]interface{}
	if i.Options == nil {
		opts = make(map[string]interface{}, 0)
	} else {
		opts = i.Options
	}
	if i.Name != "" {
		opts["name"] = i.Name
	}
	if i.Unique {
		opts["unique"] = true
	}
	o := make([]string, 0, len(opts))
	for k, v := range opts {
		vv, _ := json.Marshal(v)
		o = append(o, fmt.Sprintf("%s: %s", k, string(vv)))
	}
	sort.SliceStable(o, func(i, j int) bool { return o[i] < o[j] })
	if len(i.Columns) > 1 {
		cols := make([]string, len(i.Columns))
		for k, v := range i.Columns {
			cols[k] = `"` + v + `"`
		}
		return fmt.Sprintf(`t.Index([%s], {%s})`, strings.Join(cols, ", "), strings.Join(o, ", "))
	}
	return fmt.Sprintf(`t.Index("%s", {%s})`, i.Columns[0], strings.Join(o, ", "))
}

func (f fizzer) AddIndex(table string, columns interface{}, options Options) error {
	t := NewTable(table, nil)
	if err := t.Index(columns, options); err != nil {
		return err
	}
	f.add(f.Bubbler.AddIndex(t))
	return nil
}

func (f fizzer) DropIndex(table, name string) {
	f.add(f.Bubbler.DropIndex(Table{
		Name: table,
		Indexes: []Index{
			{Name: name},
		},
	}))
}

func (f fizzer) RenameIndex(table, old, new string) {
	f.add(f.Bubbler.RenameIndex(Table{
		Name: table,
		Indexes: []Index{
			{Name: old},
			{Name: new},
		},
	}))
}
