#ifndef PLATFORM_H
#define PLATFORM_H

// We have to include the .c files, because
// Go doesn't support static linking, and
// we want to distribute this as a single binary
// so we just build all the code in.

#if defined LINUX
#include "linux/tray.c"
#elif defined WIN32
#include "windows/tray.c"
#elif defined DARWIN
#include "darwin/dock.m"
#endif

#endif // PLATFORM_H
