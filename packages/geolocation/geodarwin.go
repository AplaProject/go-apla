// +build darwin
// +build !arm !arm64

package geolocation

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreLocation

#import <Foundation/Foundation.h>
#import <CoreLocation/CoreLocation.h>

@interface Location : NSObject <CLLocationManagerDelegate>
{
    BOOL ready;
}

@property (nonatomic, strong)CLLocationManager *manager;
@property (nonatomic, strong)NSTimer *timer;

@end

@implementation Location

- (instancetype)init {
    if (self = [super init]) {
        _manager = [[CLLocationManager alloc] init];
        _manager.delegate = self;
        ready = NO;
    }

    return self;
}

- (void)dealloc {
    [_timer invalidate];
    [_timer release];
    [_manager release];
    [super dealloc];
}

- (void)launch
{
    [_manager startUpdatingLocation];

    _timer = [NSTimer scheduledTimerWithTimeInterval:0.1
                                     target:self
                                   selector:@selector(checkIfUpdated:)
                                   userInfo:nil
                                    repeats:YES];
    [_timer release];
    [[NSRunLoop currentRunLoop] addTimer:_timer forMode:NSDefaultRunLoopMode];
    while (!ready) {
        NSDate *nextDate = [NSDate dateWithTimeIntervalSinceNow:1.0];
        [[NSRunLoop currentRunLoop] runUntilDate:nextDate];
    }
}

- (void)checkIfUpdated:(NSTimer *)timer
{
    if (_manager.location != nil) {
        [timer invalidate];
        [timer release];
        ready = YES;
        NSLog(@"invalidate the timer %@", timer);
        timer = nil;
    }
}


- (void)locationManager:(CLLocationManager *)manager didFailWithError:(NSError *)error
{
    NSLog(@"Error %@", error.userInfo);
    [_timer invalidate];
    [_timer release];
}

@end



char* getLocation() {
        Location *location = [[Location alloc] init];
	[location launch];
        NSString *str = [NSString stringWithFormat:@"%f, %f", location.manager.location.coordinate.latitude,
                         location.manager.location.coordinate.longitude];

        return (char*)[str UTF8String];
}

*/
import "C"

import (
	"strings"
	"strconv"
	"errors"
	"fmt"
)

func goString(s *C.char) string {
	return C.GoString(s)
}

func getLocation() (*coordinates, error) {
	str := goString(C.getLocation())
	sCoords := strings.Split(str, ", ")
	if len(sCoords) != 2 {
		return nil, errors.New("Wrong coordinates")
	}

	fmt.Println("Calling CLLocation()")
	lat, _ := strconv.ParseFloat(sCoords[0], 64)
	lng, _ := strconv.ParseFloat(sCoords[1], 64)

	return &coordinates{
		Latitude:lat,
		Longitude:lng,
	}, nil
}