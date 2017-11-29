// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package smart

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// SmartContract is storing smart contract data
type SmartContract struct {
	VDE           bool
	VM            *script.VM
	TxSmart       tx.SmartContract
	TxData        map[string]interface{}
	TxContract    *Contract
	TxCost        int64           // Maximum cost of executing contract
	TxUsedCost    decimal.Decimal // Used cost of CPU resources
	BlockData     *utils.BlockData
	TxHash        []byte
	PublicKeys    [][]byte
	DbTransaction *model.DbTransaction
}

var (
	funcCallsDB = map[string]struct{}{
		"DBInsert":    {},
		"DBSelect":    {},
		"DBUpdate":    {},
		"DBUpdateExt": {},
	}
	extendCost = map[string]int64{
		"AddressToId":        10,
		"ColumnCondition":    50,
		"CompileContract":    100,
		"Contains":           10,
		"ContractAccess":     50,
		"ContractConditions": 50,
		"ContractsList":      10,
		"CreateColumn":       50,
		"CreateTable":        100,
		"EcosysParam":        10,
		"Eval":               10,
		"EvalCondition":      20,
		"FlushContract":      50,
		"HMac":               50,
		"Join":               10,
		"JSONToMap":          50,
		"Sha256":             50,
		"IdToAddress":        10,
		"IsContract":         10,
		"Len":                5,
		"Join":               10,
		"Replace":            10,
		"PermColumn":         50,
		"PermTable":          100,
		"Substr":             10,
		"Size":               10,
		"ToLower":            10,
		"TrimSpace":          10,
		"TableConditions":    100,
		"UpdateLang":         10,
		"ValidateCondition":  30,
	}
)

func getCost(name string) int64 {
	if val, ok := extendCost[name]; ok {
		return val
	}
	return -1
}

// EmbedFuncs is extending vm with embedded functions
func EmbedFuncs(vm *script.VM) {
	vmExtend(vm, &script.ExtendData{Objects: map[string]interface{}{
		"AddressToId":        AddressToID,
		"ColumnCondition":    ColumnCondition,
		"CompileContract":    CompileContract,
		"Contains":           strings.Contains,
		"ContractAccess":     ContractAccess,
		"ContractConditions": ContractConditions,
		"ContractsList":      contractsList,
		"CreateColumn":       CreateColumn,
		"CreateTable":        CreateTable,
		"DBInsert":           DBInsert,
		"DBSelect":           DBSelect,
		"DBUpdate":           DBUpdate,
		"DBUpdateExt":        DBUpdateExt,
		"EcosysParam":        EcosysParam,
		"Eval":               Eval,
		"EvalCondition":      EvalCondition,
		"Float":              Float,
		"FlushContract":      FlushContract,
		"HMac":               HMac,
		"Join":               Join,
		"JSONToMap":          JSONToMap,
		"IdToAddress":        IDToAddress,
		"Int":                Int,
		"Join":               Join,
		"IsContract":         IsContract,
		"Len":                Len,
		"Money":              Money,
		"PermColumn":         PermColumn,
		"PermTable":          PermTable,
		"Str":                Str,
		"Substr":             Substr,
		"Replace":            Replace,
		"Size":               Size,
		"Sha256":             Sha256,
		"ToLower":            strings.ToLower,
		"TrimSpace":          strings.TrimSpace,
		"TableConditions":    TableConditions,
		"UpdateLang":         UpdateLang,
		"ValidateCondition":  ValidateCondition,
		//   VDE functions only
		"HTTPRequest":  HTTPRequest,
		"GetMapKeys":   GetMapKeys,
		"SortedKeys":   SortedKeys,
		"Date":         Date,
		"HTTPPostJSON": HTTPPostJSON,
	}, AutoPars: map[string]string{
		`*smart.SmartContract`: `sc`,
	}})
	vmExtendCost(vm, getCost)
	vmFuncCallsDB(vm, funcCallsDB)
}

func getTableName(sc *SmartContract, tblname string, ecosystem int64) string {
	if tblname[0] < '1' || tblname[0] > '9' || !strings.Contains(tblname, `_`) {
		prefix := converter.Int64ToStr(ecosystem)
		if sc.VDE {
			prefix += `_vde`
		}
		tblname = fmt.Sprintf(`%s_%s`, prefix, tblname)
	}
	return tblname
}

func getDefTableName(sc *SmartContract, tblname string) string {
	return getTableName(sc, tblname, sc.TxSmart.EcosystemID)
}

func accessContracts(sc *SmartContract, names ...string) bool {
	var prefix string
	if !sc.VDE {
		prefix = `@1`
	} else {
		prefix = fmt.Sprintf(`@%d`, sc.TxSmart.EcosystemID)
	}
	for _, item := range names {
		if sc.TxContract.Name == prefix+item {
			return true
		}
	}
	return false
}

// CompileContract is compiling contract
func CompileContract(sc *SmartContract, code string, state, id, token int64) (interface{}, error) {
	if !accessContracts(sc, `NewContract`, `EditContract`, `Import`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("CompileContract can be only called from NewContract or EditContract")
		return 0, fmt.Errorf(`CompileContract can be only called from NewContract or EditContract`)
	}
	return VMCompileBlock(sc.VM, code, &script.OwnerInfo{StateID: uint32(state), WalletID: id, TokenID: token})
}

// ContractAccess checks whether the name of the executable contract matches one of the names listed in the parameters.
func ContractAccess(sc *SmartContract, names ...interface{}) bool {
	for _, iname := range names {
		switch name := iname.(type) {
		case string:
			if len(name) > 0 {
				if name[0] != '@' {
					name = fmt.Sprintf(`@%d`, sc.TxSmart.EcosystemID) + name
				}
				if sc.TxContract.StackCont[len(sc.TxContract.StackCont)-1] == name {
					return true
				}
			}
		}
	}
	return false
}

// ContractConditions calls the 'conditions' function for each of the contracts specified in the parameters
func ContractConditions(sc *SmartContract, names ...interface{}) (bool, error) {
	for _, iname := range names {
		name := iname.(string)
		if len(name) > 0 {
			contract := VMGetContract(sc.VM, name, uint32(sc.TxSmart.EcosystemID))
			if contract == nil {
				contract = VMGetContract(sc.VM, name, 0)
				if contract == nil {
					log.WithFields(log.Fields{"contract_name": name, "type": consts.NotFound}).Error("Unknown contract")
					return false, fmt.Errorf(`Unknown contract %s`, name)
				}
			}
			block := contract.GetFunc(`conditions`)
			if block == nil {
				log.WithFields(log.Fields{"contract_name": name, "type": consts.EmptyObject}).Error("There is not conditions in contract")
				return false, fmt.Errorf(`There is not conditions in contract %s`, name)
			}
			_, err := VMRun(sc.VM, block, []interface{}{}, &map[string]interface{}{`ecosystem_id`: int64(sc.TxSmart.EcosystemID),
				`key_id`: sc.TxSmart.KeyID, `sc`: sc})
			if err != nil {
				return false, err
			}
		} else {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty contract name in ContractConditions")
			return false, fmt.Errorf(`empty contract name in ContractConditions`)
		}
	}
	return true, nil
}

func contractsList(value string) []interface{} {
	list := ContractsList(value)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = reflect.ValueOf(list[i]).Interface()
	}
	return result
}

// CreateTable is creating smart contract table
func CreateTable(sc *SmartContract, name string, columns, permissions string) error {
	var err error
	if !accessContracts(sc, `NewTable`, `Import`) {
		return fmt.Errorf(`CreateTable can be only called from NewTable`)
	}
	tableName := getDefTableName(sc, name)

	var cols []map[string]string
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling columns to JSON")
		return err
	}
	indexes := make([]string, 0)

	colsSQL := ""
	colperm := make(map[string]string)
	colList := make(map[string]bool)
	for _, data := range cols {
		colname := strings.ToLower(data[`name`])
		if colList[colname] {
			return fmt.Errorf(`There are the same columns`)
		}
		colList[colname] = true
		var colType string
		colDef := ``
		switch data[`type`] {
		case "varchar":
			colType = `varchar(102400)`
		case "character":
			colType = `character(1)`
			colDef = `NOT NULL DEFAULT '0'`
		case "number":
			colType = `bigint`
			colDef = `NOT NULL DEFAULT '0'`
		case "datetime":
			colType = `timestamp`
		case "double":
			colType = `double precision`
		case "money":
			colType = `decimal (30, 0)`
			colDef = `NOT NULL DEFAULT '0'`
		default:
			colType = data[`type`]
		}
		colsSQL += `"` + colname + `" ` + colType + " " + colDef + " ,\n"
		colperm[colname] = data[`conditions`]
		if data[`index`] == "1" {
			indexes = append(indexes, colname)
		}
	}
	colout, err := json.Marshal(colperm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling columns to JSON")
		return err
	}
	if sc.VDE {
		err = model.CreateVDETable(sc.DbTransaction, tableName, strings.TrimRight(colsSQL, ",\n"))
	} else {
		err = model.CreateTable(sc.DbTransaction, tableName, colsSQL)
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE tables")
		return err
	}

	for _, index := range indexes {
		err := model.CreateIndex(sc.DbTransaction, tableName+"_"+index, tableName, index)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE indexes")
			return err
		}
	}
	var perm map[string]string
	permlist := make(map[string]string)
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling permissions to JSON")
		return err
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		permlist[v] = perm[v]
	}
	permout, err := json.Marshal(permlist)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling permissions to JSON")
		return err
	}
	prefix, name := PrefixName(tableName)
	var state string
	if !sc.VDE {
		state = `@1`
	}
	id, err := model.GetNextID(sc.DbTransaction, getDefTableName(sc, `tables`))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next ID")
		return err
	}

	t := &model.TableVDE{
		ID:          id,
		Name:        name,
		Columns:     string(colout),
		Permissions: string(permout),
		Conditions:  fmt.Sprintf(`ContractAccess("%sEditTable")`, state),
	}
	t.SetTablePrefix(prefix)
	err = t.Create(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("insert vde table info")
		return err
	}
	return nil
}

// DBInsert inserts a record into the specified database table
func DBInsert(sc *SmartContract, tblname string, params string, val ...interface{}) (qcost int64, ret int64, err error) {
	tblname = getDefTableName(sc, tblname)
	if err = sc.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var ind int
	var lastID string
	if ind, err = model.NumIndexes(tblname); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("num indexes")
		return
	}
	if len(val) == 0 {
		err = fmt.Errorf(`values are undefined`)
		return
	}
	if reflect.TypeOf(val[0]) == reflect.TypeOf([]interface{}{}) {
		val = val[0].([]interface{})
	}
	qcost, lastID, err = sc.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, !sc.VDE)
	if ind > 0 {
		qcost *= int64(ind)
	}
	if err == nil {
		ret, _ = strconv.ParseInt(lastID, 10, 64)
	}
	return
}

// DBSelect returns an array of values of the specified columns when there is selection of data 'offset', 'limit', 'where'
func DBSelect(sc *SmartContract, tblname string, columns string, id int64, order string, offset, limit, ecosystem int64,
	where string, params []interface{}) (int64, []interface{}, error) {

	var (
		err  error
		rows *sql.Rows
	)
	if len(columns) == 0 {
		columns = `*`
	}
	if len(order) == 0 {
		order = `id`
	}
	where = strings.Replace(converter.Escape(where), `$`, `?`, -1)
	if id != 0 {
		where = fmt.Sprintf(`id='%d'`, id)
		limit = 1
	}
	if limit == 0 {
		limit = 25
	}
	if limit < 0 || limit > 250 {
		limit = 250
	}
	if ecosystem == 0 {
		ecosystem = sc.TxSmart.EcosystemID
	}
	tblname = getTableName(sc, tblname, ecosystem)
	rows, err = model.DBConn.Table(tblname).Select(columns).Where(where, params...).Order(order).
		Offset(offset).Limit(limit).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting rows from table")
		return 0, nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rows columns")
		return 0, nil, err
	}
	values := make([][]byte, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	result := make([]interface{}, 0, 50)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("scanning next row")
			return 0, nil, err
		}
		row := make(map[string]string)
		for i, col := range values {
			var value string
			if col != nil {
				value = string(col)
			}
			row[cols[i]] = value
		}
		result = append(result, reflect.ValueOf(row).Interface())
	}
	return 0, result, nil
}

// DBUpdate updates the item with the specified id in the table
func DBUpdate(sc *SmartContract, tblname string, id int64, params string, val ...interface{}) (qcost int64, err error) {
	tblname = getDefTableName(sc, tblname)
	if err = sc.AccessTable(tblname, "update"); err != nil {
		return
	}
	if strings.Contains(tblname, `_reports_`) {
		err = fmt.Errorf(`Access denied to report table`)
		return
	}
	columns := strings.Split(params, `,`)
	if err = sc.AccessColumns(tblname, columns); err != nil {
		return
	}
	qcost, _, err = sc.selectiveLoggingAndUpd(columns, val, tblname, []string{`id`}, []string{converter.Int64ToStr(id)}, !sc.VDE)
	return
}

// EcosysParam returns the value of the specified parameter for the ecosystem
func EcosysParam(sc *SmartContract, name string) string {
	val, _ := model.Single(`SELECT value FROM "`+getDefTableName(sc, `parameters`)+`" WHERE name = ?`, name).String()
	return val
}

// Eval evaluates the condition
func Eval(sc *SmartContract, condition string) error {
	if len(condition) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("The condition is empty")
		return fmt.Errorf(`The condition is empty`)
	}
	ret, err := sc.EvalIf(condition)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("eval condition")
		return err
	}
	if !ret {
		log.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
		return fmt.Errorf(`Access denied`)
	}
	return nil
}

// FlushContract is flushing contract
func FlushContract(sc *SmartContract, iroot interface{}, id int64, active bool) error {
	if !accessContracts(sc, `NewContract`, `EditContract`, `Import`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("FlushContract can be only called from NewContract or EditContract")
		return fmt.Errorf(`FlushContract can be only called from NewContract or EditContract`)
	}
	root := iroot.(*script.Block)
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			root.Children[i].Info.(*script.ContractInfo).Owner.TableID = id
			root.Children[i].Info.(*script.ContractInfo).Owner.Active = active
		}
	}
	VMFlushBlock(sc.VM, root)
	return nil
}

// IsContract returns true if there is the specified contract
func IsContract(sc *SmartContract, name string, state int64) bool {
	return VMGetContract(sc.VM, name, uint32(state)) != nil
}

// Len returns the length of the slice
func Len(in []interface{}) int64 {
	if in == nil {
		return 0
	}
	return int64(len(in))
}

// PermTable is changing permission of table
func PermTable(sc *SmartContract, name, permissions string) error {
	if !accessContracts(sc, `EditTable`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("EditTable can be only called from @1EditTable")
		return fmt.Errorf(`PermTable can be only called from EditTable`)
	}
	var perm map[string]string
	permlist := make(map[string]string)
	err := json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling table permissions to json")
		return err
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		permlist[v] = perm[v]
	}
	permout, err := json.Marshal(permlist)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling permission list to json")
		return err
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`permissions`}, []interface{}{string(permout)},
		getDefTableName(sc, `tables`), []string{`name`}, []string{name}, !sc.VDE)
	return err
}

// TableConditions is contract func
func TableConditions(sc *SmartContract, name, columns, permissions string) (err error) {
	isEdit := len(columns) == 0

	if isEdit {
		if !accessContracts(sc, `EditTable`) {
			log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("TableConditions can be only called from @1EditTable")
			return fmt.Errorf(`TableConditions can be only called from EditTable`)
		}
	} else if !accessContracts(sc, `NewTable`, `Import`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("TableConditions can be only called from @1NewTable")
		return fmt.Errorf(`TableConditions can be only called from NewTable or Import`)
	}

	prefix := converter.Int64ToStr(sc.TxSmart.EcosystemID)
	if sc.VDE {
		prefix += `_vde`
	}

	t := &model.Table{}
	t.SetTablePrefix(prefix)
	exists, err := t.ExistsByName(sc.DbTransaction, name)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("table is exists")
		return err
	}
	if isEdit {
		if !exists {
			log.WithFields(log.Fields{"table_name": name, "type": consts.NotFound}).Error("table does not exists")
			return fmt.Errorf(`table %s doesn't exist`, name)
		}
	} else if exists {
		log.WithFields(log.Fields{"table_name": name, "type": consts.Found}).Error("table exists")
		return fmt.Errorf(`table %s exists`, name)
	}

	var perm map[string]string
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling permissions from json")
		return
	}
	if len(perm) != 3 {
		log.WithFields(log.Fields{"size": len(perm), "type": consts.InvalidObject}).Error("permissions must contain insert, new_column, and update")
		return fmt.Errorf(`Permissions must contain "insert", "new_column", "update"`)
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		if len(perm[v]) == 0 {
			log.WithFields(log.Fields{"condition_type": v, "type": consts.EmptyObject}).Error("condition is empty")
			return fmt.Errorf(`%v condition is empty`, v)
		}
		if err = VMCompileEval(sc.VM, perm[v], uint32(sc.TxSmart.EcosystemID)); err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("compile evaluating permissions")
			return err
		}
	}
	if isEdit {
		if err = sc.AccessTable(name, `update`); err != nil {
			if err = sc.AccessRights(`changing_tables`, false); err != nil {
				return err
			}
		}
		return nil
	}

	var cols []map[string]string
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling columns permissions from json")
		return
	}
	if len(cols) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Columns are empty")
		return fmt.Errorf(`len(cols) == 0`)
	}
	if len(cols) > syspar.GetMaxColumns() {
		log.WithFields(log.Fields{"size": len(cols), "max_size": syspar.GetMaxColumns(), "type": consts.ParameterExceeded}).Error("Too many columns")
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	var indexes int
	for _, data := range cols {
		if len(data[`name`]) == 0 || len(data[`type`]) == 0 {
			log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("wrong column")
			return fmt.Errorf(`worng column`)
		}
		itype := data[`type`]
		if itype != `varchar` && itype != `number` && itype != `datetime` && itype != `text` &&
			itype != `bytea` && itype != `double` && itype != `money` && itype != `character` {
			log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect type")
			return fmt.Errorf(`incorrect type`)
		}
		if len(data[`conditions`]) == 0 {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Conditions is empty")
			return fmt.Errorf(`Conditions is empty`)
		}
		if err = VMCompileEval(sc.VM, data[`conditions`], uint32(sc.TxSmart.EcosystemID)); err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError}).Error("compile eval conditions")
			return err
		}
		if data[`index`] == `1` {
			if itype != `varchar` && itype != `number` && itype != `datetime` {
				log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect index type")
				return fmt.Errorf(`incorrect index type`)
			}
			indexes++
		}

	}
	if indexes > syspar.GetMaxIndexes() {
		log.WithFields(log.Fields{"size": indexes, "max_size": syspar.GetMaxIndexes, "type": consts.ParameterExceeded}).Error("Too many indexes")
		return fmt.Errorf(`Too many indexes. Limit is %d`, syspar.GetMaxIndexes())
	}

	if err := sc.AccessRights("new_table", false); err != nil {
		return err
	}

	return nil
}

// ValidateCondition checks if the condition can be compiled
func ValidateCondition(sc *SmartContract, condition string, state int64) error {
	if len(condition) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("conditions cannot be empty")
		return fmt.Errorf("Conditions cannot be empty")
	}
	return VMCompileEval(sc.VM, condition, uint32(state))
}

// ColumnCondition is contract func
func ColumnCondition(sc *SmartContract, tableName, name, coltype, permissions, index string) error {
	if !accessContracts(sc, `NewColumn`, `EditColumn`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("ColumnConditions can be only called from @1NewColumn")
		return fmt.Errorf(`ColumnCondition can be only called from NewColumn or EditColumn`)
	}
	isExist := strings.HasSuffix(sc.TxContract.Name, `EditColumn`)
	tEx := &model.Table{}
	prefix := converter.Int64ToStr(sc.TxSmart.EcosystemID)
	if sc.VDE {
		prefix += `_vde`
	}
	tEx.SetTablePrefix(prefix)
	name = strings.ToLower(name)

	exists, err := tEx.IsExistsByPermissionsAndTableName(name, tableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("querying that table is exists by permissions and table name")
		return err
	}
	if isExist {
		if !exists {
			log.WithFields(log.Fields{"column_name": name, "type": consts.NotFound}).Error("column does not exists")
			return fmt.Errorf(`column %s doesn't exists`, name)
		}
	} else if exists {
		log.WithFields(log.Fields{"column_name": name, "type": consts.Found}).Error("column exists")
		return fmt.Errorf(`column %s exists`, name)
	}
	if len(permissions) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Permissions are empty")
		return fmt.Errorf(`Permissions is empty`)
	}
	if err = VMCompileEval(sc.VM, permissions, uint32(sc.TxSmart.EcosystemID)); err != nil {
		return err
	}
	tblName := getDefTableName(sc, tableName)
	if isExist {
		return sc.AccessTable(tblName, `update`)
	}
	count, err := model.GetColumnCount(tblName)
	if err != nil {
		log.WithFields(log.Fields{"table": tblName, "type": consts.DBError}).Error("counting table columns")
		return err
	}
	if count >= int64(syspar.GetMaxColumns()) {
		log.WithFields(log.Fields{"size": count, "max_size": syspar.GetMaxColumns(), "type": consts.ParameterExceeded}).Error("Too many columns")
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	if coltype != `varchar` && coltype != `number` && coltype != `datetime` && coltype != `character` &&
		coltype != `text` && coltype != `bytea` && coltype != `double` && coltype != `money` {
		log.WithFields(log.Fields{"column_type": coltype, "type": consts.InvalidObject}).Error("Unknown column type")
		return fmt.Errorf(`incorrect type`)
	}
	if index == `1` {
		count, err := model.NumIndexes(tblName)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("num indexes")
			return err
		}
		if count >= syspar.GetMaxIndexes() {
			log.WithFields(log.Fields{"size": count, "max_size": syspar.GetMaxIndexes(), "type": consts.ParameterExceeded}).Error("Too many indexes")
			return fmt.Errorf(`Too many indexes. Limit is %d`, syspar.GetMaxIndexes())
		}
		if coltype != `varchar` && coltype != `number` && coltype != `datetime` {
			log.WithFields(log.Fields{"column_type": coltype, "type": consts.InvalidObject}).Error("incorrect index type")
			return fmt.Errorf(`incorrect index type`)
		}
	}

	return sc.AccessTable(tblName, "new_column")
}

// CreateColumn is creating column
func CreateColumn(sc *SmartContract, tableName, name, coltype, permissions, index string) error {
	if !accessContracts(sc, `NewColumn`) {
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("CreateColumn can be only called from @1NewColumn")
		return fmt.Errorf(`CreateColumn can be only called from NewColumn`)
	}
	name = strings.ToLower(name)
	tblname := getDefTableName(sc, tableName)

	var colType string
	switch coltype {
	case "varchar":
		colType = `varchar(102400)`
	case "number":
		colType = `bigint NOT NULL DEFAULT '0'`
	case "character":
		colType = `character(1) NOT NULL DEFAULT '0'`
	case "datetime":
		colType = `timestamp`
	case "double":
		colType = `double precision`
	case "money":
		colType = `decimal (30, 0) NOT NULL DEFAULT '0'`
	default:
		colType = coltype
	}
	err := model.AlterTableAddColumn(sc.DbTransaction, tblname, name, colType)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("adding column to the table")
		return err
	}

	if index == "1" {
		err = model.CreateIndex(sc.DbTransaction, tblname+"_"+name+"_index", tblname, name)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating index for table")
			return err
		}
	}
	tables := getDefTableName(sc, `tables`)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err = model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		return err
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting columns from the table")
		return err
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling columns to json")
		return err
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName}, !sc.VDE)
	if err != nil {
		return err
	}

	return nil
}

// PermColumn is contract func
func PermColumn(sc *SmartContract, tableName, name, permissions string) error {
	if !accessContracts(sc, `EditColumn`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("EditColumn can be only called from @1EditColumn")
		return fmt.Errorf(`EditColumn can be only called from EditColumn`)
	}
	name = strings.ToLower(name)
	tables := getDefTableName(sc, `tables`)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err := model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("querying columns by table name")
		return err
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling columns permissions from json")
		return err
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling column permissions to json")
		return err
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName}, !sc.VDE)
	return err
}

// AddressToID converts the string representation of the wallet number to a numeric
func AddressToID(input string) (addr int64) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return 0
	}
	if input[0] == '-' {
		addr, _ = strconv.ParseInt(input, 10, 64)
	} else if strings.Count(input, `-`) == 4 {
		addr = converter.StringToAddress(input)
	} else {
		uaddr, _ := strconv.ParseUint(input, 10, 64)
		addr = int64(uaddr)
	}
	if !converter.IsValidAddress(converter.AddressToString(addr)) {
		return 0
	}
	return
}

// IDToAddress converts the identifier of account to a string of the form XXXX -...- XXXX
func IDToAddress(id int64) (out string) {
	out = converter.AddressToString(id)
	if !converter.IsValidAddress(out) {
		out = `invalid`
	}
	return
}

func HMac(key, data string, raw_output bool) (ret string, err error) {
	hash, err := crypto.GetHMAC(key, data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("getting HMAC")
		return ``, err
	}
	if raw_output {
		return string(hash), nil
	} else {
		return hex.EncodeToString(hash), nil
	}
}

//Returns the array of keys of the map
func GetMapKeys(in map[string]interface{}) []interface{} {
	keys := make([]interface{}, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	return keys
}

//Returns the sorted array of keys of the map
func SortedKeys(m map[string]interface{}) []interface{} {
	i, sorted := 0, make([]string, len(m))
	for k := range m {
		sorted[i] = k
		i++
	}
	sort.Strings(sorted)

	ret := make([]interface{}, len(sorted))
	for k, v := range sorted {
		ret[k] = v
	}
	return ret
}

//Formats timestamp to specified date format
func Date(time_format string, timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(time_format)
}

// HTTPRequest sends http request
func HTTPRequest(requrl, method string, headers map[string]interface{},
	params map[string]interface{}) (string, error) {

	var ioform io.Reader

	form := &url.Values{}
	client := &http.Client{}
	for key, v := range params {
		form.Set(key, fmt.Sprint(v))
	}
	if len(*form) > 0 {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, requrl, ioform)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("new http request")
		return ``, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, v := range headers {
		req.Header.Set(key, fmt.Sprint(v))
	}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("http request")
		return ``, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading http answer")
		return ``, err
	}
	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("http status code")
		return ``, fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return string(data), nil
}

// HTTPPostJSON sends post http request with json
func HTTPPostJSON(requrl string, headers map[string]interface{}, json_str string) (string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer([]byte(json_str)))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("new http request")
		return ``, err
	}

	for key, v := range headers {
		req.Header.Set(key, fmt.Sprint(v))
	}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("http request")
		return ``, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading http answer")
		return ``, err
	}
	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("http status code")
		return ``, fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return string(data), nil
}
