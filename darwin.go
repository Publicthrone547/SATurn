package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// quitFromOutside reports whether the current quit request arrived as an
// Apple event (log out, restart, shut down, a script, the Dock). A quit from
// the app's own menu or Cmd+Q does not carry one.
static int quitFromOutside(void) {
	NSAppleEventDescriptor *ev = [[NSAppleEventManager sharedAppleEventManager] currentAppleEvent];
	return ev != nil ? 1 : 0;
}

static void activateApp(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		[NSApp activateIgnoringOtherApps:YES];
	});
}

static int reduceMotion(void) {
	return [[NSWorkspace sharedWorkspace] accessibilityDisplayShouldReduceMotion] ? 1 : 0;
}
*/
import "C"

func quitFromOutside() bool { return C.quitFromOutside() == 1 }

// activateApp brings the menu bar app to the front so its window gets focus.
func activateApp() { C.activateApp() }

func reduceMotion() bool { return C.reduceMotion() == 1 }
