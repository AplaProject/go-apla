package model

type SystemParameters struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (sp *SystemParameters) Get(name string) error {
	return DBConn.Where("name = ?").First(sp).Error
}

func GetAllSystemParameters() ([]SystemParameters, error) {
	parameters := new([]SystemParameters)
	if err := DBConn.Find(parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}
