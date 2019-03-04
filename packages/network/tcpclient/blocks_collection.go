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

package tcpclient

import (
	"context"
	"encoding/binary"
	"errors"
	"io"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"
	log "github.com/sirupsen/logrus"
)

var ErrorEmptyBlockBody = errors.New("block is empty")
var ErrorWrongSizeBytes = errors.New("wrong size bytes")

const hasVal = "has value"
const hasntVal = "has not value"

const sizeBytesLength = 4

// GetBlocksBodies send GetBodiesRequest returns channel of binary blocks data
func GetBlocksBodies(ctx context.Context, host string, blockID int64, reverseOrder bool) (<-chan []byte, error) {
	conn, err := newConnection(host)
	if err != nil {
		return nil, err
	}

	// send the type of data
	rt := &network.RequestType{Type: network.RequestTypeBlockCollection}
	if err = rt.Write(conn); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data type block body to connection")
		return nil, err
	}

	req := &network.GetBodiesRequest{
		BlockID:      uint32(blockID),
		ReverseOrder: reverseOrder,
	}

	if err = req.Write(conn); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending blocks bodies request")
		return nil, err
	}

	blocksCount, err := network.ReadInt(conn)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on getting blocks count")
		return nil, err
	}

	if blocksCount == 0 {
		log.Warnf("host: %s does'nt contains block", host)
		return nil, nil
	}

	blocksChan, errChan := GetBlockBodiesChan(ctx, conn, blocksCount)
	go func() {
		for err := range errChan {
			if err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on reading block bodies")
			}
		}
	}()

	return blocksChan, nil
}

func GetBlockBodiesChan(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)

	sizeBuf := make([]byte, sizeBytesLength)
	var bodyBuf []byte

	afterBodyProcessed := func(done <-chan struct{}) {
		<-done
		BytesPool.Put(bodyBuf)
	}

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			src.Close()
			go afterBodyProcessed(ctx.Done())
		}()

		dataSize, err := network.ReadInt(src)
		if err != nil {
			errChan <- err
			return
		}

		bodyBuf = BytesPool.Get(dataSize)
		var bodyStartIndx int64

		for i := 0; i < int(blocksCount); i++ {

			_, err := io.ReadFull(src, sizeBuf)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading size of block data")
				errChan <- err
				return
			}

			size, intErr := binary.Uvarint(sizeBuf)
			if intErr < 0 {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": ErrorWrongSizeBytes}).Error("on convert size body")
				errChan <- ErrorWrongSizeBytes
				return
			}

			bodyEndIndx := bodyStartIndx + int64(size)
			body := bodyBuf[bodyStartIndx:bodyEndIndx]
			if readed, err := io.ReadFull(src, body); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "size": size, "readed": readed, "error": err}).Error("on reading block body")
				errChan <- err
				return
			}

			bodyStartIndx = bodyEndIndx
			rawBlocksCh <- body
		}
	}()

	return rawBlocksCh, errChan
}
