package model

type RbFullNodes struct {
	RbID                int64  `gorm:"primary_key;not null"`
	FullNodesWalletJson []byte `gorm:"not_null"`
	BlockID             int64  `gorm:"primary_key;not null"`
	PrevRbID            int64  `gorm:"not null"`
}

func (r *RbFullNodes) Create() error {
	return DBConn.Create(r).Error
}

func (r *RbFullNodes) GetByRbID(id int64) error {
	return DBConn.Where("rb_id = ?", id).First(&r).Error
}
