// +build windows

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
	if ver&0xff == 6 && (ver&0xff00)>>8 <= 1 {
		return 6
	}
	return 7
}

func tray() {
	go func() {
		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://localhost:7079")
	}()
}

func enterLoop() {
	trayhost.EnterLoop("DayLight", iconData)
}
