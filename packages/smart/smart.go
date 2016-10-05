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
	"fmt"
	//	"reflect"
	//	"strings"

	//"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/script"
	//"github.com/DayLightProject/go-daylight/packages/utils"
)

type Contract struct {
	Name   string
	Called uint32
	//	parser *Parser //interface{}
	Block *script.Block
}

const (
	CALL_INIT  = 0x01
	CALL_FRONT = 0x02
	CALL_MAIN  = 0x04

	CNTOFF = 128 // offset of contract id
)

var (
	smartVM *script.VM
)

func init() {
	smartVM = script.VMInit(map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
		/*		"DBInsert":   DBInsert,
				"Balance":    Balance,
				"StateParam": StateParam,*/
	}, map[string]string{
		`*parser.Parser`: `parser`,
	})

	contract := `
contract TXCitizenRequest {
	tx {
		PublicKey  bytes
		StateId    int
		FirstName  string
		MiddleName string "optional"
		LastName   string
	}
	func init {
//		Println("TXCitizenRequest init" + $FirstName, $citizen, "/", $wallet,"=", Balance($wallet))
	}
	func front {
//		Println("TXCitizenRequest front" + $MiddleName, StateParam($StateId, "citizenship_price"))
		if 10000 {
			error "not enough money"
		}
	}
	func main {
		
//		Println("TXCitizenRequest main" + $LastName)
	}

}

contract TXNewCitizen {
			func front {
//				Println("NewCitizen Front", $citizen, $state, $PublicKey )
			}
			func main {
//				Println("NewCitizen Main", $type, $citizen, $block )
//				DBInsert(Sprintf( "%d_citizens", $state), "public_key,block_id", $PublicKey, $block)
			}
}`
	if err := Compile(contract); err != nil {
		fmt.Println(`SMART ERROR`, err)
	}
}

// Compiles contract source code
func Compile(src string) error {
	return smartVM.Compile([]rune(src))
}

// Returns true if the contract exists
func GetContract(name string /*, data interface{}*/) *Contract {
	obj, ok := smartVM.Objects[name]
	//	fmt.Println(`Get`, ok, obj, obj.Type, script.OBJ_CONTRACT)
	if ok && obj.Type == script.OBJ_CONTRACT {
		return &Contract{Name: name /*parser: p,*/, Block: obj.Value.(*script.Block)}
	}
	return nil
}

// Returns true if the contract exists
func GetContractById(id int32 /*, p *Parser*/) *Contract {
	idcont := id - CNTOFF
	if len(smartVM.Children) <= int(idcont) || smartVM.Children[idcont].Type != script.OBJ_CONTRACT {
		return nil
	}
	return &Contract{Name: smartVM.Children[idcont].Info.(*script.ContractInfo).Name,
		/*parser: p,*/ Block: smartVM.Children[idcont]}
}

func (contract *Contract) getFunc(name string) *script.Block {
	if block, ok := (*contract).Block.Objects[name]; ok && block.Type == script.OBJ_FUNC {
		return block.Value.(*script.Block)
	}
	return nil
}

func (contract *Contract) getExtend() *map[string]interface{} {
	/*	head := contract.parser.TxPtr.(*consts.TXHeader) //consts.HeaderNew(contract.parser.TxPtr)
		var citizenId, walletId int64
		if head.StateId > 0 {
			citizenId = head.UserId
		} else {
			walletId = head.UserId
		}
		block := int64(0)
		if contract.parser.BlockData != nil {
			block = contract.parser.BlockData.BlockId
		}
		extend := map[string]interface{}{`type`: head.Type, `time`: head.Type, `state`: head.StateId,
			`block`: block, `citizen`: citizenId, `wallet`: walletId,
			`parser`: contract.parser}*/
	extend := map[string]interface{}{}
	/*	for key, val := range contract.parser.TxData {
		extend[key] = val
	}*/
	/*	v := reflect.ValueOf(contract.parser.TxPtr).Elem()
		t := v.Type()
		for i := 1; i < t.NumField(); i++ {
			extend[t.Field(i).Name] = v.Field(i).Interface()
		}*/
	//fmt.Println(`Extend`, extend)
	return &extend
}

func (contract *Contract) Call(flags int) (err error) {
	methods := []string{`init`, `front`, `main`}
	extend := contract.getExtend()
	for i := uint32(0); i < 3; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := contract.getFunc(methods[i])
			if cfunc == nil {
				continue
			}
			rt := smartVM.RunInit()
			contract.Called = 1 << i
			_, err = rt.Run(cfunc, nil, extend)
			if err != nil {
				return
			}
		}
	}
	return
}

// Pre-defined functions
/*
func DBInsert(p *Parser, tblname string, params string, val ...interface{}) (err error) { // map[string]interface{}) {
	fmt.Println(`DBInsert`, tblname, params, val, len(val))
	err = p.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, true)
	return
}

func Balance(wallet_id int64) (float64, error) {
	return utils.DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Float64()
}

func StateParam(idstate int64, name string) (string, error) {
	return utils.DB.Single(`SELECT value FROM `+utils.Int64ToStr(idstate)+`_state_parameters WHERE name = ?`,
		name).String()
}


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
