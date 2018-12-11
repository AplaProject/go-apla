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

package tcpserver

import (
	"net"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"

	log "github.com/sirupsen/logrus"
)

const BlocksPerRequest = 5

// Type7 writes the body of the specified block
// blocksCollection and queue_parser_blocks daemons send the request through p.GetBlocks()
func Type7(request *network.GetBodiesRequest, w net.Conn) error {
	var blocks []*blockchain.BlockWithHash
	var err error
	order := 1
	if request.ReverseOrder {
		order = -1
	}
	blocks, err = blockchain.GetNBlocksFrom(nil, request.BlockHash, int(BlocksPerRequest), order)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Error("Error getting 1000 blocks from block_hash")
		if err := network.WriteInt(0, w); err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending 0 requested blocks")
		}
		return err
	}

	if len(blocks) == 0 {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "block_hash": request.BlockHash}).Warn("Requesting nonexistent blocks from block_hash")
		return nil
	}
	if err := network.WriteInt(int64(len(blocks)), w); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending requested blocks count")
		return err
	}

	l, err := lenOfBlockData(blocks)
	if err != nil {
		return err
	}
	if err := network.WriteInt(l, w); err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on sending requested blocks data length")
		return err
	}

	for _, b := range blocks {
		data, err := b.Block.Marshal()
		if err != nil {
			return err
		}
		br := &network.GetBodyResponse{Data: data}
		if err := br.Write(w); err != nil {
			return err
		}
	}

	return nil
}

func lenOfBlockData(blocks []*blockchain.BlockWithHash) (int64, error) {
	var length int64
	for i := 0; i < len(blocks); i++ {
		data, err := blocks[i].Block.Marshal()
		if err != nil {
			return 0, err
		}
		length += int64(len(data))
	}

	return length, nil
}
