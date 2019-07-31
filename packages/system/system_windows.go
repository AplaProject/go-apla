// +build windows

// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package system

import (
	"os"
)

/*
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>
#include <TlHelp32.h>

void kill_childproc( DWORD myprocID) {
	PROCESSENTRY32 pe;

	memset(&pe, 0, sizeof(PROCESSENTRY32));
	pe.dwSize = sizeof(PROCESSENTRY32);

	HANDLE hSnap = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
	if (Process32First(hSnap, &pe))
	{
	    BOOL bContinue = TRUE;

	    while (bContinue)
	    {
	        if (pe.th32ParentProcessID == myprocID && memcmp( pe.szExeFile, "tmp_", 4 ) != 0 &&
				memcmp(pe.szExeFile, "apla", 4) != 0)
	        {
	            HANDLE hChildProc = OpenProcess(PROCESS_ALL_ACCESS, FALSE, pe.th32ProcessID);

	            if (hChildProc)
	            {
					kill_childproc(GetProcessId(hChildProc));
	                TerminateProcess(hChildProc, 1);
	                CloseHandle(hChildProc);
	            }
	        }
	        bContinue = Process32Next(hSnap, &pe);
	    }
	}
}
*/
import "C"

// lstrcmp( pe.szExeFile, TEXT("tmp_daylight.exe")) != 0 && lstrcmp( pe.szExeFile, TEXT("daylight.exe")) != 0

func killChildProc() {
	C.kill_childproc(C.DWORD(os.Getpid()))
}
