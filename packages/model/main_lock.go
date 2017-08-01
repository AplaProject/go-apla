package model

type MainLock struct {
	LockTime   int32  `gorm:"not_null"`
	ScriptName string `gorm:"not_null;size:100"`
	Info       string `gorm:"not_null"`
	Uniq       int8   `gorm:"not_null"`
}

func (ml *MainLock) TableName() string {
	return "main_lock"
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

func (ml *MainLock) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["lock_time"] = string(ml.LockTime)
	result["script_name"] = ml.ScriptName
	result["info"] = ml.Info
	result["uniq"] = string(ml.Uniq)
	return result
}
