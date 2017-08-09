package model

type RbFullNode struct {
	RbID                int64  `gorm:"primary_key;not null"`
	FullNodesWalletJson []byte `gorm:"not null"`
	BlockID             int64  `gorm:"primary_key;not null"`
	PrevRbID            int64  `gorm:"not null"`
}

func (r *RbFullNode) Create() error {
	return DBConn.Create(r).Error
}

func (r *RbFullNode) GetByRbID(id int64) error {
	return DBConn.Where("rb_id = ?", id).First(r).Error
}
