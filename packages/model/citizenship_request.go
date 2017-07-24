package model

type CitizenshipRequests struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null"`
	PublickKey  []byte `gorm:"column:public_key_0"`
	DltWalletID int64  `gorm:"not null"`
	Name        []byte
	Approved    int64 `gorm:"not null"`
	BlockID     int64 `gorm:"not null"`
	RbID        int64 `gorm:"not null"`
}

func (cr *CitizenshipRequests) SetTableName(tablePrefix int64) {
	cr.tableName = string(tablePrefix) + "_citizenship_requests"
}

func (cr *CitizenshipRequests) TableName() string {
	return cr.tableName
}

func (cr *CitizenshipRequests) GetByWallet(walletID int64) error {
	return DBConn.Where("dlt_wallet_id = ?", walletID).Find(cr).Error
}

func (cr *CitizenshipRequests) GetByWalletOrdered(walletID int64) error {
	return DBConn.Order("id desc").Where("dlt_wallet_id = ?", walletID).Find(cr).Error
}

func (cr *CitizenshipRequests) GetUnapproved(startID int64) error {
	return DBConn.Order("id desc").Where("approved = 0 and id > ", startID).First(cr).Error
}

func (cr *CitizenshipRequests) ToStringMap() map[string]string {
	result := make(map[string]string)
	result["id"] = string(cr.ID)
	result["public_key"] = string(cr.PublickKey)
	result["dlt_wallet_id"] = string(cr.DltWalletID)
	result["name"] = string(cr.Name)
	result["approved"] = string(cr.Approved)
	result["block_id"] = string(cr.BlockID)
	result["rb_id"] = string(cr.RbID)
	return result
}
