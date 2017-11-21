package model

import "time"

// StopDaemon is model
type StopDaemon struct {
	StopTime int64 `gorm:"not null"`
}

// TableName returns name of table
func (sd *StopDaemon) TableName() string {
	return "stop_daemons"
}

// Create is creating record of model
func (sd *StopDaemon) Create() error {
	return DBConn.Create(sd).Error
}

func (sd *StopDaemon) Delete() error {
	return DBConn.Delete(&StopDaemon{}).Error
}

// Get is retrieving model from database
func (sd *StopDaemon) Get() (bool, error) {
	return isFound(DBConn.First(sd))
}

func SetStopNow() error {
	stopTime := &StopDaemon{StopTime: time.Now().Unix()}
	return stopTime.Create()
}
