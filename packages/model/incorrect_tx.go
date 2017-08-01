package model

type IncorrectTx struct {
	Time int64
	Hash []byte
	Err  string
}

func (i *IncorrectTx) Create() error {
	return DBConn.Create(i).Error
}
