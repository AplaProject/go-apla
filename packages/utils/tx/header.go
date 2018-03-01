package tx

import (
	"github.com/GenesisKernel/go-genesis/packages/modes"
)

// Header is contain header data
type Header struct {
	NodeMode      modes.NodeMode
	Type          int
	Time          int64
	EcosystemID   int64
	KeyID         int64
	NodePosition  int64
	PublicKey     []byte
	BinSignatures []byte
}
