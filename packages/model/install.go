package model

type Install struct {
	Progress string `gorm:not null;size:10`
}

func (i *Install) Get() error {
	return DBConn.Find(i).Error
}

func (i *Install) Save() error {
	return DBConn.Save(i).Error
}

func (i *Install) Create() error {
	return DBConn.Create(i).Error
}
