// +build darwin
// +build arm arm64

package main

import (
	"github.com/c-darwin/mobile/app"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework GLKit -framework UIKit
#import <UIKit/UIKit.h>
#import <Foundation/Foundation.h>
#import <GLKit/GLKit.h>

void ShowMessX(void) {
	UIAlertView* alert = [[UIAlertView alloc] initWithTitle:@"1111111111111" message:@"Это простой UIAlertView, он просто показывает сообщение" delegate:nil cancelButtonTitle:@"OK" otherButtonTitles: nil];
	[alert show];
	[alert release];
}

*/
import "C"

func main() {
	//     go func() {
	//       C.ShowMessX();
	//     }()

	app.Main(func(a app.App) {
		for _ = range a.Events() {

		}
	})
}


