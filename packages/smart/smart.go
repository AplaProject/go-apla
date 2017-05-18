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
	"fmt"
	"strconv"
	"strings"

	//"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	//"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

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
	CALL_INIT     = 0x01
	CALL_FRONT    = 0x02
	CALL_MAIN     = 0x04
	CALL_ROLLBACK = 0x08
)

var (
	smartVM *script.VM
)

func init() {
	smartVM = script.NewVM()
	smartVM.Extern = true
	smartVM.Extend(&script.ExtendData{map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
		"TxJson":  TxJson,
		"Float":   Float,
		"Money":   script.ValueToDecimal,
	}, map[string]string{
		`*smart.Contract`: `contract`,
	}})
}

func Pref2state(prefix string) (state uint32) {
	if prefix != `global` {
		val, _ := strconv.ParseUint(prefix, 10, 32)
		state = uint32(val)
	}
	return
}

func ExternOff() {
	smartVM.FlushExtern()
}

// Compiles contract source code
func Compile(src, prefix string, active bool, tblid int64) error {
	return smartVM.Compile([]rune(src), Pref2state(prefix), active, tblid)
}

func CompileBlock(src, prefix string, active bool, tblid int64) (*script.Block, error) {
	return smartVM.CompileBlock([]rune(src), Pref2state(prefix), active, tblid)
}

func CompileEval(src string, prefix uint32) error {
	return smartVM.CompileEval(src, prefix)
}

func EvalIf(src, prefix string, extend *map[string]interface{}) (bool, error) {
	return smartVM.EvalIf(src, Pref2state(prefix), extend)
}

func FlushBlock(root *script.Block) {
	smartVM.FlushBlock(root)
}

func ExtendCost(ext func(string) int64) {
	smartVM.ExtCost = ext
}

func Extend(ext *script.ExtendData) {
	smartVM.Extend(ext)
}

func Run(block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	var extcost int64
	cost := script.CostDefault
	if ecost, ok := (*extend)[`txcost`]; ok {
		cost = ecost.(int64)
	}
	rt := smartVM.RunInit(cost)
	ret, err = rt.Run(block, params, extend)
	if ecost, ok := (*extend)[`txcost`]; ok && cost > ecost.(int64) {
		extcost = cost - ecost.(int64)
	}
	(*extend)[`txcost`] = rt.Cost() - extcost
	return
}

func ActivateContract(tblid int64, prefix string, active bool) {
	if prefix == `global` {
		prefix = `0`
	}
	for i, item := range smartVM.Block.Children {
		if item != nil && item.Type == script.ObjContract {
			cinfo := item.Info.(*script.ContractInfo)
			if cinfo.TableID == tblid && strings.HasPrefix(cinfo.Name, `@`+prefix) &&
				(len(cinfo.Name) > len(prefix)+1 && cinfo.Name[len(prefix)+1] > '9') {
				smartVM.Children[i].Info.(*script.ContractInfo).Active = active
			}
		}
	}
}

// Returns true if the contract exists
func GetContract(name string, state uint32 /*, data interface{}*/) *Contract {
	name = script.StateName(state, name)
	obj, ok := smartVM.Objects[name]
	//	fmt.Println(`Get`, ok, obj, obj.Type, script.ObjContract)
	if ok && obj.Type == script.ObjContract {
		return &Contract{Name: name, Block: obj.Value.(*script.Block)}
	}
	return nil
}

func GetUsedContracts(name string, state uint32, full bool) []string {
	contract := GetContract(name, state)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Used == nil {
		return nil
	}
	ret := make([]string, 0)
	used := make(map[string]bool)
	for key := range contract.Block.Info.(*script.ContractInfo).Used {
		ret = append(ret, key)
		used[key] = true
		if full {
			sub := GetUsedContracts(key, state, full)
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

// Returns true if the contract exists
func GetContractById(id int32 /*, p *Parser*/) *Contract {
	idcont := id // - CNTOFF
	if len(smartVM.Children) <= int(idcont) || smartVM.Children[idcont].Type != script.ObjContract {
		return nil
	}
	return &Contract{Name: smartVM.Children[idcont].Info.(*script.ContractInfo).Name,
		/*parser: p,*/ Block: smartVM.Children[idcont]}
}

func (contract *Contract) GetFunc(name string) *script.Block {
	if block, ok := (*contract).Block.Objects[name]; ok && block.Type == script.ObjFunc {
		return block.Value.(*script.Block)
	}
	return nil
}

func TxJson(contract *Contract) string {
	lines := make([]string, 0)
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		switch fitem.Type.String() {
		case `string`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `int64`:
			lines = append(lines, fmt.Sprintf(`"%s": %d`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `[]uint8`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name,
				hex.EncodeToString((*(*contract).Extend)[fitem.Name].([]byte))))
		}
	}
	return `{` + strings.Join(lines, ",\r\n") + `}`
}

func Float(v interface{}) (ret float64) {
	switch value := v.(type) {
	case int64:
		ret = float64(value)
	case string:
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			ret = val
		}
	}
	return
}

// Pre-defined functions
/*
func CheckAmount() {
	amount, err := p.Single(`SELECT value FROM `+utils.Int64ToStr().TxVars[`state_code`]+`_state_parameters WHERE name = ?`, "citizenship_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	amountAndCommission, err := p.checkSenderDLT(amount, consts.COMMISSION)
	if err != nil {
		return p.ErrInfo(err)
	}
	if amount > amountAndCommission {
		return p.ErrInfo("incorrect amount")
	}
	// вычитаем из wallets_buffer
	// amount_and_commission взято из check_sender_money()
	err = p.updateWalletsBuffer(amountAndCommission)
	if err != nil {
		return p.ErrInfo(err)
	}

}
*/
