// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
	test := syspar.SysString(syspar.Test)
	jsonResponse(w, &NetworkResult{
		NetworkID:     converter.Int64ToStr(conf.Config.NetworkID),
		CentrifugoURL: conf.Config.Centrifugo.URL,
		Test:          test != `0` && test != `false`,
		FullNodes:     GetNodesJSON(),
	})
}
