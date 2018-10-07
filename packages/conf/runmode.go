package conf

type RunMode string

// VDEManager const label for running mode
const vdeMaster RunMode = "VDEMaster"

// VDE const label for running mode
const vde RunMode = "VDE"

// VDE const label for running mode
const node RunMode = "NONE"

// IsVDEMaster returns true if mode equal vdeMaster
func (rm RunMode) IsVDEMaster() bool {
	return rm == vdeMaster
}

// IsVDE returns true if mode equal vde
func (rm RunMode) IsVDE() bool {
	return rm == vde
}

// IsNode returns true if mode not equal to any VDE
func (rm RunMode) IsNode() bool {
	return rm == node
}

// IsSupportingVDE returns true if mode support vde
func (rm RunMode) IsSupportingVDE() bool {
	return rm.IsVDE() || rm.IsVDEMaster()
}
