package model

type LogTransaction struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int32  `gorm:"not null"`
}

func (lt *LogTransaction) IsExists() (bool, error) {
	query := DBConn.First(lt)
	return !query.RecordNotFound(), query.Error
}

func (lt *LogTransaction) Delete() error {
	return DBConn.Delete(lt).Error
}

func (lt *LogTransaction) Get() error {
	return DBConn.First(lt).Error
}

func (lt *LogTransaction) Create() error {
	return DBConn.Create(lt).Error
}
