package model

import (
	"database/sql"
	"encoding/json"

	"github.com/AplaProject/go-apla/packages/converter"
)

const rollbackColumnName = "rb_id"

// Rollback is model
type Rollback struct {
	RbID    int64  `gorm:"primary_key;not null"`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null;type:jsonb(PostgreSQL)"`
}

// TableName returns name of table
func (Rollback) TableName() string {
	return "rollback"
}

// Get is retrieving model from database
func (r *Rollback) Get(rollbackID int64) (bool, error) {
	return isFound(DBConn.Where("rb_id = ?", rollbackID).First(r))
}

// Create is creating record of model
func (r *Rollback) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(r).Error
}

// Delete is deleting record
func (r *Rollback) Delete() error {
	return DBConn.Delete(r).Error
}

// DataMap returns rollback data as map
func (r *Rollback) DataMap() (map[string]string, error) {
	var v map[string]string
	err := json.Unmarshal([]byte(r.Data), &v)
	return v, err
}

// GetRollbackHistory returns history of rollback
func GetRollbackHistory(id int64, limit int) ([]map[string]string, error) {
	var history []map[string]string

	for i := 0; id > 0 && i < limit; i++ {
		rb := &Rollback{}
		ok, err := rb.Get(id)
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		d, err := rb.DataMap()
		if err != nil {
			return nil, err
		}

		history = append(history, d)

		id = converter.StrToInt64(d[rollbackColumnName])
	}

	return history, nil
}

// GetRollbackIDForTableRow returns rollback ID for table row
func GetRollbackIDForTableRow(table string, rowID int64) (int64, error) {
	var rbID int64
	err := DBConn.Table(table).Select(rollbackColumnName).Where("id = ?", rowID).Row().Scan(&rbID)
	if err != nil {
		// if record not found
		if err == sql.ErrNoRows {
			return rbID, nil
		}
	}
	return rbID, err
}
