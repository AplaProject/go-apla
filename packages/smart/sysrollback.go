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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"

	log "github.com/sirupsen/logrus"
)

func SysRollback(sc *SmartContract, data string) error {
	rollbackSys := &model.RollbackTx{
		BlockID:   sc.BlockData.BlockID,
		TxHash:    sc.TxHash,
		NameTable: `@system`,
		TableID:   converter.Int64ToStr(sc.TxSmart.EcosystemID),
		Data:      data,
	}
	err := rollbackSys.Create(sc.DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating system  rollback")
		return err
	}
	return nil
}

// SysRollbackTable is rolling back table
func SysRollbackTable(DbTransaction *model.DbTransaction, TxHash []byte,
	TableName, EcosystemID string) error {
	err := model.DropTable(DbTransaction, TableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
		return err
	}
	t := model.Table{}
	t.SetTablePrefix(EcosystemID)
	if strings.HasPrefix(TableName, EcosystemID+`_`) {
		TableName = TableName[len(EcosystemID)+1:]
	}
	found, err := t.Get(DbTransaction, TableName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table info")
		return err
	}
	if found {
		err = t.Delete(DbTransaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting table")
			return err
		}
	} else {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("not found table info")
	}
	return nil
}

// SysRollbackColumn is rolling back column
func SysRollbackColumn(DbTransaction *model.DbTransaction, TxHash []byte,
	TableName, Name, EcosystemID string) error {
	Name = converter.EscapeSQL(strings.ToLower(Name))
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(DbTransaction, TxHash, fmt.Sprintf("%s_tables", EcosystemID))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column from rollback table")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("column record in rollback table")
		// if there is not such hash then NewColumn was faulty. Do nothing.
		return nil
	}
	return model.AlterTableDropColumn(TableName, Name)
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

func SysRollbackNewContract(value, EcosystemID string) error {
	contractList, err := script.ContractsList(value)
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

// RollbackEditContract rollbacks the contract
func SysRollbackEditContract(DbTransaction *model.DbTransaction, TxHash []byte,
	EcosystemID string) error {
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(DbTransaction, TxHash, fmt.Sprintf("%s_contracts", EcosystemID))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting contract from rollback table")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("contract record in rollback table")
		// if there is not such hash then EditContract was faulty. Do nothing.
		return nil
	}
	var fields map[string]string
	err = json.Unmarshal([]byte(rollbackTx.Data), &fields)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling contract values")
		return err
	}
	if len(fields["value"]) > 0 {
		var owner *script.OwnerInfo
		for i, item := range smartVM.Block.Children {
			if item != nil && item.Type == script.ObjContract {
				cinfo := item.Info.(*script.ContractInfo)
				if cinfo.Owner.TableID == converter.StrToInt64(rollbackTx.TableID) &&
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
		return SysSetContractWallet(converter.StrToInt64(rollbackTx.TableID),
			converter.StrToInt64(EcosystemID), converter.StrToInt64(fields["wallet_id"]))
	}
	return nil
}

// SysRollbackEcosystem is rolling back ecosystem
func SysRollbackEcosystem(DbTransaction *model.DbTransaction, TxHash []byte) error {
	rollbackTx := &model.RollbackTx{}
	found, err := rollbackTx.Get(DbTransaction, TxHash, "1_ecosystems")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback tx")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("system states in rollback table")
		// if there is not such hash then NewEcosystem was faulty. Do nothing.
		return nil
	}
	lastID, err := model.GetNextID(DbTransaction, "1_ecosystems")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id")
		return err
	}
	lastID--
	if converter.StrToInt64(rollbackTx.TableID) != lastID {
		log.WithFields(log.Fields{"table_id": rollbackTx.TableID, "last_id": lastID, "type": consts.InvalidObject}).Error("incorrect ecosystem id")
		return fmt.Errorf(`Incorrect ecosystem id %s != %d`, rollbackTx.TableID, lastID)
	}

	if model.IsTable(fmt.Sprintf(`%s_vde_tables`, rollbackTx.TableID)) {
		// Drop all _local_ tables
		table := &model.Table{}
		prefix := fmt.Sprintf(`%s_vde`, rollbackTx.TableID)
		table.SetTablePrefix(prefix)
		list, err := table.GetAll(prefix)
		if err != nil {
			return err
		}
		for _, item := range list {
			err = model.DropTable(DbTransaction, fmt.Sprintf("%s_%s", prefix, item.Name))
			if err != nil {
				return err
			}
		}
		for _, name := range []string{`tables`, `parameters`} {
			err = model.DropTable(DbTransaction, fmt.Sprintf("%s_%s", prefix, name))
			if err != nil {
				return err
			}
		}
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
		`buffer_data`,
	}

	if rollbackTx.TableID == "1" {
		rbTables = append(rbTables, `node_ban_logs`, `bad_blocks`, `system_parameters`, `ecosystems`)
	}

	for _, name := range rbTables {
		err = model.DropTable(DbTransaction, fmt.Sprintf("%s_%s", rollbackTx.TableID, name))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping table")
			return err
		}
	}
	rollbackTxToDel := &model.RollbackTx{TxHash: TxHash, NameTable: "1_ecosystems"}
	err = rollbackTxToDel.DeleteByHashAndTableName(DbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback tx by hash and table name")
		return err
	}

	ecosysToDel := &model.Ecosystem{ID: lastID}
	return ecosysToDel.Delete(DbTransaction)
}

// SysRollbackActivate sets Deactive status of the contract in smartVM
func SysRollbackActivate(tblid, state string) error {
	ActivateContract(converter.StrToInt64(tblid), converter.StrToInt64(state), false)
	return nil
}

// SysRollbackDeactivate sets Active status of the contract in smartVM
func SysRollbackDeactivate(tblid, state string) error {
	ActivateContract(converter.StrToInt64(tblid), converter.StrToInt64(state), true)
	return nil
}
