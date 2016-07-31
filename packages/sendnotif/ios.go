// +build darwin
// +build arm arm64

package sendnotif

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework GLKit -framework UIKit
#import <UIKit/UIKit.h>
#import <Foundation/Foundation.h>
#import <GLKit/GLKit.h>

void
ShowMessM(char* title, char* text) {
    NSLog(@"ShowMessX<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n");
    UILocalNotification* localNotification = [[UILocalNotification alloc] init];
    //localNotification.fireDate = [NSDate dateWithTimeIntervalSinceNow:5];
    localNotification.soundName = UILocalNotificationDefaultSoundName;
    localNotification.alertTitle = [NSString stringWithUTF8String:title];
    localNotification.alertBody = [NSString stringWithUTF8String:text];
    localNotification.timeZone = [NSTimeZone systemTimeZone];
    localNotification.applicationIconBadgeNumber = 1;
    localNotification.repeatInterval = NSCalendarUnitMinute;
    [[UIApplication sharedApplication] scheduleLocalNotification:localNotification];
//    [[NSNotificationCenter defaultCenter] postNotificationName:@"timerInvoked" object:self];
    [localNotification release];
}


*/
import "C"

func SendMobileNotification(title, text string) {
	C.ShowMessM(C.CString(title), C.CString(text))
}
