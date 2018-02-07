// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package smart

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
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
)

var (
	smartVM   *script.VM
	smartVDE  map[int64]*script.VM
	smartTest = make(map[string]string)

	ErrCurrentBalance = errors.New(`current balance is not enough`)
	ErrDiffKeys       = errors.New(`Contract and user public keys are different`)
	ErrEmptyPublicKey = errors.New(`empty public key`)
	ErrFounderAccount = errors.New(`Unknown founder account`)
	ErrFuelRate       = errors.New(`Fuel rate must be greater than 0`)
	ErrIncorrectSign  = errors.New(`incorrect sign`)
	ErrInvalidValue   = errors.New(`Invalid value`)
	ErrUnknownNodeID  = errors.New(`Unknown node id`)
	ErrWrongPriceFunc = errors.New(`Wrong type of price function`)
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
func LoadContracts(transaction *model.DbTransaction) (err error) {
	var states []map[string]string
	var prefix []string
	prefix = []string{`system`}
	states, err = model.GetAll(`select id from system_states order by id`, -1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting ids from system_states")
		return err
	}
	for _, istate := range states {
		prefix = append(prefix, istate[`id`])
	}
	for _, ipref := range prefix {
		if err = LoadContract(transaction, ipref); err != nil {
			break
		}
	}
	ExternOff()
	return
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
	for _, item := range contracts {
		names := strings.Join(script.ContractsList(item[`value`]), `,`)
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
	for _, item := range contracts {
		names := strings.Join(script.ContractsList(item[`value`]), `,`)
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
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Time, `ecosystem_id`: head.EcosystemID,
		`node_position`: head.NodePosition,
		`block`:         block, `key_id`: keyID, `block_key_id`: blockKeyID,
		`parent`: ``, `txcost`: sc.GetContractLimit(), `txhash`: sc.TxHash, `result`: ``,
		`sc`: sc, `contract`: sc.TxContract, `block_time`: blockTime}
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

	if table == getDefTableName(sc, `parameters`) {
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
		err = json.Unmarshal([]byte(input), &perm)
	} else {
		perm.Update = input
	}
	return
}

// AccessColumns checks access rights to the columns
func (sc *SmartContract) AccessColumns(table string, columns *[]string, update bool) error {
	logger := sc.GetLogger()
	if table == getDefTableName(sc, `parameters`) {
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
	hcolumns := make(map[string]bool)
	err = json.Unmarshal([]byte(tables.Columns), &cols)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("getting table columns")
		return err
	}
	for _, col := range *columns {
		colname := converter.Sanitize(col, `*`)
		if !update && colname == `*` {
			for column := range cols {
				hcolumns[column] = true
			}
			break
		}
		hcolumns[colname] = true
	}
	for column, cond := range cols {
		if !hcolumns[column] && !hcolumns[`*`] {
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
				hcolumns[column] = false
			}
		}
	}
	if !update {
		retColumn := make([]string, 0)
		for key, val := range hcolumns {
			if val && key != `*` {
				retColumn = append(retColumn, key)
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
		`key_id`: sc.TxSmart.KeyID, `sc`: sc,
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
		if len(wallet.PublicKey) > 0 {
			public = wallet.PublicKey
		}
		if sc.TxSmart.Type == 258 { // UpdFullNodes
			node := syspar.GetNode(sc.TxSmart.KeyID)
			if node == nil {
				logger.WithFields(log.Fields{"user_id": sc.TxSmart.KeyID, "type": consts.NotFound}).Error("unknown node id")
				return retError(ErrUnknownNodeID)
			}
			public = node.Public
		}
		if len(public) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty public key")
			return retError(ErrEmptyPublicKey)
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
			return retError(ErrIncorrectSign)
		}
		if sc.TxSmart.EcosystemID > 0 && !sc.VDE && !*utils.PrivateBlockchain {
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
				return retError(ErrFuelRate)
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
					return retError(ErrCurrentBalance)
				}
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
				return retError(err)
			}
			if !isActive && !bytes.Equal(wallet.PublicKey, payWallet.PublicKey) && !bytes.Equal(sc.TxSmart.PublicKey, payWallet.PublicKey) && sc.TxSmart.SignedBy == 0 {
				return retError(ErrDiffKeys)
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
					if _, ok := ret[0].(int64); !ok {
						logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong result type of price function")
						return retError(ErrWrongPriceFunc)
					}
					price = ret[0].(int64)
				} else {
					logger.WithFields(log.Fields{"type": consts.TypeError}).Error("Wrong type of price function")
					return retError(ErrWrongPriceFunc)
				}
			}
			sizeFuel = syspar.GetSizeFuel() * int64(len(sc.TxSmart.Data)) / 1024
			if amount.Cmp(decimal.New(sizeFuel+price, 0).Mul(fuelRate)) <= 0 {
				logger.WithFields(log.Fields{"type": consts.NoFunds}).Error("current balance is not enough")
				return retError(ErrCurrentBalance)
			}
		}
	}
	before := (*sc.TxContract.Extend)[`txcost`].(int64) + price

	// Payment for the size
	(*sc.TxContract.Extend)[`txcost`] = (*sc.TxContract.Extend)[`txcost`].(int64) - sizeFuel

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
	sc.TxUsedCost = decimal.New(before-(*sc.TxContract.Extend)[`txcost`].(int64), 0)
	sc.TxContract.TxPrice = price
	if (*sc.TxContract.Extend)[`result`] != nil {
		result = fmt.Sprint((*sc.TxContract.Extend)[`result`])
		if len(result) > 255 {
			result = result[:255]
		}
	}

	if (flags&CallAction) != 0 && sc.TxSmart.EcosystemID > 0 && !sc.VDE && !*utils.PrivateBlockchain {
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
		walletTable := fmt.Sprintf(`%d_keys`, sc.TxSmart.TokenEcosystem)
		if _, _, ierr := sc.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{apl.Sub(commission)}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(toID)}, true, true); ierr != nil {
			if ierr != errUpdNotExistRecord {
				return retError(ierr)
			}
			apl = commission
		}
		if _, _, ierr := sc.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{commission}, walletTable, []string{`id`},
			[]string{syspar.GetCommissionWallet(sc.TxSmart.TokenEcosystem)}, true, true); ierr != nil {
			if ierr != errUpdNotExistRecord {
				return retError(ierr)
			}
			apl = apl.Sub(commission)
		}
		if _, _, ierr := sc.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{apl}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(fromID)}, true, true); ierr != nil {
			return retError(ierr)
		}
		logger.WithFields(log.Fields{"commission": commission}).Debug("Paid commission")
	}
	if err != nil {
		return retError(err)
	}
	return result, nil
}
