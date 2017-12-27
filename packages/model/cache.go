package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/AplaProject/go-apla/packages/consts"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	OpPlus  = "+"
	OpMinus = "-"
)

type DBCache interface {
	Fill(*DbTransaction) error
	Flush(*DbTransaction) error
}

type KeyCache struct {
	Keys      map[int64]map[int64]*Key
	Rollbacks []*RollbackTx
	lock      sync.RWMutex
}

var KeysCache *KeyCache = &KeyCache{}

func keysTableName(prefix int64) string {
	if prefix == 0 {
		prefix = 1
	}
	return fmt.Sprintf("%d_keys", prefix)
}

func (k *KeyCache) Fill(tr *DbTransaction) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.Keys = map[int64]map[int64]*Key{}
	k.Rollbacks = []*RollbackTx{}

	ids, err := GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting all system states id")
		return err
	}
	for _, ecosystemID := range ids {
		k.Keys[ecosystemID] = map[int64]*Key{}
		keys := new([]*Key)
		if err := GetDB(tr).Table(keysTableName(ecosystemID)).Find(keys).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Selecting all keys from keys table")
			return err
		}
		for _, key := range *keys {
			k.Keys[ecosystemID][key.ID] = key
		}
	}
	return nil
}

func (k *KeyCache) get(ecosystemID, keyID int64) (bool, *Key, error) {
	if _, ok := k.Keys[ecosystemID]; ok {
		if key, ok2 := k.Keys[ecosystemID][keyID]; ok2 {
			return true, key, nil
		}
	}
	key := &Key{}
	key.SetTablePrefix(ecosystemID)
	if found, err := key.Get(keyID); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting key")
		return false, nil, err
	} else if !found {
		return false, nil, nil
	}
	if _, ok := k.Keys[ecosystemID]; !ok {
		k.Keys[ecosystemID] = map[int64]*Key{}
	}
	k.Keys[ecosystemID][keyID] = key
	return true, key, nil
}

func (k *KeyCache) Get(ecosystemID, keyID int64) (bool, *Key, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	return k.get(ecosystemID, keyID)
}

func (k *KeyCache) OpAmount(ecosystemID, keyID int64, op string, amount decimal.Decimal, blockID int64, txHash []byte) (bool, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	found, key, err := k.get(ecosystemID, keyID)
	if err != nil {
		return false, err
	}
	if found {
		data, err := json.Marshal(map[string]string{"amount": key.Amount})
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling amount to JSON")
		}
		rb := &RollbackTx{
			BlockID:   blockID,
			TxHash:    txHash,
			NameTable: keysTableName(ecosystemID),
			TableID:   strconv.FormatInt(keyID, 10),
			Data:      string(data),
		}
		k.Rollbacks = append(k.Rollbacks, rb)
		d, err := decimal.NewFromString(key.Amount)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": key.Amount}).Error("converting key amount to decimal")
			return false, err
		}
		if op == OpPlus {
			d = d.Add(amount)
		} else if op == OpMinus {
			d = d.Sub(amount)
		}
		k.Keys[ecosystemID][key.ID].Amount = d.String()
		return true, nil
	}
	return false, nil
}

func (k *KeyCache) Flush(tr *DbTransaction) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	queries := []string{}
	for ecosystemID, _ := range k.Keys {
		updQuery := `UPDATE "%s" SET amount = %s WHERE id = %d `
		for _, key := range k.Keys[ecosystemID] {
			queries = append(queries, fmt.Sprintf(updQuery, keysTableName(ecosystemID), key.Amount, key.ID))
		}
	}
	if len(queries) > 0 {
		resultQuery := strings.Join(queries, ";") + ";"
		if err := GetDB(tr).Exec(resultQuery).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("batch updating keys from cache")
			return err
		}
	}
	rollbackQuery := "INSERT INTO rollback_tx(block_id, tx_hash, table_name, table_id, data) VALUES %s;"
	valuesTpl := `(%d, %s, '%s','%s','%s')`
	values := []string{}
	for _, rb := range k.Rollbacks {
		txHashStr := `decode('` + hex.EncodeToString(rb.TxHash) + `','HEX')`
		values = append(values, fmt.Sprintf(valuesTpl, rb.BlockID, txHashStr, rb.NameTable, rb.TableID, rb.Data))
	}
	if len(values) > 0 {
		valuesQuery := strings.Join(values, ",")
		if err := GetDB(tr).Exec(fmt.Sprintf(rollbackQuery, valuesQuery)).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("batch inserting keys from cache")
			return err
		}
	}
	k.Keys = map[int64]map[int64]*Key{}
	k.Rollbacks = []*RollbackTx{}
	return nil
}
