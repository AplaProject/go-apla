/*
Package trayhost is a library for placing a Go
application in the task bar (system tray,
notification area, or dock) in a consistent
manner across multiple platforms. Currently,
there is built-in support for Windows, Mac OSX,
and Linux systems that support GTK+ 3 status
icons (including Gnome 2, KDE 4, Cinnamon,
MATE and other desktop environments).

The indended usage is for applications that
utilize web technology for the user interface, but
require access to the client system beyond what
is offered in a browser sandbox (for instance,
an application that requires access to the user's
file system).

The library places a tray icon on the host system's
task bar that can be used to open a URL, giving users
easy access to the web-based user interface.

Further information can be found at the project's
home at http://github.com/cratonica/trayhost

Clint Caywood

http://github.com/cratonica/trayhost
*/
package trayhost

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/op/go-logging"
)

/*
#cgo linux pkg-config: gtk+-3.0
#cgo linux CFLAGS: -DLINUX
#cgo windows CFLAGS: -DWIN32
#cgo darwin CFLAGS: -DDARWIN -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "platform/platform.h"
*/
import "C"

var isExiting bool
var urlPtr unsafe.Pointer
var log = logging.MustGetLogger("controllers")

// Run the host system's event loop
func EnterLoop(title string, imageData []byte) {
	defer C.free(urlPtr)

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	// Copy the image data into unmanaged memory
	cImageData := C.malloc(C.size_t(len(imageData)))
	defer C.free(cImageData)
	var cImageDataSlice []C.uchar
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&cImageDataSlice))
	sliceHeader.Cap = len(imageData)
	sliceHeader.Len = len(imageData)
	sliceHeader.Data = uintptr(cImageData)

	for i, v := range imageData {
		cImageDataSlice[i] = C.uchar(v)
	}

	// Enter the loop
	C.native_loop(cTitle, &cImageDataSlice[0], C.uint(len(imageData)))

	// If reached, user clicked Exit
	isExiting = true
	if model.DBConn != nil {
		sd := &model.StopDaemon{StopTime: time.Now().Unix()}
		err := sd.Create()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
	}
}

// Set the URL that the tray icon will open in a browser
func SetUrl(url string) {
	if isExiting {
		return
	}
	cs := C.CString(url)
	C.free(urlPtr)
	urlPtr = unsafe.Pointer(cs)
	C.set_url(cs)
}
