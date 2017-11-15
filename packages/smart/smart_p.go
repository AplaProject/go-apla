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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/utils"

	/*	"bytes"
		"database/sql"
		"reflect"


		"github.com/AplaProject/go-apla/packages/templatev2"
		"github.com/AplaProject/go-apla/packages/utils"

		"github.com/jinzhu/gorm"*/
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	funcCallsDBP = map[string]struct{}{
		"DBInsert":       struct{}{},
		"DBUpdate":       struct{}{},
		"DBUpdateExt":    struct{}{},
		"DBSelect":       struct{}{},
		"DBInt":          struct{}{},
		"DBRowExt":       struct{}{},
		"DBRow":          struct{}{},
		"DBStringExt":    struct{}{},
		"DBIntExt":       struct{}{},
		"DBFreeRequest":  struct{}{},
		"DBStringWhere":  struct{}{},
		"DBIntWhere":     struct{}{},
		"DBAmount":       struct{}{},
		"DBInsertReport": struct{}{},
		"UpdateSysParam": struct{}{},
		"FindEcosystem":  struct{}{},
	}
	extendCostP = map[string]int64{
		"AddressToId":       10,
		"IdToAddress":       10,
		"NewState":          1000, // ?? What cost must be?
		"Sha256":            50,
		"PubToID":           10,
		"EcosysParam":       10,
		"SysParamString":    10,
		"SysParamInt":       10,
		"SysFuel":           10,
		"ValidateCondition": 30,
		"EvalCondition":     20,
		"HasPrefix":         10,
		"Contains":          10,
		"Replace":           10,
		"Join":              10,
		"UpdateLang":        10,
		"Size":              10,
		"Substr":            10,
		"ContractsList":     10,
		"IsContract":        10,
		"CompileContract":   100,
		"FlushContract":     50,
		"Eval":              10,
		"Activate":          10,
		"CreateEcosystem":   100,
		"RollbackEcosystem": 100,
		"TableConditions":   100,
		"CreateTable":       100,
		"RollbackTable":     100,
		"PermTable":         100,
		"ColumnCondition":   50,
		"CreateColumn":      50,
		"RollbackColumn":    50,
		"PermColumn":        50,
		"JSONToMap":         50,
	}
)

//SignRes contains the data of the signature
type SignRes struct {
	Param string `json:"name"`
	Text  string `json:"text"`
}

// TxSignJSON is a structure for additional signs of transaction
type TxSignJSON struct {
	ForSign string    `json:"forsign"`
	Field   string    `json:"field"`
	Title   string    `json:"title"`
	Params  []SignRes `json:"params"`
}

func init() {
	Extend(&script.ExtendData{Objects: map[string]interface{}{
		"DBInsert":           DBInsert,
		"DBUpdate":           DBUpdate,
		"DBUpdateExt":        DBUpdateExt,
		"DBSelect":           DBSelect,
		"DBInt":              DBInt,
		"DBRowExt":           DBRowExt,
		"DBRow":              DBRow,
		"DBStringExt":        DBStringExt,
		"DBFreeRequest":      DBFreeRequest,
		"DBIntExt":           DBIntExt,
		"DBStringWhere":      DBStringWhere,
		"DBIntWhere":         DBIntWhere,
		"AddressToId":        AddressToID,
		"IdToAddress":        IDToAddress,
		"DBAmount":           DBAmount,
		"ContractAccess":     ContractAccess,
		"ContractConditions": ContractConditions,
		"EcosysParam":        EcosysParam,
		"SysParamString":     SysParamString,
		"SysParamInt":        SysParamInt,
		"SysFuel":            SysFuel,
		"Int":                Int,
		"Str":                Str,
		"Money":              Money,
		"Float":              Float,
		"Len":                Len,
		"Join":               Join,
		"Sha256":             Sha256,
		"PubToID":            PubToID,
		"HexToBytes":         HexToBytes,
		"LangRes":            LangRes,
		"DBInsertReport":     DBInsertReport,
		"UpdateSysParam":     UpdateSysParam,
		"ValidateCondition":  ValidateCondition,
		"EvalCondition":      EvalCondition,
		"HasPrefix":          strings.HasPrefix,
		"Contains":           strings.Contains,
		"Replace":            Replace,
		"FindEcosystem":      FindEcosystem,
		"CreateEcosystem":    CreateEcosystem,
		"RollbackEcosystem":  RollbackEcosystem,
		"CreateTable":        CreateTable,
		"RollbackTable":      RollbackTable,
		"PermTable":          PermTable,
		"TableConditions":    TableConditions,
		"ColumnCondition":    ColumnCondition,
		"CreateColumn":       CreateColumn,
		"RollbackColumn":     RollbackColumn,
		"PermColumn":         PermColumn,
		"UpdateLang":         UpdateLang,
		"Size":               Size,
		"Substr":             Substr,
		"ContractsList":      ContractsList,
		"IsContract":         IsContract,
		"CompileContract":    CompileContract,
		"FlushContract":      FlushContract,
		"Eval":               Eval,
		"Activate":           Activate,
		"JSONToMap":          JSONToMap,
		"check_signature":    CheckSignature, // system function
	}, AutoPars: map[string]string{
		`*parser.SmartContract`: `parser`,
	}})
	ExtendCost(getCostP)
	FuncCallsDB(funcCallsDBP)
}

func getCostP(name string) int64 {
	if val, ok := extendCostP[name]; ok {
		return val
	}
	return -1
}

// UpdateSysParam updates the system parameter
func UpdateSysParam(sc *SmartContract, name, value, conditions string) (int64, error) {
	var (
		fields []string
		values []interface{}
	)

	par := &model.SystemParameter{}
	_, err := par.Get(name)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("system parameter get")
		return 0, err
	}
	cond := par.Conditions
	if len(cond) > 0 {
		ret, err := sc.EvalIf(cond)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("evaluating conditions")
			return 0, err
		}
		if !ret {
			log.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
			return 0, fmt.Errorf(`Access denied`)
		}
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := CompileEval(conditions, 0); err != nil {
			log.WithFields(log.Fields{"error": err, "conditions": conditions, "state_id": 0, "type": consts.EvalError}).Error("compiling eval")
			return 0, err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty value and condition")
		return 0, fmt.Errorf(`empty value and condition`)
	}
	_, _, err = sc.selectiveLoggingAndUpd(fields, values, "system_parameters", []string{"name"}, []string{name}, !sc.VDE)
	if err != nil {
		return 0, err
	}
	err = syspar.SysUpdate()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
		return 0, err
	}
	return 0, nil
}

// DBUpdateExt updates the record in the specified table. You can specify 'where' query in params and then the values for this query
func DBUpdateExt(sc *SmartContract, tblname string, column string, value interface{}, params string, val ...interface{}) (qcost int64, err error) { // map[string]interface{}) {
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
	qcost, _, err = sc.selectiveLoggingAndUpd(columns, val, tblname, []string{column}, []string{fmt.Sprint(value)}, !sc.VDE)
	return
}

// DBInt returns the numeric value of the column for the record with the specified id
func DBInt(sc *SmartContract, tblname string, name string, id int64) (int64, int64, error) {
	tblname = getDefTableName(sc, tblname)

	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, 0, err
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id).Int64()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting db int")
	}
	return cost, res, err
}

// DBRowExt returns one row from the table StringExt
func DBRowExt(sc *SmartContract, tblname string, columns string, id interface{}, idname string) (int64, map[string]string, error) {

	tblname = getDefTableName(sc, tblname)

	isBytea := GetBytea(tblname)
	if isBytea[idname] {
		switch id.(type) {
		case string:
			if vbyte, err := hex.DecodeString(id.(string)); err == nil {
				id = vbyte
			}
		}
	}
	query := `select ` + converter.Sanitize(columns, ` ,()*`) + ` from ` + converter.EscapeName(tblname) + ` where ` + converter.EscapeName(idname) + `=?`
	cost, err := model.GetQueryTotalCost(query, id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, nil, err
	}
	res, err := model.GetOneRow(query, id).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
	}

	return cost, res, err
}

// DBRow returns one row from the table StringExt
func DBRow(sc *SmartContract, tblname string, columns string, id int64) (int64, map[string]string, error) {
	tblname = getDefTableName(sc, tblname)

	query := `select ` + converter.Sanitize(columns, ` ,()*`) + ` from ` + converter.EscapeName(tblname) + ` where id=?`
	cost, err := model.GetQueryTotalCost(query, id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, nil, err
	}
	res, err := model.GetOneRow(query, id).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
	}

	return cost, res, err
}

// DBStringExt returns the value of 'name' column for the record with the specified value of the 'idname' field
func DBStringExt(sc *SmartContract, tblname string, name string, id interface{}, idname string) (int64, string, error) {
	tblname = getDefTableName(sc, tblname)

	isBytea := GetBytea(tblname)
	if isBytea[idname] {
		switch id.(type) {
		case string:
			if vbyte, err := hex.DecodeString(id.(string)); err == nil {
				id = vbyte
			}
		}
	}

	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where `+converter.EscapeName(idname)+`=?`, id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, "", err
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where `+converter.EscapeName(idname)+`=?`, id).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting dbstring ext")
	}
	return cost, res, err
}

// DBIntExt returns the numeric value of the 'name' column for the record with the specified value of the 'idname' field
func DBIntExt(sc *SmartContract, tblname string, name string, id interface{}, idname string) (cost int64, ret int64, err error) {
	var val string
	var qcost int64

	tblname = getDefTableName(sc, tblname)
	qcost, val, err = DBStringExt(sc, tblname, name, id, idname)
	if err != nil {
		return 0, 0, err
	}
	if len(val) == 0 {
		return 0, 0, nil
	}
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("converting DBStringExt result from string to int")
	}
	return qcost, res, err
}

// DBFreeRequest is a free function that is needed to find the record with the specified value in the 'idname' column.
func DBFreeRequest(sc *SmartContract, tblname string, id interface{}, idname string) (int64, error) {
	if sc.TxContract.FreeRequest {
		log.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("DBFreeRequest can be executed only once")
		return 0, fmt.Errorf(`DBFreeRequest can be executed only once`)
	}
	sc.TxContract.FreeRequest = true
	cost, ret, err := DBStringExt(sc, tblname, idname, id, idname)
	if err != nil {
		return 0, err
	}
	if len(ret) > 0 || ret == fmt.Sprintf(`%v`, id) {
		return 0, nil
	}
	return cost, fmt.Errorf(`DBFreeRequest: cannot find %v in %s of %s`, id, idname, tblname)
}

// DBStringWhere returns the column value based on the 'where' condition and 'params' values for this condition
func DBStringWhere(sc *SmartContract, tblname string, name string, where string, params ...interface{}) (int64, string, error) {

	tblname = getDefTableName(sc, tblname)

	selectQuery := `select ` + converter.EscapeName(name) + ` from ` + converter.EscapeName(tblname) + ` where ` + strings.Replace(converter.Escape(where), `$`, `?`, -1)
	qcost, err := model.GetQueryTotalCost(selectQuery, params...)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, "", err
	}
	res, err := model.Single(selectQuery, params).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing single query")
		return 0, "", err
	}
	return qcost, res, err
}

// DBIntWhere returns the column value based on the 'where' condition and 'params' values for this condition
func DBIntWhere(sc *SmartContract, tblname string, name string, where string, params ...interface{}) (cost int64, ret int64, err error) {
	var val string
	cost, val, err = DBStringWhere(sc, tblname, name, where, params...)
	if err != nil {
		return 0, 0, err
	}
	if len(val) == 0 {
		return 0, 0, nil
	}
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("convertion DBStringWhere result from string to int")
	}
	return cost, res, err
}

// DBAmount returns the value of the 'amount' column for the record with the 'id' value in the 'column' column
func DBAmount(sc *SmartContract, tblname, column string, id int64) (int64, decimal.Decimal) {

	tblname = getDefTableName(sc, tblname)

	balance, err := model.Single("SELECT amount FROM "+converter.EscapeName(tblname)+" WHERE "+converter.EscapeName(column)+" = ?", id).String()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("executing single query")
		return 0, decimal.New(0, 0)
	}
	val, err := decimal.NewFromString(balance)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConvertionError}).Error("converting balance from string to decimal")
	}
	return 0, val
}

// SysParamString returns the value of the system parameter
func SysParamString(name string) string {
	return syspar.SysString(name)
}

// SysParamInt returns the value of the system parameter
func SysParamInt(name string) int64 {
	return syspar.SysInt64(name)
}

// SysFuel returns the fuel rate
func SysFuel(state int64) string {
	return syspar.GetFuelRate(state)
}

// Int converts a string to a number
func Int(val string) int64 {
	return converter.StrToInt64(val)
}

// Str converts the value to a string
func Str(v interface{}) (ret string) {
	switch val := v.(type) {
	case float64:
		ret = fmt.Sprintf(`%f`, val)
	default:
		ret = fmt.Sprintf(`%v`, val)
	}
	return
}

// Money converts the value into a numeric type for money
func Money(v interface{}) (ret decimal.Decimal) {
	return script.ValueToDecimal(v)
}

// Float converts the value to float64
func Float(v interface{}) (ret float64) {
	return script.ValueToFloat(v)
}

func Join(input []interface{}, sep string) string {
	var ret string
	for i, item := range input {
		if i > 0 {
			ret += sep
		}
		ret += fmt.Sprintf(`%v`, item)
	}
	return ret
}

// Sha256 returns SHA256 hash value
func Sha256(text string) string {
	hash, err := crypto.Hash([]byte(text))
	if err != nil {
		log.WithFields(log.Fields{"value": text, "error": err, "type": consts.CryptoError}).Fatal("hashing text")
	}
	hash = converter.BinToHex(hash)
	return string(hash)
}

// PubToID returns a numeric identifier for the public key specified in the hexadecimal form.
func PubToID(hexkey string) int64 {
	pubkey, err := hex.DecodeString(hexkey)
	if err != nil {
		log.WithFields(log.Fields{"value": hexkey, "error": err, "type": consts.CryptoError}).Error("decoding hexkey to string")
		return 0
	}
	return crypto.Address(pubkey)
}

// HexToBytes converts the hexadecimal representation to []byte
func HexToBytes(hexdata string) ([]byte, error) {
	return hex.DecodeString(hexdata)
}

// LangRes returns the language resource
func LangRes(sc *SmartContract, idRes, lang string) string {
	ret, _ := language.LangText(idRes, int(sc.TxSmart.EcosystemID), lang)
	return ret
}

// DBInsertReport inserts a record into the specified report table
func DBInsertReport(sc *SmartContract, tblname string, params string, val ...interface{}) (qcost int64, ret int64, err error) {
	names := strings.Split(getDefTableName(sc, tblname), `_`)
	state := converter.StrToInt64(names[0])
	if state != int64(sc.TxSmart.EcosystemID) {
		err = fmt.Errorf(`Wrong state in DBInsertReport`)
		return
	}
	if !model.IsNodeState(state, ``) {
		return
	}
	tblname = names[0] + `_reports_` + strings.Join(names[1:], `_`)
	if err = sc.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var lastID string
	qcost, lastID, err = sc.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, !sc.VDE)
	if err == nil {
		ret, _ = strconv.ParseInt(lastID, 10, 64)
	}
	return
}

// EvalCondition gets the condition and check it
func EvalCondition(sc *SmartContract, table, name, condfield string) error {
	conditions, err := model.Single(`SELECT `+converter.EscapeName(condfield)+` FROM "`+getDefTableName(sc, table)+
		`" WHERE name = ?`, name).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing single query")
		return err
	}
	if len(conditions) == 0 {
		log.WithFields(log.Fields{"type": consts.NotFound, "name": name}).Error("Record not found")
		return fmt.Errorf(`Record %s has not been found`, name)
	}
	return Eval(sc, conditions)
}

// Replace replaces old substrings to new substrings
func Replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// FindEcosystem checks if there is an ecosystem with the specified name
func FindEcosystem(sc *SmartContract, country string) (int64, int64, error) {
	query := `SELECT id FROM system_states where name=?`
	cost, err := model.GetQueryTotalCost(query, country)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, 0, err
	}
	id, err := model.Single(query, country).Int64()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing single query")
		return 0, 0, err
	}
	return cost, id, nil
}

// CreateEcosystem creates a new ecosystem
func CreateEcosystem(sc *SmartContract, wallet int64, name string) (int64, error) {
	if sc.TxContract.Name != `@1NewEcosystem` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("CreateEcosystem can be only called from @1NewEcosystem")
		return 0, fmt.Errorf(`CreateEcosystem can be only called from @1NewEcosystem`)
	}
	_, id, err := sc.selectiveLoggingAndUpd([]string{`name`}, []interface{}{
		name,
	}, `system_states`, nil, nil, !sc.VDE)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError}).Error("CreateEcosystem")
		return 0, err
	}
	err = model.ExecSchemaEcosystem(converter.StrToInt(id), wallet, name)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return 0, err
	}
	err = LoadContract(sc.DbTransaction, id)
	if err != nil {
		return 0, err
	}
	return converter.StrToInt64(id), err
}

func RollbackEcosystem(sc *SmartContract) error {
	if sc.TxContract.Name != `@1NewEcosystem` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackEcosystem can be only called from @1NewEcosystem")
		return fmt.Errorf(`RollbackEcosystem can be only called from @1NewEcosystem`)
	}
	rollbackTx := &model.RollbackTx{}
	err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, "system_states")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback tx")
		return err
	}
	lastID, err := model.GetNextID(sc.DbTransaction, `system_states`)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id")
		return err
	}
	lastID--
	if converter.StrToInt64(rollbackTx.TableID) != lastID {
		log.WithFields(log.Fields{"table_id": rollbackTx.TableID, "last_id": lastID, "type": consts.InvalidObject}).Error("incorrect ecosystem id")
		return fmt.Errorf(`Incorrect ecosystem id %s != %d`, rollbackTx.TableID, lastID)
	}
	if model.IsTable(fmt.Sprintf(`%s_vde_tables`, rollbackTx.TableID)) {
		// Drop all _local_ tables
		table := &model.Table{}
		prefix := fmt.Sprintf(`%s_vde`, rollbackTx.TableID)
		table.SetTablePrefix(prefix)
		list, err := table.GetAll(prefix)
		if err != nil {
			return err
		}
		for _, item := range list {
			err = model.DropTable(sc.DbTransaction, fmt.Sprintf("%s_%s", prefix, item.Name))
			if err != nil {
				return err
			}
		}
		for _, name := range []string{`tables`, `parameters`} {
			err = model.DropTable(sc.DbTransaction, fmt.Sprintf("%s_%s", prefix, name))
			if err != nil {
				return err
			}
		}
	}

	for _, name := range []string{`menu`, `pages`, `languages`, `signatures`, `tables`,
		`contracts`, `parameters`, `blocks`, `history`, `keys`} {
		err = model.DropTable(sc.DbTransaction, fmt.Sprintf("%s_%s", rollbackTx.TableID, name))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
			return err
		}
	}
	rollbackTxToDel := &model.RollbackTx{TxHash: sc.TxHash, NameTable: "system_states"}
	err = rollbackTxToDel.DeleteByHashAndTableName(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback tx by hash and table name")
		return err
	}
	ssToDel := &model.SystemState{ID: lastID}
	return ssToDel.Delete(sc.DbTransaction)
}

func RollbackTable(sc *SmartContract, name string) error {
	if sc.TxContract.Name != `@1NewTable` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackTable can be only called from @1NewTable")
		return fmt.Errorf(`RollbackTable can be only called from @1NewTable`)
	}
	err := model.DropTable(sc.DbTransaction, fmt.Sprintf("%d_%s", sc.TxSmart.EcosystemID, name))
	t := &model.Table{Name: name}
	err = t.Delete()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting table")
		return err
	}
	return nil
}

func RollbackColumn(sc *SmartContract, tableName, name string) error {
	if sc.TxContract.Name != `@1NewColumn` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackColumn can be only called from @1NewColumn")
		return fmt.Errorf(`RollbackColumn can be only called from @1NewColumn`)
	}
	return model.AlterTableDropColumn(fmt.Sprintf(`%d_%s`, sc.TxSmart.EcosystemID, tableName), name)
}

// UpdateLang updates language resource
func UpdateLang(sc *SmartContract, name, trans string) {
	language.UpdateLang(int(sc.TxSmart.EcosystemID), name, trans)
}

// Size returns the length of the string
func Size(s string) int64 {
	return int64(len(s))
}

// Substr returns the substring of the string
func Substr(s string, off int64, slen int64) string {
	ilen := int64(len(s))
	if off < 0 || slen < 0 || off > ilen {
		return ``
	}
	if off+slen > ilen {
		return s[off:]
	}
	return s[off : off+slen]
}

// ActivateContract sets Active status of the contract in smartVM
func Activate(sc *SmartContract, tblid int64, state int64) error {
	if sc.TxContract.Name != `@1ActivateContract` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("ActivateContract can be only called from @1ActivateContract")
		return fmt.Errorf(`ActivateContract can be only called from @1ActivateContract`)
	}
	ActivateContract(tblid, state, true)
	return nil
}

/*
// GetContractLimit returns the default maximal cost of contract
func (p *SmartContract) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if len(p.TxSmart.MaxSum) > 0 {
		p.TxCost = converter.StrToInt64(p.TxSmart.MaxSum)
	} else {
		cost, _ := templatev2.StateParam(p.TxSmart.EcosystemID, `max_sum`)
		if len(cost) > 0 {
			p.TxCost = converter.StrToInt64(cost)
		}
	}
	if p.TxCost == 0 {
		p.TxCost = script.CostDefault // fuel
	}
	return p.TxCost
}

func (p *SmartContract) getExtend() *map[string]interface{} {
	head := p.TxSmart //consts.HeaderNew(contract.parser.TxPtr)
	var keyID int64
	keyID = int64(head.KeyID)
	// test
	block := int64(0)
	blockTime := int64(0)
	blockKeyID := int64(0)
	if p.BlockData != nil {
		block = p.BlockData.BlockID
		blockKeyID = p.BlockData.KeyID
		blockTime = p.BlockData.Time
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Time, `node_position`: head.NodePosition, `ecosystem_id`: head.EcosystemID,
		`block`: block, `key_id`: keyID, `block_key_id`: blockKeyID,
		`parent`: ``, `txcost`: p.GetContractLimit(), `txhash`: p.TxHash, `result`: ``,
		`parser`: p, `contract`: p.TxContract, `block_time`: blockTime}
	//, `vars`: make(map[string]interface{})
	for key, val := range p.TxData {
		extend[key] = val
	}

	return &extend
}

// StackCont adds an element to the stack of contract call or removes the top element when name is empty
func StackCont(p interface{}, name string) {
	cont := p.(*SmartContract).TxContract
	if len(name) > 0 {
		cont.StackCont = append(cont.StackCont, name)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	return
}

// CallContract calls the contract functions according to the specified flags
func (p *SmartContract) CallContract(flags int) (err error) {
	logger := p.GetLogger()
	var (
		public                 []byte
		sizeFuel, toID, fromID int64
		fuelRate               decimal.Decimal
	)
	payWallet := &model.Key{}
	p.TxContract.Extend = p.getExtend()
	var price int64

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	p.TxContract.StackCont = []string{p.TxContract.Name}
	(*p.TxContract.Extend)[`stack_cont`] = StackCont

	if flags&smart.CallRollback == 0 && (flags&smart.CallAction) != 0 {
		toID = p.BlockData.KeyID
		fromID = p.TxSmart.KeyID
		if len(p.TxSmart.PublicKey) > 0 && string(p.TxSmart.PublicKey) != `null` {
			public = p.TxSmart.PublicKey
		}
		wallet := &model.Key{}
		wallet.SetTablePrefix(p.TxSmart.EcosystemID)
		err := wallet.Get(p.TxSmart.KeyID)
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
			return err
		}
		if len(wallet.PublicKey) > 0 {
			public = wallet.PublicKey
		}
		if p.TxSmart.Type == 258 { // UpdFullNodes
			node := syspar.GetNode(p.TxSmart.KeyID)
			if node == nil {
				logger.WithFields(log.Fields{"user_id": p.TxSmart.KeyID, "type": consts.NotFound}).Error("unknown node id")
				return fmt.Errorf("unknown node id")
			}
			public = node.Public
		}
		if len(public) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty public key")
			return fmt.Errorf("empty public key")
		}
		p.PublicKeys = append(p.PublicKeys, public)
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.TxData[`forsign`].(string), p.TxSmart.BinSignatures, false)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("checking tx data sign")
			return err
		}
		if !CheckSignResult {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect sign")
			return fmt.Errorf("incorrect sign")
		}
		if p.TxSmart.EcosystemID > 0 {
			if p.TxSmart.TokenEcosystem == 0 {
				p.TxSmart.TokenEcosystem = 1
			}
			fuelRate, err = decimal.NewFromString(syspar.GetFuelRate(p.TxSmart.TokenEcosystem))
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.TxSmart.TokenEcosystem}).Error("converting ecosystem fuel rate from string to decimal")
				return err
			}
			if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
				logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("Fuel rate must be greater than 0")
				return fmt.Errorf(`Fuel rate must be greater than 0`)
			}
			if len(p.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(p.TxSmart.PayOver)
				if err != nil {
					log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.TxSmart.TokenEcosystem}).Error("converting tx smart pay over from string to decimal")
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			if p.TxContract.Block.Info.(*script.ContractInfo).Owner.Active {
				fromID = p.TxContract.Block.Info.(*script.ContractInfo).Owner.WalletID
				p.TxSmart.TokenEcosystem = p.TxContract.Block.Info.(*script.ContractInfo).Owner.TokenID
			} else if len(p.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(p.TxSmart.PayOver)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.TxSmart.TokenEcosystem}).Error("converting tx smart pay over from string to decimal")
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			payWallet.SetTablePrefix(p.TxSmart.TokenEcosystem)
			if err = payWallet.Get(fromID); err != nil {
				if err == gorm.ErrRecordNotFound {
					return fmt.Errorf(`current balance is not enough`)
				} else {
					logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
					return err
				}
			}
			if !bytes.Equal(wallet.PublicKey, payWallet.PublicKey) && !bytes.Equal(p.TxSmart.PublicKey, payWallet.PublicKey) {
				return fmt.Errorf(`Token and user public keys are different`)
			}
			amount, err := decimal.NewFromString(payWallet.Amount)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
				return err
			}
			if cprice := p.TxContract.GetFunc(`price`); cprice != nil {
				var ret []interface{}
				if ret, err = smart.Run(cprice, nil, p.TxContract.Extend); err != nil {
					return err
				} else if len(ret) == 1 {
					if _, ok := ret[0].(int64); !ok {
						logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong result type of price function")
						return fmt.Errorf(`Wrong result type of price function`)
					}
					price = ret[0].(int64)
				} else {
					logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong type of price function")
					return fmt.Errorf(`Wrong type of price function`)
				}
			}
			sizeFuel = syspar.GetSizeFuel() * int64(len(p.TxSmart.Data)) / 1024
			if amount.Cmp(decimal.New(sizeFuel+price, 0).Mul(fuelRate)) <= 0 {
				logger.WithFields(log.Fields{"tyoe": consts.NoFunds}).Error("current balance is not enough")
				return fmt.Errorf(`current balance is not enough`)
			}
		}
	}
	before := (*p.TxContract.Extend)[`txcost`].(int64) + price

	// Payment for the size
	(*p.TxContract.Extend)[`txcost`] = (*p.TxContract.Extend)[`txcost`].(int64) - sizeFuel

	p.TxContract.FreeRequest = false
	for i := uint32(0); i < 4; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := p.TxContract.GetFunc(methods[i])
			if cfunc == nil {
				continue
			}
			p.TxContract.Called = 1 << i
			_, err = smart.Run(cfunc, nil, p.TxContract.Extend)
			if err != nil {
				before -= price
				break
			}
		}
	}
	p.TxUsedCost = decimal.New(before-(*p.TxContract.Extend)[`txcost`].(int64), 0)
	p.TxContract.TxPrice = price
	if (flags&smart.CallAction) != 0 && p.TxSmart.EcosystemID > 0 {
		apl := p.TxUsedCost.Mul(fuelRate)
		wltAmount, err := decimal.NewFromString(payWallet.Amount)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
			return err
		}
		if wltAmount.Cmp(apl) < 0 {
			apl = wltAmount
		}
		commission := apl.Mul(decimal.New(syspar.SysInt64(`commission_size`), 0)).Div(decimal.New(100, 0)).Floor()
		walletTable := fmt.Sprintf(`%d_keys`, p.TxSmart.TokenEcosystem)
		if _, _, err := p.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{apl}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(fromID)}, true); err != nil {
			return err
		}
		// TODO: add checking for key_id "toID". If key not exists it led to fork
		if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{apl.Sub(commission)}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(toID)}, true); err != nil {
			return err
		}
		if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{commission}, walletTable, []string{`id`},
			[]string{syspar.GetCommissionWallet(p.TxSmart.TokenEcosystem)}, true); err != nil {
			return err
		}
		logger.WithFields(log.Fields{"commission": commission}).Debug("Paid commission")
	}
	return
}

func checkReport(tblname string) error {
	if strings.Contains(tblname, `_reports_`) {
		log.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied to report table")
		return fmt.Errorf(`Access denied to report table`)
	}
	return nil
}

// DBString returns the value of the field of the record with the specified id
func DBString(tblname string, name string, id int64) (int64, string, error) {
	if err := checkReport(tblname); err != nil {
		return 0, ``, err
	}
	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting query total cost")
		return 0, "", nil
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting dbstring")
	}
	return cost, res, err
}

func TableName(p *SmartContract, tblname string) string {
	tblname = strings.Trim(converter.EscapeName(tblname), `"`)
	if tblname[0] >= '1' && tblname[0] <= '9' && strings.Contains(tblname, `_`) {
		return tblname
	}
	return fmt.Sprintf(`%d_%s`, p.TxEcosystemID, tblname)
}


// IsGovAccount checks whether the specified account is the owner of the state
func IsGovAccount(p *SmartContract, citizen int64) bool {
	return converter.StrToInt64(StateVal(p, `founder_account`)) == citizen
}
*/
// CheckSignature checks the additional signatures for the contract
func CheckSignature(i *map[string]interface{}, name string) error {
	state, name := script.ParseContract(name)
	pref := converter.Int64ToStr(int64(state))
	if state == 0 {
		pref = `global`
	}
	//	fmt.Println(`CheckSignature`, i, state, name)
	p := (*i)[`parser`].(*SmartContract)
	value, err := model.Single(`select value from "`+pref+`_signatures" where name=?`, name).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing single query")
		return err
	}
	if len(value) == 0 {
		return nil
	}
	hexsign, err := hex.DecodeString((*i)[`Signature`].(string))
	if len(hexsign) == 0 || err != nil {
		log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("comverting signature to hex")
		return fmt.Errorf(`wrong signature`)
	}

	var sign TxSignJSON
	err = json.Unmarshal([]byte(value), &sign)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling sign")
		return err
	}
	wallet := (*i)[`key_id`].(int64)
	if wallet == 0 {
		wallet = (*i)[`citizen`].(int64)
	}
	forsign := fmt.Sprintf(`%d,%d`, uint64((*i)[`time`].(int64)), uint64(wallet))
	for _, isign := range sign.Params {
		forsign += fmt.Sprintf(`,%v`, (*i)[isign.Param])
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forsign, hexsign, true)
	if err != nil {
		return err
	}
	if !CheckSignResult {
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect signature")
		return fmt.Errorf(`incorrect signature ` + forsign)
	}
	return nil
}

/*
func checkWhere(tblname string, where string, order string) (string, string, error) {
	if len(order) > 0 {
		order = ` order by ` + converter.EscapeName(order)
	}
	return strings.Replace(converter.Escape(where), `$`, `?`, -1), order, nil
}

*/

func JSONToMap(input string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(input), &ret)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling json to map")
		return nil, err
	}
	return ret, nil
}
