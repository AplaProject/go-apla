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

package parser

import (
	"encoding/json"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewStateParser struct {
	*Parser
	NewState *tx.NewState
}

func (p *NewStateParser) Init() error {
	newState := &tx.NewState{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newState); err != nil {
		return p.ErrInfo(err)
	}
	p.NewState = newState
	return nil
}

func (p *NewStateParser) Validate() error {
	err := p.generalCheck(`new_state`, &p.NewState.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string][]interface{}{"state_name": []interface{}{p.NewState.StateName}, "currency_name": []interface{}{p.NewState.CurrencyName}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewState.ForSign(), p.NewState.Header.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	country := string(p.NewState.StateName)
	if exist, err := IsState(p.DbTransaction, country); err != nil {
		return p.ErrInfo(err)
	} else if exist > 0 {
		return fmt.Errorf(`State %s already exists`, country)
	}

	return nil
}

func (p *NewStateParser) Main(country, currency string) (id string, err error) {
	systemState := &model.SystemState{}
	_, err = systemState.GetLast(p.DbTransaction)
	if err != nil {
		return
	}

	systemState.ID++
	systemState.RbID = 0
	err = systemState.Create(p.DbTransaction)
	if err != nil {
		return
	}
	id = converter.Int64ToStr(systemState.ID)
	rollbackTx := model.RollbackTx{BlockID: p.BlockData.BlockID, TxHash: p.TxHash, NameTable: "system_states", TableID: id}
	err = rollbackTx.Create(p.DbTransaction)
	if err != nil {
		return
	}
	err = model.CreateStateTable(p.DbTransaction, id)
	if err != nil {
		return
	}
	sid := `ContractConditions("MainCondition")` //`$citizen == ` + utils.Int64ToStr(p.TxWalletID) // id + `_citizens.id=` + utils.Int64ToStr(p.TxWalletID)
	psid := sid                                  //fmt.Sprintf(`Eval(StateParam(%s, "main_conditions"))`, id) //id+`_state_parameters.main_conditions`
	err = model.CreateStateConditions(p.DbTransaction, id, sid, psid, currency, country, p.TxWalletID)
	if err != nil {
		return
	}
	err = model.CreateSmartContractTable(p.DbTransaction, id)
	if err != nil {
		return
	}
	sc := &model.SmartContract{
		Name: "MainCondition",
		Value: []byte(`contract MainCondition {
			data {}
			conditions {
			    if(StateVal("gov_account")!=$citizen)
			    {
				warning "Sorry, you don't have access to this action."
			    }
		        }
			action {}
		}`),
		WalletID: p.TxWalletID,
		Active:   "1"}
	sc.SetTablePrefix(id)
	err = sc.Create(p.DbTransaction)
	if err != nil {
		return
	}
	scu := &model.SmartContract{}
	scu.SetTablePrefix(id)
	err = scu.UpdateConditions(p.DbTransaction, sid)
	if err != nil {
		return
	}

	err = model.CreateStateTablesTable(p.DbTransaction, id)
	if err != nil {
		return
	}
	mainCondition := `ContractConditions("MainCondition")`
	updateConditions := map[string]string{"public_key_0": mainCondition}
	perm := Permissions{
		GeneralUpdate: mainCondition,
		Update:        updateConditions,
		Insert:        mainCondition,
		NewColumn:     mainCondition,
	}
	jsonPermissions, err := json.Marshal(perm)
	if err != nil {
		return
	}
	t := &model.Table{
		Name: id + "_citizens",
		ColumnsAndPermissions: string(jsonPermissions),
		Conditions:            psid,
	}
	t.SetTablePrefix(id)
	err = t.Create(p.DbTransaction)
	if err != nil {
		return
	}

	err = model.CreateStatePagesTable(p.DbTransaction, id)
	if err != nil {
		log.Errorf("can't create state tables: %s", err)
		return
	}
	dashboardValue := `FullScreen(1)
	If(StateVal(type_office))
	Else:
	Title : Basic Apps
	Divs: col-md-4
			Divs: panel panel-default elastic
				Divs: panel-body text-center fill-area flexbox-item-grow
					Divs: flexbox-item-grow flex-center
						Divs: pv-lg
						Image("/static/img/apps/money.png", Basic, center-block img-responsive img-circle img-thumbnail thumb96 )
						DivsEnd:
						P(h4,Basic Apps)
						P(text-left,"Election and Assign, Polling, Messenger, Simple Money System")
					DivsEnd:
				DivsEnd:
				Divs: panel-footer
					Divs: clearfix
						Divs: pull-right
							BtnPage(app-basic, Install,'',btn btn-primary lang)
						DivsEnd:
					DivsEnd:
				DivsEnd:
			DivsEnd:
		DivsEnd:
	IfEnd:
	PageEnd:
`
	governmentValue := `FullScreen(1)
If(StateVal(type_office))
Else:
Title : Basic Apps
Divs: col-md-4
		Divs: panel panel-default elastic
			Divs: panel-body text-center fill-area flexbox-item-grow
				Divs: flexbox-item-grow flex-center
					Divs: pv-lg
					Image("/static/img/apps/money.png", Basic, center-block img-responsive img-circle img-thumbnail thumb96 )
					DivsEnd:
					P(h4,Basic Apps)
					P(text-left,"Election and Assign, Polling, Messenger, Simple Money System")
				DivsEnd:
			DivsEnd:
			Divs: panel-footer
				Divs: clearfix
					Divs: pull-right
						BtnPage(app-basic, Install,'',btn btn-primary lang)
					DivsEnd:
				DivsEnd:
			DivsEnd:
		DivsEnd:
	DivsEnd:
IfEnd:
PageEnd:
`
	firstPage := &model.Page{
		Name:       "dashboard_default",
		Value:      dashboardValue,
		Menu:       "menu_default",
		Conditions: sid,
	}
	firstPage.SetTablePrefix(id)
	err = firstPage.Create(p.DbTransaction)
	if err != nil {
		return
	}
	secondPage := &model.Page{
		Name:       "government",
		Value:      governmentValue,
		Menu:       "government",
		Conditions: sid,
	}
	secondPage.SetTablePrefix(id)
	err = secondPage.Create(p.DbTransaction)
	if err != nil {
		return
	}

	err = model.CreateStateMenuTable(p.DbTransaction, id)
	if err != nil {
		return
	}
	firstMenu := &model.Menu{
		Name: "menu_default",
		Value: `MenuItem(Dashboard, dashboard_default)
 MenuItem(Government dashboard, government)`,
		Conditions: sid,
	}
	firstMenu.SetTablePrefix(id)
	err = firstMenu.Create(p.DbTransaction)
	if err != nil {
		return
	}
	secondMenu := &model.Menu{
		Name: `government`,
		Value: `MenuItem(Citizen dashboard, dashboard_default)
MenuItem(Government dashboard, government)
MenuGroup(Admin tools,admin)
MenuItem(Tables,sys-listOfTables)
MenuItem(Smart contracts, sys-contracts)
MenuItem(Interface, sys-interface)
MenuItem(App List, sys-app_catalog)
MenuItem(Export, sys-export_tpl)
MenuItem(Wallet,  sys-edit_wallet)
MenuItem(Languages, sys-languages)
MenuItem(Signatures, sys-signatures)
MenuItem(Gen Keys, sys-gen_keys)
MenuEnd:
MenuBack(Welcome)`,
		Conditions: sid,
	}
	secondMenu.SetTablePrefix(id)
	err = secondMenu.Create(p.DbTransaction)
	if err != nil {
		return
	}

	err = model.CreateCitizensStateTable(p.DbTransaction, id)
	if err != nil {
		return
	}

	dltWallet := &model.DltWallet{}
	err = dltWallet.GetWalletTransaction(p.DbTransaction, p.TxWalletID)
	if err != nil {
		return
	}

	citizen := &model.Citizen{ID: p.TxWalletID, PublicKey: dltWallet.PublicKey}
	citizen.SetTablePrefix(id)
	err = citizen.Create(p.DbTransaction)
	if err != nil {
		return
	}
	err = model.CreateLanguagesStateTable(p.DbTransaction, id)
	if err != nil {
		return
	}
	err = model.CreateStateDefaultLanguages(p.DbTransaction, id, sid)
	if err != nil {
		return
	}

	err = model.CreateSignaturesStateTable(p.DbTransaction, id)
	if err != nil {
		return
	}

	err = model.CreateStateAppsTable(p.DbTransaction, id)
	if err != nil {
		return
	}

	err = model.CreateStateAnonymsTable(p.DbTransaction, id)
	if err != nil {
		return
	}

	err = template.LoadContract(p.DbTransaction, id)
	return
}

func (p *NewStateParser) Action() error {
	country := string(p.NewState.StateName)
	currency := string(p.NewState.CurrencyName)
	_, err := p.Main(country, currency)
	if err != nil {
		return p.ErrInfo(err)
	}
	dltWallet := &model.DltWallet{}
	err = dltWallet.GetWalletTransaction(p.DbTransaction, p.TxWalletID)
	if err != nil {
		return p.ErrInfo(err)
	} else if len(p.NewState.Header.PublicKey) > 30 && len(dltWallet.PublicKey) == 0 {
		_, _, err = p.selectiveLoggingAndUpd([]string{"public_key_0"}, []interface{}{converter.HexToBin(p.NewState.Header.PublicKey)}, "dlt_wallets",
			[]string{"wallet_id"}, []string{converter.Int64ToStr(p.TxWalletID)}, true)
	}
	return err
}

func (p *NewStateParser) Rollback() error {
	rollbackTx := &model.RollbackTx{}
	err := rollbackTx.Get(p.DbTransaction, p.TxHash, "system_states")
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.autoRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	for _, name := range []string{`menu`, `pages`, `citizens`, `languages`, `signatures`, `tables`,
		`smart_contracts`, `state_parameters`, `apps`, `anonyms`} {
		err = model.DropTable(p.DbTransaction, fmt.Sprintf("%s_%s", rollbackTx.TableID, name))
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	rollbackTxToDel := &model.RollbackTx{TxHash: p.TxHash, NameTable: "system_states"}
	err = rollbackTxToDel.DeleteByHashAndTableName(p.DbTransaction)
	if err != nil {
		return p.ErrInfo(err)
	}

	ssToDel := &model.SystemState{}
	_, err = ssToDel.GetLast(p.DbTransaction)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = ssToDel.Delete(p.DbTransaction)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p NewStateParser) Header() *tx.Header {
	return &p.NewState.Header
}
