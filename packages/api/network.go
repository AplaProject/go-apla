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

package api

import (
	"net/http"
	"strconv"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
)

type FullNodeJSON struct {
	TCPAddress string `json:"tcp_address"`
	APIAddress string `json:"api_address"`
	KeyID      string `json:"key_id"`
	PublicKey  string `json:"public_key"`
	UnbanTime  string `json:"unban_time,er"`
	Stopped    bool   `json:"stopped"`
}

type NetworkResult struct {
	NetworkID     string         `json:"network_ud"`
	CentrifugoURL string         `json:"centrifugo_url"`
	Test          bool           `json:"test"`
	Private       bool           `json:"private"`
	FullNodes     []FullNodeJSON `json:"full_nodes"`
}

func GetNodesJSON() []FullNodeJSON {
	nodes := make([]FullNodeJSON, 0)
	for _, node := range syspar.GetNodes() {
		nodes = append(nodes, FullNodeJSON{
			TCPAddress: node.TCPAddress,
			APIAddress: node.APIAddress,
			KeyID:      strconv.FormatInt(node.KeyID, 10),
			PublicKey:  crypto.PubToHex(node.PublicKey),
			UnbanTime:  strconv.FormatInt(node.UnbanTime.Unix(), 10),
		})
	}
	return nodes
}

func getNetworkHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, &NetworkResult{
		NetworkID:     converter.Int64ToStr(conf.Config.NetworkID),
		CentrifugoURL: conf.Config.Centrifugo.URL,
		Test:          syspar.IsTestMode(),
		Private:       syspar.IsPrivateBlockchain(),
		FullNodes:     GetNodesJSON(),
	})
}
