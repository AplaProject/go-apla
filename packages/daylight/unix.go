// +build linux freebsd darwin
// +build 386 amd64

package daylight

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"syscall"
)

func KillPid(pid string) error {
	err := syscall.Kill(utils.StrToInt(pid), syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}
