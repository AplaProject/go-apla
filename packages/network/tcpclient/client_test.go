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

package tcpclient

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"strings"
	"testing"

	"github.com/AplaProject/go-apla/packages/network"
)

var inputs = make([][]byte, 0, 100)

func init() {
	for i := 0; i < 100; i++ {
		inputs = append(inputs, []byte(strings.Repeat("B", rand.Intn(194334))))
	}
}

type BufCloser struct {
	*bytes.Buffer
}

func (bc BufCloser) Close() error {
	bc.Reset()
	return nil
}

func BenchmarkGetBlockBodiesWithChanReadAll(t *testing.B) {
	var bts []byte
	r := BufCloser{bytes.NewBuffer(bts)}

	// dataLen := 4
	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 100; i++ {
			resp := network.GetBodyResponse{
				Data: inputs[i],
			}

			resp.Write(r)
		}

		ctxDone, cancel := context.WithCancel(context.Background())

		t.StartTimer()
		blocksC, errC := GetBlockBodiesChanReadAll(ctxDone, r, 100)
		// blocksC, errC := GetBlockBodiesChan(ctxDone, r, 100)
		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			item = item[:0]
		}
		cancel()
	}
}

//===================================================GetBlockBodiesChanByBlock

// 500	   2264475 ns/op	  109001 B/op	     108 allocs/op with pool size 12832256
func BenchmarkGetBlockBodiesChanByBlockWithSyncPool(t *testing.B) {
	var bts []byte
	r := BufCloser{bytes.NewBuffer(bts)}

	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 100; i++ {
			resp := network.GetBodyResponse{
				Data: inputs[i],
			}

			// fmt.Println("lenData", len(inputs[i]))
			resp.Write(r)
		}

		ctxDone, cancel := context.WithCancel(context.Background())

		t.StartTimer()
		blocksC, errC := GetBlockBodiesChanByBlock(ctxDone, r, 100)

		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			item = item[:0]
		}
		cancel()
	}

}

func BenchmarkGetBlockBodiesChanByBlockWithBytePool(t *testing.B) {
	var bts []byte
	r := BufCloser{bytes.NewBuffer(bts)}

	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()

		var dataSize int64
		for i := 0; i < 100; i++ {
			dataSize += int64(len(inputs[i]))
		}
		network.WriteInt(dataSize, r)
		// fmt.Println("sending data size", dataSize)

		for i := 0; i < 100; i++ {
			resp := network.GetBodyResponse{
				Data: inputs[i],
			}

			// fmt.Println("lenData", len(inputs[i]))
			resp.Write(r)
		}

		ctxDone, cancel := context.WithCancel(context.Background())

		t.StartTimer()
		blocksC, errC := GetBlockBodiesChanByBlockWithBytePool(ctxDone, r, 100)

		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			// fmt.Println(len(item))
			item = item[:0]
		}
		cancel()
	}
}

func BenchmarkGetBlockBodiesWithChanReadToStruct(t *testing.B) {
	var bts []byte
	r := BufCloser{bytes.NewBuffer(bts)}

	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 100; i++ {
			resp := network.GetBodyResponse{
				Data: inputs[i],
			}

			// fmt.Println("lenData", len(inputs[i]))
			resp.Write(r)
		}

		ctx := context.Background()

		t.StartTimer()
		blocksC, errC := GetBlockBodiesChan(ctx, r, 100)

		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			item = item[:0]
		}
	}
}

func BenchmarkGetBlockBodiesAsSlice(t *testing.B) {
	var bts []byte
	r := BufCloser{bytes.NewBuffer(bts)}

	// dataLen := 4
	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 100; i++ {
			resp := network.GetBodyResponse{
				Data: inputs[i],
			}

			resp.Write(r)
		}

		ctxDone, cancel := context.WithCancel(context.Background())

		t.StartTimer()
		blocks, err := GetBlockBodiesReadAll(ctxDone, r, 100)
		if err != nil {
			fmt.Println(err)
		}

		for i := 0; i < len(blocks); i++ {
			blocks[i] = blocks[i][:0]
		}
		cancel()
	}
}

//==============================================
// func TestReadSize(t *testing.T) {
// 	bts := []byte{}
// 	buf := bytes.NewBuffer(bts)

// 	resp := network.GetBodyResponse{
// 		Data: []byte(strings.Repeat("B", 152627)),
// 	}

// 	resp.Write(buf)

// 	val, err := binary.ReadUvarint(buf)
// 	require.NoError(t, err)
// 	fmt.Println(val)
// 	// data := buf.Bytes()
// 	// size, intErr := binary.Uvarint(data[:4])
// 	// fmt.Println(size, intErr)

// }

// func TestBinary(t *testing.T) {

// 	buf := []byte{}
// 	bb := bytes.NewBuffer(buf)
// 	for _, x := range []uint64{1, 2, 127, 128, 255, 152627} {
// 		mb := make([]byte, 4)
// 		_ = binary.PutUvarint(mb, x)
// 		bb.Write(mb)
// 		// fmt.Printf("%x\n", buf[:n])
// 	}

// 	resBts := bb.Bytes()
// 	fmt.Println(resBts)
// 	var pos int
// 	for i := 0; i < 6; i++ {
// 		valBts := resBts[pos : pos+4]
// 		pos += 4
// 		fmt.Println("valBts", valBts)
// 		value, re := binary.Uvarint(valBts)
// 		fmt.Println(value, "readed", re)
// 	}
// }

// //100000	     17333 ns/op	     576 B/op	      19 allocs/op

// func BenchmarkGetBlockBodiesWithBuffer(t *testing.B) {
// 	ctx := context.Background()
// 	var bts []byte
// 	r := BufCloser{
// 		Buffer: bytes.NewBuffer(bts),
// 	}

// 	byteString := []byte(strings.Repeat("A", 32))
// 	t.ResetTimer()
// 	for j := 0; j < t.N; j++ {
// 		t.StopTimer()
// 		for i := 0; i < 5; i++ {
// 			resp := network.GetBodyResponse{
// 				Data: byteString,
// 			}

// 			resp.Write(r)
// 		}
// 		fmt.Println("[", j, "]")
// 		t.StartTimer()

// 		blocksC, errC := GetBlockBodiesChanWithPool(ctx, r, 5)
// 		go func() {
// 			err := <-errC
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 		}()

// 		for item := range blocksC {
// 			fmt.Println(string(item))
// 			BlockBodyPool.putBytes(item)
// 		}

// 	}
// }
