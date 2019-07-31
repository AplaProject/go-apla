// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package transaction

import (
	"bytes"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

const (
	errUnknownContract = `Cannot find %s contract`
)

func CreateContract(contractName string, keyID int64, params map[string]interface{},
	privateKey []byte) error {
	ecosysID, _ := converter.ParseName(contractName)
	if ecosysID == 0 {
		ecosysID = 1
	}
	contract := smart.GetContract(contractName, uint32(ecosysID))
	if contract == nil {
		return fmt.Errorf(errUnknownContract, contractName)
	}
	sc := tx.SmartContract{
		Header: tx.Header{
			ID:          int(contract.Block.Info.(*script.ContractInfo).ID),
			Time:        time.Now().Unix(),
			EcosystemID: ecosysID,
			KeyID:       keyID,
			NetworkID:   conf.Config.NetworkID,
		},
		Params: params,
	}
	txData, _, err := tx.NewTransaction(sc, privateKey)
	if err == nil {
		rtx := &RawTransaction{}
		if err = rtx.Unmarshall(bytes.NewBuffer(txData)); err == nil {
			err = model.SendTx(rtx, sc.KeyID)
		}
	}
	return err
}
