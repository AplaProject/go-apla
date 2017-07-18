package model

type MainLock struct {
	LockTime   int32  `gorm:"not_null"`
	ScriptName string `gorm:"not_null;size:100"`
	Info       string `gorm:"not_null"`
	Uniq       int8   `gorm:"not_null"`
}

func (ml *MainLock) Delete() error {
	query := DBConn.Delete(&MainLock{})
	if query.Error != nil && !query.RecordNotFound() {
		return query.Error
	}
	return nil
}

func (ml *MainLock) Get() error {
	return DBConn.First(ml).Error
}

func (ml *MainLock) Create() error {
	return DBConn.Create(ml).Error
}
