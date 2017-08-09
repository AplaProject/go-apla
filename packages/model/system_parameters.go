package model

type SystemParameter struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp SystemParameter) TableName() string {
	return "system_parameters"
}

func (sp *SystemParameter) Get(name string) error {
	return DBConn.Where("name = ?").First(sp).Error
}

func (sp *SystemParameter) GetJSONField(jsonField string, name string) (string, error) {
	var result string
	err := DBConn.Table("system_parameters").Where("name = ?", name).Select(jsonField).Find(result).Error
	return result, err
}

func (sp *SystemParameter) GetValueParameterByName(name, value string) (string, error) {
	var result string
	err := DBConn.Raw(`SELECT value->'`+value+`' FROM system_parameters WHERE name = ?`, name).Scan(&result).Error
	if err != nil {
		return "", err
	}
	return result, nil
}

func GetAllSystemParameters() ([]SystemParameter, error) {
	parameters := new([]SystemParameter)
	if err := DBConn.Find(parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}

func (sp *SystemParameter) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = sp.Name
	result["value"] = sp.Value
	result["conditions"] = sp.Conditions
	result["rb_id"] = string(sp.RbID)
	return result
}
