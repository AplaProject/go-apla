package model

type SystemRestoreAccess struct {
	ID        int64  `gorm:"primary_key;not_null"`
	CitizenID int64  `gorm:"not null"`
	StateID   int64  `gorm:"not null"`
	Active    int64  `gorm:"not null"`
	Time      int64  `gorm:"not null"`
	Close     int64  `gorm:"not null"`
	Secret    string `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
}

func (sra *SystemRestoreAccess) TableName() string {
	return "system_restore_access"
}

func (sra *SystemRestoreAccess) Get(stateID int64) (bool, error) {
	return isFound(DBConn.Where("state_id = ?", stateID).First(sra))
}

func (sra *SystemRestoreAccess) GetWithUserID(userID int64, stateID int64) error {
	return DBConn.Where("user_id  =  ? AND state_id = ?", userID, stateID).Error
}
