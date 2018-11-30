package daemonsctl

import (
	"context"

	"github.com/GenesisKernel/go-genesis/packages/modes"
)

// RunAllDaemons start daemons, load contracts and tcpserver
func RunAllDaemons(ctx context.Context) error {
	loader := modes.GetDaemonLoader()

	return loader.Load(ctx)
}
