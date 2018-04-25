package packages

// PrivateBlockchain const label for running mode
const privateBlockchain RunMode = "PrivateBlockchain"

// PublicBlockchain const label for running mode
const publicBlockchain RunMode = "PublicBlockchain"

// VDEManager const label for running mode
const vdeMaster RunMode = "VDEMaster"

// VDE const label for running mode
const vde RunMode = "VDE"

type RunMode string

// IsPublicBlockchain returns true if mode equil PublicBlockchain
func (rm RunMode) IsPublicBlockchain() bool {
	return rm == publicBlockchain
}

// IsPrivateBlockchain returns true if mode equil PrivateBlockchain
func (rm RunMode) IsPrivateBlockchain() bool {
	return rm == privateBlockchain
}

// IsVDEMaster returns true if mode equil vdeMaster
func (rm RunMode) IsVDEMaster() bool {
	return rm == vdeMaster
}

// IsVDE returns true if mode equil vde
func (rm RunMode) IsVDE() bool {
	return rm == vde
}

// IsSupportingVDE returns true if mode support vde
func (rm RunMode) IsSupportingVDE() bool {
	return rm.IsVDE() || rm.IsVDEMaster()
}
