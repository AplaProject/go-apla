// Copyright 2018 The go-daylight Authors
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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"

	log "github.com/sirupsen/logrus"
)

const (
	SysName = `@system`
)

type SysRollData struct {
	Type        string `json:"type,omitempty"`
	EcosystemID int64  `json:"ecosystem,omitempty"`
	ID          int64  `json:"id,omitempty"`
	Data        string `json:"data,omitempty"`
	TableName   string `json:"table,omitempty"`
}

func SysRollback(sc *SmartContract, data SysRollData) error {
	out, err := marshalJSON(data, `marshaling sys rollback`)
	if err != nil {
		return err
	}
	rollbackSys := &model.RollbackTx{
		BlockID:   sc.BlockData.BlockID,
		TxHash:    sc.TxHash,
		NameTable: SysName,
		TableID:   converter.Int64ToStr(sc.TxSmart.EcosystemID),
		Data:      string(out),
	}
	err = rollbackSys.Create(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating system  rollback")
		return err
	}
	return nil
}

// SysRollbackTable is rolling back table
func SysRollbackTable(DbTransaction *model.DbTransaction, sysData SysRollData) error {
	err := model.DropTable(DbTransaction, sysData.TableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
		return err
	}
	return nil
}

// SysRollbackColumn is rolling back column
func SysRollbackColumn(DbTransaction *model.DbTransaction, sysData SysRollData) error {
	return model.AlterTableDropColumn(DbTransaction, sysData.TableName, sysData.Data)
}

// SysRollbackContract performs rollback for the contract
func SysRollbackContract(name string, EcosystemID int64) error {
	vm := GetVM()
	if c := VMGetContract(vm, name, uint32(EcosystemID)); c != nil {
		id := c.Block.Info.(*script.ContractInfo).ID
		if int(id) < len(vm.Children) {
			vm.Children = vm.Children[:id]
		}
		delete(vm.Objects, c.Name)
	}

	return nil
}

func SysRollbackNewContract(sysData SysRollData, EcosystemID string) error {
	contractList, err := script.ContractsList(sysData.Data)
	if err != nil {
		return err
	}
	for _, contract := range contractList {
		if err := SysRollbackContract(contract, converter.StrToInt64(EcosystemID)); err != nil {
			return err
		}
	}
	return nil
}

// SysFlushContract is flushing contract
func SysFlushContract(iroot interface{}, id int64, active bool) error {
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
	VMFlushBlock(GetVM(), root)
	return nil
}

// SysSetContractWallet changes WalletID of the contract in smartVM
func SysSetContractWallet(tblid, state int64, wallet int64) error {
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

// SysRollbackEditContract rollbacks the contract
func SysRollbackEditContract(transaction *model.DbTransaction, sysData SysRollData,
	EcosystemID string) error {

	query := fmt.Sprintf(`select * from 1_contracts where id=?`, sysData.ID)
	fields, err := model.GetOneRowTransaction(transaction, query).String()
	if err != nil {
		return err
	}
	if len(fields["value"]) > 0 {
		var owner *script.OwnerInfo
		for i, item := range smartVM.Block.Children {
			if item != nil && item.Type == script.ObjContract {
				cinfo := item.Info.(*script.ContractInfo)
				if cinfo.Owner.TableID == sysData.ID &&
					cinfo.Owner.StateID == uint32(converter.StrToInt64(EcosystemID)) {
					owner = smartVM.Children[i].Info.(*script.ContractInfo).Owner
					break
				}
			}
		}
		if owner == nil {
			err = errContractNotFound
			log.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("getting existing contract")
			return err
		}
		wallet := owner.WalletID
		if len(fields["wallet_id"]) > 0 {
			wallet = converter.StrToInt64(fields["wallet_id"])
		}
		root, err := VMCompileBlock(GetVM(), fields["value"],
			&script.OwnerInfo{StateID: uint32(owner.StateID), WalletID: wallet, TokenID: owner.TokenID})
		if err != nil {
			log.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("compiling contract")
			return err
		}
		err = SysFlushContract(root, owner.TableID, owner.Active)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("flushing contract")
			return err
		}
	} else if len(fields["wallet_id"]) > 0 {
		return SysSetContractWallet(sysData.ID, converter.StrToInt64(EcosystemID),
			converter.StrToInt64(fields["wallet_id"]))
	}
	return nil
}

// SysRollbackEcosystem is rolling back ecosystem
func SysRollbackEcosystem(DbTransaction *model.DbTransaction, sysData SysRollData) error {
	tables := make([]string, 0)
	for table := range model.FirstEcosystemTables {
		tables = append(tables, table)
		err := model.Delete(DbTransaction, `1_`+table, fmt.Sprintf(`where ecosystem='%d'`, sysData.ID))
		if err != nil {
			return err
		}
	}
	if sysData.ID == 1 {
		tables = append(tables, `node_ban_logs`, `bad_blocks`, `system_parameters`, `ecosystems`)
		for _, name := range tables {
			err := model.DropTable(DbTransaction, fmt.Sprintf("%d_%s", sysData.ID, name))
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
				return err
			}
		}
	} else {
		if err := SysRollbackContract(`MainCondition`, sysData.ID); err != nil {
			return err
		}
	}
	return nil
}

// SysRollbackActivate sets Deactive status of the contract in smartVM
func SysRollbackActivate(sysData SysRollData) error {
	ActivateContract(sysData.ID, sysData.EcosystemID, false)
	return nil
}

// SysRollbackDeactivate sets Active status of the contract in smartVM
func SysRollbackDeactivate(sysData SysRollData) error {
	ActivateContract(sysData.ID, sysData.EcosystemID, true)
	return nil
}
