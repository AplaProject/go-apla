// +build linux freebsd

package main

import (
	"time"
)

func tray() {

}

func enterLoop() {
	time.Sleep(3600*24*90 * time.Second)
}