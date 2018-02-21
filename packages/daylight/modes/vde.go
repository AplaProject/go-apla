package modes

import "github.com/GenesisKernel/go-genesis/packages/conf"

func InitVDEMode(config *conf.VDEConfig) *VDE {
	mode := &VDE{
		VDEConfig: config,
	}

	return mode
}

type VDE struct {
	*conf.VDEConfig
}
