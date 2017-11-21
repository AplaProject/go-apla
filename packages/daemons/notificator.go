package daemons

import (
	"context"

	"github.com/AplaProject/go-apla/packages/notificator"
)

// Notificate is sending notifications
func Notificate(d *daemon, ctx context.Context) error {
	notificator.SendNotifications()
	return nil
}
