// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package system

import (
	"os"
	//	"time"
	"github.com/go-thrust/thrust"
)

func finish(exit int, isthrust bool) {
	killChildProc()
	if isthrust {
		thrust.Exit()
	}
	//	time.Sleep(1*time.Second)
	if exit != 0 {
		os.Exit(exit)
	}

}

// Finish closes the program
func Finish(exit int) {
	finish(exit, false)
}

// FinishThrust closes thrust shell program
func FinishThrust(exit int) {
	finish(exit, true)
}
