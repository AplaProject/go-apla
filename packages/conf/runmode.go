package conf

type RunMode string

// OBSManager const label for running mode
const obsMaster RunMode = "OBSMaster"

// OBS const label for running mode
const obs RunMode = "OBS"

// OBS const label for running mode
const node RunMode = "NONE"

// IsOBSMaster returns true if mode equal obsMaster
func (rm RunMode) IsOBSMaster() bool {
	return rm == obsMaster
}

// IsOBS returns true if mode equal obs
func (rm RunMode) IsOBS() bool {
	return rm == obs
}

// IsNode returns true if mode not equal to any OBS
func (rm RunMode) IsNode() bool {
	return rm == node
}

// IsSupportingOBS returns true if mode support obs
func (rm RunMode) IsSupportingOBS() bool {
	return rm.IsOBS() || rm.IsOBSMaster()
}
