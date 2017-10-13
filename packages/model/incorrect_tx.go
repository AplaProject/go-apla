package model

type IncorrectTx struct {
	Time int64
	Hash []byte
	Err  string
}

func (IncorrectTx) TableName() string {
	return "incorrect_tx"
}

func (i *IncorrectTx) Create() error {
	return DBConn.Create(i).Error
}
