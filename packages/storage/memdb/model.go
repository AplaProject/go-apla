package memdb

import (
	"encoding/json"

	"github.com/AplaProject/memdb"
)

type Model interface {
	PrimaryKey() string
}

func (t *Transaction) FindModel(m Model) (exists bool, err error) {
	var data string
	data, err = t.tx.Get(m.PrimaryKey())
	if err != nil {
		if err == memdb.ErrNotFound {
			err = nil
		}
		return
	}
	exists = true
	err = json.Unmarshal([]byte(data), m)
	return
}

func (t *Transaction) InsertModel(m Model) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return t.tx.Set(m.PrimaryKey(), string(data))
}

func (t *Transaction) UpdateModel(m Model) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = t.tx.Update(m.PrimaryKey(), string(data))
	if err != nil {
		return err
	}
	return nil
}
