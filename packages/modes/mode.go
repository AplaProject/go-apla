package modes

import (
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/julienschmidt/httprouter"
)

type ModeType int

const (
	TypeBlockchain ModeType = 1
	TypeVDE        ModeType = 2
	TypeVDEMaster  ModeType = 3
)

// NodeMode allows implement different startup modes
type NodeMode interface {
	Start(exitFunc func(int), gormInit func(conf.DBConfig))
	Stop()
	DaemonList() []string
	API() *httprouter.Router
	Type() ModeType
}
