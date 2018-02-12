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
	"reflect"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/language"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	funcCallsDBP = map[string]struct{}{
		"DBInsert":         {},
		"DBUpdate":         {},
		"DBUpdateSysParam": {},
		"DBUpdateExt":      {},
		"DBSelect":         {},
	}

	extendCostSysParams = map[string]string{
		"AddressToId":       "extend_cost_address_to_id",
		"IdToAddress":       "extend_cost_id_to_address",
		"NewState":          "extend_cost_new_state",
		"Sha256":            "extend_cost_sha256",
		"PubToID":           "extend_cost_pub_to_id",
		"EcosysParam":       "extend_cost_ecosys_param",
		"SysParamString":    "extend_cost_sys_param_string",
		"SysParamInt":       "extend_cost_sys_param_int",
		"SysFuel":           "extend_cost_sys_fuel",
		"ValidateCondition": "extend_cost_validate_condition",
		"EvalCondition":     "extend_cost_eval_condition",
		"HasPrefix":         "extend_cost_has_prefix",
		"Contains":          "extend_cost_contains",
		"Replace":           "extend_cost_replace",
		"Join":              "extend_cost_join",
		"UpdateLang":        "extend_cost_update_lang",
		"Size":              "extend_cost_size",
		"Substr":            "extend_cost_substr",
		"ContractsList":     "extend_cost_contracts_list",
		"IsObject":          "extend_cost_is_object",
		"CompileContract":   "extend_cost_compile_contract",
		"FlushContract":     "extend_cost_flush_contract",
		"Eval":              "extend_cost_eval",
		"Len":               "extend_cost_len",
		"Activate":          "extend_cost_activate",
		"Deactivate":        "extend_cost_deactivate",
		"CreateEcosystem":   "extend_cost_create_ecosystem",
		"TableConditions":   "extend_cost_table_conditions",
		"CreateTable":       "extend_cost_create_table",
		"PermTable":         "extend_cost_perm_table",
		"ColumnCondition":   "extend_cost_column_condition",
		"CreateColumn":      "extend_cost_create_column",
		"PermColumn":        "extend_cost_perm_column",
		"JSONToMap":         "extend_cost_json_to_map",
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
	EmbedFuncs(smartVM, script.VMTypeSmart)
}

func getCostP(name string) int64 {
	if key, ok := extendCostSysParams[name]; ok && syspar.HasSys(key) {
		return syspar.SysInt64(key)
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
	found, err := par.Get(name)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("system parameter get")
		return 0, err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("system parameter get")
		return 0, fmt.Errorf(`Parameter %s has not been found`, name)
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
			return 0, errAccessDenied
		}
	}
	if len(value) > 0 {
		var (
			ok, checked bool
			list        [][]string
		)
		ival := converter.StrToInt64(value)
	check:
		switch name {
		case `gap_between_blocks`:
			ok = ival > 0 && ival < 86400
		case `rb_blocks_1`, `number_of_nodes`:
			ok = ival > 0 && ival < 1000
		case `ecosystem_price`, `contract_price`, `column_price`, `table_price`, `menu_price`,
			`page_price`, `commission_size`:
			ok = ival >= 0
		case `max_block_size`, `max_tx_size`, `max_tx_count`, `max_columns`, `max_indexes`,
			`max_block_user_tx`, `max_fuel_tx`, `max_fuel_block`:
			ok = ival > 0
		case `fuel_rate`, `full_nodes`, `commission_wallet`:
			err := json.Unmarshal([]byte(value), &list)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling system param")
				return 0, err
			}
			for _, item := range list {
				switch name {
				case `fuel_rate`, `commission_wallet`:
					if len(item) != 2 || converter.StrToInt64(item[0]) <= 0 ||
						(name == `fuel_rate` && converter.StrToInt64(item[1]) <= 0) ||
						(name == `commission_wallet` && converter.StrToInt64(item[1]) == 0) {
						break check
					}
				case `full_nodes`:
					if len(item) != 3 {
						break check
					}
					key := converter.StrToInt64(item[1])
					if key == 0 || len(item[2]) != 128 || !converter.ValidateIPv4(item[0]) {
						break check
					}
				}
			}
			checked = true
		default:
			if strings.HasPrefix(name, `extend_cost_`) {
				ok = ival >= 0
				break
			}
			checked = true
		}
		if !checked && (!ok || converter.Int64ToStr(ival) != value) {
			log.WithFields(log.Fields{"type": consts.InvalidObject, "value": value, "name": name}).Error(ErrInvalidValue.Error())
			return 0, ErrInvalidValue
		}
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
	_, _, err = sc.selectiveLoggingAndUpd(fields, values, "system_parameters", []string{"id"}, []string{converter.Int64ToStr(par.ID)}, !sc.VDE && sc.Rollback, false)
	if err != nil {
		return 0, err
	}
	err = syspar.SysUpdate(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
		return 0, err
	}
	sc.SysUpdate = true
	return 0, nil
}

// DBUpdateExt updates the record in the specified table. You can specify 'where' query in params and then the values for this query
func DBUpdateExt(sc *SmartContract, tblname string, column string, value interface{},
	params string, val ...interface{}) (qcost int64, err error) {
	tblname = getDefTableName(sc, tblname)
	if err = sc.AccessTable(tblname, "update"); err != nil {
		return
	}
	if strings.Contains(tblname, `_reports_`) {
		err = fmt.Errorf(`Access denied to report table`)
		return
	}
	columns := strings.Split(params, `,`)
	if err = sc.AccessColumns(tblname, &columns, true); err != nil {
		return
	}
	qcost, _, err = sc.selectiveLoggingAndUpd(columns, val, tblname, []string{column}, []string{fmt.Sprint(value)}, !sc.VDE && sc.Rollback, false)
	return
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

// Join is joining input with separator
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

// Split splits the input string to array
func Split(input, sep string) []interface{} {
	out := strings.Split(input, sep)
	result := make([]interface{}, len(out))
	for i, val := range out {
		result[i] = reflect.ValueOf(val).Interface()
	}
	return result
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
	ret, _ := language.LangText(idRes, int(sc.TxSmart.EcosystemID), lang, sc.VDE)
	return ret
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

// CreateEcosystem creates a new ecosystem
func CreateEcosystem(sc *SmartContract, wallet int64, name string) (int64, error) {
	if sc.TxContract.Name != `@1NewEcosystem` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("CreateEcosystem can be only called from @1NewEcosystem")
		return 0, fmt.Errorf(`CreateEcosystem can be only called from @1NewEcosystem`)
	}
	_, id, err := sc.selectiveLoggingAndUpd(nil, nil, `system_states`, nil, nil, !sc.VDE && sc.Rollback, false)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError}).Error("CreateEcosystem")
		return 0, err
	}
	var sp model.StateParameter
	sp.SetTablePrefix(`1`)
	found, err := sp.Get(sc.DbTransaction, `founder_account`)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting founder")
		return 0, err
	}
	if !found || len(sp.Value) == 0 {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": ErrFounderAccount}).Error("founder not found")
		return 0, ErrFounderAccount
	}
	err = model.ExecSchemaEcosystem(sc.DbTransaction, converter.StrToInt(id), wallet, name,
		converter.StrToInt64(sp.Value))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return 0, err
	}
	err = LoadContract(sc.DbTransaction, id)
	if err != nil {
		return 0, err
	}
	sc.Rollback = false
	_, _, err = DBInsert(sc, id+"_pages", "name,value,menu,conditions", "default_page",
		SysParamString("default_ecosystem_page"), "default_menu", `ContractConditions("MainCondition")`)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return 0, err
	}
	_, _, err = DBInsert(sc, id+"_menu", "name,value,title,conditions", "default_menu",
		SysParamString("default_ecosystem_menu"), "default", `ContractConditions("MainCondition")`)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return 0, err
	}
	var (
		ret []interface{}
		pub string
	)
	_, ret, err = DBSelect(sc, "1_keys", "pub", wallet, `id`, 0, 1, 0, ``, []interface{}{})
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting pub key")
		return 0, err
	}
	if Len(ret) > 0 {
		pub = ret[0].(map[string]string)[`pub`]
	}
	_, _, err = DBInsert(sc, id+"_keys", "id,pub", wallet, pub)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return 0, err
	}
	return converter.StrToInt64(id), err
}

// RollbackEcosystem is rolling back ecosystem
func RollbackEcosystem(sc *SmartContract) error {
	if sc.TxContract.Name != `@1NewEcosystem` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackEcosystem can be only called from @1NewEcosystem")
		return fmt.Errorf(`RollbackEcosystem can be only called from @1NewEcosystem`)
	}
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, "system_states")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback tx")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("system states in rollback table")
		// if there is not such hash then NewEcosystem was faulty. Do nothing.
		return nil
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
		`contracts`, `parameters`, `blocks`, `history`, `keys`, `sections`, `member`, `roles_list`,
		`roles_assign`, `notifications`} {
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

// RollbackTable is rolling back table
func RollbackTable(sc *SmartContract, name string) error {
	if sc.TxContract.Name != `@1NewTable` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackTable can be only called from @1NewTable")
		return fmt.Errorf(`RollbackTable can be only called from @1NewTable`)
	}
	tableName := getDefTableName(sc, name)
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, tableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback table")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("table record in rollback table")
		// if there is not such hash then NewTable was faulty. Do nothing.
		return nil
	}
	err = rollbackTx.DeleteByHashAndTableName(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting record from rollback table")
		return err
	}

	err = model.DropTable(sc.DbTransaction, fmt.Sprintf("%d_%s", sc.TxSmart.EcosystemID, name))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
		return err
	}
	t := model.Table{}
	t.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID))
	found, err = t.Get(sc.DbTransaction, name)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table info")
		return err
	}
	if found {
		err = t.Delete(sc.DbTransaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting table")
			return err
		}
	} else {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("not found table info")
	}
	return nil
}

// RollbackColumn is rolling back column
func RollbackColumn(sc *SmartContract, tableName, name string) error {
	if sc.TxContract.Name != `@1NewColumn` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("RollbackColumn can be only called from @1NewColumn")
		return fmt.Errorf(`RollbackColumn can be only called from @1NewColumn`)
	}
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, fmt.Sprintf("%d_tables", sc.TxSmart.EcosystemID))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column from rollback table")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("column record in rollback table")
		// if there is not such hash then NewColumn was faulty. Do nothing.
		return nil
	}
	return model.AlterTableDropColumn(fmt.Sprintf(`%d_%s`, sc.TxSmart.EcosystemID, tableName), name)
}

// UpdateLang updates language resource
func UpdateLang(sc *SmartContract, name, trans string) {
	language.UpdateLang(int(sc.TxSmart.EcosystemID), name, trans, sc.VDE)
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

// Activate sets Active status of the contract in smartVM
func Activate(sc *SmartContract, tblid int64, state int64) error {
	if sc.TxContract.Name != `@1ActivateContract` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("ActivateContract can be only called from @1ActivateContract")
		return fmt.Errorf(`ActivateContract can be only called from @1ActivateContract`)
	}
	ActivateContract(tblid, state, true)
	return nil
}

// DeactivateContract sets Active status of the contract in smartVM
func Deactivate(sc *SmartContract, tblid int64, state int64) error {
	if sc.TxContract.Name != `@1DeactivateContract` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("DeactivateContract can be only called from @1DeactivateContract")
		return fmt.Errorf(`DeactivateContract can be only called from @1DeactivateContract`)
	}
	ActivateContract(tblid, state, false)
	return nil
}

// CheckSignature checks the additional signatures for the contract
func CheckSignature(i *map[string]interface{}, name string) error {
	state, name := script.ParseContract(name)
	pref := converter.Int64ToStr(int64(state))
	sc := (*i)[`sc`].(*SmartContract)
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
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("comverting signature to hex")
		return fmt.Errorf(`wrong signature`)
	}

	var sign TxSignJSON
	err = json.Unmarshal([]byte(value), &sign)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling sign")
		return err
	}
	wallet := (*i)[`key_id`].(int64)
	forsign := fmt.Sprintf(`%d,%d`, uint64((*i)[`time`].(int64)), uint64(wallet))
	for _, isign := range sign.Params {
		val := (*i)[isign.Param]
		if val == nil {
			val = ``
		}
		forsign += fmt.Sprintf(`,%v`, val)
	}

	CheckSignResult, err := utils.CheckSign(sc.PublicKeys, forsign, hexsign, true)
	if err != nil {
		return err
	}
	if !CheckSignResult {
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect signature")
		return fmt.Errorf(`incorrect signature ` + forsign)
	}
	return nil
}

// JSONToMap is converting json to map
func JSONToMap(input string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(input), &ret)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling json to map")
		return nil, err
	}
	return ret, nil
}
