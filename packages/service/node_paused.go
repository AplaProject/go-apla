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

package service

import "sync"

const (
	NoPause PauseType = 0

	PauseTypeUpdatingBlockchain PauseType = 1 + iota
	PauseTypeStopingNetwork
)

// np contains the reason why a node should not generating blocks
var np = &NodePaused{PauseType: NoPause}

type PauseType int

type NodePaused struct {
	mutex sync.RWMutex

	PauseType PauseType
}

func (np *NodePaused) Set(pt PauseType) {
	np.mutex.Lock()
	defer np.mutex.Unlock()

	np.PauseType = pt
}

func (np *NodePaused) Unset() {
	np.mutex.Lock()
	defer np.mutex.Unlock()

	np.PauseType = NoPause
}

func (np *NodePaused) Get() PauseType {
	np.mutex.RLock()
	defer np.mutex.RUnlock()

	return np.PauseType
}

func (np *NodePaused) IsSet() bool {
	np.mutex.RLock()
	defer np.mutex.RUnlock()

	return np.PauseType != NoPause
}

func IsNodePaused() bool {
	return np.IsSet()
}

func PauseNodeActivity(pt PauseType) {
	np.Set(pt)
}

func NodePauseType() PauseType {
	return np.Get()
}
