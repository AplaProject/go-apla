package daemons

import (
	"context"

	"github.com/AplaProject/go-apla/packages/notificator"
)

// Notificate is sending notifications
func Notificate(ctx context.Context, d *daemon) error {
	notificator.SendNotifications()
	return nil
}
