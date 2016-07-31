// +build windows

package main

import (
	"github.com/trayhost"
)

/*
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>
int w_ver() {
	DWORD dwMajorVersion = 0;
	DWORD dwVersion = 0;
	dwVersion = GetVersion();
	//dwMajorVersion = (DWORD)(LOBYTE(LOWORD(dwVersion)));
	//return dwMajorVersion;
	return dwVersion;
}*/
import "C"

func winVer() int {
	ver := int(C.w_ver())
	if ver & 0xff == 6 && ( ver & 0xff00 ) >> 8 <= 1 {
		return 6
	}
	return 7
}

func tray() {
	go func() {
		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://localhost:8089")
	}()
}

func enterLoop() {
	trayhost.EnterLoop("Dcoin", iconData)
}