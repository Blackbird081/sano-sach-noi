//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

// Chép ảnh PNG vào clipboard của macOS (dán được vào Facebook, Zalo, Messenger).
static int sanoCopyPNG(const void *bytes, int n) {
	__block int ok = 0;
	NSData *data = [NSData dataWithBytes:bytes length:n];
	void (^run)(void) = ^{
		NSImage *img = [[NSImage alloc] initWithData:data];
		if (img == nil) return;
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];
		ok = [pb writeObjects:@[img]] ? 1 : 0;
		if (ok) [pb setData:data forType:NSPasteboardTypePNG];
	};
	if ([NSThread isMainThread]) run();
	else dispatch_sync(dispatch_get_main_queue(), run);
	return ok;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

const copyImageSupported = true

func copyImage(png []byte) error {
	if len(png) == 0 {
		return errors.New("ảnh trống")
	}
	if C.sanoCopyPNG(unsafe.Pointer(&png[0]), C.int(len(png))) == 0 {
		return errors.New("không chép được ảnh vào clipboard")
	}
	return nil
}
