package model

import "strconv"

type Rollback struct {
	RbID    int64  `gorm:"primary_key;not null"`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null;type:jsonb(PostgreSQL)"`
}

func (Rollback) TableName() string {
	return "rollback"
}

func (r *Rollback) Get(rollbackID int64) error {
	return DBConn.Where("rb_id = ?", rollbackID).First(r).Error
}

func (r *Rollback) GetRollbacks(limit int) ([]Rollback, error) {
	rollbacks := new([]Rollback)
	err := DBConn.Limit(limit).Find(&rollbacks).Error
	if err != nil {
		return nil, err
	}
	return *rollbacks, err
}

func (r *Rollback) Create(transaction *DbTransaction) error {
	return getDB(transaction).Create(r).Error
}

func (r *Rollback) Delete() error {
	return DBConn.Delete(r).Error
}

func (r *Rollback) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["rb_id"] = strconv.FormatInt(r.RbID, 10)
	result["block_id"] = strconv.FormatInt(r.BlockID, 10)
	result["data"] = r.Data
	return result
}
