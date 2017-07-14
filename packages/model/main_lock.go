package model

type MainLock struct {
	LockTime   int32  `gorm:"not_null"`
	ScriptName string `gorm:"not_null;size:100"`
	Info       string `gorm:"not_null"`
	Uniq       int8   `gorm:"not_null"`
}

func DeleteMainLock() error {
	if err := DBConn.Delete(&MainLock{}).Error; err != nil {
		return err
	}
	return nil
}
