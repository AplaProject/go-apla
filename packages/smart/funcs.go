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
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/scheduler"
	"github.com/GenesisKernel/go-genesis/packages/scheduler/contract"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/satori/go.uuid"

	"github.com/shopspring/decimal"
)

const nodeBanNotificationHeader = "Your node was banned"

type permTable struct {
	Insert    string `json:"insert"`
	Update    string `json:"update"`
	NewColumn string `json:"new_column"`
	Read      string `json:"read,omitempty"`
	Filter    string `json:"filter,omitempty"`
}

type permColumn struct {
	Update string `json:"update"`
	Read   string `json:"read,omitempty"`
}

// SmartContract is storing smart contract data
type SmartContract struct {
	VDE           bool
	Rollback      bool
	SysUpdate     bool
	VM            *script.VM
	TxSmart       tx.SmartContract
	TxData        map[string]interface{}
	TxContract    *Contract
	TxFuel        int64           // The fuel of executing contract
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
		"SetPubKey":   {},
	}
	// map for table name to parameter with conditions
	tableParamConditions = map[string]string{
		"pages":      "changing_page",
		"menu":       "changing_menu",
		"signatures": "changing_signature",
		"contracts":  "changing_contracts",
		"blocks":     "changing_blocks",
		"languages":  "changing_language",
		"tables":     "changing_tables",
	}
	typeToPSQL = map[string]string{
		`json`:      `jsonb`,
		`varchar`:   `varchar(102400)`,
		`character`: `character(1) NOT NULL DEFAULT '0'`,
		`number`:    `bigint NOT NULL DEFAULT '0'`,
		`datetime`:  `timestamp`,
		`double`:    `double precision`,
		`money`:     `decimal (30, 0) NOT NULL DEFAULT '0'`,
		`text`:      `text`,
		`bytea`:     `bytea`,
	}
)

// EmbedFuncs is extending vm with embedded functions
func EmbedFuncs(vm *script.VM, vt script.VMType) {
	f := map[string]interface{}{
		"AddressToId":          AddressToID,
		"ColumnCondition":      ColumnCondition,
		"CompileContract":      CompileContract,
		"Contains":             strings.Contains,
		"ContractAccess":       ContractAccess,
		"ContractConditions":   ContractConditions,
		"ContractsList":        contractsList,
		"CreateColumn":         CreateColumn,
		"CreateTable":          CreateTable,
		"DBInsert":             DBInsert,
		"DBSelect":             DBSelect,
		"DBUpdate":             DBUpdate,
		"DBUpdateSysParam":     UpdateSysParam,
		"DBUpdateExt":          DBUpdateExt,
		"EcosysParam":          EcosysParam,
		"AppParam":             AppParam,
		"SysParamString":       SysParamString,
		"SysParamInt":          SysParamInt,
		"SysFuel":              SysFuel,
		"Eval":                 Eval,
		"EvalCondition":        EvalCondition,
		"Float":                Float,
		"FlushContract":        FlushContract,
		"GetContractByName":    GetContractByName,
		"GetContractById":      GetContractById,
		"HMac":                 HMac,
		"Join":                 Join,
		"JSONToMap":            JSONDecode, // Deprecated
		"JSONDecode":           JSONDecode,
		"JSONEncode":           JSONEncode,
		"IdToAddress":          IDToAddress,
		"Int":                  Int,
		"IsObject":             IsObject,
		"Len":                  Len,
		"Money":                Money,
		"PermColumn":           PermColumn,
		"PermTable":            PermTable,
		"Random":               Random,
		"Split":                Split,
		"Str":                  Str,
		"Substr":               Substr,
		"Replace":              Replace,
		"Size":                 Size,
		"Sha256":               Sha256,
		"PubToID":              PubToID,
		"HexToBytes":           HexToBytes,
		"LangRes":              LangRes,
		"HasPrefix":            strings.HasPrefix,
		"ValidateCondition":    ValidateCondition,
		"TrimSpace":            strings.TrimSpace,
		"ToLower":              strings.ToLower,
		"CreateEcosystem":      CreateEcosystem,
		"RollbackEcosystem":    RollbackEcosystem,
		"RollbackTable":        RollbackTable,
		"TableConditions":      TableConditions,
		"RollbackColumn":       RollbackColumn,
		"CreateLanguage":       CreateLanguage,
		"EditLanguage":         EditLanguage,
		"Activate":             Activate,
		"Deactivate":           Deactivate,
		"SetContractWallet":    SetContractWallet,
		"RollbackContract":     RollbackContract,
		"RollbackEditContract": RollbackEditContract,
		"check_signature":      CheckSignature,
		"RowConditions":        RowConditions,
		"UUID":                 UUID,
		"DecodeBase64":         DecodeBase64,
		"EncodeBase64":         EncodeBase64,
		"MD5":                  MD5,
		"EditEcosysName":       EditEcosysName,
		"GetColumnType":        GetColumnType,
		"GetType":              GetType,
		"AllowChangeCondition": AllowChangeCondition,
		"StringToBytes":        StringToBytes,
		"BytesToString":        BytesToString,
		"SetPubKey":            SetPubKey,
		"NewMoney":             NewMoney,
	}

	switch vt {
	case script.VMTypeVDE:
		f["HTTPRequest"] = HTTPRequest
		f["GetMapKeys"] = GetMapKeys
		f["SortedKeys"] = SortedKeys
		f["Date"] = Date
		f["HTTPPostJSON"] = HTTPPostJSON
		f["ValidateCron"] = ValidateCron
		f["UpdateCron"] = UpdateCron
		vmFuncCallsDB(vm, funcCallsDB)
	case script.VMTypeSmart:
		f["GetBlock"] = GetBlock
		f["UpdateNodesBan"] = UpdateNodesBan
		f["DBSelectMetrics"] = DBSelectMetrics
		f["DBCollectMetrics"] = DBCollectMetrics
		ExtendCost(getCostP)
		FuncCallsDB(funcCallsDBP)
	}

	vmExtend(vm, &script.ExtendData{Objects: f, AutoPars: map[string]string{
		`*smart.SmartContract`: `sc`,
	}})
}

func GetTableName(sc *SmartContract, tblname string, ecosystem int64) string {
	if len(tblname) > 0 && tblname[0] == '@' {
		return strings.ToLower(tblname[1:])
	}
	return strings.ToLower(fmt.Sprintf(`%s_%s`, converter.Int64ToStr(ecosystem), tblname))
}

func getDefTableName(sc *SmartContract, tblname string) string {
	return GetTableName(sc, tblname, sc.TxSmart.EcosystemID)
}

func accessContracts(sc *SmartContract, names ...string) bool {
	for _, item := range names {
		if sc.TxContract.Name == `@1`+item {
			return true
		}
	}
	return false
}

// CompileContract is compiling contract
func CompileContract(sc *SmartContract, code string, state, id, token int64) (interface{}, error) {
	if err := validateAccess(`CompileContract`, sc, nNewContract, nEditContract, nImport); err != nil {
		return nil, err
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
					return false, logErrorValue(fmt.Errorf(eUnknownContract, name), consts.NotFound,
						"Unknown contract", name)
				}
			}
			block := contract.GetFunc(`conditions`)
			if block == nil {
				return false, logErrorValue(fmt.Errorf(eContractCondition, name), consts.EmptyObject,
					"There is not conditions in contract", name)
			}
			_, err := VMRun(sc.VM, block, []interface{}{}, &map[string]interface{}{`ecosystem_id`: int64(sc.TxSmart.EcosystemID),
				`key_id`: sc.TxSmart.KeyID, `sc`: sc, `original_contract`: ``, `this_contract`: ``, `role_id`: sc.TxSmart.RoleID})
			if err != nil {
				return false, err
			}
		} else {
			return false, logError(errEmptyContract, consts.EmptyObject, "ContractConditions")
		}
	}
	return true, nil
}

func contractsList(value string) ([]interface{}, error) {
	list, err := script.ContractsList(value)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = reflect.ValueOf(list[i]).Interface()
	}
	return result, err
}

// CreateTable is creating smart contract table
func CreateTable(sc *SmartContract, name, columns, permissions string, applicationID int64) error {
	var err error

	if err := validateAccess(`CreateTable`, sc, nNewTable, nImport); err != nil {
		return err
	}

	if len(name) == 0 {
		return fmt.Errorf("The table name cannot be empty")
	}

	if len(name) > 0 && name[0] == '@' {
		return errTableName
	}

	tableName := getDefTableName(sc, name)
	if model.IsTable(tableName) {
		return fmt.Errorf("table %s exists", name)
	}

	var cols []interface{}
	if err = json.Unmarshal([]byte(columns), &cols); err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, "unmarshalling columns to JSON", columns)
	}

	colsSQL := ""
	colperm := make(map[string]string)
	colList := make(map[string]bool)
	for _, icol := range cols {
		var data map[string]interface{}
		switch v := icol.(type) {
		case string:
			err = json.Unmarshal([]byte(v), &data)
			if err != nil {
				logErrorValue(err, consts.JSONUnmarshallError,
					"unmarshalling columns permissions from json", v)
				return err
			}
		default:
			data = v.(map[string]interface{})
		}
		colname := strings.ToLower(data[`name`].(string))
		if colList[colname] {
			return errSameColumns
		}

		sqlColType, err := columnType(data["type"].(string))
		if err != nil {
			return err
		}

		colList[colname] = true
		colsSQL += `"` + colname + `" ` + sqlColType + " ,\n"
		condition := ``
		switch v := data[`conditions`].(type) {
		case string:
			condition = v
		case map[string]interface{}:
			out, err := json.Marshal(v)
			if err != nil {
				return logError(err, consts.JSONMarshallError, "marshalling conditions to json")
			}
			condition = string(out)
		}
		colperm[colname] = condition
	}
	colout, err := json.Marshal(colperm)
	if err != nil {
		return logError(err, consts.JSONMarshallError, "marshalling columns to json")
	}
	err = model.CreateTable(sc.DbTransaction, tableName, strings.TrimRight(colsSQL, ",\n"))
	if err != nil {
		return logErrorDB(err, "creating tables")
	}

	var perm permTable
	if err = json.Unmarshal([]byte(permissions), &perm); err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, "unmarshalling permissions to JSON",
			permissions)
	}
	permout, err := json.Marshal(perm)
	if err != nil {
		return logError(err, consts.JSONMarshallError, "marshalling permissions to JSON")
	}
	prefix, name := PrefixName(tableName)
	id, err := model.GetNextID(sc.DbTransaction, getDefTableName(sc, `tables`))
	if err != nil {
		return logErrorDB(err, "getting next ID")
	}

	t := &model.TableVDE{
		ID:          id,
		Name:        name,
		Columns:     string(colout),
		Permissions: string(permout),
		Conditions:  `ContractAccess("@1EditTable")`,
		AppID:       applicationID,
	}
	t.SetTablePrefix(prefix)
	err = t.Create(sc.DbTransaction)
	if err != nil {
		return logErrorDB(err, "insert vde table info")
	}
	rollbackTx := &model.RollbackTx{
		BlockID:   sc.BlockData.BlockID,
		TxHash:    sc.TxHash,
		NameTable: tableName,
		TableID:   converter.Int64ToStr(id),
	}
	if err = rollbackTx.Create(sc.DbTransaction); err != nil {
		return logErrorDB(err, "creating CreateTable rollback")
	}
	return nil
}

func columnType(colType string) (string, error) {
	if sqlColType, ok := typeToPSQL[colType]; ok {
		return sqlColType, nil
	}
	return ``, fmt.Errorf(eColumnType, colType)
}

// DBInsert inserts a record into the specified database table
func DBInsert(sc *SmartContract, tblname string, params string, val ...interface{}) (qcost int64, ret int64, err error) {
	if tblname == "system_parameters" {
		return 0, 0, fmt.Errorf("system parameters access denied")
	}

	tblname = getDefTableName(sc, tblname)
	if err = sc.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var ind int
	var lastID string
	if ind, err = model.NumIndexes(tblname); err != nil {
		err = logErrorDB(err, "num indexes")
		return
	}
	if len(val) == 0 {
		err = errValues
		return
	}
	if reflect.TypeOf(val[0]) == reflect.TypeOf([]interface{}{}) {
		val = val[0].([]interface{})
	}
	qcost, lastID, err = sc.insert(strings.Split(params, `,`), val, tblname)
	if ind > 0 {
		qcost *= int64(ind)
	}
	if err == nil {
		ret, _ = strconv.ParseInt(lastID, 10, 64)
	}
	return
}

// PrepareColumns replaces jsonb fields -> in the list of columns for db selecting
// For example, name,doc->title => name,doc::jsonb->>'title' as "doc.title"
func PrepareColumns(columns string) string {
	colList := make([]string, 0)
	for _, icol := range strings.Split(columns, `,`) {
		if strings.Contains(icol, `->`) {
			colfield := strings.Split(icol, `->`)
			if len(colfield) == 2 {
				icol = fmt.Sprintf(`%s::jsonb->>'%s' as "%[1]s.%[2]s"`, colfield[0], colfield[1])
			} else {
				icol = fmt.Sprintf(`%s::jsonb#>>'{%s}' as "%[1]s.%[3]s"`, colfield[0],
					strings.Join(colfield[1:], `,`), strings.Join(colfield[1:], `.`))
			}
		}
		colList = append(colList, icol)
	}
	return strings.Join(colList, `,`)
}

func PrepareWhere(where string) string {
	whereSlice := regexp.MustCompile(`->([\w\d_]+)`).FindAllStringSubmatchIndex(where, -1)
	startWhere := 0
	out := ``
	for i := 0; i < len(whereSlice); i++ {
		slice := whereSlice[i]
		if len(slice) != 4 {
			continue
		}
		if i < len(whereSlice)-1 && slice[1] == whereSlice[i+1][0] {
			colsWhere := []string{where[slice[2]:slice[3]]}
			from := slice[0]
			for i < len(whereSlice)-1 && slice[1] == whereSlice[i+1][0] {
				i++
				slice = whereSlice[i]
				if len(slice) != 4 {
					break
				}
				colsWhere = append(colsWhere, where[slice[2]:slice[3]])
			}
			out += fmt.Sprintf(`%s::jsonb#>>'{%s}'`, where[startWhere:from], strings.Join(colsWhere, `,`))
			startWhere = slice[3]
		} else {
			out += fmt.Sprintf(`%s->>'%s'`, where[startWhere:slice[0]], where[slice[2]:slice[3]])
			startWhere = slice[3]
		}
	}
	if len(out) > 0 {
		return out + where[startWhere:]
	}
	return where
}

// DBSelect returns an array of values of the specified columns when there is selection of data 'offset', 'limit', 'where'
func DBSelect(sc *SmartContract, tblname string, columns string, id int64, order string, offset, limit, ecosystem int64,
	where string, params []interface{}) (int64, []interface{}, error) {

	var (
		err  error
		rows *sql.Rows
		perm map[string]string
	)
	if len(columns) == 0 {
		columns = `*`
	}
	columns = strings.ToLower(columns)
	if len(order) == 0 {
		order = `id`
	}
	where = PrepareWhere(strings.Replace(converter.Escape(where), `$`, `?`, -1))
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
	tblname = GetTableName(sc, tblname, ecosystem)
	if sc.VDE {
		perm, err = sc.AccessTablePerm(tblname, `read`)
		if err != nil {
			return 0, nil, err
		}
		cols := strings.Split(columns, `,`)
		if err = sc.AccessColumns(tblname, &cols, false); err != nil {
			return 0, nil, err
		}
		columns = strings.Join(cols, `,`)
	}
	columns = PrepareColumns(columns)

	rows, err = model.GetDB(sc.DbTransaction).Table(tblname).Select(columns).Where(where, params...).Order(order).
		Offset(offset).Limit(limit).Rows()
	if err != nil {
		return 0, nil, logErrorDB(err, "selecting rows from table")
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return 0, nil, logErrorDB(err, "getting rows columns")
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
			return 0, nil, logErrorDB(err, "scanning next row")
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
	if perm != nil && len(perm[`filter`]) > 0 {
		fltResult, err := VMEvalIf(sc.VM, perm[`filter`], uint32(sc.TxSmart.EcosystemID),
			&map[string]interface{}{
				`data`: result, `original_contract`: ``, `this_contract`: ``,
				`ecosystem_id`: sc.TxSmart.EcosystemID,
				`key_id`:       sc.TxSmart.KeyID, `sc`: sc,
				`block_time`: 0, `time`: sc.TxSmart.Time})
		if err != nil {
			return 0, nil, err
		}
		if !fltResult {
			return 0, nil, errAccessDenied
		}
	}
	return 0, result, nil
}

// DBUpdate updates the item with the specified id in the table
func DBUpdate(sc *SmartContract, tblname string, id int64, params string, val ...interface{}) (qcost int64, err error) {
	if tblname == "system_parameters" {
		return 0, fmt.Errorf("system parameters access denied")
	}

	tblname = getDefTableName(sc, tblname)
	if err = sc.AccessTable(tblname, "update"); err != nil {
		return
	}
	columns := strings.Split(params, `,`)
	if err = sc.AccessColumns(tblname, &columns, true); err != nil {
		return
	}
	qcost, _, err = sc.update(columns, val, tblname, []string{`id`}, []string{converter.Int64ToStr(id)})
	return
}

// EcosysParam returns the value of the specified parameter for the ecosystem
func EcosysParam(sc *SmartContract, name string) string {
	val, _ := model.Single(`SELECT value FROM "`+getDefTableName(sc, `parameters`)+`" WHERE name = ?`, name).String()
	return val
}

// AppParam returns the value of the specified app parameter for the ecosystem
func AppParam(sc *SmartContract, app int64, name string) (string, error) {
	ap := &model.AppParam{}
	ap.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID))
	_, err := ap.Get(sc.DbTransaction, app, name)
	if err != nil {
		return ``, logErrorDB(err, "getting app param")
	}
	return ap.Value, nil
}

// Eval evaluates the condition
func Eval(sc *SmartContract, condition string) error {
	if len(condition) == 0 {
		return logError(errEmptyCond, consts.EmptyObject, "The condition is empty")
	}
	ret, err := sc.EvalIf(condition)
	if err != nil {
		return logError(err, consts.EvalError, "eval condition")
	}
	if !ret {
		return logError(errAccessDenied, consts.AccessDenied, "Access denied")
	}
	return nil
}

// FlushContract is flushing contract
func FlushContract(sc *SmartContract, iroot interface{}, id int64, active bool) error {
	if err := validateAccess(`FlushContract`, sc, nNewContract, nEditContract, nImport); err != nil {
		return err
	}
	root := iroot.(*script.Block)
	if id != 0 {
		if len(root.Children) != 1 || root.Children[0].Type != script.ObjContract {
			return errOneContract
		}
	}
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			root.Children[i].Info.(*script.ContractInfo).Owner.TableID = id
			root.Children[i].Info.(*script.ContractInfo).Owner.Active = active
		}
	}
	VMFlushBlock(sc.VM, root)
	return nil
}

// IsObject returns true if there is the specified contract
func IsObject(sc *SmartContract, name string, state int64) bool {
	return VMObjectExists(sc.VM, name, uint32(state))
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
	if err := validateAccess(`PermTable`, sc, nEditTable); err != nil {
		return err
	}
	var perm permTable
	err := json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError,
			"unmarshalling table permissions to json", permissions)
	}
	permout, err := json.Marshal(perm)
	if err != nil {
		return logError(err, consts.JSONMarshallError, "marshalling permission list to json")
	}
	_, _, err = sc.update([]string{`permissions`}, []interface{}{string(permout)},
		getDefTableName(sc, `tables`), []string{`name`}, []string{strings.ToLower(name)})
	return err
}

// TableConditions is contract func
func TableConditions(sc *SmartContract, name, columns, permissions string) (err error) {
	isEdit := len(columns) == 0
	name = strings.ToLower(name)
	if isEdit {
		if err := validateAccess(`TableConditions`, sc, nEditTable); err != nil {
			return err
		}
	} else if err := validateAccess(`TableConditions`, sc, nNewTable, nImport); err != nil {
		return err
	}

	prefix := converter.Int64ToStr(sc.TxSmart.EcosystemID)

	t := &model.Table{}
	t.SetTablePrefix(prefix)
	exists, err := t.ExistsByName(sc.DbTransaction, name)
	if err != nil {
		return logErrorDB(err, "table exists")
	}
	if isEdit {
		if !exists {
			return logErrorValue(fmt.Errorf(eTableNotFound, name), consts.NotFound,
				"table does not exists", name)
		}
	} else if exists {
		return logErrorValue(fmt.Errorf(eTableExists, name), consts.Found, "table exists", name)
	}

	var perm permTable
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, "unmarshalling permissions from json",
			permissions)
	}
	v := reflect.ValueOf(perm)
	for i := 0; i < v.NumField(); i++ {
		cond := v.Field(i).Interface().(string)
		name := v.Type().Field(i).Name
		if len(cond) == 0 && name != `Read` && name != `Filter` {
			return logError(fmt.Errorf(eEmptyCond, name), consts.EmptyObject, "condition is empty")
		}
		if err = VMCompileEval(sc.VM, cond, uint32(sc.TxSmart.EcosystemID)); err != nil {
			return logError(err, consts.EvalError, "compile evaluating permissions")
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

	var cols []interface{}
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError,
			"unmarshalling columns permissions from json", columns)
	}
	if len(cols) == 0 {
		return logError(errUndefColumns, consts.EmptyObject, "Columns are empty")
	}
	if len(cols) > syspar.GetMaxColumns() {
		return logError(fmt.Errorf(eManyColumns, syspar.GetMaxColumns()), consts.ParameterExceeded,
			"Too many columns")
	}
	for _, icol := range cols {
		var data map[string]interface{}
		switch v := icol.(type) {
		case string:
			err = json.Unmarshal([]byte(v), &data)
			if err != nil {
				logErrorValue(err, consts.JSONUnmarshallError,
					"unmarshalling columns permissions from json", v)
				return
			}
		default:
			data = v.(map[string]interface{})
		}
		if data[`name`] == nil || data[`type`] == nil {
			return logError(errWrongColumn, consts.InvalidObject, `wrong column`)
		}
		if len(typeToPSQL[data[`type`].(string)]) == 0 {
			return logError(errIncorrectType, consts.InvalidObject, `incorrect type`)
		}
		condition := ``
		switch v := data[`conditions`].(type) {
		case string:
			condition = v
		case map[string]interface{}:
			out, err := json.Marshal(v)
			if err != nil {
				return logError(err, consts.JSONMarshallError, "marshalling conditions to json")
			}
			condition = string(out)
		}
		perm, err := getPermColumns(condition)
		if err != nil {
			return logError(err, consts.EmptyObject, "Conditions is empty")
		}
		if len(perm.Update) == 0 {
			return logError(errConditionEmpty, consts.EmptyObject, "Update condition is empty")
		}
		if err = VMCompileEval(sc.VM, perm.Update, uint32(sc.TxSmart.EcosystemID)); err != nil {
			return logError(err, consts.EvalError, "compile update conditions")
		}
		if len(perm.Read) > 0 {
			if err = VMCompileEval(sc.VM, perm.Read, uint32(sc.TxSmart.EcosystemID)); err != nil {
				return logError(err, consts.EvalError, "compile read conditions")
			}
		}

	}
	if err := sc.AccessRights("new_table", false); err != nil {
		return err
	}

	return nil
}

// ValidateCondition checks if the condition can be compiled
func ValidateCondition(sc *SmartContract, condition string, state int64) error {
	if len(condition) == 0 {
		return logError(errConditionEmpty, consts.EmptyObject, "conditions cannot be empty")
	}
	return VMCompileEval(sc.VM, condition, uint32(state))
}

// ColumnCondition is contract func
func ColumnCondition(sc *SmartContract, tableName, name, coltype, permissions string) error {
	name = strings.ToLower(name)
	tableName = strings.ToLower(tableName)

	if err := validateAccess(`ColumnCondition`, sc, nNewColumn, nEditColumn); err != nil {
		return err
	}

	isExist := strings.HasSuffix(sc.TxContract.Name, nEditColumn)
	tEx := &model.Table{}
	prefix := converter.Int64ToStr(sc.TxSmart.EcosystemID)
	tEx.SetTablePrefix(prefix)

	exists, err := tEx.IsExistsByPermissionsAndTableName(sc.DbTransaction, name, tableName)
	if err != nil {
		return logErrorDB(err, "querying that table is exists by permissions and table name")
	}
	if isExist {
		if !exists {
			return logError(fmt.Errorf(eColumnNotExist, name), consts.NotFound, "column does not exists")
		}
	} else if exists {
		return logError(fmt.Errorf(eColumnExist, name), consts.Found, "column exists")
	}
	if len(permissions) == 0 {
		return logError(errPermEmpty, consts.EmptyObject, "Permissions are empty")
	}
	perm, err := getPermColumns(permissions)
	if err = VMCompileEval(sc.VM, perm.Update, uint32(sc.TxSmart.EcosystemID)); err != nil {
		return err
	}
	if len(perm.Read) > 0 {
		if err = VMCompileEval(sc.VM, perm.Read, uint32(sc.TxSmart.EcosystemID)); err != nil {
			return err
		}
	}
	tblName := getDefTableName(sc, tableName)
	if isExist {
		return sc.AccessTable(tblName, `update`)
	}
	count, err := model.GetColumnCount(tblName)
	if err != nil {
		return logErrorDB(err, "counting table columns")
	}
	if count >= int64(syspar.GetMaxColumns()) {
		return logError(fmt.Errorf(eManyColumns, syspar.GetMaxColumns()), consts.ParameterExceeded,
			"Too many columns")
	}
	if len(typeToPSQL[coltype]) == 0 {
		return logErrorValue(errIncorrectType, consts.InvalidObject, "Unknown column type", coltype)
	}
	return sc.AccessTable(tblName, "new_column")
}

// AllowChangeCondition check acces to change condition throught supper contract
func AllowChangeCondition(sc *SmartContract, tblname string) error {
	if param, ok := tableParamConditions[tblname]; ok {
		return sc.AccessRights(param, false)
	}
	return nil
}

// RowConditions checks conditions for table row by id
func RowConditions(sc *SmartContract, tblname string, id int64, conditionOnly bool) error {
	escapedTableName := converter.EscapeName(getDefTableName(sc, tblname))
	condition, err := model.GetRowConditionsByTableNameAndID(escapedTableName, id)
	if err != nil {
		return logErrorDB(err, "executing row condition query")
	}

	if len(condition) == 0 {
		return logErrorValue(fmt.Errorf(eItemNotFound, id), consts.NotFound, "record not found",
			tblname)
	}

	if err := Eval(sc, condition); err != nil {
		if err == errAccessDenied && conditionOnly {
			return AllowChangeCondition(sc, tblname)
		}

		return err
	}

	return nil
}

// CreateColumn is creating column
func CreateColumn(sc *SmartContract, tableName, name, colType, permissions string) error {
	if err := validateAccess(`CreateColumn`, sc, nNewColumn); err != nil {
		return err
	}
	name = strings.ToLower(name)
	tableName = strings.ToLower(tableName)
	tblname := getDefTableName(sc, tableName)

	sqlColType, err := columnType(colType)
	if err != nil {
		return err
	}

	err = model.AlterTableAddColumn(sc.DbTransaction, tblname, name, sqlColType)
	if err != nil {
		return logErrorDB(err, "adding column to the table")
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
		return logErrorDB(err, "selecting columns from the table")
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		return logError(err, consts.JSONMarshallError, "marshalling permissions to json")
	}
	_, _, err = sc.update([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName})
	if err != nil {
		return err
	}

	return nil
}

// SetPubKey updates the publis key
func SetPubKey(sc *SmartContract, id int64, pubKey []byte) (qcost int64, err error) {
	if err = validateAccess(`SetPubKey`, sc, nNewUser); err != nil {
		return
	}
	if len(pubKey) == consts.PubkeySizeLength*2 {
		pubKey, err = hex.DecodeString(string(pubKey))
		if err != nil {
			return 0, logError(err, consts.ConversionError, "decoding public key from hex")
		}
	}
	qcost, _, err = sc.update([]string{`pub`}, []interface{}{pubKey},
		getDefTableName(sc, `keys`), []string{`id`}, []string{converter.Int64ToStr(id)})
	return
}

func NewMoney(sc *SmartContract, id int64, amount, comment string) (err error) {
	if err = validateAccess(`NewMoney`, sc, nNewUser); err != nil {
		return err
	}
	_, _, err = sc.insert([]string{`id`, `amount`}, []interface{}{id, amount},
		getDefTableName(sc, `keys`))
	if err == nil {
		var block int64
		if sc.BlockData != nil {
			block = sc.BlockData.BlockID
		}
		_, _, err = sc.insert([]string{`sender_id`, `recipient_id`, `amount`,
			`comment`, `block_id`, `txhash`},
			[]interface{}{0, id, amount, comment, block, sc.TxHash}, getDefTableName(sc, `history`))
	}
	return err
}

// PermColumn is contract func
func PermColumn(sc *SmartContract, tableName, name, permissions string) error {
	if err := validateAccess(`PermColumn`, sc, nEditColumn); err != nil {
		return err
	}
	name = strings.ToLower(name)
	tableName = strings.ToLower(tableName)
	tables := getDefTableName(sc, `tables`)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err := model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		return logErrorDB(err, "querying columns by table name")
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError,
			"unmarshalling columns permissions from json", temp.Columns)
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		return logError(err, consts.JSONMarshallError, "marshalling column permissions to json")
	}
	_, _, err = sc.update([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName})
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

// HMac returns HMAC hash as raw or hex string
func HMac(key, data string, raw_output bool) (ret string, err error) {
	hash, err := crypto.GetHMAC(key, data)
	if err != nil {
		return ``, logError(err, consts.CryptoError, "getting HMAC")
	}
	if raw_output {
		return string(hash), nil
	} else {
		return hex.EncodeToString(hash), nil
	}
}

// GetMapKeys returns the array of keys of the map
func GetMapKeys(in map[string]interface{}) []interface{} {
	keys := make([]interface{}, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	return keys
}

// SortedKeys returns the sorted array of keys of the map
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

// Date formats timestamp to specified date format
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
		return ``, logError(err, consts.NetworkError, "new http request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, v := range headers {
		req.Header.Set(key, fmt.Sprint(v))
	}
	resp, err := client.Do(req)
	if err != nil {
		return ``, logError(err, consts.NetworkError, "http request")
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ``, logError(err, consts.IOError, "reading http answer")
	}
	if resp.StatusCode != http.StatusOK {
		return ``, logError(fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data))),
			consts.NetworkError, "http status code")
	}
	return string(data), nil
}

// HTTPPostJSON sends post http request with json
func HTTPPostJSON(requrl string, headers map[string]interface{}, json_str string) (string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer([]byte(json_str)))
	if err != nil {
		return ``, logError(err, consts.NetworkError, "new http request")
	}

	for key, v := range headers {
		req.Header.Set(key, fmt.Sprint(v))
	}
	resp, err := client.Do(req)
	if err != nil {
		return ``, logError(err, consts.NetworkError, "http request")
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ``, logError(err, consts.IOError, "reading http answer")
	}
	if resp.StatusCode != http.StatusOK {
		return ``, logError(fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data))),
			consts.NetworkError, "http status code")
	}
	return string(data), nil
}

// Random returns a random value between min and max
func Random(min int64, max int64) (int64, error) {
	if min < 0 || max < 0 || min >= max {
		return 0, logError(fmt.Errorf(eWrongRandom, min, max), consts.InvalidObject, "getting random")
	}
	return min + rand.New(rand.NewSource(time.Now().Unix())).Int63n(max-min), nil
}

func ValidateCron(cronSpec string) (err error) {
	_, err = scheduler.Parse(cronSpec)
	return
}

func UpdateCron(sc *SmartContract, id int64) error {
	cronTask := &model.Cron{}
	cronTask.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID))

	ok, err := cronTask.Get(id)
	if err != nil {
		return logErrorDB(err, "get cron record")
	}

	if !ok {
		return nil
	}

	err = scheduler.UpdateTask(&scheduler.Task{
		ID:       cronTask.UID(),
		CronSpec: cronTask.Cron,
		Handler: &contract.ContractHandler{
			Contract: cronTask.Contract,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func UpdateNodesBan(smartContract *SmartContract, timestamp int64) error {
	now := time.Unix(timestamp, 0)

	badBlocks := &model.BadBlocks{}
	banRequests, err := badBlocks.GetNeedToBanNodes(now, syspar.GetIncorrectBlocksPerDay())
	if err != nil {
		logError(err, consts.DBError, "get nodes need to be banned")
		return err
	}

	fullNodes := syspar.GetNodes()
	var updFullNodes bool
	for i, fullNode := range fullNodes {
		// Removing ban in case ban time has already passed
		if fullNode.UnbanTime.Unix() > 0 && now.After(fullNode.UnbanTime) {
			fullNode.UnbanTime = time.Unix(0, 0)
			updFullNodes = true
		}

		// Setting ban time if we have ban requests for the current node from 51% of all nodes.
		// Ban request is mean that node have added more or equal N(system parameter) of bad blocks
		for _, banReq := range banRequests {
			if banReq.ProducerNodeId == fullNode.KeyID && banReq.Count >= int64((len(fullNodes)/2)+1) {
				fullNode.UnbanTime = now.Add(syspar.GetNodeBanTime())

				blocks, err := badBlocks.GetNodeBlocks(fullNode.KeyID, now)
				if err != nil {
					return logErrorDB(err, "getting node bad blocks for removing")
				}

				for _, b := range blocks {
					if _, err := DBUpdate(smartContract, "@1_bad_blocks", b.ID, "deleted", "1"); err != nil {
						return logErrorValue(err, consts.DBError, "deleting bad block",
							converter.Int64ToStr(b.ID))
					}
				}

				banMessage := fmt.Sprintf(
					"%d/%d nodes voted for ban with %d or more blocks each",
					banReq.Count,
					len(fullNodes),
					syspar.GetIncorrectBlocksPerDay(),
				)

				_, _, err = DBInsert(
					smartContract,
					"@1_node_ban_logs",
					"node_id,banned_at,ban_time,reason",
					fullNode.KeyID,
					now.Format(time.RFC3339),
					int64(syspar.GetNodeBanTime()/time.Millisecond), // in ms
					banMessage,
				)

				if err != nil {
					return logErrorValue(err, consts.DBError, "inserting log to node_ban_log",
						converter.Int64ToStr(banReq.ProducerNodeId))
				}

				_, _, err = DBInsert(
					smartContract,
					"notifications",
					"recipient->member_id,notification->type,notification->header,notification->body",
					fullNode.KeyID,
					model.NotificationTypeSingle,
					nodeBanNotificationHeader,
					banMessage,
				)

				if err != nil {
					logErrorValue(err, consts.DBError, "inserting log to node_ban_log",
						converter.Int64ToStr(banReq.ProducerNodeId))
					return logErrorValue(err, consts.DBError, "sending notification to node owner",
						converter.Int64ToStr(banReq.ProducerNodeId))

				}

				updFullNodes = true
			}
		}

		fullNodes[i] = fullNode
	}

	if updFullNodes {
		data, err := json.Marshal(fullNodes)
		if err != nil {
			return logError(err, consts.JSONMarshallError, "marshalling full nodes")
		}

		_, err = UpdateSysParam(smartContract, syspar.FullNodes, string(data), "")
		if err != nil {
			return logErrorDB(err, "updating full nodes")
		}
	}

	return nil
}

func GetBlock(blockID int64) (map[string]int64, error) {
	block := model.Block{}
	ok, err := block.Get(blockID)
	if err != nil {
		return nil, logErrorDB(err, "getting block")
	}
	if !ok {
		return nil, nil
	}

	return map[string]int64{
		"id":     block.ID,
		"time":   block.Time,
		"key_id": block.KeyID,
	}, nil
}

// UUID returns new uuid
func UUID(sc *SmartContract) string {
	return uuid.Must(uuid.NewV4()).String()
}

// DecodeBase64 decodes base64 string
func DecodeBase64(input string) (out string, err error) {
	var bin []byte
	bin, err = base64.StdEncoding.DecodeString(input)
	if err == nil {
		out = string(bin)
	}
	return
}

// EncodeBase64 encodes string in base64
func EncodeBase64(input string) (out string) {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// MD5 returns md5 hash sum of data
func MD5(data interface{}) (string, error) {
	var b []byte

	switch v := data.(type) {
	case []uint8:
		b = []byte(v)
	case string:
		b = []byte(v)
	default:
		return "", logError(fmt.Errorf(eUnsupportedType, v), consts.ConversionError,
			"converting to bytes")
	}

	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:]), nil
}

// GetColumnType returns the type of the column
func GetColumnType(sc *SmartContract, tableName, columnName string) (string, error) {
	return model.GetColumnType(getDefTableName(sc, tableName), columnName)
}

// GetType returns the name of the type of the value
func GetType(val interface{}) string {
	if val == nil {
		return `nil`
	}
	return reflect.TypeOf(val).String()
}

// StringToBytes converts string to bytes
func StringToBytes(src string) []byte {
	return []byte(src)
}

// BytesToString converts bytes to string
func BytesToString(src []byte) string {
	return string(src)
}
