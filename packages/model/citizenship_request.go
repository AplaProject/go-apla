package model

type CitizenshipRequest struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null"`
	PublicKey  []byte `gorm:"column:public_key_0"`
	DltWalletID int64  `gorm:"not null"`
	Name        string
	Approved    int64 `gorm:"not null"`
	BlockID     int64 `gorm:"not null"`
	RbID        int64 `gorm:"not null"`
}

func (cr *CitizenshipRequest) SetTablePrefix(tablePrefix string) {
	cr.tableName = tablePrefix + "_citizenship_requests"
}

func (cr *CitizenshipRequest) TableName() string {
	return cr.tableName
}

func (cr *CitizenshipRequest) GetByWallet(walletID int64) error {
	return DBConn.Where("dlt_wallet_id = ?", walletID).Find(cr).Error
}

func (cr *CitizenshipRequest) GetByWalletOrdered(walletID int64) error {
	return DBConn.Order("id desc").Where("dlt_wallet_id = ?", walletID).Find(cr).Error
}

func (cr *CitizenshipRequest) GetUnapproved(startID int64) error {
	return DBConn.Order("id desc").Where("approved = 0 and id > ", startID).First(cr).Error
}

func (cr *CitizenshipRequest) ToStringMap() map[string]string {
	result := make(map[string]string)
	result["id"] = string(cr.ID)
	result["public_key"] = string(cr.PublicKey)
	result["dlt_wallet_id"] = string(cr.DltWalletID)
	result["name"] = string(cr.Name)
	result["approved"] = string(cr.Approved)
	result["block_id"] = string(cr.BlockID)
	result["rb_id"] = string(cr.RbID)
	return result
}
