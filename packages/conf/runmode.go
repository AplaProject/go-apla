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

package conf

type RunMode string

// VDEManager const label for running mode
const vdeMaster RunMode = "VDEMaster"

// VDE const label for running mode
const vde RunMode = "VDE"

// VDE const label for running mode
const node RunMode = "NONE"

// IsVDEMaster returns true if mode equal vdeMaster
func (rm RunMode) IsVDEMaster() bool {
	return rm == vdeMaster
}

// IsVDE returns true if mode equal vde
func (rm RunMode) IsVDE() bool {
	return rm == vde
}

// IsNode returns true if mode not equal to any VDE
func (rm RunMode) IsNode() bool {
	return rm == node
}

// IsSupportingVDE returns true if mode support vde
func (rm RunMode) IsSupportingVDE() bool {
	return rm.IsVDE() || rm.IsVDEMaster()
}
