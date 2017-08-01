package model

type Rollback struct {
	RbID    int64  `gorm:"primary_key;not null"`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null"`
}

func (r *Rollback) Get(rollbackID int64) error {
	return DBConn.Where("rb_id = ?", rollbackID).First(r).Error
}

func (r *Rollback) GetRollbacks(limit int) ([]Rollback, error) {
	rollbacks := new([]Rollback)
	err := DBConn.Limit(limit).Find(rollbacks).Error
	if err != nil {
		return nil, err
	}
	return *rollbacks, err
}

func (r *Rollback) Create() error {
	return DBConn.Create(r).Error
}

func (r *Rollback) Delete() error {
	return DBConn.Delete(r).Error
}

func (r *Rollback) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["rb_id"] = string(r.RbID)
	result["block_id"] = string(r.BlockID)
	result["data"] = r.Data
	return result
}
