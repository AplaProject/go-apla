package smart

import (
	"github.com/AplaProject/go-apla/packages/storage/memdb"
	"github.com/AplaProject/go-apla/packages/types"
)

const (
	opRead   = "read"
	opInsert = "insert"
	opUpdate = "update"

	concatKeyDelim = ':'
	memCollections = "collections"
	keyCols        = "columns"
)

func MemCollectionCreate(sc *SmartContract, table string, params *types.Map) error {
	table = memTablePrefix(sc, table)
	return MemInsert(sc, memCollections, table, params)
}

func MemCollectionUpdate(sc *SmartContract, table string, params *types.Map) error {
	table = memTablePrefix(sc, table)
	return MemUpdate(sc, memCollections, table, params)
}

func MemInsert(sc *SmartContract, table, key string, val *types.Map) error {
	table = memTablePrefix(sc, table)

	perms, err := loadMemPerms(sc, table)
	if err != nil {
		return err
	}

	if err := perms.AllowOperation(opInsert); err != nil {
		return err
	}

	tr := sc.MultiTr.Get("mem").(*memdb.Transaction)
	return tr.Set(concatKey(table, key), val)
}

func MemGet(sc *SmartContract, table, key string) (*types.Map, error) {
	table = memTablePrefix(sc, table)

	perms, err := loadMemPerms(sc, table)
	if err != nil {
		return nil, err
	}

	if err := perms.AllowOperation(opRead); err != nil {
		return nil, err
	}

	return memGet(sc, concatKey(table, key))
}

func MemUpdate(sc *SmartContract, table, key string, val *types.Map) error {
	table = memTablePrefix(sc, table)

	perms, err := loadMemPerms(sc, table)
	if err != nil {
		return err
	}

	if err = perms.AllowOperation(opUpdate); err != nil {
		return err
	}

	if err = perms.AllowChangeColumns(val); err != nil {
		return err
	}

	// tableKey := concatKey(table, key)
	// cur, err := memGet(sc, tableKey)
	// if err != nil {
	// 	return err
	// }

	// if err = allowPropCondition(sc, cur, keyCond); err != nil {
	// 	return err
	// }

	tr := sc.MultiTr.Get("mem").(*memdb.Transaction)
	return tr.Update(concatKey(table, key), val)
}

func memGet(sc *SmartContract, key string) (*types.Map, error) {
	tr := sc.MultiTr.Get("mem").(*memdb.Transaction)

	val, err := tr.Get(key)
	if err != nil {
		if tr.IsFound(err) {
			return nil, err
		}
	}
	return val, nil
}

func concatKey(prefix, key string) string {
	return prefix + string(concatKeyDelim) + key
}

func memTablePrefix(sc *SmartContract, table string) string {
	return GetTableName(sc, table)
}

type memPerms struct {
	sc   *SmartContract
	op   *types.Map
	cols *types.Map
}

func (mt *memPerms) AllowOperation(op string) error {
	return allowPropCondition(mt.sc, mt.op, op)
}

func (mt *memPerms) AllowChangeColumns(cols *types.Map) (err error) {
	if cols == nil {
		return nil
	}

	for _, col := range cols.Keys() {
		err = allowPropCondition(mt.sc, mt.cols, col)
		if err != nil {
			return
		}
	}
	return
}

func loadMemPerms(sc *SmartContract, key string) (*memPerms, error) {
	perms, err := memGet(sc, memCollections+string(concatKeyDelim)+key)
	if err != nil {
		return nil, err
	}

	mt := &memPerms{
		sc: sc,
	}

	if perms == nil {
		return mt, nil
	}

	mt.op = perms

	if v, ok := perms.Get(keyCols); ok {
		mt.cols, ok = v.(*types.Map)
	}

	return mt, nil
}

func allowPropCondition(sc *SmartContract, m *types.Map, key string) error {
	if m == nil {
		return nil
	}

	var cond string
	if v, ok := m.Get(key); ok {
		cond, ok = v.(string)
	}

	if len(cond) == 0 {
		return nil
	}

	ok, err := sc.EvalIf(cond)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return errAccessDenied
}
