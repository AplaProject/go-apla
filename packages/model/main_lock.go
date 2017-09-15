package model

import (
	"strconv"
	"time"
)

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
	return DBConn.Delete(&MainLock{}).Error
}

func MainLockDelete(scriptName string) error {
	query := DBConn.Where("script_name=?", scriptName).Delete(&MainLock{})
	if query.Error != nil && !query.RecordNotFound() {
		return query.Error
	}
	return nil
}

func (ml *MainLock) Save() error {
	return DBConn.Save(ml).Error
}

func MainLockUpdate() error {
	return DBConn.Model(&MainLock{}).Update("LockTime", int32(time.Now().Unix())).Error
}

func (ml *MainLock) Get() (bool, error) {
	query := DBConn.First(ml)
	if query.RecordNotFound() {
		return false, nil
	}
	if query.Error != nil {
		return false, query.Error
	}
	return true, nil
}

func (ml *MainLock) Create() error {
	return DBConn.Create(ml).Error
}

func (ml *MainLock) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["lock_time"] = strconv.FormatInt(int64(ml.LockTime), 10)
	result["script_name"] = ml.ScriptName
	result["info"] = ml.Info
	result["uniq"] = strconv.FormatInt(int64(ml.Uniq), 10)
	return result
}

func MainLockCreateTable() error {
	return DBConn.CreateTable(&MainLock{}).Error
}
