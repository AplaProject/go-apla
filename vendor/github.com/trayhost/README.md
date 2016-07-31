TrayHost
========

__TrayHost__ is a library for placing a __Go__ application in the task bar (system tray, notification area, or dock) in a consistent manner across multiple platforms. Currently, there is built-in support for __Windows__, __Mac OSX__, and __Linux__ systems that support GTK+ 3 status icons (including Gnome 2, KDE 4, Cinnamon, MATE and other desktop environments).

The intended usage is for applications that utilize web technology for the user interface, but require access to the client system beyond what is offered in a browser sandbox (for instance, an application that requires access to the user's file system).

The library places a tray icon on the host system's task bar that can be used to open a URL, giving users easy access to the web-based user interface. 

API docs can be found [here](http://godoc.org/github.com/cratonica/trayhost)

The Interesting Part
----------------------
```go
import (
    "fmt"
    "github.com/cratonica/trayhost"
    "runtime"
)

func main() {
    // EnterLoop must be called on the OS's main thread
    runtime.LockOSThread()

    go func() {
        // Run your application/server code in here. Most likely you will
        // want to start an HTTP server that the user can hit with a browser
        // by clicking the tray icon.

        // Be sure to call this to link the tray icon to the target url
        trayhost.SetUrl("http://localhost:8080")
    }()

    // Enter the host system's event loop
    trayhost.EnterLoop("My Go App", iconData)

    // This is only reached once the user chooses the Exit menu item
    fmt.Println("Exiting")
}
```

Build Environment
--------------------------
Before continuing, make sure that your GOPATH environment variable is set, and that you have Git and Mercurial installed and that __go__, __git__, and __hg__ are in your PATH.

Cross-compilation is not currently supported, so you will need access to a machine running the platform that you wish to target. 

Generally speaking, make sure that your system is capable of doing [cgo](http://golang.org/doc/articles/c_go_cgo.html) builds.

#### Linux
In addition to the essential GNU build tools, you will need to have the GTK+ 3.0 development headers installed.

#### Windows
To do cgo builds, you will need to install [MinGW](http://www.mingw.org/). In order to prevent the terminal window from appearing when your application runs, build with:

    go build -ldflags -H=windowsgui

#### Mac OSX
__Note__: TrayHost requires __Go 1.1__ when targetting Mac OSX, or linking will fail due to issues with previous versions of Go and Mach-O binaries.

You'll need the "Command Line Tools for Xcode", which can be installed using Xcode. You should be able to run the __cc__ command from a terminal window.

Installing
-----------
Once your build environment is configured, go get the library:

    go get github.com/cratonica/trayhost

If all goes well, you shouldn't get any errors.

Using
-----
Use the included __example_test.go__ file as a template to get going.  OSX will throw a runtime error if __EnterLoop__ is called on a child thread, so the first thing you must do is lock the OS thread. Your application code will need to run on a child goroutine. __SetUrl__ can be called lazily if you need to take some time to determine what port you are running on. 

Before it will build, you will need to pick an icon for display in the system tray.

#### Generating the Tray Icon
Included in the project is a tool for generating the icon that gets displayed in the system tray. An icon sized 64x64 pixels should suffice, but there aren't any restrictions here as the system will take care of fitting it (just don't get carried away). 

Icons are embedded into the application by generating a Go array containing the byte data using the [2goarray](http://github.com/cratonica/2goarray) tool, which will automatically be installed if it is missing. The generated .go file will be compiled into the output program, so there is no need to distribute the icon with the program. If you want to embed more resources, check out the [embed](http://github.com/cratonica/embed) project.

#### Linux/OSX
From your project root, run __make_icon.sh__, followed by the path to a __PNG__ file to use. For example:

    $GOPATH/src/github.com/cratonica/trayhost/make_icon.sh ~/Documents/MyIcon.png

This will generate a file called __iconunix.go__ and set its build options so it won't be built in Windows.

#### Windows
From the project root, run __make_icon.bat__, followed by the path to a __Windows ICO__ file to use. If you need to create an ICO file, the online tool [ConvertICO](http://convertico.com/) can do this painlessly. 

Example:

    %GOPATH%\src\github.com\cratonica\trayhost\make_icon.bat C:\MyIcon.ico

This will generate a file called __iconwin.go__ and set its build options so it will only be built in Windows.
