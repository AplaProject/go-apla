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
	"reflect"
	"strconv"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
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
	TxPrice       int64   // custom price for citizens
	TxGovAccount  int64   // state wallet
	EGSRate       float64 // money/EGS rate
	TableAccounts string
	StackCont     []string // Stack of called contracts
	Extend        *map[string]interface{}
	Block         *script.Block
}

const (
	// CallInit is a flag for calling init function of the contract
	CallInit = 0x01
	// CallCondition is a flag for calling condition function of the contract
	CallCondition = 0x02
	// CallAction is a flag for calling action function of the contract
	CallAction = 0x04
	// CallRollback is a flag for calling rollback function of the contract
	CallRollback = 0x08

	// MaxPrice is a maximal value that price function can return
	MaxPrice = 100000000000000000
)

var (
	smartVM   *script.VM
	smartVDE  map[int64]*script.VM
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
	smartVDE = make(map[int64]*script.VM)
}

// GetVM is returning smart vm
func GetVM(vde bool, ecosystemID int64) *script.VM {
	if vde {
		if v, ok := smartVDE[ecosystemID]; ok {
			return v
		}
		return nil
	}
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

func VMCompileEval(vm *script.VM, src string, prefix uint32) error {
	return vm.CompileEval(src, prefix)
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
	var extcost int64
	cost := script.CostDefault
	if ecost, ok := (*extend)[`txcost`]; ok {
		cost = ecost.(int64)
	}
	rt := vm.RunInit(cost)
	ret, err = rt.Run(block, params, extend)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("running block in smart vm")
	}
	if ecost, ok := (*extend)[`txcost`]; ok && cost > ecost.(int64) {
		extcost = cost - ecost.(int64)
	}
	(*extend)[`txcost`] = rt.Cost() - extcost
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
	if len(vm.Children) <= int(idcont) || vm.Children[idcont].Type != script.ObjContract {
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
	if sc.TxContract.Name != `@1EditContract` {
		log.WithFields(log.Fields{"type": consts.IncorrectCallingContract}).Error("SetContractWallet can be only called from @1EditContract")
		return fmt.Errorf(`SetContractWallet can be only called from @1EditContract`)
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting ids from ecosystems")
		return err
	}

	defer ExternOff()
	if err := LoadContract(transaction, "system"); err != nil {
		return err
	}

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
	code := `func DBFind(table string).Columns(columns string).Where(where string, params ...)
	.WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
   return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
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

func DBRow(table string).Columns(columns string).Where(where string, params ...)
   .WhereId(id int).Order(order string).Ecosystem(ecosystem int) map {
   
   var result array
   result = DBFind(table).Columns(columns).Where(where, params...).WhereId(id).Order(order).Ecosystem(ecosystem)

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
	contracts, err = model.GetAllTransaction(transaction, `select * from "`+prefix+`_contracts" order by id`, -1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting all transactions from contracts")
		return err
	}
	state := uint32(converter.StrToInt64(prefix))
	LoadSysFuncs(smartVM, int(state))
	for _, item := range contracts {
		list, err := script.ContractsList(item[`value`])
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Getting ContractsList")
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
			log.WithFields(log.Fields{"type": consts.EvalError, "names": names, "error": err}).Error("Load Contract")
		} else {
			log.WithFields(log.Fields{"contract_name": names, "contract_id": item["id"], "contract_active": item["active"]}).Info("OK Loading Contract")
		}
	}
	LoadVDEContracts(transaction, prefix)
	return
}

func LoadVDEContracts(transaction *model.DbTransaction, prefix string) (err error) {
	var contracts []map[string]string

	if !model.IsTable(prefix + `_vde_contracts`) {
		return
	}
	contracts, err = model.GetAllTransaction(transaction, `select * from "`+prefix+`_vde_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	state := converter.StrToInt64(prefix)
	vm := newVM()
	EmbedFuncs(vm, script.VMTypeVDE)
	smartVDE[state] = vm
	LoadSysFuncs(vm, int(state))
	for _, item := range contracts {
		list, err := script.ContractsList(item[`value`])
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Getting ContractsList")
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
			log.WithFields(log.Fields{"names": names, "error": err}).Error("Load VDE Contract")
		} else {
			log.WithFields(log.Fields{"names": names, "contract_id": item["id"]}).Info("OK Load VDE Contract")
		}
	}

	return
}

func (sc *SmartContract) getExtend() *map[string]interface{} {
	var block, blockTime, blockKeyID int64

	head := sc.TxSmart
	keyID := int64(head.KeyID)
	if sc.BlockData != nil {
		block = sc.BlockData.BlockID
		blockKeyID = sc.BlockData.KeyID
		blockTime = sc.BlockData.Time
	}
	extend := map[string]interface{}{
		`type`:              head.Type,
		`time`:              head.Time,
		`ecosystem_id`:      head.EcosystemID,
		`node_position`:     head.NodePosition,
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
		`role_id`:           head.RoleID,
	}

	for key, val := range sc.TxData {
		extend[key] = val
	}

	return &extend
}

// StackCont adds an element to the stack of contract call or removes the top element when name is empty
func StackCont(sc interface{}, name string) {
	cont := sc.(*SmartContract).TxContract
	if len(name) > 0 {
		cont.StackCont = append(cont.StackCont, name)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	return
}

func PrefixName(table string) (prefix, name string) {
	name = table
	if off := strings.IndexByte(table, '_'); off > 0 && table[0] >= '0' && table[0] <= '9' {
		prefix = table[:off]
		if strings.HasPrefix(table[off+1:], `vde_`) {
			prefix += `_vde`
			off += 4
		}
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

	if table == getDefTableName(sc, `parameters`) || table == getDefTableName(sc, `app_params`) {
		if sc.TxSmart.KeyID == converter.StrToInt64(EcosysParam(sc, `founder_account`)) {
			return tablePermission, nil
		}
		logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
		return tablePermission, errAccessDenied
	}

	if isCustom, err := sc.IsCustomTable(table); err != nil {
		logger.WithFields(log.Fields{"table": table, "error": err, "type": consts.DBError}).Error("checking custom table")
		return tablePermission, err
	} else if !isCustom {
		return tablePermission, fmt.Errorf(table + ` is not a custom table`)
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
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "error": err, "type": consts.EvalError}).Error("evaluating table permissions for action")
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
	_, err := sc.AccessTablePerm(table, action)
	return err
}

func getPermColumns(input string) (perm permColumn, err error) {
	if strings.HasPrefix(input, `{`) {
		if err = json.Unmarshal([]byte(input), &perm); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "source": input}).Error("on perm columns")
		}
	} else {
		perm.Update = input
	}
	return
}

type colAccess struct {
	ok       bool
	original string
}

// AccessColumns checks access rights to the columns
func (sc *SmartContract) AccessColumns(table string, columns *[]string, update bool) error {
	logger := sc.GetLogger()
	if table == getDefTableName(sc, `parameters`) || table == getDefTableName(sc, `app_params`) {
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
		return fmt.Errorf(eTableNotFound, table)
	}
	var cols map[string]string
	// Every item of checkColumns has 'ok' boolean value. If it equals false then the key-column
	// doesn't have read/update access rights.
	checkColumns := make(map[string]colAccess)
	err = json.Unmarshal([]byte(tables.Columns), &cols)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("getting table columns")
		return err
	}
	for _, col := range *columns {
		colname := converter.Sanitize(col, `*->`)
		if strings.Contains(colname, `->`) {
			colname = colname[:strings.Index(colname, `->`)]
		}
		if !update && colname == `*` {
			for column := range cols {
				checkColumns[column] = colAccess{true, column}
			}
			break
		}
		checkColumns[colname] = colAccess{true, colname}
	}
	_, isall := checkColumns[`*`]
	for column, cond := range cols {
		if ca, ok := checkColumns[column]; (!ok || !ca.ok) && !isall {
			continue
		}
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
				logger.WithFields(log.Fields{"condition": cond, "column": column,
					"type": consts.EvalError}).Error("evaluating condition")
				return err
			}
			if !ret {
				if update {
					return errAccessDenied
				}
				checkColumns[column] = colAccess{false, ``}
			}
		}
	}
	if !update {
		retColumn := make([]string, 0)
		for key, val := range checkColumns {
			if val.ok && key != `*` {
				retColumn = append(retColumn, val.original)
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
	if sc.VDE {
		prefix += `_vde`
	}

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
		return fmt.Errorf(`There is not %s in parameters`, condition)
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

func GetBytea(db *model.DbTransaction, table string) map[string]bool {
	isBytea := make(map[string]bool)
	colTypes, err := model.GetAllTx(db, `select column_name, data_type from information_schema.columns where table_name=?`, -1, table)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		return isBytea
	}
	for _, icol := range colTypes {
		isBytea[icol[`column_name`]] = icol[`column_name`] != `conditions` && icol[`data_type`] == `bytea`
	}
	return isBytea
}

// GetContractLimit returns the default maximal cost of contract
func (sc *SmartContract) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if !sc.VDE {
		if len(sc.TxSmart.MaxSum) > 0 {
			sc.TxCost = converter.StrToInt64(sc.TxSmart.MaxSum)
		} else {
			cost := EcosysParam(sc, `max_sum`)
			if len(cost) > 0 {
				sc.TxCost = converter.StrToInt64(cost)
			}
		}
	}
	if sc.TxCost == 0 {
		sc.TxCost = script.CostDefault // fuel
	}
	return sc.TxCost
}

// CallContract calls the contract functions according to the specified flags
func (sc *SmartContract) CallContract(flags int) (string, error) {
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

	retError := func(err error) (string, error) {
		eText := err.Error()
		if !strings.HasPrefix(eText, `{`) {
			err = script.SetVMError(`panic`, eText)
		}
		return ``, err
	}

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	sc.TxContract.StackCont = []string{sc.TxContract.Name}
	(*sc.TxContract.Extend)[`stack_cont`] = StackCont
	sc.VM = GetVM(sc.VDE, sc.TxSmart.EcosystemID)
	if (flags&CallRollback) == 0 && (flags&CallAction) != 0 {
		if !sc.VDE {
			toID = sc.BlockData.KeyID
			fromID = sc.TxSmart.KeyID
		}
		if len(sc.TxSmart.PublicKey) > 0 && string(sc.TxSmart.PublicKey) != `null` {
			public = sc.TxSmart.PublicKey
		}
		wallet := &model.Key{}
		wallet.SetTablePrefix(sc.TxSmart.EcosystemID)
		signedBy := sc.TxSmart.KeyID
		if sc.TxSmart.SignedBy != 0 {
			signedBy = sc.TxSmart.SignedBy
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
		if sc.TxSmart.Type == 258 { // UpdFullNodes
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
		CheckSignResult, err = utils.CheckSign(sc.PublicKeys, sc.TxData[`forsign`].(string), sc.TxSmart.BinSignatures, false)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("checking tx data sign")
			return retError(err)
		}
		if !CheckSignResult {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect sign")
			return retError(errIncorrectSign)
		}
		if sc.TxSmart.EcosystemID > 0 && !sc.VDE && !conf.Config.PrivateBlockchain {
			if sc.TxSmart.TokenEcosystem == 0 {
				sc.TxSmart.TokenEcosystem = 1
			}
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
			if len(sc.TxSmart.PayOver) > 0 {
				payOver, err = decimal.NewFromString(sc.TxSmart.PayOver)
				if err != nil {
					log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": sc.TxSmart.TokenEcosystem}).Error("converting tx smart pay over from string to decimal")
					return retError(err)
				}
				fuelRate = fuelRate.Add(payOver)
			}
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
			var amount decimal.Decimal
			amount, err = decimal.NewFromString(payWallet.Amount)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
				return retError(err)
			}
			if cprice := sc.TxContract.GetFunc(`price`); cprice != nil {
				var ret []interface{}
				if ret, err = VMRun(sc.VM, cprice, nil, sc.TxContract.Extend); err != nil {
					return retError(err)
				} else if len(ret) == 1 {
					switch reflect.TypeOf(ret[0]).String() {
					case `int64`:
						price = ret[0].(int64)
						if price > MaxPrice {
							return retError(errMaxPrice)
						}
						if price < 0 {
							return retError(errNegPrice)
						}
					case script.Decimal:
						if ret[0].(decimal.Decimal).GreaterThan(decimal.New(MaxPrice, 0)) {
							return retError(errMaxPrice)
						}
						if ret[0].(decimal.Decimal).LessThan(decimal.New(0, 0)) {
							return retError(errNegPrice)
						}
						price = converter.StrToInt64(ret[0].(decimal.Decimal).String())
					default:
						logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong result type of price function")
						return retError(errWrongPriceFunc)
					}
				} else {
					logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong type of price function")
					return retError(errWrongPriceFunc)
				}
			}
			sizeFuel = syspar.GetSizeFuel() * int64(len(sc.TxSmart.Data)) / 1024
			if amount.Cmp(decimal.New(sizeFuel+price, 0).Mul(fuelRate)) <= 0 {
				logger.WithFields(log.Fields{"type": consts.NoFunds}).Error("current balance is not enough")
				return retError(errCurrentBalance)
			}
		}
	}
	before := (*sc.TxContract.Extend)[`txcost`].(int64) + price

	// Payment for the size
	(*sc.TxContract.Extend)[`txcost`] = (*sc.TxContract.Extend)[`txcost`].(int64) - sizeFuel

	_, nameContract := script.ParseContract(sc.TxContract.Name)
	(*sc.TxContract.Extend)[`original_contract`] = nameContract
	(*sc.TxContract.Extend)[`this_contract`] = nameContract

	sc.TxContract.FreeRequest = false
	for i := uint32(0); i < 4; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := sc.TxContract.GetFunc(methods[i])
			if cfunc == nil {
				continue
			}
			sc.TxContract.Called = 1 << i
			_, err = VMRun(sc.VM, cfunc, nil, sc.TxContract.Extend)
			if err != nil {
				before -= price
				break
			}
		}
	}
	sc.TxFuel = before - (*sc.TxContract.Extend)[`txcost`].(int64) - price
	sc.TxUsedCost = decimal.New(before-(*sc.TxContract.Extend)[`txcost`].(int64), 0)
	sc.TxContract.TxPrice = price
	if (*sc.TxContract.Extend)[`result`] != nil {
		result = fmt.Sprint((*sc.TxContract.Extend)[`result`])
		if len(result) > 255 {
			result = result[:255]
		}
	}

	if (flags&CallAction) != 0 && sc.TxSmart.EcosystemID > 0 && !sc.VDE && !conf.Config.PrivateBlockchain {
		apl := sc.TxUsedCost.Mul(fuelRate)
		wltAmount, ierr := decimal.NewFromString(payWallet.Amount)
		if ierr != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": ierr, "value": payWallet.Amount}).Error("converting pay wallet amount from string to decimal")
			return retError(ierr)
		}
		if wltAmount.Cmp(apl) < 0 {
			apl = wltAmount
		}

		commission := apl.Mul(decimal.New(syspar.SysInt64(`commission_size`), 0)).Div(decimal.New(100, 0)).Floor()
		walletTable := model.KeyTableName(sc.TxSmart.TokenEcosystem)
		historyTable := model.HistoryTableName(sc.TxSmart.TokenEcosystem)
		comment := fmt.Sprintf("Commission for execution of %s contract", sc.TxContract.Name)
		fromIDString := converter.Int64ToStr(fromID)

		payCommission := func(toID string, sum decimal.Decimal) error {
			_, _, err := sc.selectiveLoggingAndUpd(
				[]string{"+amount"}, []interface{}{sum}, walletTable,
				[]string{"id"}, []string{toID},
				true, true,
			)
			if err != nil {
				return err
			}

			_, _, err = sc.selectiveLoggingAndUpd(
				[]string{"sender_id", "recipient_id", "amount", "comment", "block_id", "txhash"},
				[]interface{}{fromIDString, toID, sum, comment, sc.BlockData.BlockID, sc.TxHash},
				historyTable, nil, nil, true, false,
			)
			if err != nil {
				return err
			}

			return nil
		}

		if err := payCommission(converter.Int64ToStr(toID), apl.Sub(commission)); err != nil {
			if err != errUpdNotExistRecord {
				return retError(err)
			}
			apl = commission
		}

		if err := payCommission(syspar.GetCommissionWallet(sc.TxSmart.TokenEcosystem), commission); err != nil {
			if err != errUpdNotExistRecord {
				return retError(err)
			}
			apl = apl.Sub(commission)
		}

		if _, _, ierr := sc.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{apl}, walletTable, []string{`id`},
			[]string{fromIDString}, true, true); ierr != nil {
			return retError(ierr)
		}
		logger.WithFields(log.Fields{"commission": commission}).Debug("Paid commission")
	}
	if err != nil {
		return retError(err)
	}
	return result, nil
}
