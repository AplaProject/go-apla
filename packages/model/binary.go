package model

import (
	"fmt"
	"github.com/GenesisKernel/go-genesis/packages/converter"
)

const BinaryTableSuffix = "_binaries"

// Binary represents record of {prefix}_binaries table
type Binary struct {
	ecosystem int64
	ID        int64
	Name      string
	Data      []byte
	Hash      string
	MimeType  string
}

// SetTablePrefix is setting table prefix
func (b *Binary) SetTablePrefix(prefix string) {
	b.ecosystem = converter.StrToInt64(prefix)
}

// SetTableName sets name of table
func (b *Binary) SetTableName(tableName string) {
	_, ecosystem, _ := RealNameEcosystem(tableName)
	b.ecosystem = ecosystem
}

// TableName returns name of table
func (b *Binary) TableName() string {
	if b.ecosystem == 0 {
		b.ecosystem = 1
	}
	return `1_binaries`
}

// Get is retrieving model from database
func (b *Binary) Get(appID, memberID int64, name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and app_id = ? AND member_id = ? AND name = ?",
		b.ecosystem, appID, memberID, name).Select("id,name,hash").First(b))
}

// Link returns link to binary data
func (b *Binary) Link() string {
	return fmt.Sprintf(`/data/%d_%s/%d/%s/%s`, b.ecosystem, b.TableName(), b.ID, "data", b.Hash)
}

// GetByID is retrieving model from db by id
func (b *Binary) GetByID(id int64) (bool, error) {
	return isFound(DBConn.Where("id=?", id).First(b))
}
