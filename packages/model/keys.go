package model

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
)

var ecosystemNotFoundErr = errors.New("ecosystem not found")

type recordStatus int8

const (
	New      recordStatus = iota
	Updated               = iota
	Original              = iota
)

type KeyWithHistory struct {
	Key     Key
	History Key
	Status  recordStatus
}

func NewBufferedKeys() *bufferedKeys {
	return &bufferedKeys{keys: make(map[int64]map[int64]*KeyWithHistory)}
}

type bufferedKeys struct {
	keys        map[int64]map[int64]*KeyWithHistory
	rwMutex     sync.RWMutex
	updateMutex sync.Mutex
}

func loadEcosystemKeys(ecosystemID int64) (*[]Key, error) {
	var keys []Key
	err := DBConn.Raw(fmt.Sprintf(`select * from "%d_keys";`, ecosystemID)).Scan(&keys).Error
	if err != nil {
		return nil, err
	}
	return &keys, nil
}

func loadVal(ecosystemID int64, keyID int64) (Key, error) {
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
	newEcosystemBuffer := make(map[int64]*KeyWithHistory, len(*keys))
	for _, k := range *keys {
		newEcosystemBuffer[k.ID] = &KeyWithHistory{Key: k, History: k, Status: Original}
	}

	bk.keys[tablePrefix] = newEcosystemBuffer
	return nil
}

func (bk *bufferedKeys) updateKeyCache(tablePrefix int64, id int64) error {
	bk.updateMutex.Lock()
	defer bk.updateMutex.Unlock()
	key, err := loadVal(tablePrefix, id)
	if err != nil {
		return err
	}
	val := &KeyWithHistory{Key: key, History: key, Status: Original}
	bk.keys[tablePrefix][id] = val
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
		err = bk.updateKeyCache(tablePrefix, id)
		if err != nil && err != gorm.ErrRecordNotFound {
			return result, false, err
		}
		if err == gorm.ErrRecordNotFound {
			return result, false, nil
		}
	}

	result = bk.keys[tablePrefix][id].Key
	return result, true, nil
}

func (bk *bufferedKeys) UpdateKey(tablePrefix int64, id int64, key Key) (found bool, err error) {
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
	bk.keys[tablePrefix][id].Key = key
	if bk.keys[tablePrefix][id].Status != New {
		bk.keys[tablePrefix][id].Status = Updated
	}
	bk.rwMutex.Unlock()
	return true, nil
}

func (bk *bufferedKeys) PushKey(tablePrefix int64, id int64, key Key) (found bool, err error) {
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
	bk.rwMutex.RUnlock()

	bk.rwMutex.Lock()
	bk.keys[tablePrefix][id] = &KeyWithHistory{Key: key, History: Key{}, Status: New}
	bk.rwMutex.Unlock()
	return true, nil
}

func (k Key) GenerateUpdateSQL(ecosystemID int64, oldKey *Key) string {
	updateQuery := fmt.Sprintf(` UPDATE "%d_keys" SET`, ecosystemID)
	if !bytes.Equal(k.PublicKey, oldKey.PublicKey) {
		updateQuery += fmt.Sprintf(` pub = '%s',`, hex.EncodeToString(k.PublicKey))
	}
	if k.Amount != oldKey.Amount {
		updateQuery += fmt.Sprintf(` amount = %s`, k.Amount)
	}
	updateQuery += fmt.Sprintf(` where id = %d;`, k.ID)
	return updateQuery
}

func (k Key) GenerateInsertSQL(ecosystemID int64) string {
	return fmt.Sprintf(`INSERT INTO "%d_keys" VALUES (%d, '%s', %s);`,
		ecosystemID, k.ID, hex.EncodeToString(k.PublicKey), k.Amount)
}

func (bk *bufferedKeys) Flush(transaction *DbTransaction) error {
	updateQueries := ""
	insertQueries := ""
	for ecosystemID, table := range bk.keys {
		for _, key := range table {
			if key.Status == New {
				updateQueries += key.Key.GenerateInsertSQL(ecosystemID)
			} else if key.Status == Updated {
				insertQueries += key.Key.GenerateUpdateSQL(ecosystemID, &key.History)
			}
			key.History = key.Key
			key.Status = Original
		}
	}
	insertQueries += updateQueries
	if len(insertQueries) > 0 {
		err := GetDB(transaction).Exec(insertQueries).Error
		if err != nil {
			return err
		}
	}

	return nil
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
