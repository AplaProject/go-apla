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
	"unicode/utf8"

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
	"github.com/GenesisKernel/go-genesis/packages/vdemanager"
	"github.com/satori/go.uuid"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	nodeBanNotificationHeader = "Your node was banned"
	historyLimit              = 250
)

var BOM = []byte{0xEF, 0xBB, 0xBF}

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
	FullAccess    bool
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
	Rand          *rand.Rand
}

// AppendStack adds an element to the stack of contract call or removes the top element when name is empty
func (sc *SmartContract) AppendStack(contract string) error {
	cont := sc.TxContract
	if len(contract) > 0 {
		for _, item := range cont.StackCont {
			if item == contract {
				return fmt.Errorf(eContractLoop, contract)
			}
		}
		cont.StackCont = append(cont.StackCont, contract)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	(*sc.TxContract.Extend)["stack"] = cont.StackCont
	return nil
}

var (
	funcCallsDB = map[string]struct{}{
		"DBInsert":    {},
		"DBSelect":    {},
		"DBUpdate":    {},
		"DBUpdateExt": {},
		"SetPubKey":   {},
	}
	extendCost = map[string]int64{
		"AddressToId":                  10,
		"ColumnCondition":              50,
		"Contains":                     10,
		"ContractAccess":               50,
		"ContractConditions":           50,
		"ContractName":                 10,
		"CreateColumn":                 50,
		"CreateTable":                  100,
		"CreateLanguage":               50,
		"EditLanguage":                 50,
		"CreateContract":               60,
		"UpdateContract":               60,
		"EcosysParam":                  10,
		"AppParam":                     10,
		"Eval":                         10,
		"EvalCondition":                20,
		"GetContractByName":            20,
		"GetContractById":              20,
		"HMac":                         50,
		"Join":                         10,
		"JSONToMap":                    50,
		"Sha256":                       50,
		"IdToAddress":                  10,
		"Len":                          5,
		"Replace":                      10,
		"PermColumn":                   50,
		"Split":                        50,
		"PermTable":                    100,
		"Substr":                       10,
		"Size":                         10,
		"ToLower":                      10,
		"ToUpper":                      10,
		"TrimSpace":                    10,
		"TableConditions":              100,
		"ValidateCondition":            30,
		"ValidateEditContractNewValue": 10,
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
)

func getCost(name string) int64 {
	if val, ok := extendCost[name]; ok {
		return val
	}
	return -1
}

// EmbedFuncs is extending vm with embedded functions
func EmbedFuncs(vm *script.VM, vt script.VMType) {
	f := map[string]interface{}{
		"AddressToId":                  AddressToID,
		"ColumnCondition":              ColumnCondition,
		"Contains":                     strings.Contains,
		"ContractAccess":               ContractAccess,
		"ContractConditions":           ContractConditions,
		"ContractName":                 contractName,
		"ValidateEditContractNewValue": ValidateEditContractNewValue,
		"CreateColumn":                 CreateColumn,
		"CreateTable":                  CreateTable,
		"DBInsert":                     DBInsert,
		"DBSelect":                     DBSelect,
		"DBUpdate":                     DBUpdate,
		"DBUpdateSysParam":             UpdateSysParam,
		"DBUpdateExt":                  DBUpdateExt,
		"EcosysParam":                  EcosysParam,
		"AppParam":                     AppParam,
		"SysParamString":               SysParamString,
		"SysParamInt":                  SysParamInt,
		"SysFuel":                      SysFuel,
		"Eval":                         Eval,
		"EvalCondition":                EvalCondition,
		"Float":                        Float,
		"GetContractByName":            GetContractByName,
		"GetContractById":              GetContractById,
		"HMac":                         HMac,
		"Join":                         Join,
		"JSONToMap":                    JSONDecode, // Deprecated
		"JSONDecode":                   JSONDecode,
		"JSONEncode":                   JSONEncode,
		"IdToAddress":                  IDToAddress,
		"Int":                          Int,
		"Len":                          Len,
		"Money":                        Money,
		"PermColumn":                   PermColumn,
		"PermTable":                    PermTable,
		"Random":                       Random,
		"Split":                        Split,
		"Str":                          Str,
		"Substr":                       Substr,
		"Replace":                      Replace,
		"Size":                         Size,
		"Sha256":                       Sha256,
		"PubToID":                      PubToID,
		"HexToBytes":                   HexToBytes,
		"LangRes":                      LangRes,
		"HasPrefix":                    strings.HasPrefix,
		"ValidateCondition":            ValidateCondition,
		"TrimSpace":                    strings.TrimSpace,
		"ToLower":                      strings.ToLower,
		"ToUpper":                      strings.ToUpper,
		"CreateEcosystem":              CreateEcosystem,
		"RollbackEcosystem":            RollbackEcosystem,
		"CreateContract":               CreateContract,
		"UpdateContract":               UpdateContract,
		"RollbackTable":                RollbackTable,
		"TableConditions":              TableConditions,
		"RollbackColumn":               RollbackColumn,
		"CreateLanguage":               CreateLanguage,
		"EditLanguage":                 EditLanguage,
		"Activate":                     Activate,
		"Deactivate":                   Deactivate,
		"RollbackContract":             RollbackContract,
		"RollbackEditContract":         RollbackEditContract,
		"RollbackNewContract":          RollbackNewContract,
		"check_signature":              CheckSignature,
		"RowConditions":                RowConditions,
		"UUID":                         UUID,
		"DecodeBase64":                 DecodeBase64,
		"EncodeBase64":                 EncodeBase64,
		"MD5":                          MD5,
		"EditEcosysName":               EditEcosysName,
		"GetColumnType":                GetColumnType,
		"GetType":                      GetType,
		"AllowChangeCondition":         AllowChangeCondition,
		"StringToBytes":                StringToBytes,
		"BytesToString":                BytesToString,
		"SetPubKey":                    SetPubKey,
		"NewMoney":                     NewMoney,
		"GetMapKeys":                   GetMapKeys,
		"SortedKeys":                   SortedKeys,
		"Append":                       Append,
		"GetPageHistory":               GetPageHistory,
		"GetBlockHistory":              GetBlockHistory,
		"GetMenuHistory":               GetMenuHistory,
		"GetContractHistory":           GetContractHistory,
		"GetPageHistoryRow":            GetPageHistoryRow,
		"GetBlockHistoryRow":           GetBlockHistoryRow,
		"GetMenuHistoryRow":            GetMenuHistoryRow,
		"GetContractHistoryRow":        GetContractHistoryRow,
		"GetDataFromXLSX":              GetDataFromXLSX,
		"GetRowsCountXLSX":             GetRowsCountXLSX,
		"BlockTime":                    BlockTime,
	}

	switch vt {
	case script.VMTypeVDE:
		f["HTTPRequest"] = HTTPRequest
		f["Date"] = Date
		f["HTTPPostJSON"] = HTTPPostJSON
		f["ValidateCron"] = ValidateCron
		f["UpdateCron"] = UpdateCron
		vmExtendCost(vm, getCost)
		vmFuncCallsDB(vm, funcCallsDB)
	case script.VMTypeVDEMaster:
		f["HTTPRequest"] = HTTPRequest
		f["GetMapKeys"] = GetMapKeys
		f["SortedKeys"] = SortedKeys
		f["Date"] = Date
		f["HTTPPostJSON"] = HTTPPostJSON
		f["ValidateCron"] = ValidateCron
		f["UpdateCron"] = UpdateCron
		f["CreateVDE"] = CreateVDE
		f["DeleteVDE"] = DeleteVDE
		f["StartVDE"] = StartVDE
		f["StopVDEProcess"] = StopVDEProcess
		f["GetVDEList"] = GetVDEList
		vmExtendCost(vm, getCost)
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
	prefix := converter.Int64ToStr(ecosystem)
	return strings.ToLower(fmt.Sprintf(`%s_%s`, prefix, tblname))
}

func getDefTableName(sc *SmartContract, tblname string) string {
	return converter.EscapeSQL(GetTableName(sc, tblname, sc.TxSmart.EcosystemID))
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
				for i := len(sc.TxContract.StackCont) - 1; i >= 0; i-- {
					contName := sc.TxContract.StackCont[i].(string)
					if strings.HasPrefix(contName, `@`) {
						if contName == name {
							return true
						}
						break
					}
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
			vars := map[string]interface{}{`ecosystem_id`: int64(sc.TxSmart.EcosystemID),
				`key_id`: sc.TxSmart.KeyID, `sc`: sc, `original_contract`: ``, `this_contract`: ``, `role_id`: sc.TxSmart.RoleID}
			if err := sc.AppendStack(name); err != nil {
				return false, err
			}
			_, err := VMRun(sc.VM, block, []interface{}{}, &vars)
			if err != nil {
				return false, err
			}
			sc.AppendStack(``)
		} else {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty contract name in ContractConditions")
			return false, fmt.Errorf(`empty contract name in ContractConditions`)
		}
	}
	return true, nil
}

func contractName(value string) (string, error) {
	list, err := script.ContractsList(value)
	if err != nil {
		return "", err
	}
	if len(list) > 0 {
		return list[0], nil
	} else {
		return "", nil
	}
}

func ValidateEditContractNewValue(sc *SmartContract, newValue, oldValue string) error {
	list, err := script.ContractsList(newValue)
	if err != nil {
		return err
	}
	curlist, err := script.ContractsList(oldValue)
	if err != nil {
		return err
	}
	if len(list) != len(curlist) {
		return fmt.Errorf("Contract cannot be removed or inserted")
	}
	for i := 0; i < len(list); i++ {
		var ok bool
		for j := 0; j < len(curlist); j++ {
			if curlist[j] == list[i] {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("Contracts or functions names cannot be changed")
		}
	}
	return nil
}

func UpdateContract(sc *SmartContract, id int64, value, conditions, walletID string, recipient int64, active, tokenID string) error {
	if !accessContracts(sc, `EditContract`, `Import`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("UpdateContract can be only called from EditContract")
		return fmt.Errorf(`UpdateContract can be only called from EditContract`)
	}
	var pars []string
	var vals []interface{}
	ecosystemID := sc.TxSmart.EcosystemID
	var root interface{}
	if value != "" {
		var err error
		root, err = CompileContract(sc, value, ecosystemID, recipient, converter.StrToInt64(tokenID))
		if err != nil {
			return err
		}
		pars = append(pars, "value")
		vals = append(vals, value)
	}
	if conditions != "" {
		pars = append(pars, "conditions")
		vals = append(vals, conditions)
	}
	if walletID != "" {
		pars = append(pars, "wallet_id")
		vals = append(vals, recipient)
	}
	if len(vals) > 0 {
		if _, err := DBUpdate(sc, "contracts", id, strings.Join(pars, ","), vals...); err != nil {
			return err
		}
	}
	if value != "" {
		if err := FlushContract(sc, root, id, converter.StrToInt64(active) == 1); err != nil {
			return err
		}
	} else {
		if walletID != "" {
			if err := SetContractWallet(sc, id, ecosystemID, recipient); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateContract(sc *SmartContract, name, value, conditions string, walletID, tokenEcosystem, appID int64) (int64, error) {
	if !accessContracts(sc, `NewContract`, `Import`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("CreateContract can be only called from NewContract")
		return 0, fmt.Errorf(`CreateContract can be only called from NewContract`)
	}
	var id int64
	var err error

	if GetContractByName(sc, name) != 0 {
		return 0, fmt.Errorf(eContractExist, name)
	}
	root, err := CompileContract(sc, value, sc.TxSmart.EcosystemID, walletID, tokenEcosystem)
	if err != nil {
		return 0, err
	}
	_, id, err = DBInsert(sc, "contracts", "name,value,conditions,wallet_id,token_id,app_id", name, value, conditions, walletID, tokenEcosystem, appID)
	if err != nil {
		return 0, err
	}
	if err := FlushContract(sc, root, id, false); err != nil {
		return 0, err
	}
	return id, nil
}

func RollbackNewContract(sc *SmartContract, value string) error {
	contractList, err := script.ContractsList(value)
	if err != nil {
		return err
	}
	for _, contract := range contractList {
		if err := RollbackContract(sc, contract); err != nil {
			return err
		}
	}
	return nil
}

// CreateTable is creating smart contract table
func CreateTable(sc *SmartContract, name, columns, permissions string, applicationID int64) error {
	var err error
	if !accessContracts(sc, `NewTable`, `NewTableJoint`, `Import`) {
		return fmt.Errorf(`CreateTable can be only called from NewTable, NewTableJoint or Import`)
	}

	if len(name) == 0 {
		return fmt.Errorf("The table name cannot be empty")
	}

	if !converter.IsLatin(name) {
		return fmt.Errorf(eLatin, name)
	}

	tableName := getDefTableName(sc, name)
	if model.IsTable(tableName) {
		return fmt.Errorf("table %s exists", name)
	}

	var cols []interface{}
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "source": columns}).Error("unmarshalling columns to JSON")
		return err
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
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err,
					"source": v}).Error("unmarshalling columns permissions from json")
				return err
			}
		default:
			data = v.(map[string]interface{})
		}
		colname := converter.EscapeSQL(strings.ToLower(data[`name`].(string)))
		if err := checkColumnName(colname); err != nil {
			return err
		}
		if colList[colname] {
			return fmt.Errorf(`There are the same columns`)
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
				log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling conditions to json")
				return err
			}
			condition = string(out)
		}
		colperm[colname] = condition
	}
	colout, err := json.Marshal(colperm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling columns to JSON")
		return err
	}
	if sc.VDE {
		err = model.CreateVDETable(sc.DbTransaction, tableName, strings.TrimRight(colsSQL, ",\n"))
	} else {
		err = model.CreateTable(sc.DbTransaction, tableName, strings.TrimRight(colsSQL, ",\n"))
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE tables")
		return err
	}

	var perm permTable
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling permissions to JSON")
		return err
	}
	permout, err := json.Marshal(perm)
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
		AppID:       applicationID,
	}
	t.SetTablePrefix(prefix)
	err = t.Create(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("insert vde table info")
		return err
	}
	if !sc.VDE {
		rollbackTx := &model.RollbackTx{
			BlockID:   sc.BlockData.BlockID,
			TxHash:    sc.TxHash,
			NameTable: tableName,
			TableID:   converter.Int64ToStr(id),
		}
		err = rollbackTx.Create(sc.DbTransaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating CreateTable rollback")
			return err
		}
	}
	return nil
}

func columnType(colType string) (sqlColType string, err error) {
	switch colType {
	case "json":
		sqlColType = `jsonb`
	case "varchar":
		sqlColType = `varchar(102400)`
	case "character":
		sqlColType = `character(1) NOT NULL DEFAULT '0'`
	case "number":
		sqlColType = `bigint NOT NULL DEFAULT '0'`
	case "datetime":
		sqlColType = `timestamp`
	case "double":
		sqlColType = `double precision`
	case "money":
		sqlColType = `decimal (30, 0) NOT NULL DEFAULT '0'`
	case "text":
		sqlColType = "text"
	default:
		err = fmt.Errorf("Type '%s' of columns is not supported", colType)
	}

	return
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
	qcost, lastID, err = sc.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil,
		nil, !sc.VDE && sc.Rollback, false)
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
		icol = strings.TrimSpace(icol)
		if strings.Contains(icol, `->`) {
			colfield := strings.Split(icol, `->`)
			if len(colfield) == 2 {
				icol = fmt.Sprintf(`%s::jsonb->>'%s' as "%[1]s.%[2]s"`, colfield[0], colfield[1])
			} else {
				icol = fmt.Sprintf(`%s::jsonb#>>'{%s}' as "%[1]s.%[3]s"`, colfield[0],
					strings.Join(colfield[1:], `,`), strings.Join(colfield[1:], `.`))
			}
		} else if !strings.ContainsAny(icol, `:*>"`) {
			icol = `"` + icol + `"`
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

func checkNow(inputs ...string) error {
	re := regexp.MustCompile(`(now\s*\(\s*\)|localtime|current_date|current_time)`)
	for _, item := range inputs {
		if re.Match([]byte(strings.ToLower(item))) {
			return errNow
		}
	}
	return nil
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
	if err = checkNow(columns, where); err != nil {
		return 0, nil, err
	}
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

	perm, err = sc.AccessTablePerm(tblname, `read`)
	if err != nil {
		return 0, nil, err
	}
	colsList := strings.Split(columns, `,`)
	if err = sc.AccessColumns(tblname, &colsList, false); err != nil {
		return 0, nil, err
	}
	columns = strings.Join(colsList, `,`)

	columns = PrepareColumns(columns)
	rows, err = model.GetDB(sc.DbTransaction).Table(tblname).Select(columns).Where(where, params...).Order(order).
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
		row := make(map[string]interface{})
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
	if strings.Contains(tblname, `_reports_`) {
		err = fmt.Errorf(`Access denied to report table`)
		return
	}
	columns := strings.Split(params, `,`)
	if err = sc.AccessColumns(tblname, &columns, true); err != nil {
		return
	}
	qcost, _, err = sc.selectiveLoggingAndUpd(columns, val, tblname, []string{`id`}, []string{converter.Int64ToStr(id)}, !sc.VDE && sc.Rollback, true)
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting app param")
		return ``, err
	}
	return ap.Value, nil
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
		return errAccessDenied
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
	if id != 0 {
		if len(root.Children) != 1 || root.Children[0].Type != script.ObjContract {
			return fmt.Errorf(`Оnly one contract must be in the record`)
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
	var perm permTable
	err := json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling table permissions to json")
		return err
	}
	permout, err := json.Marshal(perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling permission list to json")
		return err
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`permissions`}, []interface{}{string(permout)},
		getDefTableName(sc, `tables`), []string{`name`}, []string{strings.ToLower(name)}, !sc.VDE && sc.Rollback, false)
	return err
}

// TableConditions is contract func
func TableConditions(sc *SmartContract, name, columns, permissions string) (err error) {
	isEdit := len(columns) == 0
	name = strings.ToLower(name)
	if isEdit {
		if !accessContracts(sc, `EditTable`) {
			log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("TableConditions can be only called from @1EditTable")
			return fmt.Errorf(`TableConditions can be only called from EditTable`)
		}
	} else if !accessContracts(sc, `NewTable`, `Import`, `NewTableJoint`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("TableConditions can be only called from @1NewTable, @1Import, @1NewTableJoint")
		return fmt.Errorf(`TableConditions can be only called from NewTable or Import or NewTableJoint`)
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
			return fmt.Errorf(eTableNotFound, name)
		}
	} else if exists {
		log.WithFields(log.Fields{"table_name": name, "type": consts.Found}).Error("table exists")
		return fmt.Errorf(`table %s exists`, name)
	}

	var perm permTable
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "source": permissions}).Error("unmarshalling permissions from json")
		return
	}
	v := reflect.ValueOf(perm)
	for i := 0; i < v.NumField(); i++ {
		cond := v.Field(i).Interface().(string)
		name := v.Type().Field(i).Name
		if len(cond) == 0 && name != `Read` && name != `Filter` {
			log.WithFields(log.Fields{"condition_type": name, "type": consts.EmptyObject}).Error("condition is empty")
			return fmt.Errorf(`%v condition is empty`, name)
		}
		if err = VMCompileEval(sc.VM, cond, uint32(sc.TxSmart.EcosystemID)); err != nil {
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

	var cols []interface{}
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "source": columns}).Error("unmarshalling columns permissions from json")
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
	for _, icol := range cols {
		var data map[string]interface{}
		switch v := icol.(type) {
		case string:
			err = json.Unmarshal([]byte(v), &data)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err,
					"source": v}).Error("unmarshalling columns permissions from json")
				return
			}
		default:
			data = v.(map[string]interface{})
		}
		if data[`name`] == nil || data[`type`] == nil {
			log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("wrong column")
			return fmt.Errorf(`worng column`)
		}
		itype := data[`type`].(string)
		if itype != `varchar` && itype != `number` && itype != `datetime` && itype != `text` &&
			itype != `bytea` && itype != `double` && itype != `json` && itype != `money` &&
			itype != `character` {
			log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect type")
			return fmt.Errorf(`incorrect type`)
		}
		condition := ``
		switch v := data[`conditions`].(type) {
		case string:
			condition = v
		case map[string]interface{}:
			out, err := json.Marshal(v)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling conditions to json")
				return err
			}
			condition = string(out)
		}
		perm, err := getPermColumns(condition)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Conditions is empty")
			return err
		}
		if len(perm.Update) == 0 {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Update condition is empty")
			return errConditionEmpty
		}
		if err = VMCompileEval(sc.VM, perm.Update, uint32(sc.TxSmart.EcosystemID)); err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError}).Error("compile update conditions")
			return err
		}
		if len(perm.Read) > 0 {
			if err = VMCompileEval(sc.VM, perm.Read, uint32(sc.TxSmart.EcosystemID)); err != nil {
				log.WithFields(log.Fields{"type": consts.EvalError}).Error("compile read conditions")
				return err
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
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("conditions cannot be empty")
		return fmt.Errorf("Conditions cannot be empty")
	}
	return VMCompileEval(sc.VM, condition, uint32(state))
}

// ColumnCondition is contract func
func ColumnCondition(sc *SmartContract, tableName, name, coltype, permissions string) error {
	name = converter.EscapeSQL(strings.ToLower(name))
	tableName = converter.EscapeSQL(strings.ToLower(tableName))
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

	exists, err := tEx.IsExistsByPermissionsAndTableName(sc.DbTransaction, name, tableName)
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
		log.WithFields(log.Fields{"table": tblName, "type": consts.DBError}).Error("counting table columns")
		return err
	}
	if count >= int64(syspar.GetMaxColumns()) {
		log.WithFields(log.Fields{"size": count, "max_size": syspar.GetMaxColumns(), "type": consts.ParameterExceeded}).Error("Too many columns")
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	if coltype != `varchar` && coltype != `number` && coltype != `datetime` &&
		coltype != `character` && coltype != `json` &&
		coltype != `text` && coltype != `bytea` && coltype != `double` && coltype != `money` {
		log.WithFields(log.Fields{"column_type": coltype, "type": consts.InvalidObject}).Error("Unknown column type")
		return fmt.Errorf(`incorrect type`)
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing row condition query")
		return err
	}

	if len(condition) == 0 {
		log.WithFields(log.Fields{"type": consts.NotFound, "name": tblname, "id": id}).Error("record not found")
		return fmt.Errorf("Item %d has not been found", id)
	}

	for _, v := range sc.TxContract.StackCont {
		if v == condition {
			return fmt.Errorf("Recursion detected")
		}
	}

	if err := Eval(sc, condition); err != nil {
		if err == errAccessDenied && conditionOnly {
			return AllowChangeCondition(sc, tblname)
		}

		return err
	}

	return nil
}

func checkColumnName(name string) error {
	if len(name) == 0 {
		return errEmptyColumn
	} else if name[0] >= '0' && name[0] <= '9' {
		return errWrongColumn
	}
	if !converter.IsLatin(name) {
		return fmt.Errorf(eLatin, name)
	}
	return nil
}

// CreateColumn is creating column
func CreateColumn(sc *SmartContract, tableName, name, colType, permissions string) (err error) {
	var (
		sqlColType string
		permout    []byte
	)
	if !accessContracts(sc, `NewColumn`) {
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("CreateColumn can be only called from @1NewColumn")
		return fmt.Errorf(`CreateColumn can be only called from NewColumn`)
	}
	name = converter.EscapeSQL(strings.ToLower(name))
	if err = checkColumnName(name); err != nil {
		return
	}

	tableName = strings.ToLower(tableName)
	tblname := getDefTableName(sc, tableName)

	sqlColType, err = columnType(colType)
	if err != nil {
		return
	}

	err = model.AlterTableAddColumn(sc.DbTransaction, tblname, name, sqlColType)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("adding column to the table")
		return
	}

	tables := getDefTableName(sc, `tables`)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err = model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		return
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting columns from the table")
		return
	}
	perm[name] = permissions
	permout, err = json.Marshal(perm)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling columns to json")
		return err
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName}, !sc.VDE && sc.Rollback, false)
	if err != nil {
		return
	}

	return nil
}

// SetPubKey updates the publis key
func SetPubKey(sc *SmartContract, id int64, pubKey []byte) (qcost int64, err error) {
	if !accessContracts(sc, `NewUser`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("SetPubKey can be only called from NewUser")
		return 0, fmt.Errorf(`SetPubKey can be only called from NewUser contract`)
	}
	if len(pubKey) == consts.PubkeySizeLength*2 {
		pubKey, err = hex.DecodeString(string(pubKey))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
			return
		}
	}
	qcost, _, err = sc.selectiveLoggingAndUpd([]string{`pub`}, []interface{}{pubKey},
		getDefTableName(sc, `keys`), []string{`id`}, []string{converter.Int64ToStr(id)},
		!sc.VDE && sc.Rollback, true)
	return qcost, err
}

func NewMoney(sc *SmartContract, id int64, amount, comment string) (err error) {
	if !accessContracts(sc, `NewUser`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("NewMoney can be only called from NewUser")
		return fmt.Errorf(`NewMoney can be only called from NewUser contract`)
	}
	_, _, err = sc.selectiveLoggingAndUpd([]string{`id`, `amount`}, []interface{}{id, amount},
		getDefTableName(sc, `keys`), nil, nil, !sc.VDE && sc.Rollback, false)
	if err == nil {
		var block int64
		if sc.BlockData != nil {
			block = sc.BlockData.BlockID
		}
		_, _, err = sc.selectiveLoggingAndUpd([]string{`sender_id`, `recipient_id`, `amount`,
			`comment`, `block_id`, `txhash`},
			[]interface{}{0, id, amount, comment, block, sc.TxHash},
			getDefTableName(sc, `history`), nil, nil, !sc.VDE && sc.Rollback, false)
	}
	return err
}

// PermColumn is contract func
func PermColumn(sc *SmartContract, tableName, name, permissions string) error {
	if !accessContracts(sc, `EditColumn`) {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("EditColumn can be only called from @1EditColumn")
		return fmt.Errorf(`EditColumn can be only called from EditColumn`)
	}
	name = converter.EscapeSQL(strings.ToLower(name))
	tableName = strings.ToLower(tableName)
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
		tables, []string{`name`}, []string{tableName}, !sc.VDE && sc.Rollback, false)
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

func Random(sc *SmartContract, min int64, max int64) (int64, error) {
	if min < 0 || max < 0 || min >= max {
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error("getting random")
		return 0, fmt.Errorf(`wrong random parameters %d %d`, min, max)
	}
	return min + sc.Rand.Int63n(max-min), nil
}

func ValidateCron(cronSpec string) error {
	_, err := scheduler.Parse(cronSpec)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCron(sc *SmartContract, id int64) error {
	cronTask := &model.Cron{}
	cronTask.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID) + "_vde")

	ok, err := cronTask.Get(id)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get cron record")
		return err
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get nodes need to be banned")
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
					log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node bad blocks for removing")
					return err
				}

				for _, b := range blocks {
					if _, err := DBUpdate(smartContract, "@1_bad_blocks", b.ID, "deleted", "1"); err != nil {
						log.WithFields(log.Fields{"type": consts.DBError, "id": b.ID, "error": err}).Error("deleting bad block")
						return err
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
					log.WithFields(log.Fields{"type": consts.DBError, "id": banReq.ProducerNodeId, "error": err}).Error("inserting log to node_ban_log")
					return err
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
					log.WithFields(log.Fields{"type": consts.DBError, "id": banReq.ProducerNodeId, "error": err}).Error("sending notification to node owner")
					return err
				}

				updFullNodes = true
			}
		}

		fullNodes[i] = fullNode
	}

	if updFullNodes {
		data, err := json.Marshal(fullNodes)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling full nodes")
			return err
		}

		_, err = UpdateSysParam(smartContract, syspar.FullNodes, string(data), "")
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating full nodes")
			return err
		}
	}

	return nil
}

func GetBlock(blockID int64) (map[string]int64, error) {
	block := model.Block{}
	ok, err := block.Get(blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return nil, err
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
		err := fmt.Errorf("Unsupported type %T", v)
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting to bytes")
		return "", err
	}

	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:]), nil
}

// GetColumnType returns the type of the column
func GetColumnType(sc *SmartContract, tableName, columnName string) (string, error) {
	return model.GetColumnType(getDefTableName(sc, tableName), columnName)
}

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
	if bytes.HasPrefix(src, BOM) && utf8.Valid(src[len(BOM):]) {
		return string(src[len(BOM):])
	}
	return string(src)
}

// CreateVDE allow create new VDE throw vdemanager
func CreateVDE(sc *SmartContract, name, dbUser, dbPassword string, port int64) error {
	return vdemanager.Manager.CreateVDE(name, dbUser, dbPassword, int(port))
}

// DeleteVDE delete vde
func DeleteVDE(sc *SmartContract, name string) error {
	return vdemanager.Manager.DeleteVDE(name)
}

// StartVDE run VDE process
func StartVDE(sc *SmartContract, name string) error {
	return vdemanager.Manager.StartVDE(name)
}

// StopVDEProcess stops VDE process
func StopVDEProcess(sc *SmartContract, name string) error {
	return vdemanager.Manager.StopVDE(name)
}

// GetVDEList returns list VDE process with statuses
func GetVDEList(sc *SmartContract) (map[string]string, error) {
	return vdemanager.Manager.ListProcess()
}

func GetHistory(transaction *model.DbTransaction, ecosystem int64, tableName string,
	id, idRollback int64) ([]interface{}, error) {
	table := fmt.Sprintf(`%d_%s`, ecosystem, tableName)
	rows, err := model.GetDB(transaction).Table(table).Where("id=?", id).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get current values")
		return nil, err
	}
	if !rows.Next() {
		return nil, errNotFound
	}
	defer rows.Close()
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get columns")
		return nil, err
	}
	values := make([][]byte, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	err = rows.Scan(scanArgs...)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("scan values")
		return nil, err
	}
	var value string
	curVal := make(map[string]string)
	for i, col := range values {
		if col == nil {
			value = "NULL"
		} else {
			value = string(col)
		}
		curVal[columns[i]] = value
	}
	rollbackList := []interface{}{}
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTxsByTableIDAndTableName(converter.Int64ToStr(id),
		table, historyLimit)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback history")
		return nil, err
	}
	for _, tx := range *txs {
		if len(rollbackList) > 0 {
			prev := rollbackList[len(rollbackList)-1].(map[string]string)
			prev[`block_id`] = converter.Int64ToStr(tx.BlockID)
			prev[`id`] = converter.Int64ToStr(tx.ID)
			block := model.Block{}
			if ok, err := block.Get(tx.BlockID); ok {
				prev[`block_time`] = time.Unix(block.Time, 0).Format(`2006-01-02 15:04:05`)
			} else if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block time")
				return nil, err
			}
			if idRollback == tx.ID {
				return rollbackList[len(rollbackList)-1 : len(rollbackList)], nil
			}
		}
		if tx.Data == "" {
			continue
		}
		rollback := make(map[string]string)
		for k, v := range curVal {
			rollback[k] = v
		}
		if err := json.Unmarshal([]byte(tx.Data), &rollback); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollbackTx.Data from JSON")
			return nil, err
		}
		rollbackList = append(rollbackList, rollback)
		curVal = rollback
	}
	if idRollback > 0 {
		return []interface{}{}, nil
	}
	return rollbackList, nil
}

func GetBlockHistory(sc *SmartContract, id int64) ([]interface{}, error) {
	return GetHistory(sc.DbTransaction, sc.TxSmart.EcosystemID, `blocks`, id, 0)
}

func GetPageHistory(sc *SmartContract, id int64) ([]interface{}, error) {
	return GetHistory(sc.DbTransaction, sc.TxSmart.EcosystemID, `pages`, id, 0)
}

func GetMenuHistory(sc *SmartContract, id int64) ([]interface{}, error) {
	return GetHistory(sc.DbTransaction, sc.TxSmart.EcosystemID, `menu`, id, 0)
}

func GetContractHistory(sc *SmartContract, id int64) ([]interface{}, error) {
	return GetHistory(sc.DbTransaction, sc.TxSmart.EcosystemID, `contracts`, id, 0)
}

func GetHistoryRow(sc *SmartContract, tableName string, id, idRollback int64) (map[string]interface{},
	error) {
	list, err := GetHistory(sc.DbTransaction, sc.TxSmart.EcosystemID, tableName, id, idRollback)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	if len(list) > 0 {
		for key, val := range list[0].(map[string]string) {
			result[key] = val
		}
	}
	return result, nil
}

func GetBlockHistoryRow(sc *SmartContract, id, idRollback int64) (map[string]interface{}, error) {
	return GetHistoryRow(sc, `blocks`, id, idRollback)
}

func GetPageHistoryRow(sc *SmartContract, id, idRollback int64) (map[string]interface{}, error) {
	return GetHistoryRow(sc, `pages`, id, idRollback)
}

func GetMenuHistoryRow(sc *SmartContract, id, idRollback int64) (map[string]interface{}, error) {
	return GetHistoryRow(sc, `menu`, id, idRollback)
}

func GetContractHistoryRow(sc *SmartContract, id, idRollback int64) (map[string]interface{}, error) {
	return GetHistoryRow(sc, `contracts`, id, idRollback)
}

func StackOverflow(sc *SmartContract) {
	StackOverflow(sc)
}

func BlockTime(sc *SmartContract) string {
	var blockTime int64
	if sc.BlockData != nil {
		blockTime = sc.BlockData.Time
	}
	if sc.VDE {
		blockTime = time.Now().Unix()
	}
	return Date(`2006-01-02 15:04:05`, blockTime)
}
