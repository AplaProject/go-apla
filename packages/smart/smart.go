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
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/migration/vde"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// Contract contains the information about the contract.
type Contract struct {
	Name          string
	Called        uint32
	FreeRequest   bool
	TxGovAccount  int64   // state wallet
	EGSRate       float64 // money/EGS rate
	TableAccounts string
	StackCont     []interface{} // Stack of called contracts
	Extend        *map[string]interface{}
	Block         *script.Block
}

func (c *Contract) Info() *script.ContractInfo {
	return c.Block.Info.(*script.ContractInfo)
}

const (
	// MaxPrice is a maximal value that price function can return
	MaxPrice = 100000000000000000

	CallDelayedContract = "@1CallDelayedContract"
	NewUserContract     = "@1NewUser"
	NewBadBlockContract = "@1NewBadBlock"
)

var (
	smartVM   *script.VM
	smartTest = make(map[string]string)
)

func testValue(name string, v ...interface{}) {
	smartTest[name] = fmt.Sprint(v...)
}

// GetTestValue returns the test value of the specified key
func GetTestValue(name string) string {
	return smartTest[name]
}

// GetLogger is returning logger
func (sc SmartContract) GetLogger() *log.Entry {
	var name string
	if sc.TxContract != nil {
		name = sc.TxContract.Name
	}
	return log.WithFields(log.Fields{"vde": sc.VDE, "name": name})
}

func newVM() *script.VM {
	vm := script.NewVM()
	vm.Extern = true
	vm.Extend(&script.ExtendData{Objects: map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
		"Float":   Float,
		"Money":   script.ValueToDecimal,
		`Test`:    testValue,
	}, AutoPars: map[string]string{
		`*smart.Contract`: `sc`,
	}})
	return vm
}

func init() {
	smartVM = newVM()
}

// GetVM is returning smart vm
func GetVM() *script.VM {
	return smartVM
}

func vmExternOff(vm *script.VM) {
	vm.FlushExtern()
}

func vmCompile(vm *script.VM, src string, owner *script.OwnerInfo) error {
	return vm.Compile([]rune(src), owner)
}

// VMCompileBlock is compiling block
func VMCompileBlock(vm *script.VM, src string, owner *script.OwnerInfo) (*script.Block, error) {
	return vm.CompileBlock([]rune(src), owner)
}

func getContractList(src string) (list []string) {
	for _, funcCond := range []string{`ContractConditions`, `ContractAccess`} {
		if strings.Contains(src, funcCond) {
			if ret := regexp.MustCompile(funcCond +
				`\(\s*(.*)\s*\)`).FindStringSubmatch(src); len(ret) == 2 {
				for _, item := range strings.Split(ret[1], `,`) {
					list = append(list, strings.Trim(item, "\"` "))
				}
			}
		}
	}
	return
}

func VMCompileEval(vm *script.VM, src string, prefix uint32) error {
	var ok bool
	if len(src) == 0 {
		return nil
	}
	allowed := []string{`0`, `1`, `true`, `false`, `ContractConditions\(\s*\".*\"\s*\)`,
		`ContractAccess\(\s*\".*\"\s*\)`, `RoleAccess\(\s*.*\s*\)`}
	for _, v := range allowed {
		re := regexp.MustCompile(`^` + v + `$`)
		if re.Match([]byte(src)) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf(eConditionNotAllowed, src)
	}
	err := vm.CompileEval(src, prefix)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`^@?[\d\w\_]+$`)
	for _, item := range getContractList(src) {
		if len(item) == 0 || !re.Match([]byte(item)) {
			return errIncorrectParameter
		}
	}
	return nil
}

func VMEvalIf(vm *script.VM, src string, state uint32, extend *map[string]interface{}) (bool, error) {
	return vm.EvalIf(src, state, extend)
}

func VMFlushBlock(vm *script.VM, root *script.Block) {
	vm.FlushBlock(root)
}

func vmExtend(vm *script.VM, ext *script.ExtendData) {
	vm.Extend(ext)
}

func VMRun(vm *script.VM, block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	var cost int64
	if ecost, ok := (*extend)[`txcost`]; ok {
		cost = ecost.(int64)
	} else {
		cost = syspar.GetMaxCost()
	}
	rt := vm.RunInit(cost)
	ret, err = rt.Run(block, params, extend)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("running block in smart vm")
	}
	(*extend)[`txcost`] = rt.Cost()
	return
}

func VMGetContract(vm *script.VM, name string, state uint32) *Contract {
	name = script.StateName(state, name)
	obj, ok := vm.Objects[name]

	if ok && obj.Type == script.ObjContract {
		return &Contract{Name: name, Block: obj.Value.(*script.Block)}
	}
	return nil
}

func VMObjectExists(vm *script.VM, name string, state uint32) bool {
	name = script.StateName(state, name)
	_, ok := vm.Objects[name]
	return ok
}

func vmGetUsedContracts(vm *script.VM, name string, state uint32, full bool) []string {
	contract := VMGetContract(vm, name, state)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Used == nil {
		return nil
	}
	ret := make([]string, 0)
	used := make(map[string]bool)
	for key := range contract.Block.Info.(*script.ContractInfo).Used {
		ret = append(ret, key)
		used[key] = true
		if full {
			sub := vmGetUsedContracts(vm, key, state, full)
			for _, item := range sub {
				if _, ok := used[item]; !ok {
					ret = append(ret, item)
					used[item] = true
				}
			}
		}
	}
	return ret
}

func VMGetContractByID(vm *script.VM, id int32) *Contract {
	idcont := id // - CNTOFF
	if len(vm.Children) <= int(idcont) {
		return nil
	}
	if vm.Children[idcont] == nil || vm.Children[idcont].Type != script.ObjContract {
		return nil
	}
	return &Contract{Name: vm.Children[idcont].Info.(*script.ContractInfo).Name,
		Block: vm.Children[idcont]}
}

func vmExtendCost(vm *script.VM, ext func(string) int64) {
	vm.ExtCost = ext
}

func vmFuncCallsDB(vm *script.VM, funcCallsDB map[string]struct{}) {
	vm.FuncCallsDB = funcCallsDB
}

// ExternOff switches off the extern compiling mode in smartVM
func ExternOff() {
	vmExternOff(smartVM)
}

// Compile compiles contract source code in smartVM
func Compile(src string, owner *script.OwnerInfo) error {
	return vmCompile(smartVM, src, owner)
}

// CompileBlock calls CompileBlock for smartVM
func CompileBlock(src string, owner *script.OwnerInfo) (*script.Block, error) {
	return VMCompileBlock(smartVM, src, owner)
}

// CompileEval calls CompileEval for smartVM
func CompileEval(src string, prefix uint32) error {
	return VMCompileEval(smartVM, src, prefix)
}

// EvalIf calls EvalIf for smartVM
func EvalIf(src string, state uint32, extend *map[string]interface{}) (bool, error) {
	return VMEvalIf(smartVM, src, state, extend)
}

// FlushBlock calls FlushBlock for smartVM
func FlushBlock(root *script.Block) {
	VMFlushBlock(smartVM, root)
}

// ExtendCost sets the cost of calling extended obj in smartVM
func ExtendCost(ext func(string) int64) {
	vmExtendCost(smartVM, ext)
}

func FuncCallsDB(funcCallsDB map[string]struct{}) {
	vmFuncCallsDB(smartVM, funcCallsDB)
}

// Extend set extended variable and functions in smartVM
func Extend(ext *script.ExtendData) {
	vmExtend(smartVM, ext)
}

// Run executes Block in smartVM
func Run(block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	return VMRun(smartVM, block, params, extend)
}

// ActivateContract sets Active status of the contract in smartVM
func ActivateContract(tblid, state int64, active bool) {
	for i, item := range smartVM.Block.Children {
		if item != nil && item.Type == script.ObjContract {
			cinfo := item.Info.(*script.ContractInfo)
			if cinfo.Owner.TableID == tblid && cinfo.Owner.StateID == uint32(state) {
				smartVM.Children[i].Info.(*script.ContractInfo).Owner.Active = active
			}
		}
	}
}

// SetContractWallet changes WalletID of the contract in smartVM
func SetContractWallet(sc *SmartContract, tblid, state int64, wallet int64) error {
	if err := validateAccess(`SetContractWallet`, sc, nEditContract); err != nil {
		return err
	}
	for i, item := range smartVM.Block.Children {
		if item != nil && item.Type == script.ObjContract {
			cinfo := item.Info.(*script.ContractInfo)
			if cinfo.Owner.TableID == tblid && cinfo.Owner.StateID == uint32(state) {
				smartVM.Children[i].Info.(*script.ContractInfo).Owner.WalletID = wallet
			}
		}
	}
	return nil
}

// GetContract returns true if the contract exists in smartVM
func GetContract(name string, state uint32) *Contract {
	return VMGetContract(smartVM, name, state)
}

// GetUsedContracts returns the list of contracts which are called from the specified contract
func GetUsedContracts(name string, state uint32, full bool) []string {
	return vmGetUsedContracts(smartVM, name, state, full)
}

// GetContractByID returns true if the contract exists
func GetContractByID(id int32) *Contract {
	return VMGetContractByID(smartVM, id)
}

// GetFunc returns the block of the specified function in the contract
func (contract *Contract) GetFunc(name string) *script.Block {
	if block, ok := (*contract).Block.Objects[name]; ok && block.Type == script.ObjFunc {
		return block.Value.(*script.Block)
	}
	return nil
}

// LoadContracts reads and compiles contracts from smart_contracts tables
func LoadContracts(transaction *model.DbTransaction) error {
	ecosystemsIds, err := model.GetAllSystemStatesIDs()
	if err != nil {
		return logErrorDB(err, "selecting ids from ecosystems")
	}

	defer ExternOff()

	for _, ecosystemID := range ecosystemsIds {
		prefix := strconv.FormatInt(ecosystemID, 10)
		if err := LoadContract(transaction, prefix); err != nil {
			return err
		}
	}

	ExternOff()
	return nil
}

func LoadSysFuncs(vm *script.VM, state int) error {
	code := `func DBFind(table string).Columns(columns string).Where(where map)
	.WhereId(id int).Order(order string).Limit(limit int).Offset(offset int) array {
   return DBSelect(table, columns, id, order, offset, limit, where)
}

func One(list array, name string) string {
   if list {
	   var row map 
	   row = list[0]
	   if Contains(name, "->") {
		   var colfield array
		   var val string
		   colfield = Split(ToLower(name), "->")
		   val = row[Join(colfield, ".")]
		   if !val && row[colfield[0]] {
			   var fields map
			   var i int
			   fields = JSONToMap(row[colfield[0]])
			   val = fields[colfield[1]]
			   i = 2
			   while i < Len(colfield) {
					if GetType(val) == "map[string]interface {}" {
						val = val[colfield[i]]
						if !val {
							break
						}
					  	i= i+1
				   	} else {
						break
				   	}
			   }
		   }
		   if !val {
			   return ""
		   }
		   return val
	   }
	   return row[name]
   }
   return nil
}

func Row(list array) map {
   var ret map
   if list {
	   ret = list[0]
   }
   return ret
}

func DBRow(table string).Columns(columns string).Where(where map)
   .WhereId(id int).Order(order string) map {
   
   var result array
   result = DBFind(table).Columns(columns).Where(where).WhereId(id).Order(order)

   var row map
   if Len(result) > 0 {
	   row = result[0]
   }

   return row
}

func ConditionById(table string, validate bool) {
   var row map
   row = DBRow(table).Columns("conditions").WhereId($Id)
   if !row["conditions"] {
	   error Sprintf("Item %d has not been found", $Id)
   }

   Eval(row["conditions"])

   if validate {
	   ValidateCondition($Conditions,$ecosystem_id)
   }
}`
	return vmCompile(vm, code, &script.OwnerInfo{StateID: uint32(state)})
}

// LoadContract reads and compiles contract of new state
func LoadContract(transaction *model.DbTransaction, prefix string) (err error) {
	var contracts []map[string]string
	contracts, err = model.GetAllTransaction(transaction,
		`select * from "1_contracts" where ecosystem = ? order by id`, -1, prefix)
	if err != nil {
		return logErrorDB(err, "selecting all transactions from contracts")
	}
	state := uint32(converter.StrToInt64(prefix))
	LoadSysFuncs(smartVM, int(state))
	for _, item := range contracts {
		list, err := script.ContractsList(item[`value`])
		if err != nil {
			return err
		}
		names := strings.Join(list, `,`)
		owner := script.OwnerInfo{
			StateID:  state,
			Active:   item[`active`] == `1`,
			TableID:  converter.StrToInt64(item[`id`]),
			WalletID: converter.StrToInt64(item[`wallet_id`]),
			TokenID:  converter.StrToInt64(item[`token_id`]),
		}
		if err = Compile(item[`value`], &owner); err != nil {
			logErrorValue(err, consts.EvalError, "Load Contract", names)
		}
	}

	return
}

func LoadVDEContracts(transaction *model.DbTransaction, prefix string) (err error) {
	var contracts []map[string]string

	contracts, err = model.GetAllTransaction(transaction,
		`select * from "1_contracts" where ecosystem=? order by id`, -1, prefix)
	if err != nil {
		return err
	}
	state := converter.StrToInt64(prefix)
	vm := GetVM()

	var vmt script.VMType
	if conf.Config.IsVDE() {
		vmt = script.VMTypeVDE
	} else if conf.Config.IsVDEMaster() {
		vmt = script.VMTypeVDEMaster
	}

	EmbedFuncs(vm, vmt)
	LoadSysFuncs(vm, int(state))
	for _, item := range contracts {
		list, err := script.ContractsList(item[`value`])
		if err != nil {
			return err
		}
		names := strings.Join(list, `,`)
		owner := script.OwnerInfo{
			StateID:  uint32(state),
			Active:   false,
			TableID:  converter.StrToInt64(item[`id`]),
			WalletID: 0,
			TokenID:  0,
		}

		if err = vmCompile(vm, item[`value`], &owner); err != nil {
			logErrorValue(err, consts.EvalError, "Load VDE Contract", names)
		}
	}

	return
}

func (sc *SmartContract) getExtend() *map[string]interface{} {
	var block, blockTime, blockKeyID, blockNodePosition int64

	head := sc.TxSmart
	keyID := int64(head.KeyID)
	if sc.BlockData != nil {
		block = sc.BlockData.BlockID
		blockKeyID = sc.BlockData.KeyID
		blockTime = sc.BlockData.Time
		blockNodePosition = sc.BlockData.NodePosition
	}
	extend := map[string]interface{}{
		`type`:              head.ID,
		`time`:              head.Time,
		`ecosystem_id`:      head.EcosystemID,
		`node_position`:     blockNodePosition,
		`block`:             block,
		`key_id`:            keyID,
		`block_key_id`:      blockKeyID,
		`parent`:            ``,
		`txcost`:            sc.GetContractLimit(),
		`txhash`:            sc.TxHash,
		`result`:            ``,
		`sc`:                sc,
		`contract`:          sc.TxContract,
		`block_time`:        blockTime,
		`original_contract`: ``,
		`this_contract`:     ``,
		`guest_key`:         vde.GuestKey,
	}

	for key, val := range sc.TxData {
		extend[key] = val
	}

	return &extend
}

func PrefixName(table string) (prefix, name string) {
	name = table
	if off := strings.IndexByte(table, '_'); off > 0 && table[0] >= '0' && table[0] <= '9' {
		prefix = table[:off]
		name = table[off+1:]
	}
	return
}

func (sc *SmartContract) IsCustomTable(table string) (isCustom bool, err error) {
	prefix, name := PrefixName(table)
	if len(prefix) > 0 {
		tables := &model.Table{}
		tables.SetTablePrefix(prefix)
		found, err := tables.Get(sc.DbTransaction, name)
		if err != nil {
			return false, err
		}
		if found {
			return true, nil
		}
	}
	return false, nil
}

// AccessTable checks the access right to the table
func (sc *SmartContract) AccessTablePerm(table, action string) (map[string]string, error) {
	var (
		err             error
		tablePermission map[string]string
	)
	logger := sc.GetLogger()
	isRead := action == `read`
	if GetTableName(sc, table) == `1_parameters` || GetTableName(sc, table) == `1_app_params` {
		if isRead || sc.TxSmart.KeyID == converter.StrToInt64(EcosysParam(sc, `founder_account`)) {
			return tablePermission, nil
		}
		logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
		return tablePermission, errAccessDenied
	}

	if isCustom, err := sc.IsCustomTable(table); err != nil {
		logger.WithFields(log.Fields{"table": table, "error": err, "type": consts.DBError}).Error("checking custom table")
		return tablePermission, err
	} else if !isCustom {
		if isRead {
			return tablePermission, nil
		}
		return tablePermission, fmt.Errorf(eNotCustomTable, table)
	}

	prefix, name := PrefixName(table)
	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	tablePermission, err = tables.GetPermissions(sc.DbTransaction, name, "")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table permissions")
		return tablePermission, err
	}
	if len(tablePermission[action]) > 0 {
		ret, err := sc.EvalIf(tablePermission[action])
		if err != nil {
			logger.WithFields(log.Fields{"table": table, "action": action, "permissions": tablePermission[action], "error": err, "type": consts.EvalError}).Error("evaluating table permissions for action")
			return tablePermission, err
		}
		if !ret {
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "type": consts.EvalError}).Error("access denied")
			return tablePermission, errAccessDenied
		}
	}
	return tablePermission, nil
}

func (sc *SmartContract) AccessTable(table, action string) error {
	if sc.FullAccess {
		return nil
	}
	_, err := sc.AccessTablePerm(table, action)
	return err
}

func getPermColumns(input string) (perm permColumn, err error) {
	if strings.HasPrefix(input, `{`) {
		err = unmarshalJSON([]byte(input), &perm, `on perm columns`)
	} else {
		perm.Update = input
	}
	return
}

// AccessColumns checks access rights to the columns
func (sc *SmartContract) AccessColumns(table string, columns *[]string, update bool) error {
	logger := sc.GetLogger()
	if sc.FullAccess {
		return nil
	}
	if GetTableName(sc, table) == `1_parameters` || GetTableName(sc, table) == `1_app_params` {
		if update {
			if sc.TxSmart.KeyID == converter.StrToInt64(EcosysParam(sc, `founder_account`)) {
				return nil
			}
			return errAccessDenied
		}
		return nil
	}
	prefix, name := PrefixName(table)
	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	found, err := tables.Get(sc.DbTransaction, name)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table columns")
		return err
	}
	if !found {
		if !update {
			return nil
		}
		return fmt.Errorf(eTableNotFound, table)
	}
	var cols map[string]string
	err = json.Unmarshal([]byte(tables.Columns), &cols)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("getting table columns")
		return err
	}
	colNames := make([]string, 0, len(*columns))
	for _, col := range *columns {
		if col == `*` {
			for column := range cols {
				colNames = append(colNames, column)
			}
			continue
		}
		colNames = append(colNames, col)
	}

	colList := make([]string, len(colNames))
	for i, col := range colNames {
		colname := converter.Sanitize(col, `->`)
		if strings.Contains(colname, `->`) {
			colname = colname[:strings.Index(colname, `->`)]
		}
		colList[i] = colname
	}
	checked := make(map[string]bool)
	var notaccess bool
	for i, name := range colList {
		if status, ok := checked[name]; ok {
			if !status {
				colList[i] = ``
			}
			continue
		}
		cond := cols[name]
		if len(cond) > 0 {
			perm, err := getPermColumns(cond)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.InvalidObject, "error": err}).Error("getting access columns")
				return err
			}
			if update {
				cond = perm.Update
			} else {
				cond = perm.Read
			}
			if len(cond) > 0 {
				ret, err := sc.EvalIf(cond)
				if err != nil {
					logger.WithFields(log.Fields{"condition": cond, "column": name,
						"type": consts.EvalError}).Error("evaluating condition")
					return err
				}
				checked[name] = ret
				if !ret {
					if update {
						return errAccessDenied
					}
					colList[i] = ``
					notaccess = true
				}
			}
		}
	}
	if !update && notaccess {
		retColumn := make([]string, 0)
		for i, val := range colList {
			if val != `` {
				retColumn = append(retColumn, colNames[i])
			}
		}
		if len(retColumn) == 0 {
			return errAccessDenied
		}
		*columns = retColumn
	}
	return nil
}

// AccessRights checks the access right by executing the condition value
func (sc *SmartContract) AccessRights(condition string, iscondition bool) error {
	sp := &model.StateParameter{}
	prefix := converter.Int64ToStr(sc.TxSmart.EcosystemID)

	sp.SetTablePrefix(prefix)
	_, err := sp.Get(sc.DbTransaction, condition)
	if err != nil {
		return err
	}
	conditions := sp.Value
	if iscondition {
		conditions = sp.Conditions
	}
	if len(conditions) > 0 {
		ret, err := sc.EvalIf(conditions)
		if err != nil {
			return err
		}
		if !ret {
			return errAccessDenied
		}
	} else {
		return fmt.Errorf(eNotCondition, condition)
	}
	return nil
}

// EvalIf counts and returns the logical value of the specified expression
func (sc *SmartContract) EvalIf(conditions string) (bool, error) {
	time := sc.TxSmart.Time
	blockTime := int64(0)
	if sc.BlockData != nil {
		blockTime = sc.BlockData.Time
	}
	return VMEvalIf(sc.VM, conditions, uint32(sc.TxSmart.EcosystemID), &map[string]interface{}{`ecosystem_id`: sc.TxSmart.EcosystemID,
		`key_id`: sc.TxSmart.KeyID, `sc`: sc, `original_contract`: ``, `this_contract`: ``,
		`block_time`: blockTime, `time`: time})
}

// GetContractLimit returns the default maximal cost of contract
func (sc *SmartContract) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if len(sc.TxSmart.MaxSum) > 0 {
		sc.TxCost = converter.StrToInt64(sc.TxSmart.MaxSum)
	} else {
		sc.TxCost = syspar.GetMaxCost()
	}
	return sc.TxCost
}

func (sc *SmartContract) payContract(fuelRate decimal.Decimal, payWallet *model.Key,
	fromID, toID int64) error {
	logger := sc.GetLogger()

	apl := sc.TxUsedCost.Mul(fuelRate)

	wltAmount, ierr := decimal.NewFromString(payWallet.Amount)
	if ierr != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": ierr, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
		return ierr
	}
	if wltAmount.Cmp(apl) < 0 {
		apl = wltAmount
	}

	commission := apl.Mul(decimal.New(syspar.SysInt64(`commission_size`), 0)).Div(decimal.New(100, 0)).Floor()
	walletTable := model.KeyTableName(sc.TxSmart.TokenEcosystem)
	comment := fmt.Sprintf("Commission for execution of %s contract", sc.TxContract.Name)
	fromIDString := converter.Int64ToStr(fromID)

	payCommission := func(toID string, sum decimal.Decimal) error {
		_, _, err := sc.update(
			[]string{"+amount"}, []interface{}{sum}, walletTable, "id", toID)
		if err != nil {
			return err
		}

		_, _, err = sc.insert(
			[]string{"sender_id", "recipient_id", "amount", "comment", "block_id", "txhash",
				"ecosystem"},
			[]interface{}{fromIDString, toID, sum, comment, sc.BlockData.BlockID, sc.TxHash, sc.TxSmart.TokenEcosystem},
			`1_history`)
		if err != nil {
			return err
		}

		return nil
	}

	if err := payCommission(converter.Int64ToStr(toID), apl.Sub(commission)); err != nil {
		if err != errUpdNotExistRecord {
			return err
		}
		apl = commission
	}

	if err := payCommission(syspar.GetCommissionWallet(sc.TxSmart.TokenEcosystem), commission); err != nil {
		if err != errUpdNotExistRecord {
			return err
		}
		apl = apl.Sub(commission)
	}

	if _, _, ierr := sc.update([]string{`-amount`}, []interface{}{apl}, walletTable, `id`,
		fromIDString); ierr != nil {
		return errCommission
	}
	return nil
}

func (sc *SmartContract) GetSignedBy(public []byte) (int64, error) {
	signedBy := sc.TxSmart.KeyID
	if sc.TxSmart.SignedBy != 0 {
		var isNode bool
		signedBy = sc.TxSmart.SignedBy
		fullNodes := syspar.GetNodes()
		if sc.TxContract.Name != CallDelayedContract && sc.TxContract.Name != NewUserContract &&
			sc.TxContract.Name != NewBadBlockContract {
			return 0, errDelayedContract
		}
		if len(fullNodes) > 0 {
			for _, node := range fullNodes {
				if crypto.Address(node.PublicKey) == signedBy {
					isNode = true
					break
				}
			}
		} else {
			_, NodePublicKey, err := utils.GetNodeKeys()
			if err != nil {
				return 0, err
			}
			isNode = PubToID(NodePublicKey) == signedBy
		}
		if !isNode {
			return 0, errDelayedContract
		}
	} else if len(public) > 0 && sc.TxSmart.KeyID != crypto.Address(public) {
		return 0, errDiffKeys
	}
	return signedBy, nil
}

// CallContract calls the contract functions according to the specified flags
func (sc *SmartContract) CallContract() (string, error) {
	var (
		result                        string
		err                           error
		public                        []byte
		sizeFuel, toID, fromID, price int64
		fuelRate                      decimal.Decimal
	)
	logger := sc.GetLogger()
	payWallet := &model.Key{}
	sc.TxContract.Extend = sc.getExtend()
	sc.TxSmart.TokenEcosystem = consts.TokenEcosystem

	retError := func(err error) (string, error) {
		eText := err.Error()
		if !strings.HasPrefix(eText, `{`) {
			err = script.SetVMError(`panic`, eText)
		}
		return ``, err
	}

	methods := []string{`conditions`, `action`}
	sc.AppendStack(sc.TxContract.Name)
	sc.VM = GetVM()

	if !sc.VDE {
		toID = sc.BlockData.KeyID
		fromID = sc.TxSmart.KeyID
	}
	if len(sc.TxSmart.PublicKey) > 0 && string(sc.TxSmart.PublicKey) != `null` {
		public = sc.TxSmart.PublicKey
	}
	wallet := &model.Key{}
	wallet.SetTablePrefix(sc.TxSmart.EcosystemID)
	signedBy, err := sc.GetSignedBy(public)
	if err != nil {
		return retError(err)
	}
	_, err = wallet.Get(signedBy)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
		return retError(err)
	}
	if wallet.Deleted == 1 {
		return retError(errDeletedKey)
	}
	if len(wallet.PublicKey) > 0 {
		public = wallet.PublicKey
	}
	if sc.TxSmart.ID == 258 { // UpdFullNodes
		node := syspar.GetNode(sc.TxSmart.KeyID)
		if node == nil {
			logger.WithFields(log.Fields{"user_id": sc.TxSmart.KeyID, "type": consts.NotFound}).Error("unknown node id")
			return retError(errUnknownNodeID)
		}
		public = node.PublicKey
	}
	if len(public) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty public key")
		return retError(errEmptyPublicKey)
	}
	sc.PublicKeys = append(sc.PublicKeys, public)

	var CheckSignResult bool
	CheckSignResult, err = utils.CheckSign(sc.PublicKeys, sc.TxHash, sc.TxSignature, false)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("checking tx data sign")
		return retError(err)
	}
	if !CheckSignResult {
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect sign")
		return retError(errIncorrectSign)
	}
	if sc.TxSmart.EcosystemID > 0 && !sc.VDE && !conf.Config.IsPrivateBlockchain() {
		fuelRate, err = decimal.NewFromString(syspar.GetFuelRate(sc.TxSmart.TokenEcosystem))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": sc.TxSmart.TokenEcosystem}).Error("converting ecosystem fuel rate from string to decimal")
			return retError(err)
		}
		if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
			logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("Fuel rate must be greater than 0")
			return retError(errFuelRate)
		}
		var payOver decimal.Decimal
		isActive := sc.TxContract.Block.Info.(*script.ContractInfo).Owner.Active
		if isActive {
			fromID = sc.TxContract.Block.Info.(*script.ContractInfo).Owner.WalletID
			sc.TxSmart.TokenEcosystem = sc.TxContract.Block.Info.(*script.ContractInfo).Owner.TokenID
		} else if len(sc.TxSmart.PayOver) > 0 {
			payOver, err = decimal.NewFromString(sc.TxSmart.PayOver)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": sc.TxSmart.TokenEcosystem}).Error("converting tx smart pay over from string to decimal")
				return retError(err)
			}
			fuelRate = fuelRate.Add(payOver)
		}
		payWallet.SetTablePrefix(sc.TxSmart.TokenEcosystem)
		if found, err := payWallet.Get(fromID); err != nil || !found {
			if !found {
				return retError(errCurrentBalance)
			}
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
			return retError(err)
		}
		if !isActive && !bytes.Equal(wallet.PublicKey, payWallet.PublicKey) && !bytes.Equal(sc.TxSmart.PublicKey, payWallet.PublicKey) && sc.TxSmart.SignedBy == 0 {
			return retError(errDiffKeys)
		}
		var amount, maxpay decimal.Decimal
		if len(payWallet.Amount) > 0 {
			amount, err = decimal.NewFromString(payWallet.Amount)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
				return retError(err)
			}
		} else {
			amount = decimal.New(0, 0)
		}
		if len(payWallet.Maxpay) > 0 {
			maxpay, err = decimal.NewFromString(payWallet.Maxpay)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": payWallet.Maxpay}).Error("converting pay wallet maxpay from string to decimal")
				return retError(err)
			}
		} else {
			maxpay = decimal.New(0, 0)
		}
		if maxpay.GreaterThan(decimal.New(0, 0)) && maxpay.LessThan(amount) {
			amount = maxpay
		}
		if priceName, ok := script.ContractPrices[sc.TxContract.Name]; ok {
			price = SysParamInt(priceName)
			if price > MaxPrice {
				return retError(errMaxPrice)
			}
			if price < 0 {
				return retError(errNegPrice)
			}
		}
		sizeFuel = syspar.GetSizeFuel() * sc.TxSize / 1024
		priceCost := decimal.New(price, 0)
		if amount.LessThanOrEqual(priceCost.Mul(fuelRate)) {
			logger.WithFields(log.Fields{"type": consts.NoFunds}).Error("current balance is not enough")
			return retError(errCurrentBalance)
		}
		maxCost := amount.Div(fuelRate).Floor()
		fullCost := decimal.New((*sc.TxContract.Extend)[`txcost`].(int64), 0).Add(priceCost)
		if maxCost.LessThan(fullCost) {
			(*sc.TxContract.Extend)[`txcost`] = converter.StrToInt64(maxCost.String()) - price
		}
	}

	before := (*sc.TxContract.Extend)[`txcost`].(int64)

	// Payment for the size
	(*sc.TxContract.Extend)[`txcost`] = (*sc.TxContract.Extend)[`txcost`].(int64) - sizeFuel
	if (*sc.TxContract.Extend)[`txcost`].(int64) <= 0 {
		logger.WithFields(log.Fields{"type": consts.NoFunds}).Error("current balance is not enough for payment")
		return retError(errCurrentBalance)
	}

	_, nameContract := converter.ParseName(sc.TxContract.Name)
	(*sc.TxContract.Extend)[`original_contract`] = nameContract
	(*sc.TxContract.Extend)[`this_contract`] = nameContract

	sc.TxContract.FreeRequest = false
	for i := uint32(0); i < 2; i++ {
		cfunc := sc.TxContract.GetFunc(methods[i])
		if cfunc == nil {
			continue
		}
		sc.TxContract.Called = 1 << i
		_, err = VMRun(sc.VM, cfunc, nil, sc.TxContract.Extend)
		if err != nil {
			price = 0
			break
		}
	}
	sc.TxFuel = before - (*sc.TxContract.Extend)[`txcost`].(int64)
	sc.TxUsedCost = decimal.New(sc.TxFuel+price, 0)
	if (*sc.TxContract.Extend)[`result`] != nil {
		result = fmt.Sprint((*sc.TxContract.Extend)[`result`])
		if len(result) > 255 {
			result = result[:255]
		}
	}

	if sc.TxSmart.EcosystemID > 0 && !sc.VDE && !conf.Config.IsPrivateBlockchain() {
		if ierr := sc.payContract(fuelRate, payWallet, fromID, toID); ierr != nil {
			err = ierr
		}
	}
	if err != nil {
		return retError(err)
	}
	return result, nil
}
