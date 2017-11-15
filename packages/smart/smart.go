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
	"regexp"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/templatev2"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/jinzhu/gorm"
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
)

func testValue(name string, v ...interface{}) {
	smartTest[name] = fmt.Sprint(v...)
}

// GetTestValue returns the test value of the specified key
func GetTestValue(name string) string {
	return smartTest[name]
}

func (sc SmartContract) GetLogger() *log.Entry {
	return log.WithFields(log.Fields{"vde": sc.VDE, "name": sc.TxContract.Name})
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

/*func pref2state(prefix string) (state uint32) {
	if prefix != `global` {
		if val, err := strconv.ParseUint(prefix, 10, 32); err == nil {
			state = uint32(val)
		}
	}
	return
}*/

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

func ContractsList(value string) []string {
	list := make([]string, 0)
	re := regexp.MustCompile(`contract[\s]*([\d\w_]+)[\s]*{`)
	for _, item := range re.FindAllStringSubmatch(value, -1) {
		if len(item) > 1 {
			list = append(list, item[1])
		}
	}
	return list
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
		names := strings.Join(ContractsList(item[`value`]), `,`)
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
	EmbedFuncs(vm)
	smartVDE[state] = vm
	for _, item := range contracts {
		names := strings.Join(ContractsList(item[`value`]), `,`)
		owner := script.OwnerInfo{
			StateID:  uint32(state),
			Active:   false,
			TableID:  converter.StrToInt64(item[`id`]),
			WalletID: 0,
			TokenID:  0,
		}
		if err = vmCompile(vm, item[`value`], &owner); err != nil {
			log.Error("Load VDE Contract", names, err)
			fmt.Println("Error Load VDE Contract", names, err)
		} else {
			fmt.Println("OK Load VDE Contract", names, item[`id`])
		}
	}

	return
}

// GetContractLimit returns the default maximal cost of contract
func (sc *SmartContract) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if len(sc.TxSmart.MaxSum) > 0 {
		sc.TxCost = converter.StrToInt64(sc.TxSmart.MaxSum)
	} else {
		cost, _ := templatev2.StateParam(sc.TxSmart.EcosystemID, `max_sum`)
		if len(cost) > 0 {
			sc.TxCost = converter.StrToInt64(cost)
		}
	}
	if sc.TxCost == 0 {
		sc.TxCost = script.CostDefault // fuel
	}
	return sc.TxCost
}

func (sc *SmartContract) getExtend() *map[string]interface{} {
	head := sc.TxSmart
	var citizenID, walletID int64
	citizenID = int64(head.KeyID)
	walletID = int64(head.KeyID)
	block := int64(0)
	blockTime := int64(0)
	walletBlock := int64(0)
	if sc.BlockData != nil {
		block = sc.BlockData.BlockID
		walletBlock = sc.BlockData.KeyID
		blockTime = sc.BlockData.Time
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Time, `ecosystem_id`: head.EcosystemID,
		`block`: block, `citizen`: citizenID, `key_id`: walletID, `wallet_block`: walletBlock,
		`parent`: ``, `txcost`: sc.GetContractLimit(), `txhash`: sc.TxHash, `result`: ``,
		`parser`: sc, `sc`: sc, `contract`: sc.TxContract, `block_time`: blockTime}
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

// CallContract calls the contract functions according to the specified flags
func (sc *SmartContract) CallContract(flags int) (result string, err error) {
	var (
		public          []byte
		sizeFuel, price int64
	)
	sc.TxContract.Extend = sc.getExtend()

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	sc.TxContract.StackCont = []string{sc.TxContract.Name}
	(*sc.TxContract.Extend)[`stack_cont`] = StackCont
	sc.VM = GetVM(sc.VDE, sc.TxSmart.EcosystemID)
	if (flags&CallRollback) == 0 && (flags&CallAction) != 0 {
		// TODO: insert getting toID fromID from p.CallContract
		if len(sc.TxSmart.PublicKey) > 0 && string(sc.TxSmart.PublicKey) != `null` {
			public = sc.TxSmart.PublicKey
		}
		wallet := &model.Key{}
		wallet.SetTablePrefix(sc.TxSmart.EcosystemID)
		err = wallet.Get(sc.TxSmart.KeyID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return
		}
		if len(wallet.PublicKey) > 0 {
			public = wallet.PublicKey
		}
		if sc.TxSmart.Type == 258 { // UpdFullNodes
			node := syspar.GetNode(sc.TxSmart.KeyID)
			if node == nil {
				return ``, fmt.Errorf("unknown node id")
			}
			public = node.Public
		}
		if len(public) == 0 {
			return ``, fmt.Errorf("empty public key")
		}
		sc.PublicKeys = append(sc.PublicKeys, public)
		//		fmt.Println(`CALL CONTRACT`, sc.TxData[`forsign`].(string))
		CheckSignResult, err := utils.CheckSign(sc.PublicKeys, sc.TxData[`forsign`].(string), sc.TxSmart.BinSignatures, false)
		if err != nil {
			fmt.Println(`ForSign`, sc.TxData[`forsign`].(string))
			return ``, err
		}
		if !CheckSignResult {
			return ``, fmt.Errorf("incorrect sign")
		}
		// TODO: Insert calculating balance from p.CallContract	if sc.TxSmart.EcosystemID > 0
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

	resVal := (*sc.TxContract.Extend)[`result`]
	switch v := resVal.(type) {
	case int64:
		result = strconv.FormatInt(v, 10)
	case string:
		result = v
	default:
		err = fmt.Errorf("bad transaction result")
	}
	// TODO: Insert payment from p.CallContract	if (flags&CallAction) != 0 && sc.TxSmart.EcosystemID > 0 && !sc.VDE
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

func IsCustomTable(table string) (isCustom bool, err error) {
	prefix, name := PrefixName(table)
	if len(prefix) > 0 {
		tables := &model.Table{}
		tables.SetTablePrefix(prefix)
		found, err := tables.Get(name)
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
func (sc *SmartContract) AccessTable(table, action string) error {
	logger := sc.GetLogger()

	if table == getDefTableName(sc, `parameters`) {
		if sc.TxSmart.KeyID == converter.StrToInt64(EcosysParam(sc, `founder_account`)) {
			return nil
		}
		logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
		return fmt.Errorf(`Access denied`)
	}

	if isCustom, err := IsCustomTable(table); err != nil {
		logger.WithFields(log.Fields{"table": table, "error": err, "type": consts.DBError}).Error("checking custom table")
		return err
	} else if !isCustom {
		return fmt.Errorf(table + ` is not a custom table`)
	}

	prefix, name := PrefixName(table)
	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	tablePermission, err := tables.GetPermissions(name, "")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table permissions")
		return err
	}
	if len(tablePermission[action]) > 0 {
		ret, err := sc.EvalIf(tablePermission[action])
		if err != nil {
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "error": err, "type": consts.EvalError}).Error("evaluating table permissions for action")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "type": consts.EvalError}).Error("access denied")
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}

// AccessColumns checks access rights to the columns
func (sc *SmartContract) AccessColumns(table string, columns []string) error {
	logger := sc.GetLogger()

	if table == getDefTableName(sc, `parameters`) {
		if sc.TxSmart.KeyID == converter.StrToInt64(EcosysParam(sc, `founder_account`)) {
			return nil
		}
		return fmt.Errorf(`Access denied`)
	}
	// We don't check IsCustomTable because we calls it in AccessTable
	prefix, name := PrefixName(table)

	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	columnsAndPermissions, err := tables.GetColumns(name, "")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table columns")
		return err
	}

	for _, col := range columns {
		var (
			cond string
			ok   bool
		)
		cond, ok = columnsAndPermissions[converter.Sanitize(col, ``)]
		if !ok {
			cond, ok = columnsAndPermissions[`*`]
		}
		if ok && len(cond) > 0 {
			ret, err := sc.EvalIf(cond)
			if err != nil {
				logger.WithFields(log.Fields{"condition": cond, "column": col, "type": consts.EvalError}).Error("evaluating condition")
				return err
			}
			if !ret {
				return fmt.Errorf(`Access denied`)
			}
		}
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
			return fmt.Errorf(`Access denied`)
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
		`key_id`: sc.TxSmart.KeyID, `parser`: sc, `sc`: sc,
		`block_time`: blockTime, `time`: time})
}

func GetBytea(table string) map[string]bool {
	isBytea := make(map[string]bool)
	colTypes, err := model.GetAll(`select column_name, data_type from information_schema.columns where table_name=?`, -1, table)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		return isBytea
	}
	for _, icol := range colTypes {
		isBytea[icol[`column_name`]] = icol[`column_name`] != `conditions` && icol[`data_type`] == `bytea`
	}
	return isBytea
}

func (sc *SmartContract) selectiveLoggingAndUpd(fields []string, ivalues []interface{},
	table string, whereFields, whereValues []string, generalRollback bool) (int64, string, error) {
	var (
		tableID string
		err     error
		cost    int64
	)

	if generalRollback && sc.BlockData == nil {
		return 0, ``, fmt.Errorf(`It is impossible to write to DB when Block is undefined`)
	}

	isBytea := GetBytea(table)
	for i, v := range ivalues {
		if len(fields) > i && isBytea[fields[i]] {
			switch v.(type) {
			case string:
				if vbyte, err := hex.DecodeString(v.(string)); err == nil {
					ivalues[i] = vbyte
				}
			}
		}
	}

	values := converter.InterfaceSliceToStr(ivalues)

	addSQLFields := `id`
	if len(addSQLFields) > 0 {
		addSQLFields += `,`
	}
	for i, field := range fields {
		field = strings.TrimSpace(field)
		fields[i] = field
		if field[:1] == "+" || field[:1] == "-" {
			addSQLFields += field[1:len(field)] + ","
		} else if strings.HasPrefix(field, `timestamp `) {
			addSQLFields += field[len(`timestamp `):] + `,`
		} else {
			addSQLFields += field + ","
		}
	}
	addSQLWhere := ""
	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			if val := converter.StrToInt64(whereValues[i]); val != 0 {
				addSQLWhere += whereFields[i] + "= " + whereValues[i] + " AND "
			} else {
				addSQLWhere += whereFields[i] + "= '" + whereValues[i] + "' AND "
			}
		}
	}
	if len(addSQLWhere) > 0 {
		addSQLWhere = " WHERE " + addSQLWhere[0:len(addSQLWhere)-5]
	}
	if sc.VDE {
		addSQLFields = strings.TrimRight(addSQLFields, ",")
	} else {
		addSQLFields += `rb_id`
	}
	selectQuery := `SELECT ` + addSQLFields + ` FROM "` + table + `" ` + addSQLWhere
	//	fmt.Println(`Select`, selectQuery)
	selectCost, err := model.GetQueryTotalCost(selectQuery)
	if err != nil {
		return 0, tableID, err
	}
	logData, err := model.GetOneRowTransaction(sc.DbTransaction, selectQuery).String()
	if err != nil {
		return 0, tableID, err
	}
	cost += selectCost
	if whereFields != nil && len(logData) > 0 {
		/*	if whereFields != nil {
			if len(logData) == 0 {
				return tableID, fmt.Errorf(`update of the unknown record`)
			}*/
		jsonMap := make(map[string]string)
		for k, v := range logData {
			if k == `id` {
				continue
			}
			if (isBytea[k] || converter.InSliceString(k, []string{"hash", "tx_hash", "pub", "tx_hash", "public_key_0", "node_public_key"})) && v != "" {
				jsonMap[k] = string(converter.BinToHex([]byte(v)))
			} else {
				jsonMap[k] = v
			}
			if k == "rb_id" {
				k = "prev_rb_id"
			}
			if k[:1] == "+" || k[:1] == "-" {
				addSQLFields += k[1:len(k)] + ","
			} else if strings.HasPrefix(k, `timestamp `) {
				addSQLFields += k[len(`timestamp `):] + `,`
			} else {
				addSQLFields += k + ","
			}
		}
		jsonData, _ := json.Marshal(jsonMap)
		if err != nil {
			return 0, tableID, err
		}
		var rollback *model.Rollback
		if !sc.VDE {
			rollback = &model.Rollback{Data: string(jsonData), BlockID: sc.BlockData.BlockID}
			err = rollback.Create(sc.DbTransaction)
			if err != nil {
				return 0, tableID, err
			}
		}
		addSQLUpdate := ""
		for i := 0; i < len(fields); i++ {
			// utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLUpdate += fields[i] + `=decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if fields[i][:1] == "+" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `+` + values[i] + `,`
			} else if fields[i][:1] == "-" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `-` + values[i] + `,`
			} else if values[i] == `NULL` {
				addSQLUpdate += fields[i] + `= NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLUpdate += fields[i][len(`timestamp `):] + `= to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLUpdate += fields[i] + `= timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLUpdate += fields[i] + `='` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		if !sc.VDE {
			var updateQuery string
			updateQuery = `UPDATE "` + table + `" SET ` + addSQLUpdate + ` rb_id = ? ` + addSQLWhere
			//		fmt.Println(`Update`, updateQuery)
			updateCost, err := model.GetQueryTotalCost(updateQuery, rollback.RbID)
			if err != nil {
				return 0, tableID, err
			}
			cost += updateCost
			addSQLUpdate += fmt.Sprintf("rb_id = %d", rollback.RbID)
		} else {
			addSQLUpdate = strings.TrimRight(addSQLUpdate, `,`)
		}

		//		fmt.Println(`UPDATE`, table, addSQLUpdate, addSQLWhere)
		err = model.Update(sc.DbTransaction, table, addSQLUpdate, addSQLWhere)
		if err != nil {
			return 0, tableID, err
		}
		tableID = logData[`id`]
	} else {
		isID := false
		addSQLIns0 := ""
		addSQLIns1 := ""
		for i := 0; i < len(fields); i++ {
			if fields[i] == `id` {
				isID = true
			}
			if fields[i][:1] == "+" || fields[i][:1] == "-" {
				addSQLIns0 += fields[i][1:len(fields[i])] + `,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns0 += fields[i][len(`timestamp `):] + `,`
			} else {
				addSQLIns0 += fields[i] + `,`
			}
			// || utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key", "public_key_0", "node_public_key"}))
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLIns1 += `decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if values[i] == `NULL` {
				addSQLIns1 += `NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns1 += `to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLIns1 += `timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLIns1 += `'` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		if whereFields != nil && whereValues != nil {
			for i := 0; i < len(whereFields); i++ {
				if whereFields[i] == `id` {
					isID = true
				}
				addSQLIns0 += `` + whereFields[i] + `,`
				addSQLIns1 += `'` + whereValues[i] + `',`
			}
		}
		addSQLIns0 = addSQLIns0[0 : len(addSQLIns0)-1]
		addSQLIns1 = addSQLIns1[0 : len(addSQLIns1)-1]
		//		fmt.Println(`Sel Log`, "INSERT INTO "+table+" ("+addSQLIns0+") VALUES ("+addSQLIns1+")")
		if !isID {
			id, err := model.GetNextID(sc.DbTransaction, table)
			if err != nil {
				return 0, ``, err
			}
			tableID = converter.Int64ToStr(id)
			addSQLIns0 += `,id`
			addSQLIns1 += `,'` + tableID + `'`
		}

		insertQuery := `INSERT INTO "` + table + `" (` + addSQLIns0 + `) VALUES (` + addSQLIns1 + `)`

		insertCost, err := model.GetQueryTotalCost(insertQuery)
		if err != nil {
			return 0, tableID, err
		}
		cost += insertCost
		err = model.GetDB(sc.DbTransaction).Exec(insertQuery).Error
	}
	if err != nil {
		return 0, tableID, err
	}

	if generalRollback {
		rollbackTx := &model.RollbackTx{
			BlockID:   sc.BlockData.BlockID,
			TxHash:    sc.TxHash,
			NameTable: table,
			TableID:   tableID,
		}

		err = rollbackTx.Create(sc.DbTransaction)
		if err != nil {
			return 0, tableID, err
		}
	}
	return cost, tableID, nil
}
