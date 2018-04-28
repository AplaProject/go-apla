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
	"regexp"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/language"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/metric"

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
		"GetContractByName": "extend_cost_contract_by_name",
		"GetContractById":   "extend_cost_contract_by_id",
	}
)

const (
	nActivateContract   = "ActivateContract"
	nDeactivateContract = "DeactivateContract"
	nEditColumn         = "EditColumn"
	nEditContract       = "EditContract"
	nEditEcosystemName  = "EditEcosystemName"
	nEditTable          = "EditTable"
	nImport             = "Import"
	nNewColumn          = "NewColumn"
	nNewContract        = "NewContract"
	nNewEcosystem       = "NewEcosystem"
	nNewTable           = "NewTable"
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
		return 0, logErrorDB(err, "system parameter get")
	}
	if !found {
		return 0, logError(fmt.Errorf(eParamNotFound, name), consts.NotFound, "system parameter get")
	}
	cond := par.Conditions
	if len(cond) > 0 {
		ret, err := sc.EvalIf(cond)
		if err != nil {
			return 0, logError(err, consts.EvalError, "evaluating conditions")
		}
		if !ret {
			return 0, logError(errAccessDenied, consts.AccessDenied, "Access denied")
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
		case `fuel_rate`, `commission_wallet`:
			err := json.Unmarshal([]byte(value), &list)
			if err != nil {
				return 0, logErrorValue(err, consts.JSONUnmarshallError,
					"unmarshalling system param", value)
			}
			for _, item := range list {
				switch name {
				case `fuel_rate`, `commission_wallet`:
					if len(item) != 2 || converter.StrToInt64(item[0]) <= 0 ||
						(name == `fuel_rate` && converter.StrToInt64(item[1]) <= 0) ||
						(name == `commission_wallet` && converter.StrToInt64(item[1]) == 0) {
						break check
					}
				}
			}
			checked = true
		case syspar.FullNodes:
			if err := json.Unmarshal([]byte(value), &[]syspar.FullNode{}); err != nil {
				break check
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
			return 0, logErrorValue(errInvalidValue, consts.InvalidObject, errInvalidValue.Error(),
				value)
		}
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := CompileEval(conditions, 0); err != nil {
			return 0, logErrorValue(err, consts.EvalError, "compiling eval", conditions)
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty value and condition")
		return 0, fmt.Errorf(`empty value and condition`)
	}
	_, _, err = sc.update(fields, values, "1_system_parameters", []string{"id"}, []string{converter.Int64ToStr(par.ID)})
	if err != nil {
		return 0, err
	}
	err = syspar.SysUpdate(sc.DbTransaction)
	if err != nil {
		return 0, logErrorDB(err, "updating syspar")
	}
	sc.SysUpdate = true
	return 0, nil
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

// Int converts the value to a number
func Int(v interface{}) (int64, error) {
	return converter.ValueToInt(v)
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
func Money(v interface{}) (decimal.Decimal, error) {
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
func Sha256(text string) (string, error) {
	hash, err := crypto.Hash([]byte(text))
	if err != nil {
		return ``, logErrorValue(err, consts.CryptoError, "hashing text", text)
	}
	hash = converter.BinToHex(hash)
	return string(hash), nil
}

// PubToID returns a numeric identifier for the public key specified in the hexadecimal form.
func PubToID(hexkey string) int64 {
	pubkey, err := hex.DecodeString(hexkey)
	if err != nil {
		logErrorValue(err, consts.CryptoError, "decoding hexkey to string", hexkey)
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

// GetContractByName returns id of the contract with this name
func GetContractByName(sc *SmartContract, name string) int64 {
	contract := VMGetContract(sc.VM, name, uint32(sc.TxSmart.EcosystemID))
	if contract == nil {
		return 0
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	if info == nil {
		return 0
	}
	return info.Owner.TableID
}

// GetContractById returns the name of the contract with this id
func GetContractById(sc *SmartContract, id int64) string {
	_, ret, err := DBSelect(sc, "contracts", "value", id, `id`, 0, 1,
		0, ``, []interface{}{})
	if err != nil || len(ret) != 1 {
		logErrorDB(err, "getting contract name")
		return ``
	}

	re := regexp.MustCompile(`(?is)^\s*contract\s+([\d\w_]+)\s*{`)
	names := re.FindStringSubmatch(ret[0].(map[string]string)["value"])
	if len(names) != 2 {
		return ``
	}
	return names[1]
}

// EvalCondition gets the condition and check it
func EvalCondition(sc *SmartContract, table, name, condfield string) error {
	conditions, err := model.Single(`SELECT `+converter.EscapeName(condfield)+` FROM "`+getDefTableName(sc, table)+
		`" WHERE name = ?`, name).String()
	if err != nil {
		return logErrorDB(err, "executing single query")
	}
	if len(conditions) == 0 {
		return logError(fmt.Errorf(eRecordNotFound, name), consts.NotFound, "Record not found")
	}
	return Eval(sc, conditions)
}

// Replace replaces old substrings to new substrings
func Replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// CreateEcosystem creates a new ecosystem
func CreateEcosystem(sc *SmartContract, wallet int64, name string) (int64, error) {
	if err := validateAccess(`CreateEcosystem`, sc, nNewEcosystem); err != nil {
		return 0, err
	}

	var sp model.StateParameter
	sp.SetTablePrefix(`1`)
	found, err := sp.Get(sc.DbTransaction, `founder_account`)
	if err != nil {
		return 0, logErrorDB(err, "getting founder")
	}

	if !found || len(sp.Value) == 0 {
		return 0, logError(errFounderAccount, consts.NotFound, "founder not found")
	}

	id, err := model.GetNextID(sc.DbTransaction, "1_ecosystems")
	if err != nil {
		return 0, logErrorDB(err, "generating next ecosystem id")
	}

	if err = model.ExecSchemaEcosystem(sc.DbTransaction, int(id), wallet, name, converter.StrToInt64(sp.Value)); err != nil {
		return 0, logErrorDB(err, "executing ecosystem schema")
	}

	idStr := converter.Int64ToStr(id)
	if err := LoadContract(sc.DbTransaction, idStr); err != nil {
		return 0, err
	}

	sc.Rollback = false
	if _, _, err = DBInsert(sc, `@`+idStr+"_pages", "id,name,value,menu,conditions", "1", "default_page",
		SysParamString("default_ecosystem_page"), "default_menu", `ContractConditions("MainCondition")`); err != nil {
		return 0, logErrorDB(err, "inserting default page")
	}
	if _, _, err = DBInsert(sc, `@`+idStr+"_menu", "id,name,value,title,conditions", "1", "default_menu",
		SysParamString("default_ecosystem_menu"), "default", `ContractConditions("MainCondition")`); err != nil {
		return 0, logErrorDB(err, "inserting default page")
	}

	var (
		ret []interface{}
		pub string
	)
	_, ret, err = DBSelect(sc, "@1_keys", "pub", wallet, `id`, 0, 1, 0, ``, []interface{}{})
	if err != nil {
		return 0, logErrorDB(err, "getting pub key")
	}

	if Len(ret) > 0 {
		pub = ret[0].(map[string]string)[`pub`]
	}
	if _, _, err := DBInsert(sc, `@`+idStr+"_keys", "id,pub", wallet, pub); err != nil {
		return 0, logErrorDB(err, "inserting default page")
	}

	// because of we need to know which ecosystem to rollback.
	// All tables will be deleted so it's no need to rollback data from tables
	sc.Rollback = true
	if _, _, err := DBInsert(sc, "@1_ecosystems", "id,name", id, name); err != nil {
		return 0, logErrorDB(err, "insert new ecosystem to stat table")
	}

	return id, err
}

// EditEcosysName set newName for ecosystem
func EditEcosysName(sc *SmartContract, sysID int64, newName string) error {
	if err := validateAccess(`EditEcosysName`, sc, nEditEcosystemName); err != nil {
		return err
	}
	_, err := DBUpdate(sc, "@1_ecosystems", sysID, "name", newName)
	return err
}

// RollbackEcosystem is rolling back ecosystem
func RollbackEcosystem(sc *SmartContract) error {
	if err := validateAccess(`RollbackEcosystem`, sc, nNewEcosystem); err != nil {
		return err
	}

	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, "1_ecosystems")
	if err != nil {
		return logErrorDB(err, "getting rollback tx")
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("system states in rollback table")
		// if there is not such hash then NewEcosystem was faulty. Do nothing.
		return nil
	}
	lastID, err := model.GetNextID(sc.DbTransaction, "1_ecosystems")
	if err != nil {
		return logErrorDB(err, "getting next id")
	}
	lastID--
	if converter.StrToInt64(rollbackTx.TableID) != lastID {
		log.WithFields(log.Fields{"table_id": rollbackTx.TableID, "last_id": lastID, "type": consts.InvalidObject}).Error("incorrect ecosystem id")
		return fmt.Errorf(`Incorrect ecosystem id %s != %d`, rollbackTx.TableID, lastID)
	}

	rbTables := []string{
		`menu`,
		`pages`,
		`languages`,
		`signatures`,
		`tables`,
		`contracts`,
		`parameters`,
		`blocks`,
		`history`,
		`keys`,
		`sections`,
		`members`,
		`roles`,
		`roles_participants`,
		`notifications`,
		`applications`,
		`binaries`,
		`app_params`,
	}

	if rollbackTx.TableID == "1" {
		rbTables = append(rbTables, `system_parameters`, `ecosystems`)
	}

	for _, name := range rbTables {
		err = model.DropTable(sc.DbTransaction, fmt.Sprintf("%s_%s", rollbackTx.TableID, name))
		if err != nil {
			return logErrorDB(err, "dropping table")
		}
	}
	rollbackTxToDel := &model.RollbackTx{TxHash: sc.TxHash, NameTable: "1_ecosystems"}
	err = rollbackTxToDel.DeleteByHashAndTableName(sc.DbTransaction)
	if err != nil {
		return logErrorDB(err, "deleting rollback tx by hash and table name")
	}

	ecosysToDel := &model.Ecosystem{ID: lastID}
	return ecosysToDel.Delete(sc.DbTransaction)
}

// RollbackTable is rolling back table
func RollbackTable(sc *SmartContract, name string) error {
	if err := validateAccess(`RollbackTable`, sc, nNewTable); err != nil {
		return err
	}
	name = strings.ToLower(name)
	tableName := getDefTableName(sc, name)
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, tableName)
	if err != nil {
		return logErrorDB(err, "getting rollback table")
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("table record in rollback table")
		// if there is not such hash then NewTable was faulty. Do nothing.
		return nil
	}
	err = rollbackTx.DeleteByHashAndTableName(sc.DbTransaction)
	if err != nil {
		return logErrorDB(err, "deleting record from rollback table")
	}

	err = model.DropTable(sc.DbTransaction, tableName)
	if err != nil {
		return logErrorDB(err, "dropping table")
	}
	t := model.Table{}
	t.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID))
	found, err = t.Get(sc.DbTransaction, name)
	if err != nil {
		return logErrorDB(err, "getting table info")
	}
	if found {
		err = t.Delete(sc.DbTransaction)
		if err != nil {
			return logErrorDB(err, "deleting table")
		}
	} else {
		logError(err, consts.NotFound, "not found table info")
	}
	return nil
}

// RollbackColumn is rolling back column
func RollbackColumn(sc *SmartContract, tableName, name string) error {
	if err := validateAccess(`RollbackColumn`, sc, nNewColumn); err != nil {
		return err
	}
	name = strings.ToLower(name)
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, fmt.Sprintf("%d_tables", sc.TxSmart.EcosystemID))
	if err != nil {
		return logErrorDB(err, "getting column from rollback table")
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("column record in rollback table")
		// if there is not such hash then NewColumn was faulty. Do nothing.
		return nil
	}
	return model.AlterTableDropColumn(getDefTableName(sc, tableName), name)
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
	if err := validateAccess(`Activate`, sc, nActivateContract, nDeactivateContract); err != nil {
		return err
	}
	ActivateContract(tblid, state, true)
	return nil
}

// Deactivate sets Active status of the contract in smartVM
func Deactivate(sc *SmartContract, tblid int64, state int64) error {
	if err := validateAccess(`Deactivate`, sc, nActivateContract, nDeactivateContract); err != nil {
		return err
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
		return logErrorDB(err, "executing single query")
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
		return logErrorValue(err, consts.JSONUnmarshallError, "unmarshalling sign", value)
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

// RollbackContract performs rollback for the contract
func RollbackContract(sc *SmartContract, name string) error {
	if err := validateAccess(`RollbackContract`, sc, nNewContract, nImport); err != nil {
		return err
	}
	if c := VMGetContract(sc.VM, name, uint32(sc.TxSmart.EcosystemID)); c != nil {
		id := c.Block.Info.(*script.ContractInfo).ID
		if int(id) < len(sc.VM.Children) {
			sc.VM.Children = sc.VM.Children[:id]
		}
		delete(sc.VM.Objects, c.Name)
	}

	return nil
}

// DBSelectMetrics returns list of metrics by name and time interval
func DBSelectMetrics(sc *SmartContract, metric, timeInterval, aggregateFunc string) ([]interface{}, error) {
	result, err := model.GetMetricValues(metric, timeInterval, aggregateFunc)
	if err != nil {
		return nil, logErrorDB(err, "get values of metric")
	}
	return result, nil
}

// DBCollectMetrics returns actual values of all metrics
// This function used to further store these values
func DBCollectMetrics() []interface{} {
	c := metric.NewCollector(
		metric.CollectMetricDataForEcosystemTables,
		metric.CollectMetricDataForEcosystemTx,
	)
	return c.Values()
}

// RollbackEditContract rollbacks the contract
func RollbackEditContract(sc *SmartContract) error {
	if err := validateAccess(`RollbackEditContract`, sc, nEditContract); err != nil {
		return err
	}
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(sc.DbTransaction, sc.TxHash, fmt.Sprintf("%d_contracts", sc.TxSmart.EcosystemID))
	if err != nil {
		return logErrorDB(err, "getting contract from rollback table")
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("contract record in rollback table")
		// if there is not such hash then EditContract was faulty. Do nothing.
		return nil
	}
	var fields map[string]string
	err = json.Unmarshal([]byte(rollbackTx.Data), &fields)
	if err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, "unmarshalling contract values",
			rollbackTx.Data)
	}
	if len(fields["value"]) > 0 {
		var owner *script.OwnerInfo
		for i, item := range smartVM.Block.Children {
			if item != nil && item.Type == script.ObjContract {
				cinfo := item.Info.(*script.ContractInfo)
				if cinfo.Owner.TableID == converter.StrToInt64(rollbackTx.TableID) &&
					cinfo.Owner.StateID == uint32(sc.TxSmart.EcosystemID) {
					owner = smartVM.Children[i].Info.(*script.ContractInfo).Owner
					break
				}
			}
		}
		if owner == nil {
			return logError(errContractNotFound, consts.VMError, "getting existing contract")
		}
		wallet := owner.WalletID
		if len(fields["wallet_id"]) > 0 {
			wallet = converter.StrToInt64(fields["wallet_id"])
		}
		root, err := CompileContract(sc, fields["value"], int64(owner.StateID), wallet, owner.TokenID)
		if err != nil {
			return logError(err, consts.VMError, "compiling contract")
		}
		err = FlushContract(sc, root, owner.TableID, owner.Active)
		if err != nil {
			return logError(err, consts.VMError, "flushing contract")
		}
	} else if len(fields["wallet_id"]) > 0 {
		return SetContractWallet(sc, converter.StrToInt64(rollbackTx.TableID), sc.TxSmart.EcosystemID,
			converter.StrToInt64(fields["wallet_id"]))
	}
	return nil
}

// JSONDecode converts json string to object
func JSONDecode(input string) (interface{}, error) {
	var ret interface{}
	err := json.Unmarshal([]byte(input), &ret)
	if err != nil {
		return nil, logError(err, consts.JSONUnmarshallError, "unmarshalling json")
	}
	return ret, nil
}

// JSONEncode converts object to json string
func JSONEncode(input interface{}) (string, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return "", logError(err, consts.JSONMarshallError, "marshalling json")
	}
	return string(b), nil
}
