// +build android

package sendnotif

import (
	"github.com/c-darwin/mobile/notif"
)

func SendMobileNotification(title, text string) {
	notif.SendNotif(title, text)
}
