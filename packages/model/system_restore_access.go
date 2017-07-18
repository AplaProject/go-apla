package model

type SystemRestoreAccess struct {
	ID        int64  `gorm:"primary_key;not_null"`
	CitizenID int64  `gorm:"not null"`
	StateID   int64  `gorm:"not null"`
	Active    int64  `gorm:"not null"`
	Time      int32  `gorm:"not null"`
	Close     int64  `gorm:"not null"`
	Secret    string `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
}

func (sra *SystemRestoreAccess) Get(stateID int64) error {
	return DBConn.Where("state_id = ?", stateID).First(sra).Error
}

func (sra *SystemRestoreAccess) GetWithUserID(userID int64, stateID int64) error {
	return DBConn.Where("user_id  =  ? AND state_id = ?", userID, stateID).Error
}
