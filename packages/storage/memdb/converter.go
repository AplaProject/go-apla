package memdb

import (
	"encoding/json"

	"github.com/AplaProject/go-apla/packages/types"
)

func toMap(data string) (m *types.Map, err error) {
	var v map[string]interface{}
	err = json.Unmarshal([]byte(data), &v)
	if err != nil {
		return nil, err
	}

	m = types.LoadMap(v)
	return
}

func fromMap(m *types.Map) (string, error) {
	data, err := m.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func mergeMap(m, u *types.Map) {
	for _, k := range u.Keys() {
		v, _ := u.Get(k)
		m.Set(k, v)
	}
}
