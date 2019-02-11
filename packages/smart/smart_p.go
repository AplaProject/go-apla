// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package smart

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/metric"

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
		"AddressToId":        "price_exec_address_to_id",
		"IdToAddress":        "price_exec_id_to_address",
		"Sha256":             "price_exec_sha256",
		"PubToID":            "price_exec_pub_to_id",
		"EcosysParam":        "price_exec_ecosys_param",
		"SysParamString":     "price_exec_sys_param_string",
		"SysParamInt":        "price_exec_sys_param_int",
		"SysFuel":            "price_exec_sys_fuel",
		"ValidateCondition":  "price_exec_validate_condition",
		"EvalCondition":      "price_exec_eval_condition",
		"HasPrefix":          "price_exec_has_prefix",
		"Contains":           "price_exec_contains",
		"Replace":            "price_exec_replace",
		"Join":               "price_exec_join",
		"Size":               "price_exec_size",
		"Substr":             "price_exec_substr",
		"ContractsList":      "price_exec_contracts_list",
		"IsObject":           "price_exec_is_object",
		"CompileContract":    "price_exec_compile_contract",
		"FlushContract":      "price_exec_flush_contract",
		"Eval":               "price_exec_eval",
		"Len":                "price_exec_len",
		"BindWallet":         "price_exec_bind_wallet",
		"UnbindWallet":       "price_exec_unbind_wallet",
		"CreateEcosystem":    "price_exec_create_ecosystem",
		"TableConditions":    "price_exec_table_conditions",
		"CreateTable":        "price_exec_create_table",
		"PermTable":          "price_exec_perm_table",
		"ColumnCondition":    "price_exec_column_condition",
		"CreateColumn":       "price_exec_create_column",
		"PermColumn":         "price_exec_perm_column",
		"JSONToMap":          "price_exec_json_to_map",
		"GetContractByName":  "price_exec_contract_by_name",
		"GetContractById":    "price_exec_contract_by_id",
		"TxData":             "price_tx_data",
		"ExecContractByName": "price_exec_contract_by_name",
		"ExecContractById":   "price_exec_contract_by_id",
	}
)

const (
	nBindWallet        = "BindWallet"
	nUnbindWallet      = "UnbindWallet"
	nEditColumn        = "EditColumn"
	nEditContract      = "EditContract"
	nEditEcosystemName = "EditEcosystemName"
	nEditLang          = "EditLang"
	nEditLangJoint     = "EditLangJoint"
	nEditTable         = "EditTable"
	nImport            = "Import"
	nNewColumn         = "NewColumn"
	nNewContract       = "NewContract"
	nNewEcosystem      = "NewEcosystem"
	nNewLang           = "NewLang"
	nNewLangJoint      = "NewLangJoint"
	nNewTable          = "NewTable"
	nNewTableJoint     = "NewTableJoint"
	nNewUser           = "NewUser"
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
		return 0, logErrorf(eParamNotFound, name, consts.NotFound, "system parameter get")
	}
	cond := par.Conditions
	if len(cond) > 0 {
		ret, err := sc.EvalIf(cond)
		if err != nil {
			return 0, logError(err, consts.EvalError, "evaluating conditions")
		}
		if !ret {
			return 0, logErrorShort(errAccessDenied, consts.AccessDenied)
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
		case syspar.GapsBetweenBlocks:
			ok = ival > 0 && ival < 86400
		case syspar.RbBlocks1, syspar.NumberNodes:
			ok = ival > 0 && ival < 1000
		case syspar.CommissionSize:
			ok = ival >= 0
		case syspar.MaxBlockSize, syspar.MaxTxSize, syspar.MaxTxCount, syspar.MaxColumns,
			syspar.MaxIndexes, syspar.MaxBlockUserTx, syspar.MaxTxFuel, syspar.MaxBlockFuel, syspar.MaxForsignSize:
			ok = ival > 0
		case syspar.FuelRate, syspar.CommissionWallet:
			if err := unmarshalJSON([]byte(value), &list, `system param`); err != nil {
				return 0, err
			}
			for _, item := range list {
				switch name {
				case syspar.FuelRate, syspar.CommissionWallet:
					if len(item) != 2 || converter.StrToInt64(item[0]) <= 0 ||
						(name == syspar.FuelRate && converter.StrToInt64(item[1]) <= 0) ||
						(name == syspar.CommissionWallet && converter.StrToInt64(item[1]) == 0) {
						break check
					}
				}
			}
			checked = true
		case syspar.FullNodes:
			fnodes := []types.FullNode{}
			if err := json.Unmarshal([]byte(value), &fnodes); err != nil {
				break check
			}
			checked = len(fnodes) > 0
		default:
			if strings.HasPrefix(name, `extend_cost_`) || strings.HasSuffix(name, `_price`) {
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
		return 0, logErrorShort(errEmpty, consts.EmptyObject)
	}
	_, _, err = sc.update(fields, values, "1_system_parameters", "id", par.ID)
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
	if v == nil {
		return
	}
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
	pubkey, err := crypto.HexToPub(hexkey)
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
	ret, _ := language.LangText(idRes, int(sc.TxSmart.EcosystemID), lang)
	return ret
}

// NewLang creates new language
func CreateLanguage(sc *SmartContract, name, trans string) (id int64, err error) {
	if err := validateAccess(`CreateLanguage`, sc, nNewLang, nNewLangJoint, nImport); err != nil {
		return 0, err
	}
	idStr := converter.Int64ToStr(sc.TxSmart.EcosystemID)
	if _, id, err = DBInsert(sc, `@1languages`, types.LoadMap(map[string]interface{}{"name": name,
		"ecosystem": idStr, "res": trans})); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting new language")
		return 0, err
	}
	language.UpdateLang(int(sc.TxSmart.EcosystemID), name, trans)
	return id, nil
}

// EditLanguage edits language
func EditLanguage(sc *SmartContract, id int64, name, trans string) error {
	if err := validateAccess(`EditLanguage`, sc, nEditLang, nEditLangJoint, nImport); err != nil {
		return err
	}
	if _, err := DBUpdate(sc, `@1languages`, id,
		types.LoadMap(map[string]interface{}{"name": name, "res": trans})); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting new language")
		return err
	}
	language.UpdateLang(int(sc.TxSmart.EcosystemID), name, trans)
	return nil
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
	_, ret, err := DBSelect(sc, "contracts", "value", id, `id`, 0, 1, nil)
	if err != nil || len(ret) != 1 {
		logErrorDB(err, "getting contract name")
		return ``
	}

	re := regexp.MustCompile(`(?is)^\s*contract\s+([\d\w_]+)\s*{`)
	var val string
	if v, found := ret[0].(*types.Map).Get("value"); found {
		val = v.(string)
	}
	names := re.FindStringSubmatch(val)
	if len(names) != 2 {
		return ``
	}
	return names[1]
}

// EvalCondition gets the condition and check it
func EvalCondition(sc *SmartContract, table, name, condfield string) error {
	tableName := converter.ParseTable(table, sc.TxSmart.EcosystemID)
	query := `SELECT ` + converter.EscapeName(condfield) + ` FROM "` + tableName + `" WHERE name = ? and ecosystem = ?`
	conditions, err := model.Single(sc.DbTransaction, query, name, sc.TxSmart.EcosystemID).String()
	if err != nil {
		return logErrorDB(err, "executing single query")
	}
	if len(conditions) == 0 {
		return logErrorfShort(eRecordNotFound, name, consts.NotFound)
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
		return 0, logErrorShort(errFounderAccount, consts.NotFound)
	}

	id, err := model.GetNextID(sc.DbTransaction, "1_ecosystems")
	if err != nil {
		return 0, logErrorDB(err, "generating next ecosystem id")
	}

	appID, err := model.GetNextID(sc.DbTransaction, "1_applications")
	if err != nil {
		return 0, logErrorDB(err, "generating next application id")
	}

	if err = model.ExecSchemaEcosystem(sc.DbTransaction, int(id), wallet, name, converter.StrToInt64(sp.Value), appID); err != nil {
		return 0, logErrorDB(err, "executing ecosystem schema")
	}

	idStr := converter.Int64ToStr(id)
	if err := LoadContract(sc.DbTransaction, id); err != nil {
		return 0, err
	}
	if !sc.OBS {
		if err := SysRollback(sc, SysRollData{Type: "NewEcosystem", ID: id}); err != nil {
			return 0, err
		}
	}

	sc.FullAccess = true

	if _, _, err = DBInsert(sc, "@1parameters", types.LoadMap(map[string]interface{}{
		"name": "error_page", "value": "@1error_page", "conditions": `ContractConditions("@MainCondition")`,
		"ecosystem": idStr,
	})); err != nil {
		return 0, logErrorDB(err, "inserting system parameter")
	}

	if _, _, err = DBInsert(sc, "@1applications", types.LoadMap(map[string]interface{}{
		"name":       "System",
		"conditions": `ContractConditions("MainCondition")`,
		"ecosystem":  id,
	})); err != nil {
		return 0, logErrorDB(err, "inserting application")
	}
	if _, _, err = DBInsert(sc, `@1pages`, types.LoadMap(map[string]interface{}{"ecosystem": idStr,
		"name": "default_page", "app_id": appID, "value": SysParamString("default_ecosystem_page"),
		"menu": "default_menu", "conditions": `ContractConditions("@1DeveloperCondition")`})); err != nil {
		return 0, logErrorDB(err, "inserting default page")
	}
	if _, _, err = DBInsert(sc, `@1menu`, types.LoadMap(map[string]interface{}{"ecosystem": idStr,
		"name": "default_menu", "value": SysParamString("default_ecosystem_menu"), "title": "default", "conditions": `ContractConditions("@1DeveloperCondition")`})); err != nil {
		return 0, logErrorDB(err, "inserting default page")
	}

	var (
		ret []interface{}
		pub string
	)
	_, ret, err = DBSelect(sc, "@1keys", "pub", wallet, `id`, 0, 1, nil)
	if err != nil {
		return 0, logErrorDB(err, "getting pub key")
	}

	if Len(ret) > 0 {
		if v, found := ret[0].(*types.Map).Get("pub"); found {
			pub = v.(string)
		}
	}
	if _, _, err := DBInsert(sc, `@1keys`, types.LoadMap(
		map[string]interface{}{"id": wallet, "pub": pub, "ecosystem": idStr})); err != nil {
		return 0, logErrorDB(err, "inserting key")
	}

	sc.FullAccess = false
	// because of we need to know which ecosystem to rollback.
	// All tables will be deleted so it's no need to rollback data from tables
	if _, _, err := DBInsert(sc, "@1ecosystems", types.LoadMap(map[string]interface{}{
		"id":   id,
		"name": name,
	})); err != nil {
		return 0, logErrorDB(err, "insert new ecosystem to stat table")
	}
	return id, err
}

// EditEcosysName set newName for ecosystem
func EditEcosysName(sc *SmartContract, sysID int64, newName string) error {
	if err := validateAccess(`EditEcosysName`, sc, nEditEcosystemName); err != nil {
		return err
	}

	_, err := DBUpdate(sc, "@1ecosystems", sysID,
		types.LoadMap(map[string]interface{}{"name": newName}))
	return err
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

// BndWallet sets wallet_id to current wallet and updates value in vm
func BndWallet(sc *SmartContract, tblid int64, state int64) error {
	if err := validateAccess(`BindWallet`, sc, nBindWallet); err != nil {
		log.Error("BindWallet access denied")
		return err
	}

	if _, _, err := sc.update([]string{"wallet_id"}, []interface{}{sc.TxSmart.KeyID}, "1_contracts", "id", tblid); err != nil {
		log.WithFields(log.Fields{"error": err, "contract_id": tblid}).Error("on updating contract wallet")
		return err
	}

	return SetContractWallet(sc, tblid, state, sc.TxSmart.KeyID)
}

// UnbndWallet sets Active status of the contract in smartVM
func UnbndWallet(sc *SmartContract, tblid int64, state int64) error {
	if err := validateAccess(`UnbindWallet`, sc, nUnbindWallet); err != nil {
		return err
	}

	if _, _, err := sc.update([]string{"wallet_id"}, []interface{}{0}, "1_contracts", "id", tblid); err != nil {
		log.WithFields(log.Fields{"error": err, "contract_id": tblid}).Error("on updating contract wallet")
		return err
	}

	return SetContractWallet(sc, tblid, state, 0)
}

// CheckSignature checks the additional signatures for the contract
func CheckSignature(i *map[string]interface{}, name string) error {
	state, name := converter.ParseName(name)
	sc := (*i)[`sc`].(*SmartContract)
	sn := model.Signature{}
	sn.SetTablePrefix(converter.Int64ToStr(int64(state)))
	_, err := sn.Get(name)
	if err != nil {
		return logErrorDB(err, "executing single query")
	}
	if len(sn.Value) == 0 {
		return nil
	}
	hexsign, err := hex.DecodeString((*i)[`Signature`].(string))
	if len(hexsign) == 0 || err != nil {
		return logError(errWrongSignature, consts.ConversionError, "converting signature to hex")
	}

	var sign TxSignJSON
	if err = unmarshalJSON([]byte(sn.Value), &sign, `unmarshalling sign`); err != nil {
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

	CheckSignResult, err := utils.CheckSign(sc.PublicKeys, []byte(forsign), hexsign, true)
	if err != nil {
		return err
	}
	if !CheckSignResult {
		return logErrorfShort(eIncorrectSignature, forsign, consts.InvalidObject)
	}
	return nil
}

// DBSelectMetrics returns list of metrics by name and time interval
func DBSelectMetrics(sc *SmartContract, metric, timeInterval, aggregateFunc string) ([]interface{}, error) {
	if conf.Config.IsSupportingOBS() {
		return nil, ErrNotImplementedOnOBS
	}

	timeBlock := time.Unix(sc.TxSmart.Time, 0).Format(`2006-01-02 15:04:05`)
	result, err := model.GetMetricValues(metric, timeInterval, aggregateFunc, timeBlock)
	if err != nil {
		return nil, logErrorDB(err, "get values of metric")
	}
	return result, nil
}

// DBCollectMetrics returns actual values of all metrics
// This function used to further store these values
func DBCollectMetrics(sc *SmartContract) []interface{} {
	if conf.Config.IsSupportingOBS() {
		return nil
	}

	c := metric.NewCollector(
		metric.CollectMetricDataForEcosystemTables,
		metric.CollectMetricDataForEcosystemTx,
	)
	return c.Values(sc.TxSmart.Time)
}

// JSONDecode converts json string to object
func JSONDecode(input string) (ret interface{}, err error) {
	err = unmarshalJSON([]byte(input), &ret, "unmarshalling json")
	ret = types.ConvertMap(ret)
	return
}

// JSONEncodeIdent converts object to json string
func JSONEncodeIndent(input interface{}, indent string) (string, error) {
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct && reflect.TypeOf(input).String() != `*types.Map` {
		return "", logErrorfShort(eTypeJSON, input, consts.TypeError)
	}
	var (
		b   []byte
		err error
	)
	if len(indent) == 0 {
		b, err = json.Marshal(input)
	} else {
		b, err = json.MarshalIndent(input, ``, indent)
	}
	if err != nil {
		return ``, logError(err, consts.JSONMarshallError, `marshalling json`)
	}
	out := string(b)
	out = strings.Replace(out, `\u003c`, `<`, -1)
	out = strings.Replace(out, `\u003e`, `>`, -1)
	out = strings.Replace(out, `\u0026`, `&`, -1)

	return out, nil
}

// JSONEncode converts object to json string
func JSONEncode(input interface{}) (string, error) {
	return JSONEncodeIndent(input, ``)
}

// Append syn for golang 'append' function
func Append(slice []interface{}, val interface{}) []interface{} {
	return append(slice, val)
}
