package model

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
)

type bufferedKeys struct {
	keys        map[int64]map[int64]Key
	rwMutex     sync.RWMutex
	updateMutex sync.Mutex
}

var ecosystemNotFoundErr = errors.New("ecosystem not found")

func NewBufferedKeys() *bufferedKeys {
	return &bufferedKeys{keys: make(map[int64]map[int64]Key)}
}

func loadEcosystemKeys(ecosystemID int64) (*[]Key, error) {
	var keys []Key
	err := DBConn.Raw(fmt.Sprintf(`select * from "%d_keys";`, ecosystemID)).Scan(&keys).Error
	if err != nil {
		return nil, err
	}
	return &keys, nil
}

func loadKey(ecosystemID int64, keyID int64) (Key, error) {
	key := Key{}
	k := key.SetTablePrefix(ecosystemID)
	err := k.Get(keyID)
	return *k, err
}

func (bk *bufferedKeys) updateEcosystemCache(tablePrefix int64) error {
	bk.updateMutex.Lock()
	defer bk.updateMutex.Unlock()
	keys, err := loadEcosystemKeys(tablePrefix)
	if err != nil {
		return err
	}
	newEcosystemBuffer := make(map[int64]Key, len(*keys))
	for _, k := range *keys {
		newEcosystemBuffer[k.ID] = k
	}

	bk.keys[tablePrefix] = newEcosystemBuffer
	return nil
}

func (bk *bufferedKeys) updateKeyCache(tablePrefix int64, id int64) error {
	bk.updateMutex.Lock()
	defer bk.updateMutex.Unlock()
	key, err := loadKey(tablePrefix, id)
	if err != nil {
		return err
	}
	bk.keys[tablePrefix][id] = key
	return nil
}

func (bk *bufferedKeys) Initialize() error {
	IDs, err := GetAllSystemStatesIDs()
	if err != nil {
		return err
	}
	for _, ID := range IDs {
		err := bk.updateEcosystemCache(ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bk *bufferedKeys) GetKey(tablePrefix int64, id int64) (key Key, found bool, err error) {
	result := Key{}
	bk.rwMutex.RLock()
	defer bk.rwMutex.RUnlock()

	_, ok := bk.keys[tablePrefix]
	if !ok {
		err := bk.updateEcosystemCache(tablePrefix)
		if err != nil && err != gorm.ErrRecordNotFound {
			return result, false, err
		}
		if err == gorm.ErrRecordNotFound {
			return result, false, nil
		}
	}

	_, ok = bk.keys[tablePrefix][id]
	if !ok {
		fmt.Println("retrieve key with id", id)
		err = bk.updateKeyCache(tablePrefix, id)
		if err != nil && err != gorm.ErrRecordNotFound {
			return result, false, err
		}
		if err == gorm.ErrRecordNotFound {
			return result, false, nil
		}
	}

	result = bk.keys[tablePrefix][id]
	return result, true, nil
}

func (bk *bufferedKeys) SetKey(tablePrefix int64, id int64, key Key) (found bool, err error) {
	bk.rwMutex.RLock()
	_, ok := bk.keys[tablePrefix]
	if !ok {
		err := bk.updateEcosystemCache(tablePrefix)
		if err != nil && err != gorm.ErrRecordNotFound {
			return false, err
		}
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
	}

	_, ok = bk.keys[tablePrefix]
	if !ok {
		err = bk.updateKeyCache(tablePrefix, id)
		if err != nil && err != gorm.ErrRecordNotFound {
			return false, err
		}
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
	}
	bk.rwMutex.RUnlock()

	bk.rwMutex.Lock()
	bk.keys[tablePrefix][id] = key
	bk.rwMutex.Unlock()
	return true, nil
}

// Key is model
type Key struct {
	tableName string
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	if prefix == 0 {
		prefix = 1
	}
	m.tableName = fmt.Sprintf("%d_keys", prefix)
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	return m.tableName
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) error {
	return DBConn.Where("id = ?", wallet).First(m).Error
}
