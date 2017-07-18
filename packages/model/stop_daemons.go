package model

type StopDaemons struct {
	StopTime int32 `gorm:"not null"`
}

func (sd *StopDaemons) Create() error {
	return DBConn.Create(sd).Error
}

func (sd *StopDaemons) Delete() error {
	return DBConn.Delete(&StopDaemons{}).Error
}

func (sd *StopDaemons) Get() error {
	return DBConn.First(sd).Error
}
