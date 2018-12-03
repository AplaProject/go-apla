// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
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

var (
	apiErrors = map[string]string{
		`E_CONTRACT`:        `There is not %s contract`,
		`E_DBNIL`:           `DB is nil`,
		`E_DELETEDKEY`:      `The key is deleted`,
		`E_ECOSYSTEM`:       `Ecosystem %d doesn't exist`,
		`E_EMPTYPUBLIC`:     `Public key is undefined`,
		`E_KEYNOTFOUND`:     `Key has not been found`,
		`E_EMPTYSIGN`:       `Signature is undefined`,
		`E_HASHWRONG`:       `Hash is incorrect`,
		`E_HASHNOTFOUND`:    `Hash has not been found`,
		`E_HEAVYPAGE`:       `This page is heavy`,
		`E_INSTALLED`:       `Apla is already installed`,
		`E_INVALIDWALLET`:   `Wallet %s is not valid`,
		`E_LIMITFORSIGN`:    `Length of forsign is too big (%d)`,
		`E_LIMITTXSIZE`:     `The size of tx is too big (%d)`,
		`E_NOTFOUND`:        `Page not found`,
		`E_NOTINSTALLED`:    `Apla is not installed`,
		`E_PARAMNOTFOUND`:   `Parameter %s has not been found`,
		`E_PERMISSION`:      `Permission denied`,
		`E_QUERY`:           `DB query is wrong`,
		`E_RECOVERED`:       `API recovered`,
		`E_SERVER`:          `Server error`,
		`E_SIGNATURE`:       `Signature is incorrect`,
		`E_UNKNOWNSIGN`:     `Unknown signature`,
		`E_STATELOGIN`:      `%s is not a membership of ecosystem %s`,
		`E_TABLENOTFOUND`:   `Table %s has not been found`,
		`E_TOKEN`:           `Token is not valid`,
		`E_TOKENEXPIRED`:    `Token is expired by %s`,
		`E_UNAUTHORIZED`:    `Unauthorized`,
		`E_UNDEFINEVAL`:     `Value %s is undefined`,
		`E_UNKNOWNUID`:      `Unknown uid`,
		`E_VDE`:             `Virtual Dedicated Ecosystem %d doesn't exist`,
		`E_VDECREATED`:      `Virtual Dedicated Ecosystem is already created`,
		`E_REQUESTNOTFOUND`: `Request %s doesn't exist`,
		`E_UPDATING`:        `Node is updating blockchain`,
		`E_STOPPING`:        `Network is stopping`,
		`E_NOTIMPLEMENTED`:  `Not implemented`,
	}
)
