package model

import (
	"fmt"
)

// Static is model
type Static struct {
	tableName string
	ID        int64
	Name      string
	Data      []byte
	Hash      string
}

// SetTablePrefix is setting table prefix
func (s *Static) SetTablePrefix(prefix string) {
	s.tableName = prefix + "_statics"
}

// TableName returns name of table
func (s *Static) TableName() string {
	return s.tableName
}

// Get is retrieving model from database
func (s *Static) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).Select("id,name,hash").First(s))
}

func (s *Static) Link() string {
	return fmt.Sprintf(`/data/%s/%d/%s/%s`, s.TableName(), s.ID, "data", s.Hash)
}
