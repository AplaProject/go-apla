package model

type Rollback struct {
	RbID    int64  `gorm:primary_key;not null`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null"`
}

func (r *Rollback) Get(rollbackID int64) error {
	return DBConn.Where("rb_id = ?", rollbackID).First(r).Error
}

func GetRollbacks(limit int) ([]Rollback, error) {
	rollbacks := new([]Rollback)
	err := DBConn.Limit(limit).Find(rollbacks).Error
	if err != nil {
		return nil, err
	}
	return *rollbacks, err
}

/*
func (db *DCDB) GetRollbackInfo(rollbackID int64) (map[string]string, error) {
	return db.OneRow(`select r.*, b.time from rollback as r
			left join block_chain as b on b.id=r.block_id
			where r.rb_id=?`, rollbackID).String()
}
*/

func (r *Rollback) Create() error {
	return DBConn.Create(r).Error
}

func (r *Rollback) Delete() error {
	return DBConn.Delete(r).Error
}
