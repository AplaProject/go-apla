// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
