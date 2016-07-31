#import <Cocoa/Cocoa.h>

static volatile NSString *strStaticUrl;
static volatile NSURL *staticUrl;

@interface AppDelegate: NSObject <NSApplicationDelegate>
- (NSMenu *)applicationDockMenu:(NSApplication *)sender;
@end

@implementation AppDelegate
NSString *m_title;
NSMenu *m_menu;
- (id)init:(NSString *)title {
    if ((self = [super init])) {
        m_title = title;
    }
    return self;
}

- (NSMenu *)applicationDockMenu:(NSApplication *)sender {
    if (m_menu == nil) {
        m_menu = [[[NSMenu alloc] init] retain];

        NSMenuItem *openMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Open" action:@selector(openUrl) keyEquivalent:@""] autorelease];
        [m_menu addItem:openMenuItem];

        // OSX will automatically add the Quit option

    }
    return m_menu;
}
- (void)copyUrl {
    if (strStaticUrl) {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        NSArray *copiedObjects = [NSArray arrayWithObject:strStaticUrl];
        [pasteboard writeObjects:copiedObjects];
    }
}
- (void)openUrl {
    if (staticUrl) {
        [[NSWorkspace sharedWorkspace] openURL:(NSURL*)staticUrl];
    }
}
@end


void native_loop(const char *title, unsigned char *imageDataBytes, unsigned int imageDataLen) {
    [NSAutoreleasePool new];
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    NSData *imageData = [[[NSData alloc] initWithBytes:imageDataBytes length:imageDataLen] autorelease];
    NSImage *icon = [[[NSImage alloc] initWithData:imageData] autorelease];

    [NSApp setApplicationIconImage:icon];

    AppDelegate *delegate = [[[AppDelegate alloc] init:[NSString stringWithCString:title encoding:NSASCIIStringEncoding]] autorelease];
    [NSApp setDelegate:delegate];

    [NSApp activateIgnoringOtherApps:YES];
    [NSApp run];

    // I don't think this is ever reached, but for completeness...
    [staticUrl release];
    [strStaticUrl release];
}

void set_url(const char *url) {
    // Hoping these assignments are atomic...
    // Using alloc to prevent the framework from trying to autorelease these.

    strStaticUrl = [[NSString alloc] initWithBytes:url length:strlen(url) encoding:NSASCIIStringEncoding];
    staticUrl = [[NSURL alloc] initWithString:(NSString*)strStaticUrl];
}

