package tcpclient

import (
	"bytes"
	"context"
	"fmt"
	_ "net/http/pprof"
	"strings"
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/network"
)

func init() {
	InitBlockBodyBuffer(7, 60)
}

type BufCloser struct {
	*bytes.Buffer
}

func (bc BufCloser) Close() error {
	bc.Truncate(0)
	return nil
}

// 4 bytes length
//100000	     19376 ns/op	     488 B/op	      31 allocs/op
//100000	     18469 ns/op	     449 B/op	      26 allocs/op
// 32 bytes length
// 100000	     19966 ns/op	     537 B/op	      26 allocs/op
// with stopTimer on init
// 50000	     28630 ns/op	     466 B/op	      16 allocs/op
//========================
// BenchmarkGetBlockBodies-8   	16000
// 1600000
// 16000000
//   100000	     17380 ns/op	     577 B/op	      19 allocs/op
// PASS
// ok  	github.com/GenesisKernel/go-genesis/packages/network/tcpclient	8.976s
func BenchmarkGetBlockBodies(t *testing.B) {
	ctx := context.Background()
	var bts []byte
	r := BufCloser{
		Buffer: bytes.NewBuffer(bts),
	}

	byteString := []byte(strings.Repeat("A", 32))
	var lenCounter int
	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 5; i++ {
			resp := network.GetBodyResponse{
				Data: byteString,
			}

			resp.Write(r)
		}

		t.StartTimer()
		blocksC, errC := GetBlockBodiesChan(ctx, r, 5)

		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			lenCounter += len(item)
		}
	}

	fmt.Println(lenCounter)
}

//100000	     17333 ns/op	     576 B/op	      19 allocs/op

func BenchmarkGetBlockBodiesWithBuffer(t *testing.B) {
	ctx := context.Background()
	var bts []byte
	r := BufCloser{
		Buffer: bytes.NewBuffer(bts),
	}

	byteString := []byte(strings.Repeat("A", 32))
	var lenCounter int
	t.ResetTimer()
	for j := 0; j < t.N; j++ {
		t.StopTimer()
		for i := 0; i < 5; i++ {
			resp := network.GetBodyResponse{
				Data: byteString,
			}

			resp.Write(r)
		}

		t.StartTimer()

		blocksC, errC := GetBlockBodiesChanWithPool(ctx, r, 5)

		go func() {
			err := <-errC
			if err != nil {
				fmt.Println(err)
			}
		}()

		for item := range blocksC {
			lenCounter += len(item)
			BlockBodyPool.putBytes(item)
		}
	}

	fmt.Println(lenCounter)
}
