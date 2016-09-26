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

package main

import (
	"runtime"

	"github.com/DayLightProject/go-daylight/packages/daylight"
	"github.com/DayLightProject/go-daylight/packages/system"
	"github.com/go-thrust/lib/bindings/window"
)

func main() {
	runtime.LockOSThread()
	var thrustWindow *window.Window
	tray()
	go daylight.Start("", thrustWindow)
	enterLoop()
	system.Finish(0)
}
