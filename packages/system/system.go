// system
package system

import (
	"os"
//	"time"
	"github.com/go-thrust/thrust"
)

func finish( exit int, isthrust bool ) {
	killChildProc()
	if isthrust {
		thrust.Exit()
	}
//	time.Sleep(1*time.Second)
	if exit != 0 {
		os.Exit(exit)
	}
	
}

func Finish(exit int) {
	finish(exit, false)
}

func FinishThrust(exit int) {
	finish(exit, true)
}

