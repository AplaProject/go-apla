package modes

import "github.com/GenesisKernel/go-genesis/packages/conf"

// InitVDEMaster returns new master of VDE
func InitVDEMaster(config *conf.VDEMasterConfig) *VDEMaster {
	mode := &VDEMaster{
		VDEMasterConfig: config,
		VDE:             InitVDEMode(config.VDE),
	}
}

// VDEMaster represents master of VDE mode
type VDEMaster struct {
	*VDEMasterConfig
	*VDE
}
