// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
