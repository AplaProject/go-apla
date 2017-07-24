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

package controllers

import (
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type signaturesPage struct {
	Lang       map[string]string
	WalletID   int64
	CitizenID  int64
	Signatures []map[string]string
	Global     string
}

// Signatures shows the list of the additional signatures
func (c *Controller) Signatures() (string, error) {

	var err error

	global := c.r.FormValue("global")
	prefix := "global"
	if global == "" || global == "0" {
		prefix = c.StateIDStr
		global = "0"
	}

	signature := &model.Signatures{}
	rows, err := signature.GetAllOredered(prefix)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	signatures := make([]map[string]string, 0)
	for _, sign := range rows {
		signatures := append(signatures, sign.ToMap())
	}

	TemplateStr, err := makeTemplate("signatures_list", "signatures_list", &signaturesPage{
		Lang:       c.Lang,
		WalletID:   c.SessWalletID,
		CitizenID:  c.SessCitizenID,
		Signatures: signatures,
		Global:     global})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
